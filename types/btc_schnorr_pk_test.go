package types_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/babylonchain/babylon/types"
	"github.com/stretchr/testify/require"
)

func FuzzBIP340PubKey(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		_, btcPK, err := datagen.GenRandomBTCKeyPair(r)
		require.NoError(t, err)

		// btcPK -> BIP340PubKey -> btcPK
		pk := types.NewBIP340PubKeyFromBTCPK(btcPK)
		btcPK2, err := pk.ToBTCPK()
		require.NoError(t, err)
		// NOTE: we can only ensure they have the same x value.
		// there could be 2 different y values for a given x on secp256k1
		// curve. The BIP340 encoding is compressed in the sense that
		// it only contains x value but not y. pk.ToBTCPK() may choose a random
		// one of the 2 possible y values.
		require.Zero(t, btcPK.X().Cmp(btcPK2.X()))

		// pk -> bytes -> pk
		pkBytes := pk.MustMarshal()
		var pk2 types.BIP340PubKey
		_, err = types.NewBIP340PubKey(pkBytes)
		require.NoError(t, err)
		err = pk2.Unmarshal(pkBytes)
		require.NoError(t, err)
		require.Equal(t, *pk, pk2)
	})
}
