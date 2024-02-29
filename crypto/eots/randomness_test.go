package eots_test

import (
	crand "crypto/rand"
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/crypto/eots"
	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/stretchr/testify/require"
)

func FuzzEOTSSignAndVerify(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		sk, err := eots.KeyGen(r)
		require.NoError(t, err)
		pk := eots.PubGen(sk)

		sr, pr, err := eots.RandGen(crand.Reader)
		require.NoError(t, err)

		msg := datagen.GenRandomByteArray(r, 100)
		sig, err := eots.Sign(sk, sr, msg)
		require.NoError(t, err)

		err = eots.Verify(pk, pr, msg, sig)
		require.NoError(t, err)
	})
}

func FuzzBTCHeightIndex(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		// EOTS key pair
		sk, err := eots.KeyGen(r)
		require.NoError(t, err)
		pk := eots.PubGen(sk)

		// master randomness pair
		msr, mpr, err := eots.NewMasterRandPair(r)
		require.NoError(t, err)
		require.NoError(t, msr.Validate())
		require.NoError(t, mpr.Validate())

		height := uint32(datagen.RandomInt(r, 10000))

		// derive pair of randomness via master secret randomness at this height
		sr, pr, err := msr.DeriveRandPair(height)
		require.NoError(t, err)

		// derive public randomness via master public randomness at this height
		pr2, err := mpr.DerivePubRand(height)
		require.NoError(t, err)

		// assert consistency of public randomness
		require.Equal(t, pr, pr2)

		// sign EOTS using secret randomness
		msg := datagen.GenRandomByteArray(r, 100)
		sig, err := eots.Sign(sk, sr, msg)
		require.NoError(t, err)

		// verify EOTS sig using public key
		err = eots.Verify(pk, pr, msg, sig)
		require.NoError(t, err)
	})
}
