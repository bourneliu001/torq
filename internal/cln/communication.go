package cln

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/internal/graph_events"
	"github.com/lncapital/torq/internal/lightning_helpers"
	"github.com/lncapital/torq/pkg/cln_connect"
	"github.com/lncapital/torq/proto/cln"
)

const routingPolicyUpdateLimiterSeconds = 5 * 60

var (
	connectionWrapperOnce sync.Once           //nolint:gochecknoglobals
	connectionWrapper     *connectionsWrapper //nolint:gochecknoglobals
)

type connectionsWrapper struct {
	mu                 sync.Mutex
	connections        map[int]*grpc.ClientConn
	grpcAddresses      map[int]string
	certificateBytes   map[int][]byte
	keyBytes           map[int][]byte
	caCertificateBytes map[int][]byte
}

func getConnection(nodeId int) (*grpc.ClientConn, error) {
	connectionWrapperOnce.Do(func() {
		log.Debug().Msg("Loading Connection Wrapper.")
		connectionWrapper = &connectionsWrapper{
			mu:                 sync.Mutex{},
			connections:        make(map[int]*grpc.ClientConn),
			grpcAddresses:      make(map[int]string),
			certificateBytes:   make(map[int][]byte),
			keyBytes:           make(map[int][]byte),
			caCertificateBytes: make(map[int][]byte),
		}
	})

	connectionWrapper.mu.Lock()
	defer connectionWrapper.mu.Unlock()

	ncd := cache.GetNodeConnectionDetails(nodeId)

	existingConnection, exists := connectionWrapper.connections[nodeId]
	if !exists ||
		connectionWrapper.grpcAddresses[nodeId] != ncd.GRPCAddress ||
		!bytes.Equal(connectionWrapper.certificateBytes[nodeId], ncd.CertificateFileBytes) ||
		!bytes.Equal(connectionWrapper.keyBytes[nodeId], ncd.KeyFileBytes) ||
		!bytes.Equal(connectionWrapper.caCertificateBytes[nodeId], ncd.CaCertificateFileBytes) {

		conn, err := cln_connect.Connect(ncd.GRPCAddress, ncd.CertificateFileBytes, ncd.KeyFileBytes,
			ncd.CaCertificateFileBytes)
		if err != nil {
			log.Error().Err(err).Msgf("GRPC connection Failed for node id: %v", nodeId)
			return nil, errors.Wrapf(err, "Connecting to GRPC.")
		}
		connectionWrapper.connections[nodeId] = conn
		connectionWrapper.grpcAddresses[nodeId] = ncd.GRPCAddress
		connectionWrapper.certificateBytes[nodeId] = ncd.CertificateFileBytes
		connectionWrapper.keyBytes[nodeId] = ncd.KeyFileBytes
		connectionWrapper.caCertificateBytes[nodeId] = ncd.CaCertificateFileBytes
		if exists && existingConnection != nil {
			err = existingConnection.Close()
			if err != nil {
				log.Error().Err(err).Msgf("GRPC close connection failed for node id: %v", nodeId)
			}
		}
	}
	return connectionWrapper.connections[nodeId], nil
}

type lightningService struct {
	limit chan struct{}
}

func Information(ctx context.Context,
	request lightning_helpers.InformationRequest) lightning_helpers.InformationResponse {
	ctx, span := otel.Tracer(name).Start(ctx, "Information")
	defer span.End()
	responseChan := make(chan any)
	processSequential(ctx, 2, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_helpers.InformationResponse); ok {
		return res
	}
	return lightning_helpers.InformationResponse{}
}

func SignMessage(ctx context.Context,
	request lightning_helpers.SignMessageRequest) lightning_helpers.SignMessageResponse {
	ctx, span := otel.Tracer(name).Start(ctx, "SignMessage")
	defer span.End()
	responseChan := make(chan any)
	processSequential(ctx, 2, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_helpers.SignMessageResponse); ok {
		return res
	}
	return lightning_helpers.SignMessageResponse{}
}

func SignatureVerification(ctx context.Context,
	request lightning_helpers.SignatureVerificationRequest) lightning_helpers.SignatureVerificationResponse {
	ctx, span := otel.Tracer(name).Start(ctx, "SignatureVerification")
	defer span.End()
	responseChan := make(chan any)
	processSequential(ctx, 2, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_helpers.SignatureVerificationResponse); ok {
		return res
	}
	return lightning_helpers.SignatureVerificationResponse{}
}

func RoutingPolicyUpdate(ctx context.Context,
	request lightning_helpers.RoutingPolicyUpdateRequest) lightning_helpers.RoutingPolicyUpdateResponse {
	ctx, span := otel.Tracer(name).Start(ctx, "RoutingPolicyUpdate")
	defer span.End()
	responseChan := make(chan any)
	processSequential(ctx, 2, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_helpers.RoutingPolicyUpdateResponse); ok {
		return res
	}
	return lightning_helpers.RoutingPolicyUpdateResponse{}
}

func ConnectPeer(ctx context.Context,
	request lightning_helpers.ConnectPeerRequest) lightning_helpers.ConnectPeerResponse {
	ctx, span := otel.Tracer(name).Start(ctx, "ConnectPeer")
	defer span.End()
	responseChan := make(chan any)
	processConcurrent(ctx, 60, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_helpers.ConnectPeerResponse); ok {
		return res
	}
	return lightning_helpers.ConnectPeerResponse{}
}

func DisconnectPeer(ctx context.Context,
	request lightning_helpers.DisconnectPeerRequest) lightning_helpers.DisconnectPeerResponse {
	ctx, span := otel.Tracer(name).Start(ctx, "DisconnectPeer")
	defer span.End()
	responseChan := make(chan any)
	processConcurrent(ctx, 60, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_helpers.DisconnectPeerResponse); ok {
		return res
	}
	return lightning_helpers.DisconnectPeerResponse{}
}

func WalletBalance(ctx context.Context,
	request lightning_helpers.WalletBalanceRequest) lightning_helpers.WalletBalanceResponse {
	ctx, span := otel.Tracer(name).Start(ctx, "WalletBalance")
	defer span.End()
	responseChan := make(chan any)
	processSequential(ctx, 2, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_helpers.WalletBalanceResponse); ok {
		return res
	}
	return lightning_helpers.WalletBalanceResponse{}
}

func ListPeers(ctx context.Context,
	request lightning_helpers.ListPeersRequest) lightning_helpers.ListPeersResponse {
	ctx, span := otel.Tracer(name).Start(ctx, "ListPeers")
	defer span.End()
	responseChan := make(chan any)
	processConcurrent(ctx, 60, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_helpers.ListPeersResponse); ok {
		return res
	}
	return lightning_helpers.ListPeersResponse{}
}

func NewAddress(ctx context.Context,
	request lightning_helpers.NewAddressRequest) lightning_helpers.NewAddressResponse {
	ctx, span := otel.Tracer(name).Start(ctx, "NewAddress")
	defer span.End()
	responseChan := make(chan any)
	processSequential(ctx, 2, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_helpers.NewAddressResponse); ok {
		return res
	}
	return lightning_helpers.NewAddressResponse{}
}

