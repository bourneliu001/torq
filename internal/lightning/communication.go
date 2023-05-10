package lightning

import (
	"context"

	"github.com/cockroachdb/errors"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/cln"
	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/internal/lightning_helpers"
	"github.com/lncapital/torq/internal/lnd"
)

var ServiceInactiveError = errors.New("service is not active")         //nolint:gochecknoglobals
var UnsupportedOperationError = errors.New("request is not supported") //nolint:gochecknoglobals

func GetInformation(ctx context.Context, nodeId int) (lightning_helpers.InformationResponse, error) {
	request := lightning_helpers.InformationRequest{
		CommunicationRequest: lightning_helpers.CommunicationRequest{
			NodeId: nodeId,
		},
	}

	response := lightning_helpers.InformationResponse{
		Request: request,
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
	}

	nodeConnectionDetails := cache.GetNodeConnectionDetails(nodeId)
	switch nodeConnectionDetails.Implementation {
	case core.LND:
		if !cache.IsLndServiceActive(nodeId) {
			return lightning_helpers.InformationResponse{}, ServiceInactiveError
		}
		response = lnd.Information(request)
	case core.CLN:
		if !cache.IsClnServiceActive(nodeId) {
			return lightning_helpers.InformationResponse{}, ServiceInactiveError
		}
		response = cln.Information(ctx, request)
	}
	if response.Error != "" {
		return lightning_helpers.InformationResponse{}, errors.New(response.Error)
	}
	return response, nil
}

func SignMessage(ctx context.Context, nodeId int, message string, singleHash *bool) (string, error) {
	request := lightning_helpers.SignMessageRequest{
		CommunicationRequest: lightning_helpers.CommunicationRequest{
			NodeId: nodeId,
		},
		Message:    message,
		SingleHash: singleHash,
	}

	response := lightning_helpers.SignMessageResponse{
		Request: request,
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
	}

	nodeConnectionDetails := cache.GetNodeConnectionDetails(nodeId)
	switch nodeConnectionDetails.Implementation {
	case core.LND:
		if !cache.IsLndServiceActive(nodeId) {
			return "", ServiceInactiveError
		}
		response = lnd.SignMessage(request)
	case core.CLN:
		if !cache.IsClnServiceActive(nodeId) {
			return "", ServiceInactiveError
		}
		response = cln.SignMessage(ctx, request)
	}
	if response.Error != "" {
		return "", errors.New(response.Error)
	}
	return response.Signature, nil
}

func SignatureVerification(ctx context.Context, nodeId int, message string, signature string) (string, bool, error) {
	request := lightning_helpers.SignatureVerificationRequest{
		CommunicationRequest: lightning_helpers.CommunicationRequest{
			NodeId: nodeId,
		},
		Message:   message,
		Signature: signature,
	}

	response := lightning_helpers.SignatureVerificationResponse{
		Request: request,
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
	}

	nodeConnectionDetails := cache.GetNodeConnectionDetails(nodeId)
	switch nodeConnectionDetails.Implementation {
	case core.LND:
		if !cache.IsLndServiceActive(nodeId) {
			return "", false, ServiceInactiveError
		}
		response = lnd.SignatureVerification(request)
	case core.CLN:
		if !cache.IsClnServiceActive(nodeId) {
			return "", false, ServiceInactiveError
		}
		response = cln.SignatureVerification(ctx, request)
	}
	if response.Error != "" {
		return "", false, errors.New(response.Error)
	}
	return response.PublicKey, response.Valid, nil
}

func SetRoutingPolicy(ctx context.Context,
	request lightning_helpers.RoutingPolicyUpdateRequest) (lightning_helpers.RoutingPolicyUpdateResponse, error) {

	response := lightning_helpers.RoutingPolicyUpdateResponse{
		Request: request,
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
	}

	nodeConnectionDetails := cache.GetNodeConnectionDetails(request.NodeId)
	switch nodeConnectionDetails.Implementation {
	case core.LND:
		if !cache.IsLndServiceActive(request.NodeId) {
			return lightning_helpers.RoutingPolicyUpdateResponse{}, ServiceInactiveError
		}
		response = lnd.RoutingPolicyUpdate(request)
	case core.CLN:
		if !cache.IsClnServiceActive(request.NodeId) {
			return lightning_helpers.RoutingPolicyUpdateResponse{}, ServiceInactiveError
		}
		response = cln.RoutingPolicyUpdate(ctx, request)
	}
	if response.Error != "" {
		return lightning_helpers.RoutingPolicyUpdateResponse{}, errors.New(response.Error)
	}
	return response, nil
}

