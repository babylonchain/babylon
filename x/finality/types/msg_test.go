package types_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/crypto/eots"
	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/babylonchain/babylon/x/finality/types"
	"github.com/stretchr/testify/require"
)

func FuzzMsgAddFinalitySig(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		sk, err := eots.KeyGen(r)
		require.NoError(t, err)
		msr, mpr, err := eots.NewMasterRandPair(r)
		require.NoError(t, err)
		blockHeight := datagen.RandomInt(r, 10)
		blockHash := datagen.GenRandomByteArray(r, 32)

		sr, _, err := msr.DeriveRandPair(uint32(blockHeight))
		require.NoError(t, err)

		signer := datagen.GenRandomAccount().Address
		msg, err := types.NewMsgAddFinalitySig(signer, sk, sr, blockHeight, blockHash)
		require.NoError(t, err)

		// verify msg's EOTS sig against the given public randomness
		err = msg.VerifyEOTSSig(mpr)
		require.NoError(t, err)
	})
}
