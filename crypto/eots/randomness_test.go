package eots_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/crypto/eots"
	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/stretchr/testify/require"
)

func FuzzBIP32RandomnessCodec(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		// master randomness pair
		msr, mpr, err := eots.NewMasterRandPair(r)
		require.NoError(t, err)
		require.NoError(t, msr.Validate())
		require.NoError(t, mpr.Validate())

		// roundtrip of marshaling/unmarshaling msr to/from string
		msrStr := msr.MarshalBase58()
		msr2, err := eots.NewMasterSecretRandFromBase58(msrStr)
		require.NoError(t, err)
		require.Equal(t, msr.Marshal(), msr2.Marshal())

		// roundtrip of marshaling/unmarshaling msr to/from bytes
		msrBytes := msr.Marshal()
		msr2, err = eots.NewMasterSecretRand(msrBytes)
		require.NoError(t, err)
		require.Equal(t, msr.Marshal(), msr2.Marshal())

		// roundtrip of marshaling/unmarshaling mpr to/from string
		mprStr := mpr.MarshalBase58()
		mpr2, err := eots.NewMasterPublicRandFromBase58(mprStr)
		require.NoError(t, err)
		require.Equal(t, mpr.Marshal(), mpr2.Marshal())

		// roundtrip of marshaling/unmarshaling mpr to/from bytes
		mprBytes := mpr.Marshal()
		mpr2, err = eots.NewMasterPublicRand(mprBytes)
		require.NoError(t, err)
		require.Equal(t, mpr.Marshal(), mpr2.Marshal())
	})
}

func FuzzBIP32RandomnessDerivation(f *testing.F) {
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

		// ensure msr can derive mpr
		mpr2, err := msr.MasterPubicRand()
		require.NoError(t, err)
		require.NoError(t, mpr2.Validate())
		require.Equal(t, mpr, mpr2)

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