func MoveFundsOffChain(request lightning_helpers.MoveFundsOffChainRequest) lightning_helpers.MoveFundsOffChainResponse {
	responseChan := make(chan any)
	processConcurrent(context.Background(), 2, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_helpers.MoveFundsOffChainResponse); ok {
		return res
	}
	return lightning_helpers.MoveFundsOffChainResponse{}
}

func OpenChannel(ctx context.Context,
	request lightning_helpers.OpenChannelRequest) lightning_helpers.OpenChannelResponse {
	ctx, span := otel.Tracer(name).Start(ctx, "OpenChannel")
	defer span.End()
	responseChan := make(chan any)
	processConcurrent(ctx, 300, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_helpers.OpenChannelResponse); ok {
		return res
	}
	return lightning_helpers.OpenChannelResponse{}
}

func CloseChannel(ctx context.Context,
	request lightning_helpers.CloseChannelRequest) lightning_helpers.CloseChannelResponse {
	ctx, span := otel.Tracer(name).Start(ctx, "CloseChannel")
	defer span.End()
	responseChan := make(chan any)
	processConcurrent(ctx, 300, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_helpers.CloseChannelResponse); ok {
		return res
	}
	return lightning_helpers.CloseChannelResponse{}
}

func NewInvoice(ctx context.Context,
	request lightning_helpers.NewInvoiceRequest) lightning_helpers.NewInvoiceResponse {
	ctx, span := otel.Tracer(name).Start(ctx, "NewInvoice")
	defer span.End()
	responseChan := make(chan any)
	processSequential(ctx, 2, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_helpers.NewInvoiceResponse); ok {
		return res
	}
	return lightning_helpers.NewInvoiceResponse{}
}

func OnChainPayment(ctx context.Context,
	request lightning_helpers.OnChainPaymentRequest) lightning_helpers.OnChainPaymentResponse {
	ctx, span := otel.Tracer(name).Start(ctx, "OnChainPayment")
	defer span.End()
	responseChan := make(chan any)
	processSequential(ctx, 2, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_helpers.OnChainPaymentResponse); ok {
		return res
	}
	return lightning_helpers.OnChainPaymentResponse{}
}

func NewPayment(ctx context.Context,
	request lightning_helpers.NewPaymentRequest) lightning_helpers.NewPaymentResponse {
	ctx, span := otel.Tracer(name).Start(ctx, "NewPayment")
	defer span.End()
	responseChan := make(chan any)
	processConcurrent(ctx, 120, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_helpers.NewPaymentResponse); ok {
		return res
	}
	return lightning_helpers.NewPaymentResponse{}
}

func DecodeInvoice(ctx context.Context,
	request lightning_helpers.DecodeInvoiceRequest) lightning_helpers.DecodeInvoiceResponse {
	ctx, span := otel.Tracer(name).Start(ctx, "DecodeInvoice")
	defer span.End()
	responseChan := make(chan any)
	processSequential(ctx, 2, request, responseChan)
	response := <-responseChan
	if res, ok := response.(lightning_helpers.DecodeInvoiceResponse); ok {
		return res
	}
	return lightning_helpers.DecodeInvoiceResponse{}
}

const concurrentWorkLimit = 10

var serviceSequential = lightningService{limit: make(chan struct{}, 1)}                   //nolint:gochecknoglobals
var serviceConcurrent = lightningService{limit: make(chan struct{}, concurrentWorkLimit)} //nolint:gochecknoglobals

func processSequential(ctx context.Context, timeoutInSeconds int, req any, responseChan chan any) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, time.Duration(timeoutInSeconds)*time.Second)

	select {
	case <-ctx.Done():
		cancel()
		return
	case serviceSequential.limit <- struct{}{}:
	}

	go processRequestSequential(ctx, cancel, req, responseChan)
}

func processRequestSequential(ctx context.Context, cancel context.CancelFunc, req any, responseChan chan<- any) {
	defer func() {
		cancel()
		<-serviceSequential.limit
	}()

	select {
	case <-ctx.Done():
		responseChan <- nil
		return
	default:
	}

	processRequestByType(ctx, req, responseChan)
}

func processConcurrent(ctx context.Context, timeoutInSeconds int, req any, responseChan chan any) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, time.Duration(timeoutInSeconds)*time.Second)

	select {
	case <-ctx.Done():
		cancel()
		return
	case serviceConcurrent.limit <- struct{}{}:
	}

	go processRequestConcurrent(ctx, cancel, req, responseChan)
}

func processRequestConcurrent(ctx context.Context, cancel context.CancelFunc, req any, responseChan chan<- any) {
	defer func() {
		cancel()
		<-serviceConcurrent.limit
	}()

	select {
	case <-ctx.Done():
		responseChan <- nil
		return
	default:
	}

	processRequestByType(ctx, req, responseChan)
}

