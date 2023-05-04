package cln

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"testing"

	"github.com/lib/pq"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/proto/cln"
	"github.com/lncapital/torq/testutil"
)

type Transaction struct {
	Amount            int64          `db:"amount"`
	TransactionHash   string         `db:"tx_hash"`
	BlockHash         string         `db:"block_hash"`
	RawTransactionHex string         `db:"raw_tx_hex"`
	BlockHeight       uint32         `db:"block_height"`
	DestAddresses     pq.StringArray `db:"dest_addresses"`
	NodeId            int            `db:"node_id"`
}

func TestStoreTransaction(t *testing.T) {

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

	err = settings.InitializeSettingsCache(db)
	if err != nil {
		cancel()
		log.Fatal().Msgf("Problem initializing SettingsCache cache: %v", err)
	}

	err = settings.InitializeNodesCache(db)
	if err != nil {
		cancel()
		log.Fatal().Msgf("Problem initializing NodeCache cache: %v", err)
	}

	err = settings.InitializeChannelsCache(db)
	if err != nil {
		cancel()
		log.Fatal().Err(err).Msgf("Problem initializing ChannelCache cache: %v", err)
	}
	nodeId := cache.GetChannelPeerNodeIdByPublicKey(testutil.TestPublicKey1, core.Bitcoin, core.SigNet)
	nodeSettings := cache.GetNodeSettingsByNodeId(nodeId)

	expected := Transaction{
		Amount:            100000,
		TransactionHash:   "8673221e16aa288e34aacc85b7ce6389dab7467f645fe240470bff8d64c20169",
		RawTransactionHex: "020000000001019e5bfe1ff0c30f1af75ba9ba71003913f2c34959ef508a05294a093e6d7c18de0000000000fdffffff0206252600000000001600146d10885a7e02937060e3c756836a726349957d3ea02526000000000022002044ca56102e34c2c05b266aecb600f907638264356fe7e090e410e571a98f10d6024730440220120dfd75324f4bc3c8c233fa7d079a314f7c9461b5ffd293b4bf861cb39a7cc902201425b8bcb484d00e50ade0d8404c15f5df6bb41278b33a13e1c6be3e352ff26501210397f8946150e488955572b41eaca840777a65a71024f91c4549deaf7d1c789cc383000000",
		BlockHeight:       1,
		DestAddresses: []string{
			"00146d10885a7e02937060e3c756836a726349957d3e",
			"002044ca56102e34c2c05b266aecb600f907638264356fe7e090e410e571a98f10d6",
		},
		NodeId: nodeId,
	}
	transactionHash, err := hex.DecodeString(expected.TransactionHash)
	if err != nil {
		fmt.Println("Unable to convert hex to byte. ", err)
	}
	rawTransactionHex, err := hex.DecodeString(expected.RawTransactionHex)
	if err != nil {
		fmt.Println("Unable to convert hex to byte. ", err)
	}
	destAddresses0, err := hex.DecodeString(expected.DestAddresses[0])
	if err != nil {
		fmt.Println("Unable to convert hex to byte. ", err)
	}
	destAddresses1, err := hex.DecodeString(expected.DestAddresses[1])
	if err != nil {
		fmt.Println("Unable to convert hex to byte. ", err)
	}

	clnTransaction := cln.ListtransactionsTransactions{
		Hash:        transactionHash,
		Rawtx:       rawTransactionHex,
		Blockheight: expected.BlockHeight,
		Txindex:     0,
		Locktime:    0,
		Version:     0,
		Inputs:      nil,
		Outputs: []*cln.ListtransactionsTransactionsOutputs{
			{
				Index:        0,
				AmountMsat:   &cln.Amount{Msat: 60_000},
				ScriptPubKey: destAddresses0,
				ItemType:     nil,
				Channel:      nil,
			},
			{
				Index:        1,
				AmountMsat:   &cln.Amount{Msat: 40_000},
				ScriptPubKey: destAddresses1,
				ItemType:     nil,
				Channel:      nil,
			},
		},
	}
	err = storeTransaction(db, &clnTransaction, nodeSettings)
	if err != nil {
		return
	}

	if err != nil {
		testutil.Fatalf(t, "storeTransaction", err)
	}

	row := db.QueryRowx(`SELECT tx_hash, amount, block_height, dest_addresses, raw_tx_hex, node_id FROM tx LIMIT 1;`)
	if row.Err() != nil {
		testutil.Fatalf(t, "querying tx table", err)
	}

	got := Transaction{}
	err = row.StructScan(&got)
	if err != nil {
		testutil.Fatalf(t, "scanning row", err)
	}

	if !reflect.DeepEqual(got, expected) {
		testutil.Errorf(t, "Got:\n%v\nWant:\n%v\n", got, expected)
	}

}
