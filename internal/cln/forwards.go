package cln

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/internal/services_helpers"
	"github.com/lncapital/torq/proto/cln"
)

type forwardStatus int

const (
	offered = forwardStatus(iota)
	remoteFailed
	localFailed
)

type client_ListForwards interface {
	ListForwards(ctx context.Context,
		in *cln.ListforwardsRequest,
		opts ...grpc.CallOption) (*cln.ListforwardsResponse, error)
}

func SubscribeAndStoreForwards(ctx context.Context,
	client client_ListForwards,
	db *sqlx.DB,
	nodeSettings cache.NodeSettingsCache) {

	serviceType := services_helpers.ClnServiceForwardsService

	cache.SetInitializingNodeServiceState(serviceType, nodeSettings.NodeId)

	ticker := time.NewTicker(streamForwardsTickerSeconds * time.Second)
	defer ticker.Stop()
	tickerChannel := ticker.C

	err := listAndProcessForwards(ctx, db, client, serviceType, nodeSettings, true)
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
			err = listAndProcessForwards(ctx, db, client, serviceType, nodeSettings, false)
			if err != nil {
				processError(ctx, serviceType, nodeSettings, err)
				return
			}
		}
	}
}

func listAndProcessForwards(ctx context.Context, db *sqlx.DB, client client_ListForwards,
	serviceType services_helpers.ServiceType,
	nodeSettings cache.NodeSettingsCache,
	bootStrapping bool) error {

	ctx, span := otel.Tracer(name).Start(ctx, "listAndProcessForwards")
	defer span.End()

	var unprocessedShortChannelIds []string
	var closedChannelIds []int
	channels := cache.GetChannelSettingsByNodeId(nodeSettings.NodeId)
	for _, channel := range channels {
		// We don't want the forwards for closed channels when they have the flag core.ImportedForwards
		if channel.Flags.HasChannelFlag(core.ImportedForwards) {
			continue
		}
		if channel.ShortChannelId == nil {
			continue
		}
		unprocessedShortChannelIds = append(unprocessedShortChannelIds, *channel.ShortChannelId)
		if channel.Status >= core.CooperativeClosed {
			closedChannelIds = append(closedChannelIds, channel.ChannelId)
		}
	}

	ctx, span = otel.Tracer(name).Start(ctx, "listAndProcessForwards")
	span.SetAttributes(attribute.String("Status", cln.ListforwardsRequest_SETTLED.String()))
	err := processForwards(ctx, db, client, serviceType,
		cln.ListforwardsRequest_SETTLED, unprocessedShortChannelIds, nodeSettings, bootStrapping)
	if err != nil {
		return errors.Wrapf(err, "processing of forwards failed")
	}
	span.End()

	if cache.GetNodeConnectionDetails(nodeSettings.NodeId).CustomSettings.
		HasNodeConnectionDetailCustomSettings(core.ImportHtlcEvents) {

		ctx, span = otel.Tracer(name).Start(ctx, "listAndProcessForwards")
		span.SetAttributes(attribute.String("Status", cln.ListforwardsRequest_OFFERED.String()))
		err = processForwards(ctx, db, client, serviceType,
			cln.ListforwardsRequest_OFFERED, unprocessedShortChannelIds, nodeSettings, bootStrapping)
		if err != nil {
			return errors.Wrapf(err, "processing of forwards failed")
		}
		span.End()

		ctx, span = otel.Tracer(name).Start(ctx, "listAndProcessForwards")
		span.SetAttributes(attribute.String("Status", cln.ListforwardsRequest_LOCAL_FAILED.String()))
		err = processForwards(ctx, db, client, serviceType,
			cln.ListforwardsRequest_LOCAL_FAILED, unprocessedShortChannelIds, nodeSettings, bootStrapping)
		if err != nil {
			return errors.Wrapf(err, "processing of forwards failed")
		}
		span.End()

		ctx, span = otel.Tracer(name).Start(ctx, "listAndProcessForwards")
		span.SetAttributes(attribute.String("Status", cln.ListforwardsRequest_FAILED.String()))
		err = processForwards(ctx, db, client, serviceType,
			cln.ListforwardsRequest_FAILED, unprocessedShortChannelIds, nodeSettings, bootStrapping)
		if err != nil {
			return errors.Wrapf(err, "processing of forwards failed")
		}
		span.End()
	}
	for _, closedChannelId := range closedChannelIds {
		channelSetting := cache.GetChannelSettingByChannelId(closedChannelId)
		channelSetting.AddChannelFlags(core.ImportedForwards)
		cache.SetChannelFlags(closedChannelId, channelSetting.Flags)
		_, err = db.Exec(`UPDATE channel SET flags=$1, updated_on=$2 WHERE channel_id=$3`,
			channelSetting.Flags, time.Now().UTC(), channelSetting.ChannelId)
		if err != nil {
			return errors.Wrapf(err, "updating channel flag for channelId: %v, nodeId: %v",
				channelSetting.ChannelId, nodeSettings.NodeId)
		}
	}

	if bootStrapping {
		log.Info().Msgf("Initial import of transactions is done for nodeId: %v", nodeSettings.NodeId)
		cache.SetActiveNodeServiceState(serviceType, nodeSettings.NodeId)
	}
	return nil
}

