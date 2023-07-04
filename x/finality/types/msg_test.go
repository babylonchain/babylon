package types_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/crypto/eots"
	"github.com/babylonchain/babylon/testutil/datagen"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/finality/types"
	"github.com/stretchr/testify/require"
)

func FuzzMsgAddFinalitySig(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		sk, err := eots.KeyGen(r)
		require.NoError(t, err)
		sr, pr, err := eots.RandGen(r)
		require.NoError(t, err)
		blockHeight := datagen.RandomInt(r, 10)
		blockHash := datagen.GenRandomByteArray(r, 32)

		signer := datagen.GenRandomAccount().Address
		msg, err := types.NewMsgAddFinalitySig(signer, sk, sr, blockHeight, blockHash)
		require.NoError(t, err)

		// basic sanity checks
		err = msg.ValidateBasic()
		require.NoError(t, err)
		// verify msg's EOTS sig against the given public randomness
		err = msg.VerifyEOTSSig(bbn.NewSchnorrPubRandFromFieldVal(pr))
		require.NoError(t, err)
	})
}

func FuzzMsgCommitPubRandList(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		sk, _, err := datagen.GenRandomBTCKeyPair(r)
		require.NoError(t, err)

		startHeight := datagen.RandomInt(r, 10)
		numPubRand := datagen.RandomInt(r, 100) + 1
		_, msg, err := datagen.GenRandomMsgCommitPubRandList(r, sk, startHeight, numPubRand)
		require.NoError(t, err)

		// sanity checks, including verifying signature
		err = msg.ValidateBasic()
		require.NoError(t, err)
	})
}
