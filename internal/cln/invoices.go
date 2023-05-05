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

const streamInvoicesTickerSeconds = 15 * 60

type client_ListInvoices interface {
	ListInvoices(ctx context.Context,
		in *cln.ListinvoicesRequest,
		opts ...grpc.CallOption) (*cln.ListinvoicesResponse, error)
	Decode(ctx context.Context,
		in *cln.DecodeRequest,
		opts ...grpc.CallOption) (*cln.DecodeResponse, error)
}

func SubscribeAndStoreInvoices(ctx context.Context,
	client client_ListInvoices,
	db *sqlx.DB,
	nodeSettings cache.NodeSettingsCache) {

	serviceType := services_helpers.ClnServiceInvoicesService

	cache.SetInitializingNodeServiceState(serviceType, nodeSettings.NodeId)

	ticker := time.NewTicker(streamInvoicesTickerSeconds * time.Second)
	defer ticker.Stop()
	tickerChannel := ticker.C

	err := listAndProcessInvoices(ctx, db, client, serviceType, nodeSettings, true)
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
			err = listAndProcessInvoices(ctx, db, client, serviceType, nodeSettings, false)
			if err != nil {
				processError(ctx, serviceType, nodeSettings, err)
				return
			}
		}
	}
}

func listAndProcessInvoices(ctx context.Context,
	db *sqlx.DB,
	client client_ListInvoices,
	serviceType services_helpers.ServiceType,
	nodeSettings cache.NodeSettingsCache,
	bootStrapping bool) error {

	clnInvoices, err := client.ListInvoices(ctx, &cln.ListinvoicesRequest{})
	if err != nil {
		return errors.Wrapf(err, "listing invoices for nodeId: %v", nodeSettings.NodeId)
	}

	err = storeInvoices(ctx, db, client, clnInvoices.Invoices, serviceType, nodeSettings, bootStrapping)
	if err != nil {
		return errors.Wrap(err, "storing invoices")
	}

	if bootStrapping {
		log.Info().Msgf("Initial import of invoices is done for nodeId: %v", nodeSettings.NodeId)
		cache.SetActiveNodeServiceState(serviceType, nodeSettings.NodeId)
	}
	return nil
}

