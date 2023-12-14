package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	testhelper "github.com/babylonchain/babylon/testutil/helper"
	"github.com/babylonchain/babylon/x/epoching/keeper"
	"github.com/stretchr/testify/require"
)

func FuzzAppHashChain(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		var err error

		helper := testhelper.NewHelper(t)
		ctx, k := helper.Ctx, helper.App.EpochingKeeper
		// ensure that the epoch info is correct at the genesis
		epoch := k.GetEpoch(ctx)
		require.Equal(t, epoch.EpochNumber, uint64(1))
		require.Equal(t, epoch.FirstBlockHeight, uint64(1))

		// set a random epoch interval
		epochInterval := k.GetParams(ctx).EpochInterval

		// reach the end of the 1st epoch
		expectedHeight := epochInterval - 1
		for i := uint64(0); i < expectedHeight; i++ {
			ctx, err = helper.GenAndApplyEmptyBlock(r)
			require.NoError(t, err)
		}
		// ensure epoch number is 1
		epoch = k.GetEpoch(ctx)
		require.Equal(t, uint64(1), epoch.EpochNumber)

		// ensure prover and verifier are correct
		randomHeightInEpoch := uint64(r.Intn(int(expectedHeight)) + 1)
		randomAppHash, err := k.GetAppHash(ctx, randomHeightInEpoch)
		require.NoError(t, err)
		proof, err := k.ProveAppHashInEpoch(ctx, randomHeightInEpoch, epoch.EpochNumber)
		require.NoError(t, err)
		err = keeper.VerifyAppHashInclusion(randomAppHash, epoch.AppHashRoot, proof)
		require.NoError(t, err)
	})
}