func ConnectPeer(ctx context.Context, nodeId int, publicKey string, host string) (bool, error) {
	request := lightning_helpers.ConnectPeerRequest{
		CommunicationRequest: lightning_helpers.CommunicationRequest{
			NodeId: nodeId,
		},
		PublicKey: publicKey,
		Host:      host,
	}

	response := lightning_helpers.ConnectPeerResponse{
		Request: request,
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
	}

	nodeConnectionDetails := cache.GetNodeConnectionDetails(nodeId)
	switch nodeConnectionDetails.Implementation {
	case core.LND:
		if !cache.IsLndServiceActive(nodeId) {
			return false, ServiceInactiveError
		}
		response = lnd.ConnectPeer(request)
	case core.CLN:
		if !cache.IsClnServiceActive(nodeId) {
			return false, ServiceInactiveError
		}
		response = cln.ConnectPeer(ctx, request)
	}
	if response.Error != "" {
		return false, errors.New(response.Error)
	}
	return response.RequestFailCurrentlyConnected, nil
}

func DisconnectPeer(ctx context.Context, nodeId int, peerNodeId int) (bool, error) {
	request := lightning_helpers.DisconnectPeerRequest{
		CommunicationRequest: lightning_helpers.CommunicationRequest{
			NodeId: nodeId,
		},
		PeerNodeId: peerNodeId,
	}

	response := lightning_helpers.DisconnectPeerResponse{
		Request: request,
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
	}

	nodeConnectionDetails := cache.GetNodeConnectionDetails(nodeId)
	switch nodeConnectionDetails.Implementation {
	case core.LND:
		if !cache.IsLndServiceActive(nodeId) {
			return false, ServiceInactiveError
		}
		response = lnd.DisconnectPeer(request)
	case core.CLN:
		if !cache.IsClnServiceActive(nodeId) {
			return false, ServiceInactiveError
		}
		response = cln.DisconnectPeer(ctx, request)
	}
	if response.Error != "" {
		return false, errors.New(response.Error)
	}
	return response.RequestFailedCurrentlyDisconnected, nil
}

func GetWalletBalance(ctx context.Context, nodeId int) (lightning_helpers.WalletBalanceResponse, error) {
	request := lightning_helpers.WalletBalanceRequest{
		CommunicationRequest: lightning_helpers.CommunicationRequest{
			NodeId: nodeId,
		},
	}

	response := lightning_helpers.WalletBalanceResponse{
		Request: request,
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
	}

	nodeConnectionDetails := cache.GetNodeConnectionDetails(nodeId)
	switch nodeConnectionDetails.Implementation {
	case core.LND:
		if !cache.IsLndServiceActive(nodeId) {
			return lightning_helpers.WalletBalanceResponse{}, ServiceInactiveError
		}
		response = lnd.WalletBalance(request)
	case core.CLN:
		if !cache.IsClnServiceActive(nodeId) {
			return lightning_helpers.WalletBalanceResponse{}, ServiceInactiveError
		}
		response = cln.WalletBalance(ctx, request)
	}
	if response.Error != "" {
		return lightning_helpers.WalletBalanceResponse{}, errors.New(response.Error)
	}
	return response, nil
}

func ListPeers(ctx context.Context, nodeId int, latestError bool) (map[string]lightning_helpers.Peer, error) {
	request := lightning_helpers.ListPeersRequest{
		CommunicationRequest: lightning_helpers.CommunicationRequest{
			NodeId: nodeId,
		},
		LatestError: latestError,
	}

	response := lightning_helpers.ListPeersResponse{
		Request: request,
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
	}

	nodeConnectionDetails := cache.GetNodeConnectionDetails(nodeId)
	switch nodeConnectionDetails.Implementation {
	case core.LND:
		if !cache.IsLndServiceActive(nodeId) {
			return nil, ServiceInactiveError
		}
		response = lnd.ListPeers(request)
	case core.CLN:
		if !cache.IsClnServiceActive(nodeId) {
			return nil, ServiceInactiveError
		}
		response = cln.ListPeers(ctx, request)
	}
	if response.Error != "" {
		return nil, errors.New(response.Error)
	}
	return response.Peers, nil
}

