package testutil

import (
	"context"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/internal/settings"
)

const colorReset = "\033[0m"
const colorRed = "\033[31m"
const colorGreen = "\033[32m"
const colorCyan = "\033[36m"
const succeed = "\u2713"
const failed = "\u2717"

func Given(t *testing.T, txt string) {
	t.Logf("%s %s%s", colorCyan, txt, colorReset)
}

func GivenF(t *testing.T, txt string, a ...interface{}) {
	t.Logf(fmt.Sprintf(colorCyan+txt+colorReset, a))
}

func WhenF(t *testing.T, txt string, a ...interface{}) {
	t.Logf(fmt.Sprintf(colorCyan+"  "+txt+colorReset, a...))
}

func Successf(t *testing.T, txt string, a ...interface{}) {
	t.Logf(fmt.Sprintf(colorGreen+"  "+succeed+"  "+txt+colorReset, a...))
}

func Errorf(t *testing.T, txt string, a ...interface{}) {
	t.Errorf(fmt.Sprintf(colorRed+"  "+failed+"  "+txt+colorReset, a...))
}

func Fatalf(t *testing.T, txt string, a ...interface{}) {
	t.Fatalf(fmt.Sprintf(colorRed+"  "+failed+"  "+txt+colorReset, a...))
}

func HexDecodeString(s string) []byte {
	ba, err := hex.DecodeString(s)
	if err != nil {
		log.Fatal().Msgf("Unable to convert hex to byte. (%v)", err)
	}
	return ba
}

func Setup(db *sqlx.DB, cancel context.CancelFunc) (int, cache.NodeSettingsCache) {
	err := settings.InitializeSettingsCache(db)
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
	nodeId := cache.GetChannelPeerNodeIdByPublicKey(TestPublicKey1, core.Bitcoin, core.SigNet)
	nodeSettings := cache.GetNodeSettingsByNodeId(nodeId)
	return nodeId, nodeSettings
}
