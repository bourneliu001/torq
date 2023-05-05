package cln

import (
	"context"
	"database/sql"
	"encoding/hex"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/channels"
	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/internal/graph_events"
	"github.com/lncapital/torq/internal/nodes"
	"github.com/lncapital/torq/internal/services_helpers"
	"github.com/lncapital/torq/proto/cln"
)

type client_ListChannels interface {
	ListPeerChannels(ctx context.Context,
		in *cln.ListpeerchannelsRequest,
		opts ...grpc.CallOption) (*cln.ListpeerchannelsResponse, error)
	ListChannels(ctx context.Context,
		in *cln.ListchannelsRequest,
		opts ...grpc.CallOption) (*cln.ListchannelsResponse, error)
}

func SubscribeAndStoreChannels(ctx context.Context,
	client client_ListChannels,
	db *sqlx.DB,
	nodeSettings cache.NodeSettingsCache) {

	serviceType := services_helpers.ClnServiceChannelsService

	cache.SetInitializingNodeServiceState(serviceType, nodeSettings.NodeId)

	ticker := time.NewTicker(streamChannelsTickerSeconds * time.Second)
	defer ticker.Stop()
	tickerChannel := ticker.C

	err := listAndProcessChannels(ctx, db, client, serviceType, nodeSettings, true)
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
			err = listAndProcessChannels(ctx, db, client, serviceType, nodeSettings, false)
			if err != nil {
				processError(ctx, serviceType, nodeSettings, err)
				return
			}
		}
	}
}

func listAndProcessChannels(ctx context.Context, db *sqlx.DB, client client_ListChannels,
	serviceType services_helpers.ServiceType,
	nodeSettings cache.NodeSettingsCache,
	bootStrapping bool) error {

	clnPeerChannels, err := client.ListPeerChannels(ctx, &cln.ListpeerchannelsRequest{})
	if err != nil {
		return errors.Wrapf(err, "listing peer channels for nodeId: %v", nodeSettings.NodeId)
	}

	err = storePeerChannels(db, clnPeerChannels.Channels, nodeSettings)
	if err != nil {
		return errors.Wrapf(err, "storing source channels for nodeId: %v", nodeSettings.NodeId)
	}

	publicKey, err := hex.DecodeString(nodeSettings.PublicKey)
	if err != nil {
		return errors.Wrapf(err, "decoding public key for nodeId: %v", nodeSettings.NodeId)
	}
	clnChannels, err := client.ListChannels(ctx, &cln.ListchannelsRequest{
		Source: publicKey,
	})
	if err != nil {
		return errors.Wrapf(err, "listing source channels for nodeId: %v", nodeSettings.NodeId)
	}

	err = storeChannels(db, clnChannels.Channels, nodeSettings)
	if err != nil {
		return errors.Wrapf(err, "storing source channels for nodeId: %v", nodeSettings.NodeId)
	}

	clnChannels, err = client.ListChannels(ctx, &cln.ListchannelsRequest{
		Destination: publicKey,
	})
	if err != nil {
		return errors.Wrapf(err, "listing destination channels for nodeId: %v", nodeSettings.NodeId)
	}

	err = storeChannels(db, clnChannels.Channels, nodeSettings)
	if err != nil {
		return errors.Wrapf(err, "storing destination channels for nodeId: %v", nodeSettings.NodeId)
	}

	if bootStrapping {
		log.Info().Msgf("Initial import of peers is done for nodeId: %v", nodeSettings.NodeId)
		cache.SetActiveNodeServiceState(serviceType, nodeSettings.NodeId)
	}
	return nil
}

func storePeerChannels(db *sqlx.DB,
	clnPeerChannels []*cln.ListpeerchannelsChannels,
	nodeSettings cache.NodeSettingsCache) error {

	for _, clnPeerChannel := range clnPeerChannels {
		if clnPeerChannel != nil {
			peerPublicKey := hex.EncodeToString(clnPeerChannel.PeerId)
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
			_, err := processPeerChannel(db, clnPeerChannel, nodeSettings, peerNodeId, peerPublicKey)
			if err != nil {
				return errors.Wrapf(err, "process channel for nodeId: %v", nodeSettings.NodeId)
			}
		}
	}
	return nil
}

