package keeper_test

import (
	"math/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// TODO (stateful tests): slash some random validators and check if the resulting (slashed) validator sets are consistent or not
// require mocking slashing

func FuzzSlashedValSet(f *testing.F) {
	f.Add(int64(11111))
	f.Add(int64(22222))
	f.Add(int64(55555))
	f.Add(int64(12312))

	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)

		app, ctx, keeper, _, _, _ := setupTestKeeperWithValSet(t)
		getValSet := keeper.GetValidatorSet(ctx, 0)

		// slash a random set of validators
		numSlashed := rand.Intn(len(getValSet))
		for i := 0; i < numSlashed; i++ {
			idx := rand.Intn(len(getValSet))
			slashedVal := getValSet[idx]
			app.StakingKeeper.Slash(ctx, sdk.ConsAddress(slashedVal.Addr), 0, slashedVal.Power, sdk.OneDec())
			// remove the slashed validator from the validator set
			getValSet = append(getValSet[:idx], getValSet[idx+1:]...)
		}

		// go to the 1st block and thus epoch 1
		ctx = genAndApplyEmptyBlock(app, ctx)
		epochNumber := keeper.GetEpoch(ctx).EpochNumber
		require.Equal(t, uint64(1), epochNumber)

		// check whether the validator set has excluded the slashed ones or not
		getValSet2 := keeper.GetValidatorSet(ctx, epochNumber)
		require.Equal(t, len(getValSet), len(getValSet2))
		for i := range getValSet2 {
			require.Equal(t, sdk.ValAddress(getValSet[i].Addr), getValSet[i].Addr)
		}
	})
}
