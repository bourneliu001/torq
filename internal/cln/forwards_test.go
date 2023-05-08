package cln

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/internal/services_helpers"
	"github.com/lncapital/torq/proto/cln"
	"github.com/lncapital/torq/testutil"
)

type Forward struct {
	Time               time.Time `db:"time"`
	FeeMsat            uint64    `db:"fee_msat"`
	IncomingAmountMsat uint64    `db:"incoming_amount_msat"`
	IncomingChannelId  int       `db:"incoming_channel_id"`
	OutgoingAmountMsat uint64    `db:"outgoing_amount_msat"`
	OutgoingChannelId  int       `db:"outgoing_channel_id"`
	NodeId             int       `db:"node_id"`
}

type Htlc struct {
	Time               time.Time     `db:"time"`
	IncomingAmountMsat uint64        `db:"incoming_amt_msat"`
	IncomingChannelId  int           `db:"incoming_channel_id"`
	IncomingHtlcId     uint64        `db:"incoming_htlc_id"`
	OutgoingAmountMsat uint64        `db:"outgoing_amt_msat"`
	OutgoingChannelId  int           `db:"outgoing_channel_id"`
	OutgoingHtlcId     uint64        `db:"outgoing_htlc_id"`
	ClnForwardStatus   forwardStatus `db:"cln_forward_status_id"`
	NodeId             int           `db:"node_id"`
}

// stubClnListForwards
type stubClnListForwards struct {
	Forwards []*cln.ListforwardsForwards
	Status   cln.ListforwardsRequest_ListforwardsStatus
}

func (c *stubClnListForwards) ListForwards(ctx context.Context,
	in *cln.ListforwardsRequest,
	opts ...grpc.CallOption) (*cln.ListforwardsResponse, error) {

	if c.Forwards == nil {
		return &cln.ListforwardsResponse{}, nil
	}

	if in.Status != nil && *in.Status != c.Status {
		return &cln.ListforwardsResponse{}, nil
	}

	r := cln.ListforwardsResponse{
		Forwards: c.Forwards,
	}
	return &r, nil
}

func TestListForwards(t *testing.T) {

	srv, err := testutil.InitTestDBConn()
	if err != nil {
		panic(err)
	}

	db, cancel, err := srv.NewTestDatabase()
	defer cancel()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	nodeId, nodeSettings := testutil.Setup(db, cancel)

	expectedForward := getExpectedForward(nodeId)

	clnForward := constructClnForward(expectedForward)

	mclient := stubClnListForwards{
		Forwards: []*cln.ListforwardsForwards{
			&clnForward,
		},
		Status: cln.ListforwardsRequest_SETTLED,
	}

	// run it twice it should be smart enough to ignore the duplication
	for i := 0; i < 2; i++ {
		err = listAndProcessForwards(context.Background(), db, &mclient, services_helpers.ClnServiceForwardsService,
			nodeSettings, false)
		if err != nil {
			log.Fatal().Err(err).Msgf("Problem in listAndProcessForwards: %v", err)
		}
	}

	var recordCount int
	err = db.QueryRow("select count(*) from forward;").Scan(&recordCount)
	if err != nil {
		testutil.Fatalf(t, "We get an error: %v", err)
	}
	if recordCount == 0 {
		testutil.Errorf(t, "We expected to store records but stored %d", recordCount)
	}

	ncd := cache.GetNodeConnectionDetails(nodeSettings.NodeId)
	ncd.CustomSettings = ncd.CustomSettings.AddNodeConnectionDetailCustomSettings(core.ImportHtlcEvents)
	cache.SetNodeConnectionDetails(nodeSettings.NodeId, ncd)
	expectedHtlc := getExpectedHtlc(nodeId)

	clnHtlc := constructClnHtlc(expectedHtlc)

	mclient = stubClnListForwards{
		Forwards: []*cln.ListforwardsForwards{
			&clnHtlc,
		},
		Status: cln.ListforwardsRequest_FAILED,
	}

	// run it twice it should be smart enough to ignore the duplication
	for i := 0; i < 2; i++ {
		err = listAndProcessForwards(context.Background(), db, &mclient, services_helpers.ClnServiceForwardsService,
			nodeSettings, false)
		if err != nil {
			log.Fatal().Err(err).Msgf("Problem in listAndProcessForwards: %v", err)
		}
	}

	err = db.QueryRow("select count(*) from htlc_event;").Scan(&recordCount)
	if err != nil {
		testutil.Fatalf(t, "We get an error: %v", err)
	}
	if recordCount == 0 {
		testutil.Errorf(t, "We expected to store records but stored %d", recordCount)
	}

}