func storeChannels(db *sqlx.DB,
	clnChannels []*cln.ListchannelsChannels,
	nodeSettings cache.NodeSettingsCache) error {

	for _, clnChannel := range clnChannels {
		if clnChannel != nil {
			sourcePublicKey := hex.EncodeToString(clnChannel.Source)
			destinationPublicKey := hex.EncodeToString(clnChannel.Destination)
			peerNodeId := cache.GetPeerNodeIdByPublicKey(sourcePublicKey, nodeSettings.Chain, nodeSettings.Network)
			if peerNodeId == nodeSettings.NodeId {
				peerNodeId = cache.GetPeerNodeIdByPublicKey(destinationPublicKey, nodeSettings.Chain, nodeSettings.Network)
			}
			if peerNodeId == 0 {
				log.Info().Msgf("Skipping channel: Peer node is unknown for nodeId: %v", nodeSettings.NodeId)
				continue
			}
			channelId := cache.GetChannelIdByShortChannelId(&clnChannel.ShortChannelId)
			if channelId == 0 {
				log.Info().Msgf("Skipping channel: Channel is unknown for nodeId: %v", nodeSettings.NodeId)
				continue
			}

			announcingNodeId := cache.GetPeerNodeIdByPublicKey(
				sourcePublicKey, nodeSettings.Chain, nodeSettings.Network)
			connectingNodeId := cache.GetPeerNodeIdByPublicKey(
				destinationPublicKey, nodeSettings.Chain, nodeSettings.Network)

			channelEvent := graph_events.ChannelEventFromGraph{}
			channelEvent.ChannelId = channelId
			channelEvent.NodeId = nodeSettings.NodeId
			channelEvent.AnnouncingNodeId = announcingNodeId
			channelEvent.ConnectingNodeId = connectingNodeId
			channelEvent.Outbound = announcingNodeId == nodeSettings.NodeId
			channelEvent.FeeRateMilliMsat = int64(clnChannel.FeePerMillionth)
			channelEvent.FeeBaseMsat = int64(clnChannel.BaseFeeMillisatoshi)
			channelEvent.Disabled = !clnChannel.Active
			minHtlcMsat := clnChannel.HtlcMinimumMsat
			if minHtlcMsat != nil {
				channelEvent.MinHtlcMsat = (*minHtlcMsat).Msat
			}
			maxHtlcMsat := clnChannel.HtlcMaximumMsat
			if maxHtlcMsat != nil {
				channelEvent.MaxHtlcMsat = (*maxHtlcMsat).Msat
			}
			channelEvent.TimeLockDelta = clnChannel.Delay
			err := insertRoutingPolicy(db, channelEvent, nodeSettings)
			if err != nil {
				return errors.Wrapf(err, "process routing policy for nodeId: %v", nodeSettings.NodeId)
			}
		}
	}
	return nil
}

