package cln_connect

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/timeout"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/lncapital/torq/pkg/grpc_helpers"
	"github.com/lncapital/torq/pkg/prometheus"
)

// Connect connects to CLN using gRPC. DO NOT USE THIS UNLESS THE GRPC SETTINGS ARE NOT VALIDATED NOR ACTIVATED IN TORQ.
func Connect(host string,
	certificate []byte,
	key []byte,
	caCertificate []byte) (*grpc.ClientConn, error) {

	clientCrt, err := tls.X509KeyPair(certificate, key)
	if err != nil {
		return nil, errors.New("CLN credentials: failed to create X509 KeyPair")
	}

	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(caCertificate)

	serverName := "localhost"
	if strings.Contains(host, "cln") {
		serverName = "cln"
	}

	tlsConfig := &tls.Config{
		MinVersion:   tls.VersionTLS12,
		ClientAuth:   tls.RequestClientCert,
		Certificates: []tls.Certificate{clientCrt},
		RootCAs:      certPool,
		ServerName:   serverName,
	}

	clMetrics := prometheus.GetGrpcClientMetrics()
	loggerOpts := grpc_helpers.GetLoggingOptions()
	exemplarFromContext := grpc_helpers.GetExemplarFromContext()

	logger := zerolog.New(os.Stderr)

	opts := []grpc.DialOption{
		grpc.WithReturnConnectionError(),
		grpc.FailOnNonTempDialError(true),
		grpc.WithBlock(),
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
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
		return nil, fmt.Errorf("cannot dial to CLN %v", err)
	}

	return conn, nil
}