func processRequestByType(ctx context.Context, req any, responseChan chan<- any) {

	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("CLN is panicking: %v", string(debug.Stack()))

			communicationResponse := lightning_helpers.CommunicationResponse{
				Status: lightning_helpers.Inactive,
				Error:  fmt.Sprintf("CLN is panicking: %v", err),
			}
			switch r := req.(type) {
			case lightning_helpers.InformationRequest:
				responseChan <- lightning_helpers.InformationResponse{
					Request:               r,
					CommunicationResponse: communicationResponse,
				}
				return
			case lightning_helpers.SignMessageRequest:
				responseChan <- lightning_helpers.SignMessageResponse{
					Request:               r,
					CommunicationResponse: communicationResponse,
				}
				return
			case lightning_helpers.SignatureVerificationRequest:
				responseChan <- lightning_helpers.SignatureVerificationResponse{
					Request:               r,
					CommunicationResponse: communicationResponse,
				}
				return
			case lightning_helpers.RoutingPolicyUpdateRequest:
				responseChan <- lightning_helpers.RoutingPolicyUpdateResponse{
					Request:               r,
					CommunicationResponse: communicationResponse,
				}
				return
			case lightning_helpers.ConnectPeerRequest:
				responseChan <- lightning_helpers.ConnectPeerResponse{
					Request:               r,
					CommunicationResponse: communicationResponse,
				}
				return
			case lightning_helpers.DisconnectPeerRequest:
				responseChan <- lightning_helpers.DisconnectPeerResponse{
					Request:               r,
					CommunicationResponse: communicationResponse,
				}
				return
			case lightning_helpers.WalletBalanceRequest:
				responseChan <- lightning_helpers.WalletBalanceResponse{
					Request:               r,
					CommunicationResponse: communicationResponse,
				}
				return
			case lightning_helpers.ListPeersRequest:
				responseChan <- lightning_helpers.ListPeersResponse{
					Request:               r,
					CommunicationResponse: communicationResponse,
				}
				return
			case lightning_helpers.NewAddressRequest:
				responseChan <- lightning_helpers.NewAddressResponse{
					Request:               r,
					CommunicationResponse: communicationResponse,
				}
				return
			case lightning_helpers.OpenChannelRequest:
				responseChan <- lightning_helpers.OpenChannelResponse{
					Request:               r,
					CommunicationResponse: communicationResponse,
				}
				return
			case lightning_helpers.CloseChannelRequest:
				responseChan <- lightning_helpers.CloseChannelResponse{
					Request:               r,
					CommunicationResponse: communicationResponse,
				}
				return
			case lightning_helpers.NewInvoiceRequest:
				responseChan <- lightning_helpers.NewInvoiceResponse{
					Request:               r,
					CommunicationResponse: communicationResponse,
				}
				return
			case lightning_helpers.OnChainPaymentRequest:
				responseChan <- lightning_helpers.OnChainPaymentResponse{
					Request:               r,
					CommunicationResponse: communicationResponse,
				}
				return
			case lightning_helpers.NewPaymentRequest:
				responseChan <- lightning_helpers.NewPaymentResponse{
					Request:               r,
					CommunicationResponse: communicationResponse,
				}
				return
			case lightning_helpers.DecodeInvoiceRequest:
				responseChan <- lightning_helpers.DecodeInvoiceResponse{
					Request:               r,
					CommunicationResponse: communicationResponse,
				}
				return
			}
			responseChan <- nil
			return
		}
	}()

	switch r := req.(type) {
	case lightning_helpers.InformationRequest:
		responseChan <- processGetInfoRequest(ctx, r)
		return
	case lightning_helpers.SignMessageRequest:
		responseChan <- processSignMessageRequest(ctx, r)
		return
	case lightning_helpers.SignatureVerificationRequest:
		responseChan <- processSignatureVerificationRequest(ctx, r)
		return
	case lightning_helpers.RoutingPolicyUpdateRequest:
		responseChan <- processRoutingPolicyUpdateRequest(ctx, r)
		return
	case lightning_helpers.ConnectPeerRequest:
		responseChan <- processConnectPeerRequest(ctx, r)
		return
	case lightning_helpers.DisconnectPeerRequest:
		responseChan <- processDisconnectPeerRequest(ctx, r)
		return
	case lightning_helpers.WalletBalanceRequest:
		responseChan <- processWalletBalanceRequest(ctx, r)
		return
	case lightning_helpers.ListPeersRequest:
		responseChan <- processListPeersRequest(ctx, r)
		return
	case lightning_helpers.NewAddressRequest:
		responseChan <- processNewAddressRequest(ctx, r)
		return
	case lightning_helpers.OpenChannelRequest:
		responseChan <- processOpenChannelRequest(ctx, r)
		return
	case lightning_helpers.CloseChannelRequest:
		responseChan <- processCloseChannelRequest(ctx, r)
		return
	case lightning_helpers.NewInvoiceRequest:
		responseChan <- processNewInvoiceRequest(ctx, r)
		return
	case lightning_helpers.OnChainPaymentRequest:
		responseChan <- processOnChainPaymentRequest(ctx, r)
		return
	case lightning_helpers.NewPaymentRequest:
		responseChan <- processNewPaymentRequest(ctx, r)
		return
	case lightning_helpers.DecodeInvoiceRequest:
		responseChan <- processDecodeInvoiceRequest(ctx, r)
		return
	case lightning_helpers.MoveFundsOffChainRequest:
		responseChan <- processMoveFundsOffChain(ctx, r)
		return
	}

	responseChan <- nil
}

func processMoveFundsOffChain(ctx context.Context, request lightning_helpers.MoveFundsOffChainRequest) lightning_helpers.MoveFundsOffChainResponse {
	response := lightning_helpers.MoveFundsOffChainResponse{
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
		Request: request,
	}

	// short channel id from channel cache
	channel := cache.GetChannelSettingByChannelId(request.ChannelId)
	if channel.ShortChannelId == nil {
		response.Error = fmt.Sprintf("Channel %d not found in cache", request.ChannelId)
		return response
	}

	connection, err := getConnection(request.NodeId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to obtain a GRPC connection.")
		response.Error = err.Error()
		return response
	}

	route := []*cln.SendpayRoute{{
		Id:         request.RHash,
		AmountMsat: &cln.Amount{Msat: uint64(request.AmountMsat)},
		Channel:    *channel.ShortChannelId,
	}}

	p := &cln.SendpayRequest{
		Route:         route,
		PaymentHash:   request.RHash,
		Label:         nil,
		AmountMsat:    nil,
		Bolt11:        nil,
		PaymentSecret: nil,
		Partid:        nil,
		Localinvreqid: nil,
		Groupid:       nil,
	}
	resp, err := cln.NewNodeClient(connection).SendPay(ctx, p)
	if err != nil {
		response.Status = resp.Status.String()
		response.Error = err.Error()
		return response
	}

	wResp, err := cln.NewNodeClient(connection).WaitSendPay(ctx, &cln.WaitsendpayRequest{PaymentHash: request.RHash})
	if err != nil {
		response.Status = wResp.Status.String()
		response.Error = err.Error()
		return response
	}

	response.CommunicationResponse.Status = lightning_helpers.Active
	response.Status = wResp.Status.String()

	return response

}

func processGetInfoRequest(ctx context.Context,
	request lightning_helpers.InformationRequest) lightning_helpers.InformationResponse {

	ctx, span := otel.Tracer(name).Start(ctx, "processGetInfoRequest")
	defer span.End()

	response := lightning_helpers.InformationResponse{
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
		Request: request,
	}

	connection, err := getConnection(request.NodeId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to obtain a GRPC connection.")
		response.Error = err.Error()
		return response
	}

	infoRequest := cln.GetinfoRequest{}
	info, err := cln.NewNodeClient(connection).Getinfo(ctx, &infoRequest)
	if err != nil {
		response.Error = err.Error()
		return response
	}

	response.Implementation = core.CLN
	response.Status = lightning_helpers.Active
	response.PublicKey = hex.EncodeToString(info.Id)
	response.Version = info.Version
	response.Alias = info.Alias
	response.Color = hex.EncodeToString(info.Color)
	response.PendingChannelCount = int(info.NumPendingChannels)
	response.ActiveChannelCount = int(info.NumActiveChannels)
	response.InactiveChannelCount = int(info.NumInactiveChannels)
	response.PeerCount = int(info.NumPeers)
	response.BlockHeight = info.Blockheight
	response.ChainSynced = info.WarningBitcoindSync == nil || *info.WarningBitcoindSync == ""
	response.GraphSynced = info.WarningLightningdSync == nil || *info.WarningLightningdSync == ""
	//response.Addresses = info.Address
	//response.Network = info.Network
	return response
}

func processSignMessageRequest(ctx context.Context,
	request lightning_helpers.SignMessageRequest) lightning_helpers.SignMessageResponse {

	ctx, span := otel.Tracer(name).Start(ctx, "processSignMessageRequest")
	defer span.End()

	response := lightning_helpers.SignMessageResponse{
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
		Request: request,
	}

	connection, err := getConnection(request.NodeId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to obtain a GRPC connection.")
		response.Error = err.Error()
		return response
	}

	signMsgReq := cln.SignmessageRequest{
		Message: request.Message,
	}
	signMsgResp, err := cln.NewNodeClient(connection).SignMessage(ctx, &signMsgReq)
	if err != nil {
		response.Error = err.Error()
		return response
	}

	response.Status = lightning_helpers.Active
	response.Signature = signMsgResp.Zbase
	return response
}

