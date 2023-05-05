package lnd_connect

import (
	"context"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/timeout"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"gopkg.in/macaroon.v2"

	"github.com/lncapital/torq/pkg/grpc_helpers"
)

// Connect connects to LND using gRPC. DO NOT USE THIS UNLESS THE GRPC SETTINGS ARE NOT VALIDATED NOR ACTIVATED IN TORQ.
func Connect(host string, tlsCert []byte, macaroonBytes []byte) (*grpc.ClientConn, error) {

	cp := x509.NewCertPool()
	if !cp.AppendCertsFromPEM(tlsCert) {
		return nil, fmt.Errorf("credentials: failed to append certificates")
	}
	tlsCreds := credentials.NewClientTLSFromCert(cp, "")

	mac := &macaroon.Macaroon{}
	if err := mac.UnmarshalBinary(macaroonBytes); err != nil {
		return nil, fmt.Errorf("cannot unmarshal macaroon: %v", err)
	}

	macCred, err := NewMacaroonCredential(mac)
	if err != nil {
		return nil, fmt.Errorf("cannot create macaroon credentials: %v", err)
	}

	clMetrics := grpc_helpers.GetClientMetrics()
	loggerOpts := grpc_helpers.GetLoggingOptions()
	exemplarFromContext := grpc_helpers.GetExemplarFromContext()

	logger := zerolog.New(os.Stderr)

	opts := []grpc.DialOption{
		grpc.WithReturnConnectionError(),
		grpc.FailOnNonTempDialError(true),
		grpc.WithBlock(),
		grpc.WithTransportCredentials(tlsCreds),
		grpc.WithPerRPCCredentials(macCred),
		// max size to 25mb
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(grpc_helpers.RecvMsgSize)),
		grpc.WithChainUnaryInterceptor(
			timeout.UnaryClientInterceptor(grpc_helpers.UnaryTimeout),
			otelgrpc.UnaryClientInterceptor(),
			clMetrics.UnaryClientInterceptor(grpcprom.WithExemplarFromContext(exemplarFromContext)),
			logging.UnaryClientInterceptor(grpc_helpers.InterceptorLogger(logger), loggerOpts...),
		),
		grpc.WithChainStreamInterceptor(
			otelgrpc.StreamClientInterceptor(),
			clMetrics.StreamClientInterceptor(grpcprom.WithExemplarFromContext(exemplarFromContext)),
			logging.StreamClientInterceptor(grpc_helpers.InterceptorLogger(logger), loggerOpts...),
		),
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	conn, err := grpc.DialContext(ctx, host, opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot dial to LND: %v", err)
	}

	return conn, nil
}
