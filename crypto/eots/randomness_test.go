package eots_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/crypto/eots"
	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/stretchr/testify/require"
)

func FuzzBTCHeightIndex(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		msr, mpr, err := eots.NewMasterRandPair(r)
		require.NoError(t, err)
		require.NoError(t, msr.Validate())
		require.NoError(t, mpr.Validate())

		height := uint32(10000)

		// derive pair of randomness via master secret randomness
		sr, pr, err := msr.DeriveRandPair(uint32(height))
		require.NoError(t, err)

		// derive public randomness via master public randomness
		pr2, err := mpr.DerivePubRand(height)
		require.NoError(t, err)

		// assert consistency of public randomness
		require.Equal(t, pr, pr2)

		// sign EOTS using secret randomness
		sk, err := eots.KeyGen(r)
		pk := sk.PubKey()
		require.NoError(t, err)
		msg := []byte("hello world")
		sig, err := eots.Sign(sk, sr, msg)
		require.NoError(t, err)

		// verify EOTS sig using public key
		err = eots.Verify(pk, pr2, msg, sig)
		require.NoError(t, err)
	})
}
