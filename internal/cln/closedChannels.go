package cln

import (
	"context"
	"encoding/hex"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/channels"
	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/internal/nodes"
	"github.com/lncapital/torq/internal/services_helpers"
	"github.com/lncapital/torq/proto/cln"
)

type client_ListClosedChannels interface {
	ListClosedChannels(ctx context.Context,
		in *cln.ListclosedchannelsRequest,
		opts ...grpc.CallOption) (*cln.ListclosedchannelsResponse, error)
}

func SubscribeAndStoreClosedChannels(ctx context.Context,
	client client_ListClosedChannels,
	db *sqlx.DB,
	nodeSettings cache.NodeSettingsCache) {

	serviceType := services_helpers.ClnServiceClosedChannelsService

	cache.SetInitializingNodeServiceState(serviceType, nodeSettings.NodeId)

	ticker := time.NewTicker(streamClosedChannelsTickerSeconds * time.Second)
	defer ticker.Stop()
	tickerChannel := ticker.C

	err := listAndProcessClosedChannels(ctx, db, client, serviceType, nodeSettings, true)
	if err != nil {
		processError(ctx, serviceType, nodeSettings, err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			cache.SetInactiveNodeServiceState(serviceType, nodeSettings.NodeId)
			return
		case <-tickerChannel:
			err = listAndProcessClosedChannels(ctx, db, client, serviceType, nodeSettings, false)
			if err != nil {
				processError(ctx, serviceType, nodeSettings, err)
				return
			}
		}
	}
}

func listAndProcessClosedChannels(ctx context.Context,
	db *sqlx.DB,
	client client_ListClosedChannels,
	serviceType services_helpers.ServiceType,
	nodeSettings cache.NodeSettingsCache,
	bootStrapping bool) error {

	ctx, span := otel.Tracer(name).Start(ctx, "listAndProcessClosedChannels")
	defer span.End()

	clnChannels, err := client.ListClosedChannels(ctx, &cln.ListclosedchannelsRequest{})
	if err != nil {
		return errors.Wrapf(err, "listing source channels for nodeId: %v", nodeSettings.NodeId)
	}

	err = storeClosedChannels(ctx, db, clnChannels.Closedchannels, nodeSettings)
	if err != nil {
		return errors.Wrapf(err, "storing source channels for nodeId: %v", nodeSettings.NodeId)
	}

	if bootStrapping {
		log.Info().Msgf("Initial import of closed channels is done for nodeId: %v", nodeSettings.NodeId)
		cache.SetActiveNodeServiceState(serviceType, nodeSettings.NodeId)
	}
	return nil
}

func storeClosedChannels(ctx context.Context,
	db *sqlx.DB,
	clnChannels []*cln.ListclosedchannelsClosedchannels,
	nodeSettings cache.NodeSettingsCache) error {

	_, span := otel.Tracer(name).Start(ctx, "storeClosedChannels")
	defer span.End()

	for _, clnChannel := range clnChannels {
		if clnChannel != nil {
			peerPublicKey := hex.EncodeToString(clnChannel.PeerId)
			peerNodeId := cache.GetPeerNodeIdByPublicKey(peerPublicKey, nodeSettings.Chain, nodeSettings.Network)
			if peerNodeId == 0 {
				var err error
				peerNodeId, err = nodes.AddNodeWhenNew(db, nodes.Node{
					PublicKey: peerPublicKey,
					Chain:     nodeSettings.Chain,
					Network:   nodeSettings.Network,
				}, nil)
				if err != nil {
					return errors.Wrapf(err, "add new peer node for nodeId: %v", nodeSettings.NodeId)
				}
			}
			err := processClosedChannel(db, clnChannel, nodeSettings, peerNodeId)
			if err != nil {
				return errors.Wrapf(err, "process closed channel for nodeId: %v", nodeSettings.NodeId)
			}
		}
	}
	return nil
}

