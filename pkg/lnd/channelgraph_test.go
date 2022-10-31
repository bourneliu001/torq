package lnd

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"gopkg.in/guregu/null.v4"

	"github.com/lncapital/torq/internal/channels"
	"github.com/lncapital/torq/pkg/commons"
	"github.com/lncapital/torq/testutil"
)

type stubLNDSubscribeChannelGraphRPC struct {
	grpc.ClientStream
	GraphTopologyUpdate []*lnrpc.GraphTopologyUpdate
	CancelFunc          func()
}

func (s *stubLNDSubscribeChannelGraphRPC) Recv() (*lnrpc.GraphTopologyUpdate, error) {
	if len(s.GraphTopologyUpdate) == 0 {
		s.CancelFunc()
		return nil, context.Canceled
	}
	var gtu interface{}
	gtu, s.GraphTopologyUpdate = s.GraphTopologyUpdate[0], s.GraphTopologyUpdate[1:]
	if u, ok := gtu.(*lnrpc.GraphTopologyUpdate); ok {
		return u, nil
	}
	s.CancelFunc()
	return nil, context.Canceled
}

func (s *stubLNDSubscribeChannelGraphRPC) SubscribeChannelGraph(ctx context.Context, in *lnrpc.GraphTopologySubscription,
	opts ...grpc.CallOption) (lnrpc.Lightning_SubscribeChannelGraphClient, error) {

	return s, nil
}