func storeInvoices(ctx context.Context,
	db *sqlx.DB,
	client client_ListInvoices,
	clnInvoices []*cln.ListinvoicesInvoices,
	serviceType services_helpers.ServiceType,
	nodeSettings cache.NodeSettingsCache,
	bootStrapping bool) error {

	var lastImportAt *time.Time
	err := db.Get(&lastImportAt, `SELECT MAX(created_on) FROM invoice WHERE node_id=$1`, nodeSettings.NodeId)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return errors.Wrapf(err, "obtaining latest creation date for invoices of nodeId: %v", nodeSettings.NodeId)
	}

	var lastPaymentAt *time.Time
	err = db.Get(&lastPaymentAt, `SELECT MAX(settle_date) FROM invoice WHERE node_id=$1`, nodeSettings.NodeId)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return errors.Wrapf(err, "obtaining latest payment date for invoices of nodeId: %v", nodeSettings.NodeId)
	}

	counter := 0
	for _, clnInvoice := range clnInvoices {
		if clnInvoice == nil {
			continue
		}
		expiresAt := time.Unix(int64(clnInvoice.ExpiresAt), 0)
		if lastImportAt != nil && expiresAt.Before(*lastImportAt) {
			continue
		}
		if lastPaymentAt != nil && clnInvoice.PaidAt != nil {
			paidAt := time.Unix(int64(*clnInvoice.PaidAt), 0)
			if paidAt.Before(*lastPaymentAt) {
				continue
			}
		}
		err = storeInvoice(ctx, db, client, clnInvoice, nodeSettings)
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

func storeInvoice(ctx context.Context,
	db *sqlx.DB,
	client client_ListInvoices,
	clnInvoice *cln.ListinvoicesInvoices,
	nodeSettings cache.NodeSettingsCache) error {

	expiresAt := time.Unix(int64(clnInvoice.ExpiresAt), 0)
	var paidAt *time.Time
	if clnInvoice.PaidAt != nil {
		paidAtTime := time.Unix(int64(*clnInvoice.PaidAt), 0)
		paidAt = &paidAtTime
	}

	var amountMsat uint64
	if clnInvoice.AmountMsat != nil {
		amountMsat = clnInvoice.AmountMsat.Msat
	}
	var amountPaidMsat uint64
	if clnInvoice.AmountReceivedMsat != nil {
		amountPaidMsat = clnInvoice.AmountReceivedMsat.Msat
	}
	var invoiceState string
	switch clnInvoice.Status {
	case cln.ListinvoicesInvoices_UNPAID:
		invoiceState = "OPEN"
	case cln.ListinvoicesInvoices_PAID:
		invoiceState = "SETTLED"
	case cln.ListinvoicesInvoices_EXPIRED:
		invoiceState = "CANCELED"
	}

	var existingCreationDate *time.Time
	err := db.Get(&existingCreationDate,
		`SELECT creation_date FROM invoice WHERE node_id=$1 AND label=$2`, nodeSettings.NodeId, clnInvoice.Label)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return errors.Wrapf(err,
			"finding existing invoice for label: %v, nodeId: %v", clnInvoice.Label, nodeSettings.NodeId)
	}

	if existingCreationDate == nil {
		var invoiceString string
		if clnInvoice.Bolt11 != nil {
			invoiceString = *clnInvoice.Bolt11
		}
		if clnInvoice.Bolt12 != nil {
			invoiceString = *clnInvoice.Bolt12
		}
		decodedInvoice, err := client.Decode(ctx, &cln.DecodeRequest{String_: invoiceString})
		if err != nil {
			return errors.Wrapf(err,
				"decoding invoice failed for label: %v, nodeId: %v", clnInvoice.Label, nodeSettings.NodeId)
		}
		if decodedInvoice == nil {
			return errors.Wrapf(err,
				"decoding invoice failed for label: %v, nodeId: %v", clnInvoice.Label, nodeSettings.NodeId)
		}
		var creationDate *time.Time
		if decodedInvoice.CreatedAt != nil {
			creationDateTime := time.Unix(int64(*decodedInvoice.CreatedAt), 0)
			creationDate = &creationDateTime
		}
		descriptionHash := hex.EncodeToString(decodedInvoice.DescriptionHash)
		// TODO FIXME CLN fix this fallback stuff
		//if decodedInvoice.Fallbacks
		//fallback_addr
		var destinationNodeId *int
		var destinationPublicKey *string
		destinationPublicKeyString := hex.EncodeToString(decodedInvoice.Payee)
		if destinationPublicKeyString != "" {
			destinationPublicKey = &destinationPublicKeyString
			destinationNodeIdInt := cache.GetPeerNodeIdByPublicKey(
				destinationPublicKeyString, nodeSettings.Chain, nodeSettings.Network)
			if destinationNodeIdInt != 0 {
				destinationNodeId = &destinationNodeIdInt
			}
		}

		// TODO FIXME CLN: can we get HTLCs for a settled invoice?
		// TODO FIXME CLN: RoutHints, Features are missing?

		_, err = db.Exec(`INSERT INTO invoice (
				memo, label, type,
				r_preimage, r_hash,
				value_msat, settle_date, expiry, settle_index, amt_paid_msat, invoice_state,
			 	creation_date, description_hash, destination_node_id, destination_pub_key,
            	bolt11, bolt12,
				node_id, created_on, updated_on
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
			);`,
			clnInvoice.Description, clnInvoice.Label, decodedInvoice.ItemType,
			hex.EncodeToString(clnInvoice.PaymentPreimage), hex.EncodeToString(clnInvoice.PaymentHash),
			amountMsat, paidAt, expiresAt, clnInvoice.PayIndex, amountPaidMsat, invoiceState,
			creationDate, descriptionHash, destinationNodeId, destinationPublicKey,
			clnInvoice.Bolt11, clnInvoice.Bolt12,
			nodeSettings.NodeId, time.Now(), time.Now())
		if err != nil {
			return errors.Wrap(err, "Executing SQL")
		}
		return nil
	}

	_, err = db.Exec(`UPDATE invoice
			SET settle_date=$1, settle_index=$2, amt_paid_msat=$3, invoice_state=$4,
			    node_id=$5, updated_on=$6
			WHERE label=$7 AND (
			      	(settle_date IS NULL OR settle_date!=$1) OR
			      	(settle_index IS NULL OR settle_index!=$2) OR
			      	(amt_paid_msat IS NULL OR amt_paid_msat!=$3) OR
			      	(invoice_state IS NULL OR invoice_state!=$4)
			      );`,
		paidAt, clnInvoice.PayIndex, amountPaidMsat, invoiceState, nodeSettings.NodeId, time.Now(), clnInvoice.Label)
	if err != nil {
		return errors.Wrap(err, "Executing SQL")
	}
	return nil
}