func processForwards(ctx context.Context,
	db *sqlx.DB,
	client client_ListForwards,
	serviceType services_helpers.ServiceType,
	clnStatus cln.ListforwardsRequest_ListforwardsStatus,
	unprocessedShortChannelIds []string,
	nodeSettings cache.NodeSettingsCache,
	bootStrapping bool) error {

	for ix, shortChannelId := range unprocessedShortChannelIds {
		ctx, span := otel.Tracer(name).Start(ctx, "processForwards")
		span.SetAttributes(attribute.String("InChannel-ShortChannelId", shortChannelId))
		clnForwards, err := client.ListForwards(ctx, &cln.ListforwardsRequest{InChannel: &unprocessedShortChannelIds[ix], Status: &clnStatus})
		if err != nil {
			return errors.Wrapf(err, "listing %v forwards for nodeId: %v", clnStatus.String(), nodeSettings.NodeId)
		}
		err = storeForwards(ctx, db, clnStatus, clnForwards.Forwards, shortChannelId, "", nodeSettings)
		if err != nil {
			return errors.Wrapf(err, "storing %v forwards for nodeId: %v", clnStatus.String(), nodeSettings.NodeId)
		}
		span.End()

		ctx, span = otel.Tracer(name).Start(ctx, "processForwards")
		span.SetAttributes(attribute.String("OutChannel-ShortChannelId", shortChannelId))
		clnForwards, err = client.ListForwards(ctx, &cln.ListforwardsRequest{OutChannel: &unprocessedShortChannelIds[ix], Status: &clnStatus})
		if err != nil {
			return errors.Wrapf(err, "listing %v forwards for nodeId: %v", clnStatus.String(), nodeSettings.NodeId)
		}

		err = storeForwards(ctx, db, clnStatus, clnForwards.Forwards, "", shortChannelId, nodeSettings)
		if err != nil {
			return errors.Wrapf(err, "storing %v forwards for nodeId: %v", clnStatus.String(), nodeSettings.NodeId)
		}
		if bootStrapping {
			cache.SetInitializingNodeServiceState(serviceType, nodeSettings.NodeId)
		}
		span.End()
	}
	return nil
}

func storeForwards(ctx context.Context,
	db *sqlx.DB,
	clnStatus cln.ListforwardsRequest_ListforwardsStatus,
	clnForwards []*cln.ListforwardsForwards,
	incomingShortChannelId string,
	outgoingShortChannelId string,
	nodeSettings cache.NodeSettingsCache) error {

	_, span := otel.Tracer(name).Start(ctx, "storeForwards")
	defer span.End()

	var channelId int
	if incomingShortChannelId != "" {
		channelId = cache.GetChannelIdByShortChannelId(&incomingShortChannelId)
	} else {
		channelId = cache.GetChannelIdByShortChannelId(&outgoingShortChannelId)
	}

	var status forwardStatus
	var latestForward *time.Time
	var err error
	if clnStatus == cln.ListforwardsRequest_SETTLED {
		if incomingShortChannelId != "" {
			err = db.Get(&latestForward,
				`SELECT MAX(time) FROM forward WHERE node_id=$1 AND incoming_channel_id=$2;`,
				nodeSettings.NodeId, channelId)
		} else {
			err = db.Get(&latestForward,
				`SELECT MAX(time) FROM forward WHERE node_id=$1 AND outgoing_channel_id=$2;`,
				nodeSettings.NodeId, channelId)
		}
	} else {
		switch clnStatus {
		case cln.ListforwardsRequest_OFFERED:
			status = offered
		case cln.ListforwardsRequest_FAILED:
			status = remoteFailed
		case cln.ListforwardsRequest_LOCAL_FAILED:
			status = localFailed
		}
		if incomingShortChannelId != "" {
			err = db.Get(&latestForward,
				`SELECT MAX(time)
					   FROM htlc_event
					   WHERE node_id=$1 AND cln_forward_status_id=$2 AND incoming_channel_id=$3;`,
				nodeSettings.NodeId, status, channelId)
		} else {
			err = db.Get(&latestForward,
				`SELECT MAX(time)
					   FROM htlc_event
					   WHERE node_id=$1 AND cln_forward_status_id=$2 AND outgoing_channel_id=$3;`,
				nodeSettings.NodeId, status, channelId)
		}
	}
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return errors.Wrapf(err, "obtaining maximum forward time for transactions for nodeId: %v",
				nodeSettings.NodeId)
		}
	}

	if latestForward == nil {
		dummy := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		latestForward = &dummy
	}

	for _, clnForward := range clnForwards {
		if clnForward == nil {
			continue
		}
		eventTime := time.Unix(int64(clnForward.ReceivedTime), 0)
		if eventTime.Before(*latestForward) {
			continue
		}
		if clnStatus == cln.ListforwardsRequest_SETTLED {
			err = storeForward(db, clnForward, nodeSettings)
		} else {
			err = storeHtlc(db, status, clnForward, nodeSettings)
		}
		if err != nil {
			return errors.Wrapf(err, "persisting forward for nodeId: %v", nodeSettings.NodeId)
		}
	}
	return nil
}

