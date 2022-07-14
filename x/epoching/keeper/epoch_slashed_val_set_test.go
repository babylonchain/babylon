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
		t.Skip()

		rand.Seed(seed)

		app, ctx, keeper, _, _, _ := setupTestKeeperWithValSet(t)
		getValSet := keeper.GetValidatorSet(ctx, 0)

		// slash a random validator
		idx := rand.Intn(len(getValSet))
		slashedVal := getValSet[idx]
		app.StakingKeeper.Slash(ctx, sdk.ConsAddress(slashedVal.Addr), 0, slashedVal.Power, sdk.OneDec())

		// go to the 1st block and thus epoch 1
		ctx = genAndApplyEmptyBlock(app, ctx)
		epochNumber := keeper.GetEpoch(ctx).EpochNumber
		require.Equal(t, uint64(1), epochNumber)

		// check whether the validator set remains the same or not
		getValSet2 := keeper.GetValidatorSet(ctx, epochNumber)
		require.Equal(t, len(getValSet)-1, len(getValSet2))
	})
}