func processSignatureVerificationRequest(ctx context.Context,
	request lightning_helpers.SignatureVerificationRequest) lightning_helpers.SignatureVerificationResponse {

	ctx, span := otel.Tracer(name).Start(ctx, "processSignatureVerificationRequest")
	defer span.End()

	response := lightning_helpers.SignatureVerificationResponse{
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
		Request: request,
	}

	connection, err := getConnection(request.NodeId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to obtain a GRPC connection.")
		response.Error = err.Error()
		return response
	}

	verifyMsgReq := cln.CheckmessageRequest{
		Message: request.Message,
		Zbase:   request.Signature,
	}
	verifyMsgResp, err := cln.NewNodeClient(connection).CheckMessage(ctx, &verifyMsgReq)
	if err != nil {
		response.Error = err.Error()
		return response
	}
	if !verifyMsgResp.Verified {
		response.Message = "Signature is not valid"
		return response
	}

	response.Status = lightning_helpers.Active
	response.PublicKey = hex.EncodeToString(verifyMsgResp.Pubkey)
	response.Valid = verifyMsgResp.GetVerified()
	return response
}

func processRoutingPolicyUpdateRequest(ctx context.Context,
	request lightning_helpers.RoutingPolicyUpdateRequest) lightning_helpers.RoutingPolicyUpdateResponse {

	ctx, span := otel.Tracer(name).Start(ctx, "processRoutingPolicyUpdateRequest")
	defer span.End()

	response := validateRoutingPolicyUpdateRequest(request)
	if response != nil {
		return *response
	}

	channelState := cache.GetChannelState(request.NodeId, request.ChannelId, true)
	if channelState == nil {
		return lightning_helpers.RoutingPolicyUpdateResponse{
			CommunicationResponse: lightning_helpers.CommunicationResponse{
				Status: lightning_helpers.Inactive,
			},
			Request: request,
		}
	}
	if !routingPolicyUpdateRequestContainsUpdates(request, channelState) {
		return lightning_helpers.RoutingPolicyUpdateResponse{
			CommunicationResponse: lightning_helpers.CommunicationResponse{
				Status: lightning_helpers.Active,
			},
			Request: request,
		}
	}

	response = routingPolicyUpdateRequestIsRepeated(request)
	if response != nil {
		return *response
	}

	connection, err := getConnection(request.NodeId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to obtain a GRPC connection.")
		response.Error = err.Error()
		return *response
	}

	resp, err := cln.NewNodeClient(connection).SetChannel(ctx, constructPolicyUpdateRequest(request, channelState))

	// TODO FIXME TIMELOCK CANNOT BE SET VIA SetChannel

	return processRoutingPolicyUpdateResponse(request, resp, err)
}

func processRoutingPolicyUpdateResponse(request lightning_helpers.RoutingPolicyUpdateRequest,
	resp *cln.SetchannelResponse,
	err error) lightning_helpers.RoutingPolicyUpdateResponse {

	if err != nil && resp == nil {
		log.Error().Err(err).Msgf("Failed to update routing policy for channelId: %v on nodeId: %v",
			request.ChannelId, request.NodeId)
		return lightning_helpers.RoutingPolicyUpdateResponse{
			CommunicationResponse: lightning_helpers.CommunicationResponse{
				Status: lightning_helpers.Inactive,
			},
			Request: request,
		}
	}
	var failedUpdateArray []lightning_helpers.FailedRequest
	for _, failedUpdate := range resp.Channels {
		if failedUpdate.WarningHtlcmaxTooHigh != nil {
			log.Error().Msgf("Failed to update routing policy for channelId: %v on nodeId: %v (cln-grpc error: %v)",
				request.ChannelId, request.NodeId, *failedUpdate.WarningHtlcmaxTooHigh)
			failedUpdateArray = append(failedUpdateArray, lightning_helpers.FailedRequest{
				Reason: *failedUpdate.WarningHtlcmaxTooHigh,
				Error:  *failedUpdate.WarningHtlcmaxTooHigh,
			})
		}
		if failedUpdate.WarningHtlcminTooLow != nil {
			log.Error().Msgf("Failed to update routing policy for channelId: %v on nodeId: %v (cln-grpc error: %v)",
				request.ChannelId, request.NodeId, *failedUpdate.WarningHtlcminTooLow)
			failedUpdateArray = append(failedUpdateArray, lightning_helpers.FailedRequest{
				Reason: *failedUpdate.WarningHtlcminTooLow,
				Error:  *failedUpdate.WarningHtlcminTooLow,
			})
		}
	}
	if err != nil || len(failedUpdateArray) != 0 {
		log.Error().Err(err).Msgf("Failed to update routing policy for channelId: %v on nodeId: %v",
			request.ChannelId, request.NodeId)
		return lightning_helpers.RoutingPolicyUpdateResponse{
			CommunicationResponse: lightning_helpers.CommunicationResponse{
				Status: lightning_helpers.Inactive,
			},
			Request:       request,
			FailedUpdates: failedUpdateArray,
		}
	}
	return lightning_helpers.RoutingPolicyUpdateResponse{
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Active,
		},
		Request: request,
	}
}

func constructPolicyUpdateRequest(request lightning_helpers.RoutingPolicyUpdateRequest,
	channelState *cache.ChannelStateSettingsCache) *cln.SetchannelRequest {

	policyUpdateRequest := &cln.SetchannelRequest{}

	var feePpm uint32
	if request.FeeRateMilliMsat == nil {
		feePpm = uint32(channelState.LocalFeeRateMilliMsat)
	} else {
		feePpm = uint32(*request.FeeRateMilliMsat)
	}
	policyUpdateRequest.Feeppm = &feePpm

	var feeBase uint64
	if request.FeeBaseMsat == nil {
		feeBase = uint64(channelState.LocalFeeBaseMsat)
	} else {
		feeBase = uint64(*request.FeeBaseMsat)
	}
	policyUpdateRequest.Feebase = &cln.Amount{Msat: feeBase}

	var minHtlcMsat uint64
	if request.MinHtlcMsat == nil {
		minHtlcMsat = channelState.LocalMinHtlcMsat
	} else {
		minHtlcMsat = *request.MinHtlcMsat
	}
	policyUpdateRequest.Htlcmin = &cln.Amount{Msat: minHtlcMsat}

	var maxHtlcMsat uint64
	if request.MaxHtlcMsat == nil {
		maxHtlcMsat = channelState.LocalMaxHtlcMsat
	} else {
		maxHtlcMsat = *request.MaxHtlcMsat
	}
	policyUpdateRequest.Htlcmax = &cln.Amount{Msat: maxHtlcMsat}

	channelSettings := cache.GetChannelSettingByChannelId(request.ChannelId)
	policyUpdateRequest.Id = *channelSettings.ShortChannelId
	return policyUpdateRequest
}