func processClosedChannel(db *sqlx.DB,
	clnChannel *cln.ListclosedchannelsClosedchannels,
	nodeSettings cache.NodeSettingsCache,
	peerNodeId int) error {

	foi := int(clnChannel.FundingOutnum)
	fundingOutputIndex := &foi
	var fundingTransactionHash *string
	if len(clnChannel.FundingTxid) != 0 {
		fti := hex.EncodeToString(clnChannel.FundingTxid)
		fundingTransactionHash = &fti
	}
	var channelId int
	if clnChannel.ShortChannelId != nil {
		channelId = cache.GetChannelIdByShortChannelId(clnChannel.ShortChannelId)
	}
	if channelId == 0 && len(clnChannel.FundingTxid) != 0 {
		channelId = cache.GetChannelIdByFundingTransaction(fundingTransactionHash, fundingOutputIndex)
	}
	var channel channels.Channel
	if channelId == 0 {
		channel = channels.Channel{
			FundingTransactionHash: fundingTransactionHash,
			FundingOutputIndex:     fundingOutputIndex,
			FirstNodeId:            nodeSettings.NodeId,
			SecondNodeId:           peerNodeId,
		}
	} else {
		channelSettings := cache.GetChannelSettingByChannelId(channelId)
		channel = channels.Channel{
			ChannelID:              channelSettings.ChannelId,
			ShortChannelID:         channelSettings.ShortChannelId,
			FundingTransactionHash: fundingTransactionHash,
			FundingOutputIndex:     fundingOutputIndex,
			ClosingTransactionHash: channelSettings.ClosingTransactionHash,
			Capacity:               channelSettings.Capacity,
			Private:                channelSettings.Private,
			FirstNodeId:            channelSettings.FirstNodeId,
			SecondNodeId:           channelSettings.SecondNodeId,
			InitiatingNodeId:       channelSettings.InitiatingNodeId,
			AcceptingNodeId:        channelSettings.AcceptingNodeId,
			ClosingNodeId:          channelSettings.ClosingNodeId,
			Status:                 channelSettings.Status,
			FundingBlockHeight:     channelSettings.FundingBlockHeight,
			FundedOn:               channelSettings.FundedOn,
			ClosingBlockHeight:     channelSettings.ClosingBlockHeight,
			ClosedOn:               channelSettings.ClosedOn,
			Flags:                  channelSettings.Flags,
		}
	}
	channel.Private = clnChannel.Private
	switch clnChannel.Opener {
	case cln.ChannelSide_REMOTE:
		channel.InitiatingNodeId = &peerNodeId
		channel.AcceptingNodeId = &nodeSettings.NodeId
	case cln.ChannelSide_LOCAL:
		channel.InitiatingNodeId = &nodeSettings.NodeId
		channel.AcceptingNodeId = &peerNodeId
	}
	if clnChannel.ShortChannelId != nil {
		shortChannelId := *clnChannel.ShortChannelId
		channel.ShortChannelID = &shortChannelId
	}
	if clnChannel.TotalMsat != nil {
		channel.Capacity = int64(clnChannel.TotalMsat.Msat / 1_000)
	}
	if clnChannel.Closer != nil {
		switch *clnChannel.Closer {
		case cln.ChannelSide_REMOTE:
			channel.ClosingNodeId = &peerNodeId
		case cln.ChannelSide_LOCAL:
			channel.ClosingNodeId = &nodeSettings.NodeId
		}
	}
	switch clnChannel.CloseCause {
	case cln.ListclosedchannelsClosedchannels_USER:
		channel.Status = core.CooperativeClosed
	case cln.ListclosedchannelsClosedchannels_REMOTE:
		channel.Status = core.RemoteForceClosed
	case cln.ListclosedchannelsClosedchannels_LOCAL:
		channel.Status = core.LocalForceClosed
	case cln.ListclosedchannelsClosedchannels_PROTOCOL:
		// TODO FIXME CLN: This is just a guess there is no information to be found about breach
		channel.Status = core.BreachClosed
	case cln.ListclosedchannelsClosedchannels_ONCHAIN:
		// TODO FIXME CLN: This is just a guess there is no information to be found about abandoned
		channel.Status = core.AbandonedClosed
	case cln.ListclosedchannelsClosedchannels_UNKNOWN:
		// TODO FIXME CLN: This is just a guess there is no information to be found about funding cancelled
		channel.Status = core.FundingCancelledClosed
	}
	_, err := channels.AddChannelOrUpdateChannelStatus(db, nodeSettings, channel)
	if err != nil {
		return errors.Wrapf(err, "update channel data for channelId: %v, nodeId: %v",
			channelId, nodeSettings.NodeId)
	}

	// This section is to verify if there are any channels left and if not remove it as active peer.
	peerPublicKey := cache.GetNodeSettingsByNodeId(peerNodeId).PublicKey
	if cache.GetChannelPeerNodeIdByPublicKey(peerPublicKey, nodeSettings.Chain, nodeSettings.Network) != 0 {
		chans, err := channels.GetOpenChannelsForNodeId(db, nodeSettings.NodeId)
		if err != nil {
			return errors.Wrapf(err, "verify if there are any channels left for peerNodeId: %v, nodeId: %v",
				peerNodeId, nodeSettings.NodeId)
		}
		if len(chans) == 0 {
			cache.SetInactiveChannelPeerNode(peerNodeId, peerPublicKey, nodeSettings.Chain, nodeSettings.Network)
		}
	}
	return nil
}
