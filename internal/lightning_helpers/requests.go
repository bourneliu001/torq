package lightning_helpers

import (
	"encoding/hex"

	"github.com/jmoiron/sqlx"

	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/proto/cln"
	"github.com/lncapital/torq/proto/lnrpc"
)

type RebalanceOrigin int

const (
	RebalanceWorkflowNode = RebalanceOrigin(iota)
	RebalanceManual
)

type PeerSyncType int32

const (
	// PeerUnknownSync Denotes that we cannot determine the peer's current sync type.
	PeerUnknownSync PeerSyncType = 0
	// PeerActiveSync Denotes that we are actively receiving new graph updates from the peer.
	PeerActiveSync PeerSyncType = 1
	// PeerPassiveSync Denotes that we are not receiving new graph updates from the peer.
	PeerPassiveSync PeerSyncType = 2
	// PeerPinnedSync Denotes that this peer is pinned into an active sync.
	PeerPinnedSync PeerSyncType = 3
)

type AddressType int32

const (
	Unknown = AddressType(iota)
	P2WPKH  = 1
	P2WKH   = 2
	NP2WKH  = 3
	P2TR    = 4
)

type FeatureEntry struct {
	Key   uint32  `json:"key"`
	Value Feature `json:"value"`
}

type Feature struct {
	Name       string `json:"name"`
	IsRequired bool   `json:"isRequired"`
	IsKnown    bool   `json:"isKnown"`
}

type TimeStampedError struct {
	Timestamp uint64 `json:"timestamp"`
	Error     string `json:"error"`
}

type Peer struct {
	PubKey          string             `json:"pubKey"`
	Address         string             `json:"address"`
	BytesSent       uint64             `json:"bytesSent"`
	BytesRecv       uint64             `json:"bytesRecv"`
	SatSent         int64              `json:"satSent"`
	SatRecv         int64              `json:"satRecv"`
	Inbound         bool               `json:"inbound"`
	PingTime        int64              `json:"pingTime"`
	SyncType        PeerSyncType       `json:"syncType"`
	Features        []FeatureEntry     `json:"features"`
	Errors          []TimeStampedError `json:"errors"`
	FlapCount       int32              `json:"flapCount"`
	LastFlapNS      int64              `json:"lastFlapNs"`
	LastPingPayload []byte             `json:"lastPingPayload"`
}

func GetPeerLND(peer *lnrpc.Peer) Peer {
	p := Peer{
		PubKey:          peer.PubKey,
		Address:         peer.Address,
		BytesSent:       peer.BytesSent,
		BytesRecv:       peer.BytesRecv,
		SatSent:         peer.SatSent,
		SatRecv:         peer.SatRecv,
		Inbound:         peer.Inbound,
		PingTime:        peer.PingTime,
		SyncType:        PeerSyncType(peer.SyncType),
		FlapCount:       peer.FlapCount,
		LastFlapNS:      peer.LastFlapNs,
		LastPingPayload: peer.LastPingPayload,
	}

	features := make([]FeatureEntry, len(peer.Features))
	for key, feature := range peer.Features {
		features = append(features, FeatureEntry{
			Key: key,
			Value: Feature{
				Name:       feature.Name,
				IsRequired: feature.IsRequired,
				IsKnown:    feature.IsKnown,
			},
		})
	}
	p.Features = features

	timeStampedErrors := make([]TimeStampedError, len(peer.Errors))
	for _, tse := range peer.Errors {
		timeStampedErrors = append(timeStampedErrors, TimeStampedError{
			Timestamp: tse.Timestamp,
			Error:     tse.Error,
		})
	}
	p.Errors = timeStampedErrors
	return p
}

func GetPeerCLN(peer *cln.ListpeersPeers) Peer {
	p := Peer{
		PubKey: hex.EncodeToString(peer.Id),
	}
	if peer.RemoteAddr != nil {
		p.Address = *peer.RemoteAddr
	}
	if p.Address == "" && len(peer.Netaddr) != 0 {
		p.Address = peer.Netaddr[0]
	}
	return p
}

type CommunicationRequest struct {
	NodeId int `json:"nodeId"`
}

type RebalanceRequest struct {
	Origin RebalanceOrigin `json:"origin"`
	// Either manually generated number for manual rebalance or
	// WorkflowVersionNodeId for rebalance originating from workflows
	OriginId        int    `json:"originId"`
	OriginReference string `json:"originReference"`
	// Either IncomingChannelId is populated or OutgoingChannelId is.
	IncomingChannelId int `json:"incomingChannelId"`
	// Either OutgoingChannelId is populated or IncomingChannelId is.
	OutgoingChannelId int `json:"outgoingChannelIds"`
	// ONLY used for previous success rerun validation!
	ChannelIds            []int       `json:"channelIds"`
	AmountMsat            uint64      `json:"amountMsat"`
	MaximumCostMsat       uint64      `json:"maximumCostMsat"`
	MaximumConcurrency    int         `json:"maximumConcurrency"`
	WorkflowUnfocusedPath interface{} `json:"-"`
}

type RebalanceRequests struct {
	CommunicationRequest
	Requests []RebalanceRequest `json:"requests"`
}

type InformationRequest struct {
	CommunicationRequest
}

type SignMessageRequest struct {
	CommunicationRequest
	Message    string `json:"message"`
	SingleHash *bool  `json:"singleHash"`
}

type SignatureVerificationRequest struct {
	CommunicationRequest
	Message   string `json:"message"`
	Signature string `json:"signature"`
}

