package cln

import (
	"context"
	"reflect"
	"testing"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/internal/services_helpers"
	"github.com/lncapital/torq/proto/cln"
	"github.com/lncapital/torq/testutil"
)

type Peer struct {
	Address          *string                     `db:"address"`
	ConnectionStatus core.NodeConnectionStatus   `db:"connection_status"`
	Settings         *core.NodeConnectionSetting `db:"setting"`
	PeerNodeId       int                         `db:"node_id"`
	TorqNodeId       int                         `db:"torq_node_id"`
}

// stubClnListPeers
type stubClnListPeers struct {
	Peers []*cln.ListpeersPeers
}

func (c *stubClnListPeers) ListPeers(ctx context.Context,
	in *cln.ListpeersRequest,
	opts ...grpc.CallOption) (*cln.ListpeersResponse, error) {

	if c.Peers == nil {
		return &cln.ListpeersResponse{}, nil
	}

	r := cln.ListpeersResponse{
		Peers: c.Peers,
	}

	return &r, nil
}

func TestListPeers(t *testing.T) {

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

	nodeId, nodeSettings := testutil.Setup(db, cancel)

	expected := getExpectedPeer(nodeId)

	clnPeer := constructClnPeer(expected)

	mclient := stubClnListPeers{
		Peers: []*cln.ListpeersPeers{
			&clnPeer,
		},
	}

	// run it twice it should be smart enough to ignore the duplication
	for i := 0; i < 2; i++ {
		err = listAndProcessPeers(context.Background(), db, &mclient, services_helpers.ClnServicePeersService,
			nodeSettings, false)
		if err != nil {
			log.Fatal().Err(err).Msgf("Problem in listAndProcessTransactions: %v", err)
		}
	}

	var recordCount int
	err = db.QueryRow("select count(*) from node_connection_history;").Scan(&recordCount)
	if err != nil {
		testutil.Fatalf(t, "We get an error: %v", err)
	}
	if recordCount != 1 {
		testutil.Errorf(t, "We expected to store %d records but stored %d", 1, recordCount)
	}

	row := db.QueryRowx(`SELECT address, connection_status, setting, node_id, torq_node_id FROM node_connection_history LIMIT 1;`)
	if row.Err() != nil {
		testutil.Fatalf(t, "querying node_connection_history table", err)
	}

	got := Peer{}
	err = row.StructScan(&got)
	if err != nil {
		testutil.Fatalf(t, "scanning row", err)
	}

	if !reflect.DeepEqual(got, expected) {
		testutil.Errorf(t, "Got:\n%v\nWant:\n%v\n", got, expected)
	}

	// All peers got disconnected
	mclient = stubClnListPeers{
		Peers: []*cln.ListpeersPeers{},
	}

	// run it twice it should be smart enough to ignore the duplication
	for i := 0; i < 2; i++ {
		err = listAndProcessPeers(context.Background(), db, &mclient, services_helpers.ClnServicePeersService,
			nodeSettings, false)
		if err != nil {
			log.Fatal().Err(err).Msgf("Problem in listAndProcessTransactions: %v", err)
		}
	}

	err = db.QueryRow("select count(*) from node_connection_history;").Scan(&recordCount)
	if err != nil {
		testutil.Fatalf(t, "We get an error: %v", err)
	}
	if recordCount != 2 {
		testutil.Errorf(t, "We expected to store %d records but stored %d", 2, recordCount)
	}
}

func getExpectedPeer(nodeId int) Peer {
	dummyAddress := "127.0.0.1:1234"
	peerNodeId := cache.GetChannelPeerNodeIdByPublicKey(testutil.TestPublicKey2, core.Bitcoin, core.SigNet)
	expected := Peer{
		TorqNodeId:       nodeId,
		PeerNodeId:       peerNodeId,
		Address:          &dummyAddress,
		ConnectionStatus: core.NodeConnectionStatusConnected,
	}
	return expected
}

func constructClnPeer(expected Peer) cln.ListpeersPeers {
	return cln.ListpeersPeers{
		Id:         testutil.HexDecodeString(cache.GetNodeSettingsByNodeId(expected.PeerNodeId).PublicKey),
		Connected:  expected.ConnectionStatus == core.NodeConnectionStatusConnected,
		RemoteAddr: expected.Address,
	}
}
