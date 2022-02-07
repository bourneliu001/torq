package lndutil

import (
	"context"
	"encoding/json"
	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/lightningnetwork/lnd/lnrpc"
	"golang.org/x/sync/errgroup"
	"io"
	"time"
)

// SubscribeAndStoreChannelGraph Subscribes to channel updates
func SubscribeAndStoreChannelGraph(ctx context.Context, client lnrpc.LightningClient, db *sqlx.DB) error {

	req := lnrpc.GraphTopologySubscription{}
	stream, err := client.SubscribeChannelGraph(ctx, &req)

	if err != nil {
		return errors.Wrapf(err, "SubscribeAndStoreChannelGraph -> client.SubscribeChannelGraph(%v, %v)", ctx, req)
	}

	errGrp, ctx := errgroup.WithContext(ctx)

	for {

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		gpu, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return errors.Wrap(err, "SubscribeChannelEvents -> stream.Recv()")
		}

		errGrp.Go(func() error {
			err := processNodeUpdates(gpu.NodeUpdates, db)
			if err != nil {
				return err
			}
			return nil
		})

		errGrp.Go(func() error {
			err := processChannelUpdates(gpu.ChannelUpdates, db)
			if err != nil {
				return err
			}
			return nil
		})

	}

	return errGrp.Wait()
}

func processNodeUpdates(nus []*lnrpc.NodeUpdate, db *sqlx.DB) error {

	for _, nu := range nus {
		// Check if this node update is relevant to a node we have or have had a channel with
		relevant, _ := isRelevantOrOurNode(nu.IdentityKey)

		if relevant {
			ts := time.Now()
			err := insertNodeEvent(db, ts, nu.IdentityKey, nu.Alias, nu.Color,
				nu.NodeAddresses, nu.Features)
			if err != nil {
				return errors.Wrapf(err, "processNodeUpdates ->insertNodeEvent(%v, %s, %s, %s, %s, %v, %v)",
					db, ts, nu.IdentityKey, nu.Alias, nu.Color, nu.NodeAddresses, nu.Features)
			}
		}

	}

	return nil
}

func processChannelUpdates(cus []*lnrpc.ChannelEdgeUpdate, db *sqlx.DB) error {
	for _, cu := range cus {

		// Check if this channel update is relevant to one of our channels
		// And if one of our nodes is advertising the channel update (meaning
		// we have changed our the channel policy).
		ourNode := isOurNode(cu.AdvertisingNode)
		chanPoint, err := getChanPoint(cu.ChanPoint.GetFundingTxidBytes(), cu.ChanPoint.OutputIndex)
		if err != nil {
			return errors.Wrapf(err, "SubscribeChannelEvents ->getChanPoint(%b, %d)",
				cu.ChanPoint.GetFundingTxidBytes(), cu.ChanPoint.OutputIndex)
		}
		relevantChannel := isRelevantChannel(chanPoint)

		if relevantChannel {
			ts := time.Now()
			err := insertRoutingPolicy(db, ts, ourNode, cu)
			if err != nil {
				return errors.Wrapf(err, "SubscribeChannelEvents ->insertRoutingPolicy(%v, %s, %t, %v)",
					db, ts, ourNode, cu)
			}
		}

	}

	return nil
}

