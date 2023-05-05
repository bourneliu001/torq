package cln

import (
	"context"
	"encoding/hex"
	"reflect"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/services_helpers"
	"github.com/lncapital/torq/proto/cln"
	"github.com/lncapital/torq/testutil"
)

type NodeEvent struct {
	Timestamp     time.Time `db:"timestamp"`
	Alias         string    `db:"alias"`
	Color         string    `db:"color"`
	NodeAddresses string    `db:"node_addresses"`
	EventNodeId   int       `db:"event_node_id"`
	NodeId        int       `db:"node_id"`
}

// stubClnListNodes
type stubClnListNodes struct {
	Nodes             []*cln.ListnodesNodes
	PeerNodePublicKey string
}

func (c *stubClnListNodes) ListNodes(ctx context.Context,
	in *cln.ListnodesRequest,
	opts ...grpc.CallOption) (*cln.ListnodesResponse, error) {

	if c.Nodes == nil {
		return &cln.ListnodesResponse{}, nil
	}

	if in.Id != nil && hex.EncodeToString(in.Id) == c.PeerNodePublicKey {
		r := cln.ListnodesResponse{
			Nodes: c.Nodes,
		}
		return &r, nil
	}

	return &cln.ListnodesResponse{}, nil
}

func TestListNodes(t *testing.T) {

	srv, err := testutil.InitTestDBConn()
	if err != nil {
		panic(err)
	}

	db, cancel, err := srv.NewTestDatabase(true)
	defer cancel()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, nodeSettings := testutil.Setup(db, cancel)

	expectedNodeEvent := getExpectedNodeEvent(nodeSettings)

	clnNode := constructClnNode(expectedNodeEvent)

	mclient := stubClnListNodes{
		Nodes: []*cln.ListnodesNodes{
			&clnNode,
		},
		PeerNodePublicKey: testutil.TestPublicKey2,
	}

	// run it twice it should be smart enough to ignore the duplication
	for i := 0; i < 2; i++ {
		err = listAndProcessNodes(context.Background(), db, &mclient, services_helpers.ClnServiceNodesService,
			nodeSettings, false)
		if err != nil {
			log.Fatal().Err(err).Msgf("Problem in listAndProcessForwards: %v", err)
		}
	}

	var recordCount int
	err = db.QueryRow("select count(*) from node_event;").Scan(&recordCount)
	if err != nil {
		testutil.Fatalf(t, "We get an error: %v", err)
	}
	if recordCount != 1 {
		testutil.Errorf(t, "We expected to store %v records but stored %d", 1, recordCount)
	}

	row := db.QueryRowx(`SELECT timestamp, alias, color, node_addresses, event_node_id, node_id FROM node_event LIMIT 1;`)
	if row.Err() != nil {
		testutil.Fatalf(t, "querying node_event table", err)
	}

	got := NodeEvent{}
	err = row.StructScan(&got)
	if err != nil {
		testutil.Fatalf(t, "scanning row", err)
	}

	if !reflect.DeepEqual(got, expectedNodeEvent) {
		testutil.Errorf(t, "Got:\n%v\nWant:\n%v\n", got, expectedNodeEvent)
	}
}

func getExpectedNodeEvent(nodeSettings cache.NodeSettingsCache) NodeEvent {
	eventNodeId := cache.GetPeerNodeIdByPublicKey(testutil.TestPublicKey2, nodeSettings.Chain, nodeSettings.Network)
	expected := NodeEvent{
		Timestamp:     time.Now().Round(time.Second).UTC(),
		Alias:         "TEST",
		Color:         "c0c0c0",
		NodeAddresses: "[{\"port\": 1234, \"address\": \"localhost\", \"item_type\": 1}]",
		EventNodeId:   eventNodeId,
		NodeId:        nodeSettings.NodeId,
	}
	return expected
}

func constructClnNode(expected NodeEvent) cln.ListnodesNodes {
	eventNodePublicKey := cache.GetNodeSettingsByNodeId(expected.EventNodeId).PublicKey
	unixTimestamp := uint32(expected.Timestamp.Unix())
	localhost := "localhost"
	nodeAddresses := cln.ListnodesNodesAddresses{
		ItemType: cln.ListnodesNodesAddresses_IPV4,
		Port:     1234,
		Address:  &localhost,
	}
	return cln.ListnodesNodes{
		Nodeid:        testutil.HexDecodeString(eventNodePublicKey),
		LastTimestamp: &unixTimestamp,
		Alias:         &expected.Alias,
		Color:         testutil.HexDecodeString(expected.Color),
		Addresses:     []*cln.ListnodesNodesAddresses{&nodeAddresses},
	}
}