func processPeerChannel(db *sqlx.DB,
	clnPeerChannel *cln.ListpeerchannelsChannels,
	nodeSettings cache.NodeSettingsCache,
	peerNodeId int,
	peerPublicKey string) (int, error) {

	var fundingOutputIndex *int
	if clnPeerChannel.FundingOutnum != nil {
		foi := int(*clnPeerChannel.FundingOutnum)
		fundingOutputIndex = &foi
	}
	var fundingTransactionHash *string
	if len(clnPeerChannel.FundingTxid) != 0 {
		fti := hex.EncodeToString(clnPeerChannel.FundingTxid)
		fundingTransactionHash = &fti
	}
	var channelId int
	if clnPeerChannel.ShortChannelId != nil {
		channelId = cache.GetChannelIdByShortChannelId(clnPeerChannel.ShortChannelId)
	}
	if channelId == 0 && len(clnPeerChannel.FundingTxid) != 0 {
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
			FundingTransactionHash: fundingTransactionHash,
			FundingOutputIndex:     fundingOutputIndex,
			FirstNodeId:            channelSettings.FirstNodeId,
			SecondNodeId:           channelSettings.SecondNodeId,

			ChannelID:              channelSettings.ChannelId,
			ShortChannelID:         channelSettings.ShortChannelId,
			FundingBlockHeight:     channelSettings.FundingBlockHeight,
			FundedOn:               channelSettings.FundedOn,
			Capacity:               channelSettings.Capacity,
			InitiatingNodeId:       channelSettings.InitiatingNodeId,
			AcceptingNodeId:        channelSettings.AcceptingNodeId,
			Private:                channelSettings.Private,
			Status:                 channelSettings.Status,
			ClosingTransactionHash: channelSettings.ClosingTransactionHash,
			ClosingNodeId:          channelSettings.ClosingNodeId,
			ClosingBlockHeight:     channelSettings.ClosingBlockHeight,
			ClosedOn:               channelSettings.ClosedOn,
			Flags:                  channelSettings.Flags,
		}
	}
	if clnPeerChannel.ShortChannelId != nil &&
		(channel.ShortChannelID == nil || *channel.ShortChannelID != clnPeerChannel.GetShortChannelId()) {

		shortChannelId := *clnPeerChannel.ShortChannelId
		channel.ShortChannelID = &shortChannelId
	}
	if clnPeerChannel.TotalMsat != nil {
		channel.Capacity = int64(clnPeerChannel.TotalMsat.Msat / 1_000)
	}
	if clnPeerChannel.Private != nil {
		channel.Private = *clnPeerChannel.Private
	}
	if clnPeerChannel.Closer != nil {
		switch *clnPeerChannel.Closer {
		case cln.ChannelSide_REMOTE:
			channel.ClosingNodeId = &peerNodeId
		case cln.ChannelSide_LOCAL:
			channel.ClosingNodeId = &nodeSettings.NodeId
		}
	}
	if clnPeerChannel.Opener != nil {
		switch *clnPeerChannel.Opener {
		case cln.ChannelSide_REMOTE:
			channel.InitiatingNodeId = &peerNodeId
			channel.AcceptingNodeId = &nodeSettings.NodeId
		case cln.ChannelSide_LOCAL:
			channel.InitiatingNodeId = &nodeSettings.NodeId
			channel.AcceptingNodeId = &peerNodeId
		}
	}
	if clnPeerChannel.CloseTo != nil &&
		(channel.ClosingTransactionHash == nil || *channel.ClosingTransactionHash != hex.EncodeToString(clnPeerChannel.CloseTo)) {

		closeTo := hex.EncodeToString(clnPeerChannel.CloseTo)
		channel.ClosingTransactionHash = &closeTo
	}
	channelStatus := core.Opening
	if clnPeerChannel.State != nil {
		switch *clnPeerChannel.State {
		case cln.ListpeerchannelsChannels_OPENINGD,
			cln.ListpeerchannelsChannels_DUALOPEND_OPEN_INIT,
			cln.ListpeerchannelsChannels_CHANNELD_AWAITING_LOCKIN,
			cln.ListpeerchannelsChannels_DUALOPEND_AWAITING_LOCKIN:
			channelStatus = core.Opening
		case cln.ListpeerchannelsChannels_CHANNELD_NORMAL:
			channelStatus = core.Open
		case cln.ListpeerchannelsChannels_CHANNELD_SHUTTING_DOWN,
			cln.ListpeerchannelsChannels_CLOSINGD_SIGEXCHANGE,
			cln.ListpeerchannelsChannels_CLOSINGD_COMPLETE:
			channelStatus = core.CooperativeClosed
		case cln.ListpeerchannelsChannels_AWAITING_UNILATERAL:
			channelStatus = core.LocalForceClosed
		case cln.ListpeerchannelsChannels_FUNDING_SPEND_SEEN,
			cln.ListpeerchannelsChannels_ONCHAIN:
			// TODO FIXME How identify BreachClosed?
			channelStatus = core.RemoteForceClosed
		}
	}
	channel.Status = channelStatus

	// TODO FIXME CLN returns a lot more data then LND. We should probably expand our channel table!

	_, err := channels.AddChannelOrUpdateChannelStatus(db, nodeSettings, channel)
	if err != nil {
		return 0, errors.Wrapf(err, "update channel data for channelId: %v, nodeId: %v",
			channelId, nodeSettings.NodeId)
	}
	cache.SetChannelPeerNode(peerNodeId, peerPublicKey, nodeSettings.Chain, nodeSettings.Network, channelStatus)
	channelState := cache.GetChannelState(nodeSettings.NodeId, channelId, true)
	if channelState == nil {
		log.Info().Msgf("Peer channel received with unknown channel state for channelId: %v, nodeId: %v",
			channelId, nodeSettings.NodeId)
		return channelId, nil
	}
	//if clnPeerChannel.ChannelType != nil {
	//	for _, ctn := range (*clnPeerChannel.ChannelType).Names {
	//		if channelState.CommitmentType == core.CommitmentTypeUnknown {
	//			channelState.CommitmentType = core.GetCommitmentTypeForCln(ctn)
	//		}
	//	}
	//}
	var htlcs []cache.Htlc
	pendingIncomingHtlcCount := 0
	pendingIncomingHtlcAmount := int64(0)
	pendingOutgoingHtlcCount := 0
	pendingOutgoingHtlcAmount := int64(0)
	for _, htlc := range clnPeerChannel.Htlcs {
		if htlc == nil {
			continue
		}
		if htlc.Direction == nil || htlc.AmountMsat == nil || htlc.Expiry == nil || htlc.Id == nil {
			continue
		}
		amount := int64((*htlc.AmountMsat).Msat / 1_000)
		switch *htlc.Direction {
		case cln.ListpeerchannelsChannelsHtlcs_IN:
			htlcs = append(htlcs, cache.Htlc{
				Incoming:         true,
				Amount:           amount,
				HashLock:         htlc.PaymentHash,
				ExpirationHeight: *htlc.Expiry,
				HtlcIndex:        *htlc.Id,
			})
			pendingIncomingHtlcCount++
			pendingIncomingHtlcAmount += amount
		case cln.ListpeerchannelsChannelsHtlcs_OUT:
			htlcs = append(htlcs, cache.Htlc{
				Incoming:            false,
				Amount:              amount,
				HashLock:            htlc.PaymentHash,
				ExpirationHeight:    *htlc.Expiry,
				ForwardingHtlcIndex: *htlc.Id,
			})
			pendingOutgoingHtlcCount++
			pendingOutgoingHtlcAmount += amount
		}
	}
	channelState.PendingHtlcs = htlcs
	channelState.PendingIncomingHtlcCount = pendingIncomingHtlcCount
	channelState.PendingIncomingHtlcAmount = pendingIncomingHtlcAmount
	channelState.PendingOutgoingHtlcCount = pendingOutgoingHtlcCount
	channelState.PendingOutgoingHtlcAmount = pendingOutgoingHtlcAmount

	cache.SetChannelState(nodeSettings.NodeId, *channelState)
	cache.SetChannelPeerNode(peerNodeId, peerPublicKey, nodeSettings.Chain, nodeSettings.Network, channel.Status)
	return channelId, nil
}