func storeForward(db *sqlx.DB,
	clnForward *cln.ListforwardsForwards,
	nodeSettings cache.NodeSettingsCache) error {

	forwardTime := time.Unix(int64(clnForward.ReceivedTime), 0).UTC()
	var feeMsat uint64
	if clnForward.FeeMsat != nil {
		feeMsat = clnForward.FeeMsat.Msat
	}
	var inMsat uint64
	if clnForward.InMsat != nil {
		inMsat = clnForward.InMsat.Msat
	}
	var outMsat uint64
	if clnForward.OutMsat != nil {
		outMsat = clnForward.OutMsat.Msat
	}

	incomingShortChannelId := clnForward.InChannel
	incomingChannelId := cache.GetChannelIdByShortChannelId(&incomingShortChannelId)
	incomingChannelIdP := &incomingChannelId
	if incomingChannelId == 0 {
		log.Error().Msgf("Forward received for a non existing channel (incomingChannelIdP: %v)",
			incomingShortChannelId)
		incomingChannelIdP = nil
	}

	var outgoingShortChannelId string
	if clnForward.OutChannel != nil {
		outgoingShortChannelId = *clnForward.OutChannel
	}
	outgoingChannelId := cache.GetChannelIdByShortChannelId(&outgoingShortChannelId)
	outgoingChannelIdP := &outgoingChannelId
	if outgoingChannelId == 0 {
		log.Error().Msgf("Forward received for a non existing channel (outgoingShortChannelId: %v)",
			outgoingShortChannelId)
		outgoingChannelIdP = nil
	}
	// Duplication will get fixed once CLN supports paging
	_, err := db.Exec(`INSERT INTO forward
    					(time, fee_msat, incoming_amount_msat, incoming_channel_id,
    					 outgoing_amount_msat, outgoing_channel_id, node_id)
					VALUES ($1, $2, $3, $4, $5, $6, $7);`,
		forwardTime,
		feeMsat,
		inMsat,
		incomingChannelIdP,
		outMsat,
		outgoingChannelIdP,
		nodeSettings.NodeId,
	)
	if err != nil {
		return errors.Wrap(err, "Executing SQL")
	}
	return nil
}

func storeHtlc(db *sqlx.DB,
	status forwardStatus,
	clnForward *cln.ListforwardsForwards,
	nodeSettings cache.NodeSettingsCache) error {

	forwardTime := time.Unix(int64(clnForward.ReceivedTime), 0).UTC()
	var inMsat uint64
	if clnForward.InMsat != nil {
		inMsat = clnForward.InMsat.Msat
	}
	var outMsat uint64
	if clnForward.OutMsat != nil {
		outMsat = clnForward.OutMsat.Msat
	}

	incomingShortChannelId := clnForward.InChannel
	incomingChannelId := cache.GetChannelIdByShortChannelId(&incomingShortChannelId)
	incomingChannelIdP := &incomingChannelId
	if incomingChannelId == 0 {
		log.Error().Msgf("Forward received for a non existing channel (incomingChannelIdP: %v)",
			incomingShortChannelId)
		incomingChannelIdP = nil
	}

	var outgoingShortChannelId string
	if clnForward.OutChannel != nil {
		outgoingShortChannelId = *clnForward.OutChannel
	}
	outgoingChannelId := cache.GetChannelIdByShortChannelId(&outgoingShortChannelId)
	outgoingChannelIdP := &outgoingChannelId
	if outgoingChannelId == 0 {
		log.Error().Msgf("Forward received for a non existing channel (outgoingShortChannelId: %v)",
			outgoingShortChannelId)
		outgoingChannelIdP = nil
	}

	jb, err := json.Marshal(clnForward)
	if err != nil {
		log.Error().Err(err).Msgf("Marshalling HTLC forward %v", clnForward)
	}
	_, err = db.Exec(`
		INSERT INTO htlc_event
		    (time, data, event_type,
		     incoming_channel_id, incoming_amt_msat, incoming_htlc_id,
		     outgoing_channel_id, outgoing_amt_msat, outgoing_htlc_id,
		     node_id, cln_forward_status_id
		) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);`,
		forwardTime, string(jb), "ForwardFailEvent",
		incomingChannelIdP, inMsat, clnForward.InHtlcId,
		outgoingChannelIdP, outMsat, clnForward.OutHtlcId,
		nodeSettings.NodeId, status,
	)
	if err != nil {
		return errors.Wrap(err, "Executing SQL")
	}
	return nil
}
