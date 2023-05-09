package lnd

import (
	"testing"

	"github.com/btcsuite/btcd/chaincfg"
)

//func TestFetchLastInvoiceIndexes(t *testing.T) {
//	srv, err := testutil.InitTestDBConn()
//	if err != nil {
//		panic(err)
//	}
//
//	db, err := srv.NewTestDatabase()
//	if err != nil {
//		t.Fatal(err)
//	}
//
//}

func TestGetNodeNetwork(t *testing.T) {
	got := chaincfg.MainNetParams.Bech32HRPSegwit
	want := "bc"

	if got == want {
		t.Logf("Passed")
	} else {
		t.Fatalf("Failed")
	}
}
