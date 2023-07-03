package types_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/crypto/eots"
	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/stretchr/testify/require"
)

func FuzzMsgAddVote(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		sk, err := eots.KeyGen(r)
		require.NoError(t, err)

		msg, pr, err := datagen.GenRandomMsgAddVote(r, sk)
		require.NoError(t, err)

		// basic sanity checks
		err = msg.ValidateBasic()
		require.NoError(t, err)
		// verify msg's EOTS sig against the given public randomness
		err = msg.VerifyEOTSSig(pr)
		require.NoError(t, err)
	})
}

func FuzzMsgCommitPubRand(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		sk, _, err := datagen.GenRandomBTCKeyPair(r)
		require.NoError(t, err)

		msg, err := datagen.GenRandomMsgCommitPubRand(r, sk)
		require.NoError(t, err)

		// sanity checks, including verifying signature
		err = msg.ValidateBasic()
		require.NoError(t, err)
	})
}