func validateRoutingPolicyUpdateRequest(
	request lightning_helpers.RoutingPolicyUpdateRequest) *lightning_helpers.RoutingPolicyUpdateResponse {

	if request.FeeRateMilliMsat == nil &&
		request.FeeBaseMsat == nil &&
		request.MaxHtlcMsat == nil &&
		request.MinHtlcMsat == nil &&
		request.TimeLockDelta == nil {
		return &lightning_helpers.RoutingPolicyUpdateResponse{
			CommunicationResponse: lightning_helpers.CommunicationResponse{
				Status:  lightning_helpers.Active,
				Message: "Nothing changed so update is ignored",
			},
			Request: request,
		}
	}
	if request.ChannelId == 0 {
		return &lightning_helpers.RoutingPolicyUpdateResponse{
			CommunicationResponse: lightning_helpers.CommunicationResponse{
				Status: lightning_helpers.Inactive,
				Error:  "ChannelId is 0",
			},
			Request: request,
		}
	}
	if request.TimeLockDelta != nil && *request.TimeLockDelta < 18 {
		return &lightning_helpers.RoutingPolicyUpdateResponse{
			CommunicationResponse: lightning_helpers.CommunicationResponse{
				Status: lightning_helpers.Inactive,
				Error:  "TimeLockDelta is < 18",
			},
			Request: request,
		}
	}
	return nil
}

func routingPolicyUpdateRequestContainsUpdates(request lightning_helpers.RoutingPolicyUpdateRequest,
	channelState *cache.ChannelStateSettingsCache) bool {

	if request.TimeLockDelta != nil && *request.TimeLockDelta != channelState.LocalTimeLockDelta {
		return true
	}
	if request.FeeRateMilliMsat != nil && *request.FeeRateMilliMsat != channelState.LocalFeeRateMilliMsat {
		return true
	}
	if request.FeeBaseMsat != nil && *request.FeeBaseMsat != channelState.LocalFeeBaseMsat {
		return true
	}
	if request.MinHtlcMsat != nil && *request.MinHtlcMsat != channelState.LocalMinHtlcMsat {
		return true
	}
	if request.MaxHtlcMsat != nil && *request.MaxHtlcMsat != channelState.LocalMaxHtlcMsat {
		return true
	}
	return false
}

func routingPolicyUpdateRequestIsRepeated(
	request lightning_helpers.RoutingPolicyUpdateRequest) *lightning_helpers.RoutingPolicyUpdateResponse {

	rateLimitSeconds := routingPolicyUpdateLimiterSeconds
	if request.RateLimitSeconds > 0 {
		rateLimitSeconds = request.RateLimitSeconds
	}
	channelEventsFromGraph, err := graph_events.GetChannelEventFromGraph(request.Db, request.ChannelId, &rateLimitSeconds)
	if err != nil {
		return &lightning_helpers.RoutingPolicyUpdateResponse{
			CommunicationResponse: lightning_helpers.CommunicationResponse{
				Status: lightning_helpers.Inactive,
				Error:  err.Error(),
			},
			Request: request,
		}
	}

	if len(channelEventsFromGraph) > 1 {
		timeLockDelta := channelEventsFromGraph[0].TimeLockDelta
		timeLockDeltaCounter := 1
		minHtlcMsat := channelEventsFromGraph[0].MinHtlcMsat
		minHtlcMsatCounter := 1
		maxHtlcMsat := channelEventsFromGraph[0].MaxHtlcMsat
		maxHtlcMsatCounter := 1
		feeBaseMsat := channelEventsFromGraph[0].FeeBaseMsat
		feeBaseMsatCounter := 1
		feeRateMilliMsat := channelEventsFromGraph[0].FeeRateMilliMsat
		feeRateMilliMsatCounter := 1
		for i := 0; i < len(channelEventsFromGraph); i++ {
			if timeLockDelta != channelEventsFromGraph[i].TimeLockDelta {
				timeLockDeltaCounter++
				timeLockDelta = channelEventsFromGraph[i].TimeLockDelta
			}
			if minHtlcMsat != channelEventsFromGraph[i].MinHtlcMsat {
				minHtlcMsatCounter++
				minHtlcMsat = channelEventsFromGraph[i].MinHtlcMsat
			}
			if maxHtlcMsat != channelEventsFromGraph[i].MaxHtlcMsat {
				maxHtlcMsatCounter++
				maxHtlcMsat = channelEventsFromGraph[i].MaxHtlcMsat
			}
			if feeBaseMsat != channelEventsFromGraph[i].FeeBaseMsat {
				feeBaseMsatCounter++
				feeBaseMsat = channelEventsFromGraph[i].FeeBaseMsat
			}
			if feeRateMilliMsat != channelEventsFromGraph[i].FeeRateMilliMsat {
				feeRateMilliMsatCounter++
				feeRateMilliMsat = channelEventsFromGraph[i].FeeRateMilliMsat
			}
		}
		rateLimitCount := 2
		if request.RateLimitCount > 0 {
			rateLimitCount = request.RateLimitCount
		}
		if timeLockDeltaCounter >= rateLimitCount ||
			minHtlcMsatCounter >= rateLimitCount || maxHtlcMsatCounter >= rateLimitCount ||
			feeBaseMsatCounter >= rateLimitCount || feeRateMilliMsatCounter >= rateLimitCount {

			return &lightning_helpers.RoutingPolicyUpdateResponse{
				CommunicationResponse: lightning_helpers.CommunicationResponse{
					Status: lightning_helpers.Inactive,
					Error: fmt.Sprintf("Routing policy update ignored due to rate limiter for channelId: %v",
						request.ChannelId),
				},
				Request: request,
			}
		}
	}
	return nil
}

func processConnectPeerRequest(ctx context.Context,
	request lightning_helpers.ConnectPeerRequest) lightning_helpers.ConnectPeerResponse {

	ctx, span := otel.Tracer(name).Start(ctx, "processConnectPeerRequest")
	defer span.End()

	response := lightning_helpers.ConnectPeerResponse{
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
		Request:                       request,
		RequestFailCurrentlyConnected: false,
	}

	connection, err := getConnection(request.NodeId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to obtain a GRPC connection.")
		response.Error = err.Error()
		return response
	}

	host := request.Host
	var port *uint32
	if strings.Contains(request.Host, ":") {
		host = request.Host[:strings.Index(request.Host, ":")]
		portInt, err := strconv.ParseUint(request.Host[strings.Index(request.Host, ":")+1:], 10, 32)
		if err == nil {
			p := uint32(portInt)
			port = &p
		}
	}
	connectPeerRequest := cln.ConnectRequest{
		Id:   request.PublicKey,
		Host: &host,
		Port: port,
	}

	_, err = cln.NewNodeClient(connection).ConnectPeer(ctx, &connectPeerRequest)

	if err != nil {
		response.Error = err.Error()
		return response
	}

	nodeSettings := cache.GetNodeSettingsByNodeId(request.NodeId)
	eventNodeId := cache.GetPeerNodeIdByPublicKey(request.PublicKey, nodeSettings.Chain, nodeSettings.Network)
	cache.SetConnectedPeerNode(eventNodeId, request.PublicKey, nodeSettings.Chain, nodeSettings.Network)
	response.Status = lightning_helpers.Active
	return response
}

