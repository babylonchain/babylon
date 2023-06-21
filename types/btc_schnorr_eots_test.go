package types_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/babylonchain/babylon/types"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/stretchr/testify/require"
)

func FuzzSchnorrEOTSSig(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		randBytes := datagen.GenRandomByteArray(r, 32)
		var modNScalar btcec.ModNScalar
		modNScalar.PutBytesUnchecked(randBytes)

		// ModNScalar -> SchnorrEOTSSig -> ModNScalar
		sig := types.NewSchnorrEOTSSigFromModNScalar(&modNScalar)
		modNScalar2 := sig.ToModNScalar()
		require.True(t, modNScalar.Equals(modNScalar2))

		// SchnorrEOTSSig -> bytes -> SchnorrEOTSSig
		randBytes2 := sig.MustMarshal()
		sig2, err := types.NewSchnorrEOTSSig(randBytes)
		require.NoError(t, err)
		require.Equal(t, randBytes, randBytes2)
		require.Equal(t, sig, sig2)
	})
}
