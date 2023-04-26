package keeper_test

import (
	"github.com/babylonchain/babylon/testutil/datagen"
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/x/epoching/testepoching"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func FuzzEpochValSet(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 100)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		helper := testepoching.NewHelperWithValSet(t)
		ctx, keeper := helper.Ctx, helper.EpochingKeeper
		valSet := helper.StakingKeeper.GetLastValidators(helper.Ctx)
		getValSet := keeper.GetValidatorSet(ctx, 0)
		require.Equal(t, len(valSet), len(getValSet))
		for i := range getValSet {
			consAddr, err := valSet[i].GetConsAddr()
			require.NoError(t, err)
			require.Equal(t, sdk.ValAddress(consAddr), getValSet[i].GetValAddress())
		}

		// generate a random number of new blocks
		numIncBlocks := r.Uint64()%1000 + 1
		for i := uint64(0); i < numIncBlocks; i++ {
			ctx = helper.GenAndApplyEmptyBlock(r)
		}

		// check whether the validator set remains the same or not
		getValSet2 := keeper.GetValidatorSet(ctx, keeper.GetEpoch(ctx).EpochNumber)
		require.Equal(t, len(valSet), len(getValSet2))
		for i := range getValSet2 {
			consAddr, err := valSet[i].GetConsAddr()
			require.NoError(t, err)
			require.Equal(t, sdk.ValAddress(consAddr), getValSet[i].GetValAddress())
		}
	})
}

// TODO (stateful tests): create some random validators and check if the resulting validator set is consistent or not (require mocking MsgWrappedCreateValidator)