func processDisconnectPeerRequest(ctx context.Context,
	request lightning_helpers.DisconnectPeerRequest) lightning_helpers.DisconnectPeerResponse {

	ctx, span := otel.Tracer(name).Start(ctx, "processDisconnectPeerRequest")
	defer span.End()

	response := lightning_helpers.DisconnectPeerResponse{
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
		Request:                            request,
		RequestFailedCurrentlyDisconnected: false,
	}

	connection, err := getConnection(request.NodeId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to obtain a GRPC connection.")
		response.Error = err.Error()
		return response
	}

	publicKey, err := hex.DecodeString(cache.GetNodeSettingsByNodeId(request.PeerNodeId).PublicKey)
	if err != nil {
		response.Error = err.Error()
		return response
	}

	force := true
	disconnectPeerRequest := cln.DisconnectRequest{
		Id:    publicKey,
		Force: &force,
	}

	_, err = cln.NewNodeClient(connection).Disconnect(ctx, &disconnectPeerRequest)
	if err != nil {
		response.Error = err.Error()
		return response
	}

	peer := cache.GetNodeSettingsByNodeId(request.PeerNodeId)
	cache.RemoveConnectedPeerNode(peer.NodeId, peer.PublicKey, peer.Chain, peer.Network)
	response.Status = lightning_helpers.Active
	return response
}

func processWalletBalanceRequest(ctx context.Context,
	request lightning_helpers.WalletBalanceRequest) lightning_helpers.WalletBalanceResponse {

	ctx, span := otel.Tracer(name).Start(ctx, "processWalletBalanceRequest")
	defer span.End()

	response := lightning_helpers.WalletBalanceResponse{
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
		Request: request,
	}

	connection, err := getConnection(request.NodeId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to obtain a GRPC connection.")
		response.Error = err.Error()
		return response
	}

	funds, err := cln.NewNodeClient(connection).ListFunds(ctx, &cln.ListfundsRequest{})
	if err != nil {
		response.Error = err.Error()
		return response
	}

	response.Status = lightning_helpers.Active
	var unconfirmedBalance int64
	var confirmedBalance int64
	var lockedBalance int64
	for _, utxo := range funds.Outputs {
		if utxo.AmountMsat != nil {
			switch utxo.Status {
			case cln.ListfundsOutputs_UNCONFIRMED, cln.ListfundsOutputs_IMMATURE:
				unconfirmedBalance += int64(utxo.AmountMsat.Msat) / 1_000
			case cln.ListfundsOutputs_CONFIRMED:
				confirmedBalance += int64(utxo.AmountMsat.Msat) / 1_000
			case cln.ListfundsOutputs_SPENT:
			}
			if utxo.Reserved {
				lockedBalance += int64(utxo.AmountMsat.Msat) / 1_000
			}
		}
	}
	response.ReservedBalanceAnchorChan = 0
	response.UnconfirmedBalance = unconfirmedBalance
	response.ConfirmedBalance = confirmedBalance
	response.TotalBalance = confirmedBalance + unconfirmedBalance
	response.LockedBalance = lockedBalance

	return response
}

func processListPeersRequest(ctx context.Context,
	request lightning_helpers.ListPeersRequest) lightning_helpers.ListPeersResponse {

	ctx, span := otel.Tracer(name).Start(ctx, "processListPeersRequest")
	defer span.End()

	response := lightning_helpers.ListPeersResponse{
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
		Request: request,
	}

	connection, err := getConnection(request.NodeId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to obtain a GRPC connection.")
		response.Error = err.Error()
		return response
	}

	rsp, err := cln.NewNodeClient(connection).ListPeers(ctx, &cln.ListpeersRequest{})

	if err != nil {
		response.Error = err.Error()
		return response
	}

	peers := make(map[string]lightning_helpers.Peer)
	for _, peer := range rsp.Peers {
		if peer != nil && peer.Connected {
			peers[hex.EncodeToString(peer.Id)] = lightning_helpers.GetPeerCLN(peer)
		}
	}

	response.Status = lightning_helpers.Active
	response.Peers = peers

	return response
}

func processNewAddressRequest(ctx context.Context,
	request lightning_helpers.NewAddressRequest) lightning_helpers.NewAddressResponse {

	ctx, span := otel.Tracer(name).Start(ctx, "processNewAddressRequest")
	defer span.End()

	response := lightning_helpers.NewAddressResponse{
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
		Request: request,
	}

	connection, err := getConnection(request.NodeId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to obtain a GRPC connection.")
		response.Error = err.Error()
		return response
	}

	// TODO FIXME CLN implementation is temporary
	clnAddressRequest := &cln.NewaddrRequest{}
	bech32 := cln.NewaddrRequest_BECH32
	clnAddressRequest.Addresstype = &bech32

	rsp, err := cln.NewNodeClient(connection).NewAddr(ctx, clnAddressRequest)
	if err != nil {
		response.Error = err.Error()
		return response
	}
	response.Status = lightning_helpers.Active
	if rsp.P2ShSegwit != nil {
		response.Address = *rsp.P2ShSegwit
	}
	if response.Address == "" && rsp.Bech32 != nil {
		response.Address = *rsp.Bech32
	}
	return response
}

func processOpenChannelRequest(ctx context.Context,
	request lightning_helpers.OpenChannelRequest) lightning_helpers.OpenChannelResponse {

	ctx, span := otel.Tracer(name).Start(ctx, "processOpenChannelRequest")
	defer span.End()

	response := lightning_helpers.OpenChannelResponse{
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
		Request: request,
	}

	connection, err := getConnection(request.NodeId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to obtain a GRPC connection.")
		response.Error = err.Error()
		return response
	}

	openChanReq, err := prepareOpenRequest(request)
	if err != nil {
		response.Error = err.Error()
		return response
	}

	channel, err := cln.NewNodeClient(connection).FundChannel(ctx, openChanReq)
	if err != nil {
		response.Error = err.Error()
		return response
	}

	response.ChannelPoint = fmt.Sprintf("%v:%v", hex.EncodeToString(channel.Txid), channel.Outnum)
	response.ChannelStatus = core.Opening
	response.FundingTransactionHash = hex.EncodeToString(channel.Txid)
	response.FundingOutputIndex = channel.Outnum
	response.Status = lightning_helpers.Active
	return response
}

func prepareOpenRequest(request lightning_helpers.OpenChannelRequest) (*cln.FundchannelRequest, error) {
	if request.SatPerVbyte != nil && request.TargetConf != nil {
		return nil, errors.New("Cannot set both SatPerVbyte and TargetConf")
	}

	pubKeyHex, err := hex.DecodeString(request.NodePubKey)
	if err != nil {
		return nil, errors.New("error decoding public key hex")
	}

	//open channel request
	openChanReq := &cln.FundchannelRequest{
		Id: pubKeyHex,

		// This is the amount we are putting into the channel (channel size)
		Amount: &cln.AmountOrAll{Value: &cln.AmountOrAll_Amount{Amount: &cln.Amount{Msat: uint64(request.LocalFundingAmount * 1_000)}}},
	}

	// The amount to give the other node in the opening process.
	// NB: This means you will give the other node this amount of sats
	if request.PushSat != nil {
		openChanReq.PushMsat = &cln.Amount{Msat: uint64((*request.PushSat) * 1_000)}
	}

	if request.SatPerVbyte != nil {
		// TODO FIXME CLN verify
		openChanReq.Feerate = &cln.Feerate{Style: &cln.Feerate_Perkb{Perkb: uint32(*request.SatPerVbyte)}}
	}

	if request.TargetConf != nil {
		minDept := uint32(*request.TargetConf)
		openChanReq.Mindepth = &minDept
	}

	if request.Private != nil {
		announced := !*request.Private
		openChanReq.Announce = &announced
	}

	// TODO FIXME CLN verify
	//if request.MinHtlcMsat != nil {
	//	openChanReq.MinHtlcMsat = int64(*request.MinHtlcMsat)
	//}

	// TODO FIXME CLN verify
	//if request.RemoteCsvDelay != nil {
	//	openChanReq.RemoteCsvDelay = *request.RemoteCsvDelay
	//}

	if request.MinConfs != nil {
		minConf := uint32(*request.MinConfs)
		openChanReq.Minconf = &minConf
	}

	// TODO FIXME CLN verify
	//if request.SpendUnconfirmed != nil {
	//	openChanReq.SpendUnconfirmed = *request.SpendUnconfirmed
	//}

	if request.CloseAddress != nil {
		openChanReq.CloseTo = request.CloseAddress
	}
	return openChanReq, nil
}

