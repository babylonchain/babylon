package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/babylonchain/babylon/x/epoching/keeper"
	"github.com/babylonchain/babylon/x/epoching/testepoching"
	"github.com/babylonchain/babylon/x/epoching/types"
	"github.com/stretchr/testify/require"
)

func FuzzAppHashChain(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)

		helper := testepoching.NewHelper(t)
		ctx, k := helper.Ctx, helper.EpochingKeeper
		// ensure that the epoch info is correct at the genesis
		epoch := k.GetEpoch(ctx)
		require.Equal(t, epoch.EpochNumber, uint64(0))
		require.Equal(t, epoch.FirstBlockHeight, uint64(0))

		// set a random epoch interval
		epochInterval := rand.Uint64()%100 + 2 // the epoch interval should at at least 2
		k.SetParams(ctx, types.Params{
			EpochInterval: epochInterval,
		})

		// reach the end of the 1st epoch
		expectedHeight := epochInterval
		expectedAppHashs := [][]byte{}
		for i := uint64(0); i < expectedHeight; i++ {
			ctx = helper.GenAndApplyEmptyBlock()
			expectedAppHashs = append(expectedAppHashs, ctx.BlockHeader().AppHash)
		}
		// ensure epoch number is 1
		epoch = k.GetEpoch(ctx)
		require.Equal(t, uint64(1), epoch.EpochNumber)

		// ensure appHashs are same as expectedAppHashs
		appHashs, err := k.GetAllAppHashsForEpoch(ctx, epoch)
		require.NoError(t, err)
		require.Equal(t, expectedAppHashs, appHashs)

		// ensure prover and verifier are correct
		randomHeightInEpoch := uint64(rand.Intn(int(expectedHeight)) + 1)
		randomAppHash, err := k.GetAppHash(ctx, randomHeightInEpoch)
		require.NoError(t, err)
		proof, err := k.ProveAppHashInEpoch(ctx, randomHeightInEpoch, epoch.EpochNumber)
		require.NoError(t, err)
		err = keeper.VerifyAppHashInclusion(randomAppHash, epoch.AppHashRoot, proof)
		require.NoError(t, err)
	})
}