func insertRoutingPolicy(
	db *sqlx.DB,
	channelEvent graph_events.ChannelEventFromGraph,
	nodeSettings cache.NodeSettingsCache) error {

	existingChannelEvent := graph_events.ChannelEventFromGraph{}
	err := db.Get(&existingChannelEvent, `
				SELECT *
				FROM routing_policy
				WHERE channel_id=$1 AND announcing_node_id=$2 AND connecting_node_id=$3
				ORDER BY ts DESC
				LIMIT 1;`, channelEvent.ChannelId, channelEvent.AnnouncingNodeId, channelEvent.ConnectingNodeId)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return errors.Wrapf(err, "insertNodeEvent -> getPreviousChannelEvent.")
		}
	}

	// If one of our active torq nodes is announcing_node_id then the channel update was by our node
	// TODO FIXME ignore if previous update was from the same node so if announcing_node_id=node_id on previous record
	// and the current parameters are announcing_node_id!=node_id
	if existingChannelEvent.Disabled != channelEvent.Disabled ||
		existingChannelEvent.FeeBaseMsat != channelEvent.FeeBaseMsat ||
		existingChannelEvent.FeeRateMilliMsat != channelEvent.FeeRateMilliMsat ||
		existingChannelEvent.MaxHtlcMsat != channelEvent.MaxHtlcMsat ||
		existingChannelEvent.MinHtlcMsat != channelEvent.MinHtlcMsat ||
		existingChannelEvent.TimeLockDelta != channelEvent.TimeLockDelta {

		now := time.Now().UTC()
		_, err := db.Exec(`
		INSERT INTO routing_policy
			(ts,disabled,time_lock_delta,min_htlc,max_htlc_msat,fee_base_msat,fee_rate_mill_msat,
			 channel_id,announcing_node_id,connecting_node_id,node_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);`, now,
			channelEvent.Disabled, channelEvent.TimeLockDelta, channelEvent.MinHtlcMsat,
			channelEvent.MaxHtlcMsat, channelEvent.FeeBaseMsat, channelEvent.FeeRateMilliMsat,
			channelEvent.ChannelId, channelEvent.AnnouncingNodeId, channelEvent.ConnectingNodeId, nodeSettings.NodeId)
		if err != nil {
			return errors.Wrapf(err, "insertRoutingPolicy")
		}

		channelState := cache.GetChannelState(nodeSettings.NodeId, channelEvent.ChannelId, true)
		if channelState == nil {
			log.Info().Msgf("Peer channel received with unknown channel state for channelId: %v, nodeId: %v",
				channelEvent.ChannelId, nodeSettings.NodeId)
			return nil
		}
		if channelEvent.AnnouncingNodeId == nodeSettings.NodeId {
			channelState.LocalDisabled = channelEvent.Disabled
			channelState.LocalTimeLockDelta = channelEvent.TimeLockDelta
			channelState.LocalFeeBaseMsat = channelEvent.FeeBaseMsat
			channelState.LocalFeeRateMilliMsat = channelEvent.FeeRateMilliMsat
			channelState.LocalMinHtlcMsat = channelEvent.MinHtlcMsat
			channelState.LocalMaxHtlcMsat = channelEvent.MaxHtlcMsat
		} else {
			channelState.RemoteDisabled = channelEvent.Disabled
			channelState.RemoteTimeLockDelta = channelEvent.TimeLockDelta
			channelState.RemoteFeeBaseMsat = channelEvent.FeeBaseMsat
			channelState.RemoteFeeRateMilliMsat = channelEvent.FeeRateMilliMsat
			channelState.RemoteMinHtlcMsat = channelEvent.MinHtlcMsat
			channelState.RemoteMaxHtlcMsat = channelEvent.MaxHtlcMsat
		}
		cache.SetChannelState(nodeSettings.NodeId, *channelState)
	}
	return nil
}
