package keeper_test

import (
	"math/rand"
	"sort"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func FuzzSlashedValSet(f *testing.F) {
	f.Add(int64(11111))
	f.Add(int64(22222))
	f.Add(int64(55555))
	f.Add(int64(12312))

	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)

		app, ctx, keeper, _, _, _ := SetupTestKeeperWithValSet(t)
		getValSet := keeper.GetValidatorSet(ctx, 0)

		// slash a random subset of validators
		numSlashed := rand.Intn(len(getValSet))
		excpectedSlashedVals := []sdk.ValAddress{}
		for i := 0; i < numSlashed; i++ {
			idx := rand.Intn(len(getValSet))
			slashedVal := getValSet[idx]
			app.StakingKeeper.Slash(ctx, sdk.ConsAddress(slashedVal.Addr), 0, slashedVal.Power, sdk.OneDec())
			// add the slashed validator to the slashed validator set
			excpectedSlashedVals = append(excpectedSlashedVals, slashedVal.Addr)
			// remove the slashed validator from the validator set in order to avoid slashing a validator more than once
			getValSet = append(getValSet[:idx], getValSet[idx+1:]...)
		}

		// check whether the slashed validator set in DB is consistent or not
		actualSlashedVals := keeper.GetSlashedValidators(ctx, 0)
		require.Equal(t, len(excpectedSlashedVals), len(actualSlashedVals))
		sortVals(excpectedSlashedVals)
		sortVals(actualSlashedVals)
		for i := range actualSlashedVals {
			require.Equal(t, excpectedSlashedVals[i], actualSlashedVals[i])
		}

		// go to the 1st block and thus epoch 1
		ctx = genAndApplyEmptyBlock(app, ctx)
		epochNumber := keeper.GetEpoch(ctx).EpochNumber
		require.Equal(t, uint64(1), epochNumber)
		// no validator is slashed in epoch 1
		require.Empty(t, keeper.GetSlashedValidators(ctx, 1))
	})
}

func sortVals(vals []sdk.ValAddress) {
	sort.Slice(vals, func(i, j int) bool {
		return sdk.BigEndianToUint64(vals[i]) < sdk.BigEndianToUint64(vals[j])
	})
}
