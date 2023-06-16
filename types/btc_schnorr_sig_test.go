package types_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/babylonchain/babylon/types"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/stretchr/testify/require"
)

func FuzzBIP340Signature(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		btcSK, _, err := datagen.GenRandomBTCKeyPair(r)
		require.NoError(t, err)

		// sign a random msg
		msgHash := datagen.GenRandomBtcdHash(r)
		btcSig, err := schnorr.Sign(btcSK, msgHash[:])
		require.NoError(t, err)

		// btcSig -> BIP340Signature -> btcSig
		sig := types.NewBIP340SignatureFromBTCSig(btcSig)
		btcSig2, err := sig.ToBTCSig()
		require.NoError(t, err)
		require.True(t, btcSig.IsEqual(btcSig2))

		// BIP340Signature -> bytes -> BIP340Signature
		sigBytes := sig.MustMarshal()
		var sig2 types.BIP340Signature
		_, err = types.NewBIP340Signature(sigBytes)
		require.NoError(t, err)
		err = sig2.Unmarshal(sigBytes)
		require.NoError(t, err)
		require.Equal(t, sig, sig2)
	})
}
