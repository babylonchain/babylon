package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/x/epoching/testepoching"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func FuzzEpochValSet(f *testing.F) {
	f.Add(int64(11111))
	f.Add(int64(22222))
	f.Add(int64(55555))
	f.Add(int64(12312))

	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)

		helper := testepoching.NewHelperWithValSet(t)
		ctx, keeper := helper.Ctx, helper.EpochingKeeper
		valSet := helper.StakingKeeper.GetLastValidators(helper.Ctx)
		getValSet := keeper.GetValidatorSet(ctx, 0)
		require.Equal(t, len(valSet), len(getValSet))
		for i := range getValSet {
			consAddr, err := valSet[i].GetConsAddr()
			require.NoError(t, err)
			require.Equal(t, sdk.ValAddress(consAddr), getValSet[i].Addr)
		}

		// generate a random number of new blocks
		numIncBlocks := rand.Uint64()%1000 + 1
		for i := uint64(0); i < numIncBlocks; i++ {
			ctx = helper.GenAndApplyEmptyBlock()
		}

		// check whether the validator set remains the same or not
		getValSet2 := keeper.GetValidatorSet(ctx, keeper.GetEpoch(ctx).EpochNumber)
		require.Equal(t, len(valSet), len(getValSet2))
		for i := range getValSet2 {
			consAddr, err := valSet[i].GetConsAddr()
			require.NoError(t, err)
			require.Equal(t, sdk.ValAddress(consAddr), getValSet[i].Addr)
		}
	})
}

// TODO (stateful tests): create some random validators and check if the resulting validator set is consistent or not (require mocking MsgWrappedCreateValidator)