func TestSubscribeChannelGraphUpdates(t *testing.T) {
	srv, err := testutil.InitTestDBConn()
	if err != nil {
		panic(err)
	}

	db, cancel, err := srv.NewTestDatabase(true)
	// TODO FIXME WHY?
	defer time.Sleep(1 * time.Second)
	defer cancel()
	if err != nil {
		t.Fatal(err)
	}

	chanPoint := &lnrpc.ChannelPoint{FundingTxid: &lnrpc.
		ChannelPoint_FundingTxidBytes{
		FundingTxidBytes: []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}},
		OutputIndex: 1}

	chanPointStr, err := chanPointFromByte([]byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}, chanPoint.OutputIndex)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Irrelevant routing policy updates are ignored", func(t *testing.T) {

		irrelevatChannelPoint := &lnrpc.ChannelPoint{FundingTxid: &lnrpc.
			ChannelPoint_FundingTxidBytes{
			FundingTxidBytes: []byte{2, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
				1, 1, 1, 1}},
			OutputIndex: 2}

		irrelecantUpdateEvent := lnrpc.GraphTopologyUpdate{
			NodeUpdates: nil,
			ChannelUpdates: []*lnrpc.ChannelEdgeUpdate{{
				ChanId:    12345678,
				ChanPoint: irrelevatChannelPoint,
				Capacity:  2000000,
				RoutingPolicy: &lnrpc.RoutingPolicy{
					TimeLockDelta:    0,
					MinHtlc:          0,
					FeeBaseMsat:      2,
					FeeRateMilliMsat: 200,
					Disabled:         true,
					MaxHtlcMsat:      1000,
					LastUpdate:       0,
				},
				AdvertisingNode: testutil.TestPublicKey2,
				ConnectingNode:  testutil.TestPublicKey1,
			}},
			ClosedChans: nil,
		}

		result := simulateChannelGraphUpdate(t, db, &stubLNDSubscribeChannelGraphRPC{GraphTopologyUpdate: []*lnrpc.
			GraphTopologyUpdate{&irrelecantUpdateEvent}}, chanPointStr)

		if len(result) != 0 {
			testutil.Fatalf(t, "Expected to find no routing policy record stored in the database. Found %d",
				len(result))
		}
	})

	t.Run("Relevant routing policies are correctly stored", func(t *testing.T) {

		updateEvent := lnrpc.GraphTopologyUpdate{
			NodeUpdates: nil,
			ChannelUpdates: []*lnrpc.ChannelEdgeUpdate{{
				ChanId:    1111,
				ChanPoint: chanPoint,
				Capacity:  1000000,
				RoutingPolicy: &lnrpc.RoutingPolicy{
					TimeLockDelta:    0,
					MinHtlc:          0,
					FeeBaseMsat:      2,
					FeeRateMilliMsat: 100,
					Disabled:         true,
					MaxHtlcMsat:      1000,
					LastUpdate:       0,
				},
				AdvertisingNode: testutil.TestPublicKey2,
				ConnectingNode:  testutil.TestPublicKey1,
			}},
			ClosedChans: nil,
		}

		expected := routingPolicyData{
			Ts:                time.Now(),
			LNDChannelPoint:   chanPointStr,
			LNDShortChannelId: updateEvent.ChannelUpdates[0].ChanId,
			Outbound:          false,
			AnnouncingPubKey:  updateEvent.ChannelUpdates[0].AdvertisingNode,
			FeeRateMillMsat:   updateEvent.ChannelUpdates[0].RoutingPolicy.FeeRateMilliMsat,
			FeeBaseMsat:       updateEvent.ChannelUpdates[0].RoutingPolicy.FeeBaseMsat,
			MaxHtlcMsat:       updateEvent.ChannelUpdates[0].RoutingPolicy.MaxHtlcMsat,
			MinHtlc:           updateEvent.ChannelUpdates[0].RoutingPolicy.MinHtlc,
			TimeLockDelta:     updateEvent.ChannelUpdates[0].RoutingPolicy.TimeLockDelta,
			Disabled:          updateEvent.ChannelUpdates[0].RoutingPolicy.Disabled,
		}

		result := simulateChannelGraphUpdate(t, db,
			&stubLNDSubscribeChannelGraphRPC{GraphTopologyUpdate: []*lnrpc.GraphTopologyUpdate{&updateEvent}},
			chanPointStr,
		)

		if len(result) != 1 {
			testutil.Fatalf(t, "Expected to find a single routing policy record stored in the database. Found %d",
				len(result))
		}

		if result[0].AnnouncingPubKey != expected.AnnouncingPubKey {
			testutil.Errorf(t, "Incorrect announcing pub key. Expected: %v, got %v", expected.AnnouncingPubKey,
				result[0].AnnouncingPubKey)
		}

		if result[0].LNDChannelPoint != expected.LNDChannelPoint {
			testutil.Errorf(t, "Incorrect channel point. Expected: %v, got %v", expected.LNDChannelPoint, result[0].LNDChannelPoint)
		}

		if result[0].LNDShortChannelId != expected.LNDShortChannelId {
			testutil.Errorf(t, "Incorrect channel id. Expected: %v, got %v", expected.LNDShortChannelId, result[0].LNDShortChannelId)
		}

		if result[0].Disabled != expected.Disabled {
			testutil.Errorf(t, "Incorrect channel disabled state. Expected: %v, got %v", expected.Disabled,
				result[0].Disabled)
		}

		if result[0].FeeRateMillMsat != expected.FeeRateMillMsat {
			testutil.Errorf(t, "Incorrect fee rate. Expected: %v, got %v", expected.FeeRateMillMsat,
				result[0].FeeRateMillMsat)
		}

		if result[0].FeeBaseMsat != expected.FeeBaseMsat {
			testutil.Errorf(t, "Incorrect base fee state. Expected: %v, got %v", expected.FeeBaseMsat,
				result[0].FeeBaseMsat)
		}

		if result[0].MinHtlc != expected.MinHtlc {
			testutil.Errorf(t, "Incorrect min htlc. Expected: %v, got %v", expected.MinHtlc, result[0].MinHtlc)
		}

		if result[0].MaxHtlcMsat != expected.MaxHtlcMsat {
			testutil.Errorf(t, "Incorrect max htlc. Expected: %v, got %v", expected.MaxHtlcMsat, result[0].MaxHtlcMsat)
		}

		if result[0].Outbound != expected.Outbound {
			testutil.Errorf(t, "Incorrect outbound state. Expected: %v, got %v", expected.Outbound, result[0].Outbound)
		}

		if result[0].TimeLockDelta != expected.TimeLockDelta {
			testutil.Errorf(t, "Incorrect timelock delta. Expected: %v, got %v", expected.TimeLockDelta,
				result[0].TimeLockDelta)
		}

		r2 := simulateChannelGraphUpdate(t, db, &stubLNDSubscribeChannelGraphRPC{GraphTopologyUpdate: []*lnrpc.
			GraphTopologyUpdate{&updateEvent}}, chanPointStr)

		if len(r2) != 1 {
			testutil.Fatalf(t, "Expected to find a single routing policy record stored in the database. Found %d",
				len(r2))
		}

		if t.Failed() {
			t.FailNow()
		}

		secondUpdateEvent := lnrpc.GraphTopologyUpdate{
			NodeUpdates: nil,
			ChannelUpdates: []*lnrpc.ChannelEdgeUpdate{{
				ChanId:    1111,
				ChanPoint: chanPoint,
				Capacity:  2000000,
				RoutingPolicy: &lnrpc.RoutingPolicy{
					TimeLockDelta:    0,
					MinHtlc:          0,
					FeeBaseMsat:      2,
					FeeRateMilliMsat: 200,
					Disabled:         true,
					MaxHtlcMsat:      1000,
					LastUpdate:       0,
				},
				AdvertisingNode: testutil.TestPublicKey1,
				ConnectingNode:  testutil.TestPublicKey2,
			}},
			ClosedChans: nil,
		}

		e3 := routingPolicyData{
			Ts:                time.Now(),
			LNDChannelPoint:   chanPointStr,
			LNDShortChannelId: secondUpdateEvent.ChannelUpdates[0].ChanId,
			Outbound:          true,
			AnnouncingPubKey:  testutil.TestPublicKey1,
			FeeRateMillMsat:   secondUpdateEvent.ChannelUpdates[0].RoutingPolicy.FeeRateMilliMsat,
			FeeBaseMsat:       secondUpdateEvent.ChannelUpdates[0].RoutingPolicy.FeeBaseMsat,
			MaxHtlcMsat:       secondUpdateEvent.ChannelUpdates[0].RoutingPolicy.MaxHtlcMsat,
			MinHtlc:           secondUpdateEvent.ChannelUpdates[0].RoutingPolicy.MinHtlc,
			TimeLockDelta:     secondUpdateEvent.ChannelUpdates[0].RoutingPolicy.TimeLockDelta,
			Disabled:          secondUpdateEvent.ChannelUpdates[0].RoutingPolicy.Disabled,
		}

		r3 := simulateChannelGraphUpdate(t, db, &stubLNDSubscribeChannelGraphRPC{GraphTopologyUpdate: []*lnrpc.
			GraphTopologyUpdate{&secondUpdateEvent}}, chanPointStr)

		if r3[1].AnnouncingPubKey != e3.AnnouncingPubKey {
			testutil.Errorf(t, "Incorrect announcing pub key. Expected: %v, got %v", e3.AnnouncingPubKey,
				r3[1].AnnouncingPubKey)
		}

		if r3[1].Outbound != e3.Outbound {
			testutil.Errorf(t, "Incorrect outbound state. Expected: %v, got %v", e3.Outbound, r3[1].Outbound)
		}

	})

}