func NewAddress(ctx context.Context,
	request lightning_helpers.NewAddressRequest) (string, error) {
	response := lightning_helpers.NewAddressResponse{
		Request: request,
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
	}

	nodeConnectionDetails := cache.GetNodeConnectionDetails(request.NodeId)
	switch nodeConnectionDetails.Implementation {
	case core.LND:
		if !cache.IsLndServiceActive(request.NodeId) {
			return "", ServiceInactiveError
		}
		response = lnd.NewAddress(request)
	case core.CLN:
		if !cache.IsClnServiceActive(request.NodeId) {
			return "", ServiceInactiveError
		}
		response = cln.NewAddress(ctx, request)
	}
	if response.Error != "" {
		return "", errors.New(response.Error)
	}
	return response.Address, nil
}

func OpenChannel(ctx context.Context,
	request lightning_helpers.OpenChannelRequest) (lightning_helpers.OpenChannelResponse, error) {
	response := lightning_helpers.OpenChannelResponse{
		Request: request,
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
	}

	nodeConnectionDetails := cache.GetNodeConnectionDetails(request.NodeId)
	switch nodeConnectionDetails.Implementation {
	case core.LND:
		if !cache.IsLndServiceActive(request.NodeId) {
			return lightning_helpers.OpenChannelResponse{}, ServiceInactiveError
		}
		response = lnd.OpenChannel(request)
	case core.CLN:
		if !cache.IsClnServiceActive(request.NodeId) {
			return lightning_helpers.OpenChannelResponse{}, ServiceInactiveError
		}
		response = cln.OpenChannel(ctx, request)
	}
	if response.Error != "" {
		return lightning_helpers.OpenChannelResponse{}, errors.New(response.Error)
	}
	return response, nil
}

func BatchOpenChannel(ctx context.Context,
	request lightning_helpers.BatchOpenChannelRequest) (lightning_helpers.BatchOpenChannelResponse, error) {

	response := lightning_helpers.BatchOpenChannelResponse{
		Request: request,
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
	}

	nodeConnectionDetails := cache.GetNodeConnectionDetails(request.NodeId)
	switch nodeConnectionDetails.Implementation {
	case core.LND:
		if !cache.IsLndServiceActive(request.NodeId) {
			return lightning_helpers.BatchOpenChannelResponse{}, ServiceInactiveError
		}
		response = lnd.BatchOpenChannel(request)
	case core.CLN:
		if !cache.IsClnServiceActive(request.NodeId) {
			return lightning_helpers.BatchOpenChannelResponse{}, ServiceInactiveError
		}
		return lightning_helpers.BatchOpenChannelResponse{}, UnsupportedOperationError
	}
	if response.Error != "" {
		return lightning_helpers.BatchOpenChannelResponse{}, errors.New(response.Error)
	}
	return response, nil
}

func CloseChannel(ctx context.Context,
	request lightning_helpers.CloseChannelRequest) (lightning_helpers.CloseChannelResponse, error) {
	response := lightning_helpers.CloseChannelResponse{
		Request: request,
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
	}

	nodeConnectionDetails := cache.GetNodeConnectionDetails(request.NodeId)
	switch nodeConnectionDetails.Implementation {
	case core.LND:
		if !cache.IsLndServiceActive(request.NodeId) {
			return lightning_helpers.CloseChannelResponse{}, ServiceInactiveError
		}
		response = lnd.CloseChannel(request)
	case core.CLN:
		if !cache.IsClnServiceActive(request.NodeId) {
			return lightning_helpers.CloseChannelResponse{}, ServiceInactiveError
		}
		response = cln.CloseChannel(ctx, request)
	}
	if response.Error != "" {
		return lightning_helpers.CloseChannelResponse{}, errors.New(response.Error)
	}
	return response, nil
}

func NewInvoice(ctx context.Context,
	request lightning_helpers.NewInvoiceRequest) (lightning_helpers.NewInvoiceResponse, error) {
	response := lightning_helpers.NewInvoiceResponse{
		Request: request,
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
	}

	nodeConnectionDetails := cache.GetNodeConnectionDetails(request.NodeId)
	switch nodeConnectionDetails.Implementation {
	case core.LND:
		if !cache.IsLndServiceActive(request.NodeId) {
			return lightning_helpers.NewInvoiceResponse{}, ServiceInactiveError
		}
		response = lnd.NewInvoice(request)
	case core.CLN:
		if !cache.IsClnServiceActive(request.NodeId) {
			return lightning_helpers.NewInvoiceResponse{}, ServiceInactiveError
		}
		response = cln.NewInvoice(ctx, request)
	}
	if response.Error != "" {
		return lightning_helpers.NewInvoiceResponse{}, errors.New(response.Error)
	}
	return response, nil
}

