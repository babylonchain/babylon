package keeper_test

import (
	"math/rand"
	"testing"

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

		app, ctx, keeper, _, _, valSet := setupTestKeeperWithValSet(t)
		getValSet := keeper.GetValidatorSet(ctx, 0)
		require.Equal(t, len(valSet.Validators), len(getValSet))
		for i := range getValSet {
			require.Equal(t, sdk.ValAddress(valSet.Validators[i].Address), getValSet[i].Addr)
		}

		// generate a random number of new blocks
		numIncBlocks := rand.Uint64()%1000 + 1
		for i := uint64(0); i < numIncBlocks; i++ {
			ctx = nextBlock(app, ctx)
		}

		// check whether the validator set remains the same or not
		getValSet2 := keeper.GetValidatorSet(ctx, keeper.GetEpoch(ctx).EpochNumber)
		require.Equal(t, len(valSet.Validators), len(getValSet2))
		for i := range getValSet2 {
			require.Equal(t, sdk.ValAddress(valSet.Validators[i].Address), getValSet[i].Addr)
		}
	})
}