type RoutingPolicyUpdateRequest struct {
	CommunicationRequest
	Db               *sqlx.DB `json:"-"`
	RateLimitSeconds int      `json:"rateLimitSeconds"`
	RateLimitCount   int      `json:"rateLimitCount"`
	ChannelId        int      `json:"channelId"`
	FeeRateMilliMsat *int64   `json:"feeRateMilliMsat"`
	FeeBaseMsat      *int64   `json:"feeBaseMsat"`
	MaxHtlcMsat      *uint64  `json:"maxHtlcMsat"`
	MinHtlcMsat      *uint64  `json:"minHtlcMsat"`
	TimeLockDelta    *uint32  `json:"timeLockDelta"`
}

type ConnectPeerRequest struct {
	CommunicationRequest
	PublicKey string `json:"publicKey"`
	Host      string `json:"host"`
}

type DisconnectPeerRequest struct {
	CommunicationRequest
	PeerNodeId int `json:"peerNodeId"`
}

type WalletBalanceRequest struct {
	CommunicationRequest
}

type ListPeersRequest struct {
	CommunicationRequest
	LatestError bool `json:"latestError"`
}

type NewAddressRequest struct {
	CommunicationRequest
	Type AddressType `json:"type"`
	//The name of the account to generate a new address for. If empty, the default wallet account is used.
	Account string `json:"account"`
}

type OpenChannelRequest struct {
	CommunicationRequest
	SatPerVbyte        *uint64 `json:"satPerVbyte"`
	NodePubKey         string  `json:"nodePubKey"`
	Host               *string `json:"host"`
	LocalFundingAmount int64   `json:"localFundingAmount"`
	PushSat            *int64  `json:"pushSat"`
	TargetConf         *int32  `json:"targetConf"`
	Private            *bool   `json:"private"`
	MinHtlcMsat        *uint64 `json:"minHtlcMsat"`
	RemoteCsvDelay     *uint32 `json:"remoteCsvDelay"`
	MinConfs           *int32  `json:"minConfs"`
	SpendUnconfirmed   *bool   `json:"spendUnconfirmed"`
	CloseAddress       *string `json:"closeAddress"`
}

type BatchOpenChannel struct {
	NodePublicKey      string `json:"nodePublicKey"`
	LocalFundingAmount int64  `json:"localFundingAmount"`
	PushSat            *int64 `json:"pushSat"`
	Private            *bool  `json:"private"`
}

type BatchOpenChannelRequest struct {
	CommunicationRequest
	Channels    []BatchOpenChannel `json:"channels"`
	TargetConf  *int32             `json:"targetConf"`
	SatPerVbyte *int64             `json:"satPerVbyte"`
}

type CloseChannelRequest struct {
	CommunicationRequest
	Db              *sqlx.DB `json:"-"`
	ChannelId       int      `json:"channelId"`
	Force           *bool    `json:"force"`
	TargetConf      *int32   `json:"targetConf"`
	DeliveryAddress *string  `json:"deliveryAddress"`
	SatPerVbyte     *uint64  `json:"satPerVbyte"`
}

type NewInvoiceRequest struct {
	CommunicationRequest
	Memo            *string `json:"memo"`
	RPreImage       *string `json:"rPreImage"`
	ValueMsat       *int64  `json:"valueMsat"`
	Expiry          *int64  `json:"expiry"`
	FallBackAddress *string `json:"fallBackAddress"`
	Private         *bool   `json:"private"`
	IsAmp           *bool   `json:"isAmp"`
}

type ChannelStatusUpdateRequest struct {
	CommunicationRequest
	Db            *sqlx.DB `json:"-"`
	ChannelId     int
	ChannelStatus core.Status
}

type OnChainPaymentRequest struct {
	CommunicationRequest
	Address          string  `json:"address"`
	AmountSat        int64   `json:"amountSat"`
	TargetConf       *int32  `json:"targetConf"`
	SatPerVbyte      *uint64 `json:"satPerVbyte"`
	SendAll          *bool   `json:"sendAll"`
	Label            *string `json:"label"`
	MinConfs         *int32  `json:"minConfs"`
	SpendUnconfirmed *bool   `json:"spendUnconfirmed"`
}

type NewPaymentRequest struct {
	CommunicationRequest
	ProgressReportChannel chan<- interface{} `json:"-"`
	Invoice               *string            `json:"invoice"`
	TimeOutSecs           int32              `json:"timeoutSecs"`
	Dest                  *string            `json:"dest"`
	AmtMSat               *int64             `json:"amtMSat"`
	FeeLimitMsat          *int64             `json:"feeLimitMsat"`
	AllowSelfPayment      *bool              `json:"allowSelfPayment"`
}

type DecodeInvoiceRequest struct {
	CommunicationRequest
	Invoice string `json:"invoice"`
}

type MoveFundsOffChainRequest struct {
	CommunicationRequest
	ChannelId      int    `json:"channelId"`
	RHash          []byte `json:"rHash"`
	OutgoingNodeId int    `json:"outgoingNodeId"`
	IncomingNodeId int    `json:"incomingNodeId"`
	AmountMsat     int64  `json:"amountMsat"`
}

type MoveFundsOnChainRequest struct {
	CommunicationRequest
	OutgoingNodeId int   `json:"outgoingNodeId"`
	IncomingNodeId int   `json:"incomingNodeId"`
	AmountSat      int64 `json:"amountSat"`
}