func OnChainPayment(ctx context.Context,
	request lightning_helpers.OnChainPaymentRequest) (string, error) {
	response := lightning_helpers.OnChainPaymentResponse{
		Request: request,
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
	}

	nodeConnectionDetails := cache.GetNodeConnectionDetails(request.NodeId)
	switch nodeConnectionDetails.Implementation {
	case core.LND:
		if !cache.IsLndServiceActive(request.NodeId) {
			return "", ServiceInactiveError
		}
		response = lnd.OnChainPayment(request)
	case core.CLN:
		if !cache.IsClnServiceActive(request.NodeId) {
			return "", ServiceInactiveError
		}
		response = cln.OnChainPayment(ctx, request)
	}
	if response.Error != "" {
		return "", errors.New(response.Error)
	}
	return response.TxId, nil
}

func NewPayment(ctx context.Context,
	request lightning_helpers.NewPaymentRequest) (lightning_helpers.NewPaymentResponse, error) {
	response := lightning_helpers.NewPaymentResponse{
		Request: request,
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
	}

	nodeConnectionDetails := cache.GetNodeConnectionDetails(request.NodeId)
	switch nodeConnectionDetails.Implementation {
	case core.LND:
		if !cache.IsLndServiceActive(request.NodeId) {
			return lightning_helpers.NewPaymentResponse{}, ServiceInactiveError
		}
		response = lnd.NewPayment(request)
	case core.CLN:
		if !cache.IsClnServiceActive(request.NodeId) {
			return lightning_helpers.NewPaymentResponse{}, ServiceInactiveError
		}
		response = cln.NewPayment(ctx, request)
	}
	if response.Error != "" {
		return lightning_helpers.NewPaymentResponse{}, errors.New(response.Error)
	}
	return response, nil
}

func DecodeInvoice(ctx context.Context,
	request lightning_helpers.DecodeInvoiceRequest) (lightning_helpers.DecodeInvoiceResponse, error) {
	response := lightning_helpers.DecodeInvoiceResponse{
		Request: request,
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
	}

	nodeConnectionDetails := cache.GetNodeConnectionDetails(request.NodeId)
	switch nodeConnectionDetails.Implementation {
	case core.LND:
		if !cache.IsLndServiceActive(request.NodeId) {
			return lightning_helpers.DecodeInvoiceResponse{}, ServiceInactiveError
		}
		response = lnd.DecodeInvoice(request)
	case core.CLN:
		if !cache.IsClnServiceActive(request.NodeId) {
			return lightning_helpers.DecodeInvoiceResponse{}, ServiceInactiveError
		}
		response = cln.DecodeInvoice(ctx, request)
	}
	if response.Error != "" {
		return lightning_helpers.DecodeInvoiceResponse{}, errors.New(response.Error)
	}
	return response, nil
}

func ChannelStatusUpdate(ctx context.Context, request lightning_helpers.ChannelStatusUpdateRequest) error {
	response := lightning_helpers.ChannelStatusUpdateResponse{
		Request: request,
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
	}

	nodeConnectionDetails := cache.GetNodeConnectionDetails(request.NodeId)
	switch nodeConnectionDetails.Implementation {
	case core.LND:
		if !cache.IsLndServiceActive(request.NodeId) {
			return ServiceInactiveError
		}
		response = lnd.ChannelStatusUpdate(request)
	case core.CLN:
		if !cache.IsClnServiceActive(request.NodeId) {
			return ServiceInactiveError
		}
		return UnsupportedOperationError
	}
	if response.Error != "" {
		return errors.New(response.Error)
	}
	return nil
}

func MoveFundsOffChain(ctx context.Context,
	request lightning_helpers.MoveFundsOffChainRequest) (lightning_helpers.MoveFundsOffChainResponse, error) {
	response := lightning_helpers.MoveFundsOffChainResponse{
		Request: request,
		CommunicationResponse: lightning_helpers.CommunicationResponse{
			Status: lightning_helpers.Inactive,
		},
	}

	nodeConnectionDetails := cache.GetNodeConnectionDetails(request.NodeId)
	switch nodeConnectionDetails.Implementation {
	case core.LND:
		if !cache.IsLndServiceActive(request.NodeId) {
			return lightning_helpers.MoveFundsOffChainResponse{}, ServiceInactiveError
		}
		response = lnd.MoveFundsOffChain(request)
	case core.CLN:
		if !cache.IsClnServiceActive(request.NodeId) {
			return lightning_helpers.MoveFundsOffChainResponse{}, ServiceInactiveError
		}
		response = cln.MoveFundsOffChain(ctx, request)
	}
	if response.Error != "" {
		return lightning_helpers.MoveFundsOffChainResponse{}, errors.New(response.Error)
	}
	return response, nil
}