func processCloseChannelRequest(ctx context.Context,
	request lightning_helpers.CloseChannelRequest) lightning_helpers.CloseChannelResponse {

	ctx, span := otel.Tracer(name).Start(ctx, "processCloseChannelRequest")
	defer span.End()

	response := lightning_helpers.CloseChannelResponse{
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
		Request: request,
	}

	connection, err := getConnection(request.NodeId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to obtain a GRPC connection.")
		response.Error = err.Error()
		return response
	}

	closeChanReq, err := prepareCloseRequest(request)
	if err != nil {
		response.Error = err.Error()
		return response
	}

	channel, err := cln.NewNodeClient(connection).Close(ctx, closeChanReq)
	if err != nil {
		response.Error = err.Error()
		return response
	}

	response.ChannelStatus = core.Closing
	response.ClosingTransactionHash = hex.EncodeToString(channel.Txid)
	response.Status = lightning_helpers.Active
	return response
}

func prepareCloseRequest(request lightning_helpers.CloseChannelRequest) (*cln.CloseRequest, error) {
	if request.SatPerVbyte != nil && request.TargetConf != nil {
		return nil, errors.New("Cannot set both SatPerVbyte and TargetConf")
	}
	channel := cache.GetChannelSettingByChannelId(request.ChannelId)
	closeChanReq := &cln.CloseRequest{
		Id: *channel.ShortChannelId,
	}
	if request.Force != nil {
		closeChanReq.ForceLeaseClosed = request.Force
	}
	// TODO FIXME CLN verify
	//if request.TargetConf != nil {
	//	closeChanReq.TargetConf = *ccReq.TargetConf
	//}
	if request.SatPerVbyte != nil {
		closeChanReq.Feerange = append(closeChanReq.Feerange, &cln.Feerate{Style: &cln.Feerate_Perkb{Perkb: uint32(*request.SatPerVbyte)}})
		closeChanReq.Feerange = append(closeChanReq.Feerange, &cln.Feerate{Style: &cln.Feerate_Perkb{Perkb: uint32(*request.SatPerVbyte)}})
	}
	// TODO FIXME CLN verify
	//if request.DeliveryAddress != nil {
	//	closeChanReq.DeliveryAddress = *ccReq.DeliveryAddress
	//}
	return closeChanReq, nil
}

func processNewInvoiceRequest(ctx context.Context,
	request lightning_helpers.NewInvoiceRequest) lightning_helpers.NewInvoiceResponse {

	ctx, span := otel.Tracer(name).Start(ctx, "processNewInvoiceRequest")
	defer span.End()

	response := lightning_helpers.NewInvoiceResponse{
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
		Request: request,
	}

	connection, err := getConnection(request.NodeId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to obtain a GRPC connection.")
		response.Error = err.Error()
		return response
	}

	invoiceRequest, err := prepareInvoiceRequest(request)
	if err != nil {
		response.Error = err.Error()
		return response
	}

	resp, err := cln.NewNodeClient(connection).Invoice(ctx, invoiceRequest)
	if err != nil {
		response.Error = err.Error()
		return response
	}

	response.PaymentAddress = hex.EncodeToString(resp.PaymentHash)
	response.PaymentRequest = resp.Bolt11
	response.RHash = resp.PaymentHash
	response.Status = lightning_helpers.Active
	return response
}

func prepareInvoiceRequest(request lightning_helpers.NewInvoiceRequest) (*cln.InvoiceRequest, error) {
	req := &cln.InvoiceRequest{}
	if request.Memo != nil && *request.Memo != "" {
		req.Description = *request.Memo
	}
	if request.Expiry != nil {
		expiry := uint64(*request.Expiry)
		req.Expiry = &expiry
	}
	if request.FallBackAddress != nil && *request.FallBackAddress != "" {
		req.Fallbacks = append(req.Fallbacks, *request.FallBackAddress)
	}
	if request.RPreImage != nil && *request.RPreImage != "" {
		preimage, err := hex.DecodeString(*request.RPreImage)
		if err != nil {
			return nil, errors.New("decoding pre image")
		}
		req.Preimage = preimage
	}
	if request.ValueMsat != nil {
		amountMsat := uint64(*request.ValueMsat)
		req.AmountMsat = &cln.AmountOrAny{Value: &cln.AmountOrAny_Amount{Amount: &cln.Amount{Msat: amountMsat}}}
	}
	// TODO FIXME CLN Better labeling?
	req.Label = time.Now().UTC().Format("20060102.150405.000000")
	// TODO FIXME CLN AMP?
	return req, nil
}

func processOnChainPaymentRequest(ctx context.Context,
	request lightning_helpers.OnChainPaymentRequest) lightning_helpers.OnChainPaymentResponse {

	ctx, span := otel.Tracer(name).Start(ctx, "processOnChainPaymentRequest")
	defer span.End()

	response := lightning_helpers.OnChainPaymentResponse{
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
		Request: request,
	}

	connection, err := getConnection(request.NodeId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to obtain a GRPC connection.")
		response.Error = err.Error()
		return response
	}

	wr := &cln.WithdrawRequest{
		Destination: request.Address,
	}
	if request.SendAll != nil && *request.SendAll {
		wr.Satoshi = &cln.AmountOrAll{Value: &cln.AmountOrAll_All{All: true}}
	} else {
		wr.Satoshi = &cln.AmountOrAll{Value: &cln.AmountOrAll_Amount{Amount: &cln.Amount{Msat: uint64(request.AmountSat)}}}
	}
	// TODO FIXME CLN: Incorporate target conf / send unconfirmed / Label
	if request.SatPerVbyte != nil {
		wr.Feerate = &cln.Feerate{Style: &cln.Feerate_Perkb{Perkb: uint32(*request.SatPerVbyte)}}
	}
	if request.MinConfs != nil {
		minConfs := uint32(*request.MinConfs)
		wr.Minconf = &minConfs
	}

	resp, err := cln.NewNodeClient(connection).Withdraw(ctx, wr)

	if err != nil {
		response.Error = err.Error()
		return response
	}

	response.TxId = hex.EncodeToString(resp.Txid)
	response.Status = lightning_helpers.Active
	return response
}

