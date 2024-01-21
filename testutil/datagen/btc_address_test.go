package datagen_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/stretchr/testify/require"
)

func FuzzGenRandomBTCAddress(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		net := &chaincfg.SimNetParams

		addr, err := datagen.GenRandomBTCAddress(r, net)
		require.NoError(t, err)

		// validate the address encoding/decoding
		addr2, err := btcutil.DecodeAddress(addr.EncodeAddress(), net)
		require.NoError(t, err)

		// ensure the address does not change after encoding/decoding
		require.Equal(t, addr.String(), addr2.String())
	})
}
