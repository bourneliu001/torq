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
	"github.com/lncapital/torq/internal/services_helpers"
	"github.com/lncapital/torq/proto/cln"
)

const streamPaymentsTickerSeconds = 15 * 60

type client_ListPayments interface {
	ListSendPays(ctx context.Context,
		in *cln.ListsendpaysRequest,
		opts ...grpc.CallOption) (*cln.ListsendpaysResponse, error)
	Decode(ctx context.Context,
		in *cln.DecodeRequest,
		opts ...grpc.CallOption) (*cln.DecodeResponse, error)
}

func SubscribeAndStorePayments(ctx context.Context,
	client client_ListPayments,
	db *sqlx.DB,
	nodeSettings cache.NodeSettingsCache) {

	serviceType := services_helpers.ClnServicePaymentsService

	cache.SetInitializingNodeServiceState(serviceType, nodeSettings.NodeId)

	ticker := time.NewTicker(streamPaymentsTickerSeconds * time.Second)
	defer ticker.Stop()
	tickerChannel := ticker.C

	err := listAndProcessPayments(ctx, db, client, serviceType, nodeSettings, true)
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
			err = listAndProcessPayments(ctx, db, client, serviceType, nodeSettings, false)
			if err != nil {
				processError(ctx, serviceType, nodeSettings, err)
				return
			}
		}
	}
}

func listAndProcessPayments(ctx context.Context,
	db *sqlx.DB,
	client client_ListPayments,
	serviceType services_helpers.ServiceType,
	nodeSettings cache.NodeSettingsCache,
	bootStrapping bool) error {

	err := processPayments(ctx, db, client, serviceType, cln.ListsendpaysRequest_FAILED, nodeSettings, bootStrapping)
	if err != nil {
		return errors.Wrap(err, "processing failed payments")
	}
	err = processPayments(ctx, db, client, serviceType, cln.ListsendpaysRequest_PENDING, nodeSettings, bootStrapping)
	if err != nil {
		return errors.Wrap(err, "processing pending payments")
	}
	err = processPayments(ctx, db, client, serviceType, cln.ListsendpaysRequest_COMPLETE, nodeSettings, bootStrapping)
	if err != nil {
		return errors.Wrap(err, "processing completed payments")
	}

	if bootStrapping {
		log.Info().Msgf("Initial import of payments is done for nodeId: %v", nodeSettings.NodeId)
		cache.SetActiveNodeServiceState(serviceType, nodeSettings.NodeId)
	}
	return nil
}

func processPayments(ctx context.Context,
	db *sqlx.DB,
	client client_ListPayments,
	serviceType services_helpers.ServiceType,
	status cln.ListsendpaysRequest_ListsendpaysStatus,
	nodeSettings cache.NodeSettingsCache,
	bootStrapping bool) error {

	clnPayments, err := client.ListSendPays(ctx, &cln.ListsendpaysRequest{
		Status: &status,
	})
	if err != nil {
		return errors.Wrapf(err, "listing payments for nodeId: %v", nodeSettings.NodeId)
	}

	err = storePayments(ctx, db, client, clnPayments.Payments, serviceType, nodeSettings, bootStrapping)
	if err != nil {
		return errors.Wrap(err, "storing payments")
	}

	if bootStrapping {
		log.Info().Msgf("Initial import of payments is done for nodeId: %v", nodeSettings.NodeId)
		cache.SetActiveNodeServiceState(serviceType, nodeSettings.NodeId)
	}
	return nil
}

func storePayments(ctx context.Context,
	db *sqlx.DB,
	client client_ListPayments,
	clnPayments []*cln.ListsendpaysPayments,
	serviceType services_helpers.ServiceType,
	nodeSettings cache.NodeSettingsCache,
	bootStrapping bool) error {

	var lastCreatedAt *time.Time
	err := db.Get(&lastCreatedAt, `SELECT MAX(creation_timestamp) FROM payment WHERE node_id=$1`, nodeSettings.NodeId)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return errors.Wrapf(err, "obtaining latest creation date for invoices of nodeId: %v", nodeSettings.NodeId)
	}

	counter := 0
	for _, clnPayment := range clnPayments {
		if clnPayment == nil {
			continue
		}
		createdAt := time.Unix(int64(clnPayment.CreatedAt), 0)
		if createdAt.Before(*lastCreatedAt) {
			continue
		}
		err = storePayment(ctx, db, client, clnPayment, nodeSettings)
		if err != nil {
			return errors.Wrapf(err, "persisting invoice for nodeId: %v", nodeSettings.NodeId)
		}
		counter++
		if bootStrapping && counter%10_000 == 0 {
			cache.SetInitializingNodeServiceState(serviceType, nodeSettings.NodeId)
		}
	}
	return nil
}