func processNewPaymentRequest(ctx context.Context,
	request lightning_helpers.NewPaymentRequest) lightning_helpers.NewPaymentResponse {

	ctx, span := otel.Tracer(name).Start(ctx, "processNewPaymentRequest")
	defer span.End()

	response := lightning_helpers.NewPaymentResponse{
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
		Request: request,
	}

	connection, err := getConnection(request.NodeId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to obtain a GRPC connection.")
		response.Error = err.Error()
		return response
	}

	timeOutSecs := uint32(request.TimeOutSecs)
	p := &cln.PayRequest{
		// TODO FIXME CLN Add support for more of these options to our UI
		Bolt11:     *request.Invoice,
		AmountMsat: &cln.Amount{Msat: uint64(*request.AmtMSat)},
		//Label:         request.,
		//Riskfactor:    request.,
		//Maxfeepercent: request.,
		RetryFor: &timeOutSecs,
		//Maxdelay: request.,
		//Exemptfee:     request.,
		//Localinvreqid: request.,
		//Exclude:       request.,
		Maxfee: &cln.Amount{Msat: uint64(*request.FeeLimitMsat)},
		//Description:   request.,
	}

	resp, err := cln.NewNodeClient(connection).Pay(ctx, p)
	if err != nil {
		response.Error = err.Error()
		return response
	}

	response.Hash = hex.EncodeToString(resp.PaymentHash)
	response.PaymentStatus = resp.Status.String()
	response.Preimage = hex.EncodeToString(resp.PaymentPreimage)
	response.CreationDate = time.Unix(int64(resp.CreatedAt), 0)
	if resp.AmountMsat != nil {
		response.AmountMsat = int64((*resp.AmountMsat).Msat)
	}
	if resp.AmountSentMsat != nil && resp.AmountMsat != nil {
		response.FeePaidMsat = int64((*resp.AmountSentMsat).Msat - (*resp.AmountMsat).Msat)
	}
	if resp.WarningPartialCompletion != nil {
		response.FailureReason = *resp.WarningPartialCompletion
	}
	response.Status = lightning_helpers.Active
	return response
}

func processDecodeInvoiceRequest(ctx context.Context,
	request lightning_helpers.DecodeInvoiceRequest) lightning_helpers.DecodeInvoiceResponse {

	ctx, span := otel.Tracer(name).Start(ctx, "processDecodeInvoiceRequest")
	defer span.End()

	response := lightning_helpers.DecodeInvoiceResponse{
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
		Request: request,
	}

	connection, err := getConnection(request.NodeId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to obtain a GRPC connection.")
		response.Error = err.Error()
		return response
	}

	client := cln.NewNodeClient(connection)
	decodedInvoice, err := client.Decode(ctx, &cln.DecodeRequest{String_: request.Invoice})
	// TODO: Handle different error types like incorrect checksum etc to explain why the decode failed.
	if err != nil {
		response.Error = err.Error()
		return response
	}

	response = constructDecodedInvoice(decodedInvoice, response)
	if response.NodeAlias == "" {
		pk, err := hex.DecodeString(response.DestinationPubKey)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to decode for public key: %v", response.DestinationPubKey)
			return response
		}
		clnNode, err := client.ListNodes(ctx, &cln.ListnodesRequest{
			Id: pk,
		})
		if err != nil {
			log.Error().Err(err).Msgf("Failed to obtain node alias for public key: %v", response.DestinationPubKey)
			return response
		}
		if clnNode != nil && len(clnNode.Nodes) != 0 && clnNode.Nodes[0] != nil && clnNode.Nodes[0].Alias != nil {
			response.NodeAlias = *clnNode.Nodes[0].Alias
		}
	}
	return response
}

func constructDecodedInvoice(decodedInvoice *cln.DecodeResponse,
	response lightning_helpers.DecodeInvoiceResponse) lightning_helpers.DecodeInvoiceResponse {
	if decodedInvoice == nil || !decodedInvoice.Valid {
		return response
	}
	response.Status = lightning_helpers.Active
	response.DestinationPubKey = hex.EncodeToString(decodedInvoice.Payee)
	nodeSettings := cache.GetNodeSettingsByNodeId(response.Request.NodeId)
	response.NodeAlias = cache.GetNodeAlias(cache.GetPeerNodeIdByPublicKey(response.DestinationPubKey, nodeSettings.Chain, nodeSettings.Network))
	response.RHash = hex.EncodeToString(decodedInvoice.PaymentHash)
	if decodedInvoice.InvoiceAmountMsat != nil {
		response.ValueMsat = int64((*decodedInvoice.InvoiceAmountMsat).Msat)
	}
	if decodedInvoice.InvoiceFallbacks != nil {
		for _, fb := range decodedInvoice.InvoiceFallbacks {
			if response.FallbackAddr == "" {
				response.FallbackAddr = fb.GetAddress()
			}
		}
	}
	if decodedInvoice.CreatedAt != nil {
		response.CreatedAt = time.Unix(int64(*decodedInvoice.CreatedAt), 0)
		if decodedInvoice.Expiry != nil {
			response.ExpireAt = response.ExpireAt.Add(time.Duration(*decodedInvoice.Expiry) * time.Second)
		}
	}
	if decodedInvoice.Expiry != nil {
		response.Expiry = int64(*decodedInvoice.Expiry)
	}
	if decodedInvoice.MinFinalCltvExpiry != nil {
		response.CltvExpiry = int64(*decodedInvoice.MinFinalCltvExpiry)
	}
	//response.RouteHints = constructRoutes(decodedInvoice.Routes)
	response.Features = constructFeatureMap(decodedInvoice.InvoiceFeatures)
	return response
}

//func constructRoutes(routes *cln.Routes) []lightning_helpers.RouteHint {
//	var r []lightning_helpers.RouteHint
//
//	if routes != nil {
//		for _, rh := range routes.Routes {
//			var hopHints []lightning_helpers.HopHint
//			for _, hh := range rh.Hops {
//				feeBaseMsat := uint32(0)
//				if hh.FeeBaseMsat != nil {
//					feeBaseMsat = uint32((*hh.FeeBaseMsat).Msat)
//				}
//				hopHints = append(hopHints, lightning_helpers.HopHint{
//					ShortChannelId:         hh.ShortChannelId,
//					ChannelSourcePublicKey: hex.EncodeToString(hh.Pubkey),
//					FeeBaseMsat:            feeBaseMsat,
//					CltvExpiryDelta:        hh.CltvExpiryDelta,
//					FeeProportional:        hh.FeeProportionalMillionths,
//				})
//			}
//			r = append(r, lightning_helpers.RouteHint{
//				HopHints: hopHints,
//			})
//		}
//	}
//
//	return r
//}

func constructFeatureMap(features []byte) lightning_helpers.FeatureMap {
	f := lightning_helpers.FeatureMap{}
	//for n, v := range features {
	//	f[n] = lightning_helpers.Feature{
	//		Name:       v.Name,
	//		IsKnown:    v.IsKnown,
	//		IsRequired: v.IsRequired,
	//	}
	//}
	return f
}