const rpQuery = `
INSERT INTO routing_policy (ts,
	chan_id, 
	announcing_pub_key,
	chan_point,
	outbound,
	disabled,
	time_lock_delta,
	min_htlc,
	max_htlc_msat,
	fee_base_msat,
	fee_rate_mill_msat) 
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

func insertRoutingPolicy(db *sqlx.DB, ts time.Time, outbound bool, cu *lnrpc.ChannelEdgeUpdate) error {

	cp, err := getChanPoint(cu.ChanPoint.GetFundingTxidBytes(), cu.ChanPoint.GetOutputIndex())
	if err != nil {
		return err
	}

	db.Exec(rpQuery, ts, cu.ChanId, cu.AdvertisingNode, cp, outbound,
		cu.RoutingPolicy.Disabled, cu.RoutingPolicy.TimeLockDelta, cu.RoutingPolicy.MinHtlc,
		cu.RoutingPolicy.MaxHtlcMsat, cu.RoutingPolicy.FeeBaseMsat, cu.RoutingPolicy.FeeRateMilliMsat)

	return nil
}

func insertNodeEvent(db *sqlx.DB, ts time.Time, pubKey string, alias string, color string,
	na []*lnrpc.NodeAddress, f map[uint32]*lnrpc.Feature) error {

	// Create json byte object from node address map
	najb, err := json.Marshal(na)
	if err != nil {
		return errors.Wrapf(err, "insertNodeEvent -> json.Marshal(%v)", na)
	}

	// Create json byte object from features list
	fjb, err := json.Marshal(f)
	if err != nil {
		return errors.Wrapf(err, "insertNodeEvent -> json.Marshal(%v)", f)
	}

	db.Exec(`INSERT INTO node_event (timestamp, pub_key, alias, color, node_addresses, features) 
    	VALUES ($1, $2, $3, $4, $5, $6)`, ts, pubKey, alias, color, najb, fjb)

	return nil
}

var chanPointList []string

func InitChanIdList(db *sqlx.DB) error {
	q := `
		select array_agg(chan_point) as chan_point from (
			select
				last(event_type, time) as event_type,
				last(chan_point,time) as chan_point
			from channel_event
			where event_type in(0,1)
			group by chan_point
		) as t
		where t.event_type = 0;`

	err := db.QueryRowx(q).Scan(pq.Array(&chanPointList))
	if err != nil {
		return errors.Wrapf(err, "InitChanIdList -> db.QueryRow(%s).Scan(pq.Array(%v))", q, chanPointList)
	}

	return nil
}

func UpdateChanIdList(c chan string) {

	var chanPoint string

waitForUpdate:
	for {
		// Wait for new peers to enter
		chanPoint = <-c

		// Remove chanPoint to the list, if it's already present.
		// continue the outer loop and wait for new update
		for i, cp := range chanPointList {
			if cp == chanPoint {
				chanPointList = append(chanPointList[:i], chanPointList[i+1:]...)
				continue waitForUpdate
			}
		}

		// If not present add it to the chanID list
		chanPointList = append(chanPointList, chanPoint)

	}
}

var ourNodePubKeys []string

// InitOurNodesList populates the list of nodes owned by us.
func InitOurNodesList(pubkeys []string) {
	ourNodePubKeys = pubkeys
}

// pubKeyList is used to store which node and channel updates to store. We only want to store
// updates that are relevant to our channels and their nodes.
var pubKeyList []string

// InitPeerList fetches all public keys from the list of all channels. This is used to
// filter out noise from the graph updates.
func InitPeerList(db *sqlx.DB) error {
	q := `select array_agg(distinct pub_key) as all_nodes from channel_event where event_type in (0, 1);`
	err := db.QueryRow(q).Scan(pq.Array(&pubKeyList))
	if err != nil {
		return errors.Wrapf(err, "InitPeerList -> db.QueryRow(%s).Scan(pq.Array(%v))", q, pubKeyList)
	}

	return nil
}

// UpdatePeerList is meant to run as a gorouting. It adds new public keys to the pubKeyList
// and removes existing ones.
func UpdatePeerList(c chan string) {

	var pubKey string
	var present bool

	for {
		// Wait for new peers to enter
		pubKey = <-c
		// Add it to the peer list, if not already present.
		for _, p := range pubKeyList {
			if p == pubKey {
				present = true
				break
			}
		}

		// If not present add it to the public key  list
		if present == false {
			pubKeyList = append(pubKeyList, pubKey)
		}

		// Reset to false in order to allow the next public key to be added.
		present = false
	}
}

// isRelevantOrOurNode is used to check if any public key is in the pubKeyList.
// The first boolean returned indicate if the key is relevant, the second boolean
// indicates that it is one of our own nodes.
func isRelevantOrOurNode(pubKey string) (bool, bool) {

	if isOurNode(pubKey) {
		// Is relevant (first boolean), _and_ our node (second boolean).
		return true, true
	}

	if isRelevant(pubKey) {
		// Is relevant (first boolean), _but not_ our node (second boolean).
		return true, false
	}

	return false, false
}

func isRelevantChannel(chanPoint string) bool {
	for _, cid := range chanPointList {
		if cid == chanPoint {
			return true
		}
	}
	return false
}

// isRelevant is used to check if any public key is in the pubKeyList.
func isRelevant(pubKey string) bool {

	for _, p := range pubKeyList {

		// Check if any of the provided public keys equals the current public key.
		if p == pubKey {
			// If found, no reason to search further, immediately return true.
			return true
		}

	}

	return false
}

// isOurNode is used to check if the public key is from one of our own nodes.
func isOurNode(pubKey string) bool {

	for _, p := range ourNodePubKeys {

		// Check if the public key belongs to one of our nodes.
		if p == pubKey {
			// If found, no reason to search further, immediately return true, and true
			return true
		}

	}
	return false
}