func storePayment(ctx context.Context,
	db *sqlx.DB,
	client client_ListPayments,
	clnPayment *cln.ListsendpaysPayments,
	nodeSettings cache.NodeSettingsCache) error {

	createdAt := time.Unix(int64(clnPayment.CreatedAt), 0)

	var invoiceState string
	switch clnPayment.Status {
	case cln.ListsendpaysPayments_PENDING:
		invoiceState = "IN_FLIGHT"
	case cln.ListsendpaysPayments_FAILED:
		invoiceState = "FAILED"
	case cln.ListsendpaysPayments_COMPLETE:
		invoiceState = "SUCCEEDED"
	}
	var destinationNodeId *int
	var destinationPublicKey *string
	destinationPublicKeyString := hex.EncodeToString(clnPayment.Destination)
	if destinationPublicKeyString != "" {
		destinationPublicKey = &destinationPublicKeyString
		destinationNodeIdInt := cache.GetPeerNodeIdByPublicKey(
			destinationPublicKeyString, nodeSettings.Chain, nodeSettings.Network)
		if destinationNodeIdInt != 0 {
			destinationNodeId = &destinationNodeIdInt
		}
	}
	var label *string
	if clnPayment.Label != nil {
		label = clnPayment.Label
	}
	var failureReason *string
	if clnPayment.Erroronion != nil {
		failureReasonString := hex.EncodeToString(clnPayment.Erroronion)
		failureReason = &failureReasonString
	}
	var amountMsat *uint64
	if clnPayment.AmountSentMsat != nil {
		amountMsat = &clnPayment.AmountSentMsat.Msat
	}

	var existingStatus *string
	err := db.Get(&existingStatus,
		`SELECT status FROM payment WHERE node_id=$1 AND payment_hash=$2`,
		nodeSettings.NodeId, hex.EncodeToString(clnPayment.PaymentHash))
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return errors.Wrapf(err,
			"finding existing invoice for payment hash: %v, nodeId: %v",
			hex.EncodeToString(clnPayment.PaymentHash), nodeSettings.NodeId)
	}

	if existingStatus == nil {
		// TODO FIXME CLN can we get HTLCs for a settled payments?
		_, err = db.Exec(`INSERT INTO payment
    		(payment_hash, creation_timestamp, payment_preimage, label,
    		 value_msat, bolt11, bolt12,
    		 status, failure_reason, destination_node_id, destination_pub_key, description,
    		 node_id, created_on, updated_on)
			  VALUES ($1, $2, $3, $4, $5,$6, $7, $8, $9, $10, $11, $12, $13, $14, $15);`,
			hex.EncodeToString(clnPayment.PaymentHash), createdAt, hex.EncodeToString(clnPayment.PaymentPreimage), label,
			amountMsat, clnPayment.Bolt11, clnPayment.Bolt12,
			invoiceState, failureReason, destinationNodeId, destinationPublicKey, clnPayment.Description,
			nodeSettings.NodeId, time.Now(), time.Now())
		if err != nil {
			return errors.Wrap(err, "Executing SQL")
		}
		return nil
	}

	_, err = db.Exec(`UPDATE payment
			SET status=$1, failure_reason=$2, description=$3, payment_preimage=$4
			    node_id=$5, updated_on=$6
			WHERE payment_hash=$7 AND (
			      	(status IS NULL OR status!=$1) OR
			      	(failure_reason IS NULL OR failure_reason!=$2) OR
			      	(description IS NULL OR description!=$3) OR
			      	(payment_preimage IS NULL OR payment_preimage!=$4)
			      );`,
		invoiceState, failureReason, clnPayment.Description, clnPayment.PaymentPreimage,
		nodeSettings.NodeId, time.Now(), hex.EncodeToString(clnPayment.PaymentHash))
	if err != nil {
		return errors.Wrap(err, "Executing SQL")
	}
	return nil
}