type routingPolicyData struct {
	Ts                time.Time
	Outbound          bool   `db:"outbound"`
	FeeRateMillMsat   int64  `db:"fee_rate_mill_msat"`
	FeeBaseMsat       int64  `db:"fee_base_msat"`
	MaxHtlcMsat       uint64 `db:"max_htlc_msat"`
	MinHtlc           int64  `db:"min_htlc"`
	TimeLockDelta     uint32 `db:"time_lock_delta"`
	Disabled          bool   `db:"disabled"`
	ChannelId         int    `db:"channel_id"`
	AnnouncingNodeId  int    `db:"announcing_node_id"`
	AnnouncingPubKey  string `db:"announcing_public_key"`
	ConnectingNodeId  int    `db:"connecting_node_id"`
	ConnectingPubKey  string `db:"connecting_public_key"`
	NodeId            int    `db:"node_id"`
	LNDChannelPoint   string `db:"lnd_channel_point"`
	LNDShortChannelId uint64 `db:"lnd_short_channel_id"`
}

func simulateChannelGraphUpdate(t *testing.T, db *sqlx.DB, client *stubLNDSubscribeChannelGraphRPC, chanPointStr string) (r []routingPolicyData) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	errs, ctx := errgroup.WithContext(ctx)
	client.CancelFunc = cancel

	channel := channels.Channel{
		ShortChannelID:    "1111",
		FirstNodeId:       commons.GetNodeIdFromPublicKey(testutil.TestPublicKey1, commons.Bitcoin, commons.SigNet),
		SecondNodeId:      commons.GetNodeIdFromPublicKey(testutil.TestPublicKey2, commons.Bitcoin, commons.SigNet),
		LNDShortChannelID: 1111,
		LNDChannelPoint:   null.StringFrom(chanPointStr),
		Status:            channels.Open,
	}
	channelId, err := channels.AddChannelOrUpdateChannelStatus(db, channel)
	if err != nil {
		t.Fatalf("Problem adding channel %v", channel)
	}
	t.Logf("channel added with channelId: %v", channelId)

	errs.Go(func() error {
		err := SubscribeAndStoreChannelGraph(ctx, client, db,
			commons.GetNodeSettingsByNodeId(
				commons.GetNodeIdFromPublicKey(testutil.TestPublicKey1, commons.Bitcoin, commons.SigNet)))
		if err != nil {
			t.Fatalf("Problem subscribing to channel graph: %v", err)
		}
		return err
	})

	// Wait for subscriptions to complete
	err = errs.Wait()
	if err != nil {
		t.Fatal(err)
	}

	var result []routingPolicyData
	err = db.Select(&result, `
			select rp.ts,
				   rp.outbound,
				   rp.fee_rate_mill_msat,
				   rp.fee_base_msat,
				   rp.max_htlc_msat,
				   rp.min_htlc,
				   rp.time_lock_delta,
				   rp.disabled,
				   c.channel_id,
				   rp.announcing_node_id,
				   rp.connecting_node_id,
				   rp.node_id,
				   an.public_key AS announcing_public_key,
				   cn.public_key AS connecting_public_key,
				   c.lnd_short_channel_id,
				   c.lnd_channel_point
			from routing_policy rp
			JOIN channel c ON c.channel_id=rp.channel_id
			JOIN node an ON rp.announcing_node_id=an.node_id
			JOIN node cn ON rp.connecting_node_id=cn.node_id;`)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			t.Fatal("There were no routing policies but I did expect there to be some")
		}
		t.Fatalf("Problem executing sql: %v", err)
	}

	return result
}
