package lnd

import (
	"reflect"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/proto/lnrpc"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/testutil"
)

type Transaction struct {
	Amount        int64          `db:"amount"`
	BlockHash     string         `db:"block_hash"`
	BlockHeight   uint32         `db:"block_height"`
	TotalFees     int64          `db:"total_fees"`
	DestAddresses pq.StringArray `db:"dest_addresses"`
	Label         string         `db:"label"`
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

	expected := Transaction{
		Amount:      100000,
		BlockHash:   "0000000000000000000000000000000000000000000000000000000000000000",
		BlockHeight: 1,
		TotalFees:   1000,
		DestAddresses: []string{
			"sb1q3e0rpuq04nknd9gzd7kfp5tqqfuvmxd3v9aaax",
			"sb1qzfw8yz3ays09rztc9vcvpey2l2tzf2kefclmap",
		},
	}

	_, err = storeTransaction(db, &lnrpc.Transaction{
		TxHash:      "test",
		Amount:      expected.Amount,
		BlockHash:   expected.BlockHash,
		BlockHeight: int32(expected.BlockHeight),
		TimeStamp:   time.Now().Unix(),
		TotalFees:   expected.TotalFees,
		OutputDetails: []*lnrpc.OutputDetail{
			{
				OutputType:   lnrpc.OutputScriptType(1),
				Address:      expected.DestAddresses[0],
				PkScript:     "testScript1",
				OutputIndex:  0,
				Amount:       60,
				IsOurAddress: false,
			},
			{
				OutputType:   lnrpc.OutputScriptType(1),
				Address:      expected.DestAddresses[1],
				PkScript:     "testScript2",
				OutputIndex:  1,
				Amount:       40,
				IsOurAddress: true,
			},
		},
		RawTxHex:          "",
		Label:             expected.Label,
		PreviousOutpoints: nil,
	}, cache.GetChannelPeerNodeIdByPublicKey(testutil.TestPublicKey1, core.Bitcoin, core.SigNet))
	if err != nil {
		return
	}

	if err != nil {
		testutil.Fatalf(t, "storeTransaction", err)
	}

	row := db.QueryRowx(`select
		amount,
		block_hash,
		block_height,
		total_fees,
		dest_addresses,
		label
	from tx LIMIT 1;`)

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