func TestStoreForward(t *testing.T) {

	srv, err := testutil.InitTestDBConn()
	if err != nil {
		panic(err)
	}

	db, cancel, err := srv.NewTestDatabase()
	defer cancel()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	nodeId, nodeSettings := testutil.Setup(db, cancel)

	expected := getExpectedForward(nodeId)

	clnForward := constructClnForward(expected)

	err = storeForward(db, &clnForward, nodeSettings)
	if err != nil {
		testutil.Fatalf(t, "storeForward", err)
	}

	row := db.QueryRowx(`
		SELECT time, fee_msat,
		       incoming_amount_msat, incoming_channel_id,
		       outgoing_amount_msat, outgoing_channel_id, node_id
		FROM forward LIMIT 1;`)
	if row.Err() != nil {
		testutil.Fatalf(t, "querying forward table", err)
	}

	got := Forward{}
	err = row.StructScan(&got)
	if err != nil {
		testutil.Fatalf(t, "scanning row", err)
	}

	if !reflect.DeepEqual(got, expected) {
		testutil.Errorf(t, "Got:\n%v\nWant:\n%v\n", got, expected)
	}
}

func TestStoreHtlc(t *testing.T) {

	srv, err := testutil.InitTestDBConn()
	if err != nil {
		panic(err)
	}

	db, cancel, err := srv.NewTestDatabase()
	defer cancel()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	nodeId, nodeSettings := testutil.Setup(db, cancel)

	expected := getExpectedHtlc(nodeId)

	clnHtlc := constructClnHtlc(expected)

	err = storeHtlc(db, remoteFailed, &clnHtlc, nodeSettings)
	if err != nil {
		testutil.Fatalf(t, "storeForward", err)
	}

	row := db.QueryRowx(`
		SELECT time,
		       incoming_channel_id, incoming_amt_msat, incoming_htlc_id,
			   outgoing_channel_id, outgoing_amt_msat, outgoing_htlc_id,
			   node_id, cln_forward_status_id
		FROM htlc_event LIMIT 1;`)
	if row.Err() != nil {
		testutil.Fatalf(t, "querying htlc_event table", err)
	}

	got := Htlc{}
	err = row.StructScan(&got)
	if err != nil {
		testutil.Fatalf(t, "scanning row", err)
	}

	if !reflect.DeepEqual(got, expected) {
		testutil.Errorf(t, "Got:\n%v\nWant:\n%v\n", got, expected)
	}
}

func getExpectedForward(nodeId int) Forward {
	inChannelId := cache.GetChannelIdByChannelPoint(testutil.TestChannelPoint1)
	outChannelId := cache.GetChannelIdByChannelPoint(testutil.TestChannelPoint2)
	expected := Forward{
		Time:               time.Now().Round(time.Second).UTC(),
		FeeMsat:            1_000,
		IncomingAmountMsat: 1_000_000,
		IncomingChannelId:  inChannelId,
		OutgoingAmountMsat: 999_000,
		OutgoingChannelId:  outChannelId,
		NodeId:             nodeId,
	}
	return expected
}

func getExpectedHtlc(nodeId int) Htlc {
	inChannelId := cache.GetChannelIdByChannelPoint(testutil.TestChannelPoint1)
	outChannelId := cache.GetChannelIdByChannelPoint(testutil.TestChannelPoint2)
	expected := Htlc{
		Time:               time.Now().Round(time.Second).UTC(),
		IncomingAmountMsat: 1_000_000,
		IncomingChannelId:  inChannelId,
		IncomingHtlcId:     12345,
		OutgoingAmountMsat: 999_000,
		OutgoingChannelId:  outChannelId,
		OutgoingHtlcId:     67890,
		ClnForwardStatus:   remoteFailed,
		NodeId:             nodeId,
	}
	return expected
}

func constructClnForward(expected Forward) cln.ListforwardsForwards {
	feeMsat := cln.Amount{Msat: expected.FeeMsat}
	incomingAmountMsat := cln.Amount{Msat: expected.IncomingAmountMsat}
	outgoingAmountMsat := cln.Amount{Msat: expected.OutgoingAmountMsat}
	inShortChannelId := cache.GetChannelSettingByChannelId(expected.IncomingChannelId).ShortChannelId
	outShortChannelId := cache.GetChannelSettingByChannelId(expected.OutgoingChannelId).ShortChannelId
	return cln.ListforwardsForwards{
		InChannel:    *inShortChannelId,
		InMsat:       &incomingAmountMsat,
		Status:       cln.ListforwardsForwards_SETTLED,
		ReceivedTime: float64(expected.Time.Unix()),
		OutChannel:   outShortChannelId,
		Style:        nil,
		FeeMsat:      &feeMsat,
		OutMsat:      &outgoingAmountMsat,
	}
}

func constructClnHtlc(expected Htlc) cln.ListforwardsForwards {
	incomingAmountMsat := cln.Amount{Msat: expected.IncomingAmountMsat}
	outgoingAmountMsat := cln.Amount{Msat: expected.OutgoingAmountMsat}
	inShortChannelId := cache.GetChannelSettingByChannelId(expected.IncomingChannelId).ShortChannelId
	outShortChannelId := cache.GetChannelSettingByChannelId(expected.OutgoingChannelId).ShortChannelId
	return cln.ListforwardsForwards{
		InChannel:    *inShortChannelId,
		InHtlcId:     &expected.IncomingHtlcId,
		InMsat:       &incomingAmountMsat,
		Status:       cln.ListforwardsForwards_FAILED,
		ReceivedTime: float64(expected.Time.Unix()),
		OutChannel:   outShortChannelId,
		OutHtlcId:    &expected.OutgoingHtlcId,
		OutMsat:      &outgoingAmountMsat,
	}
}
