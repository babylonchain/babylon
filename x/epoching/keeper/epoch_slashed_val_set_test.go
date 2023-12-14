package keeper_test

import (
	"math/rand"
	"sort"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/babylonchain/babylon/testutil/datagen"
	testhelper "github.com/babylonchain/babylon/testutil/helper"

	"github.com/babylonchain/babylon/x/epoching/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func FuzzSlashedValSet(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		var err error

		helper := testhelper.NewHelperWithValSet(t)
		ctx, keeper, stakingKeeper := helper.Ctx, helper.App.EpochingKeeper, helper.App.StakingKeeper
		getValSet := keeper.GetValidatorSet(ctx, 1)

		// slash a random subset of validators
		numSlashed := r.Intn(len(getValSet))
		excpectedSlashedVals := []sdk.ValAddress{}
		for i := 0; i < numSlashed; i++ {
			idx := r.Intn(len(getValSet))
			slashedVal := getValSet[idx]
			_, err = stakingKeeper.Slash(ctx, slashedVal.Addr, 0, slashedVal.Power, sdkmath.LegacyOneDec())
			require.NoError(t, err)
			// add the slashed validator to the slashed validator set
			excpectedSlashedVals = append(excpectedSlashedVals, slashedVal.Addr)
			// remove the slashed validator from the validator set in order to avoid slashing a validator more than once
			getValSet = append(getValSet[:idx], getValSet[idx+1:]...)
		}

		// check whether the slashed validator set in DB is consistent or not
		actualSlashedVals := keeper.GetSlashedValidators(ctx, 1)
		require.Equal(t, len(excpectedSlashedVals), len(actualSlashedVals))
		sortVals(excpectedSlashedVals)
		actualSlashedVals = types.NewSortedValidatorSet(actualSlashedVals)
		for i := range actualSlashedVals {
			require.Equal(t, excpectedSlashedVals[i], actualSlashedVals[i].GetValAddress())
		}

		// go to epoch 2
		epochInterval := keeper.GetParams(ctx).EpochInterval
		for i := uint64(0); i < epochInterval; i++ {
			ctx, err = helper.GenAndApplyEmptyBlock(r)
			require.NoError(t, err)
		}
		epochNumber := keeper.GetEpoch(ctx).EpochNumber
		require.Equal(t, uint64(2), epochNumber)
		// no validator is slashed in epoch 1
		require.Empty(t, keeper.GetSlashedValidators(ctx, 2))
	})
}

func sortVals(vals []sdk.ValAddress) {
	sort.Slice(vals, func(i, j int) bool {
		return sdk.BigEndianToUint64(vals[i]) < sdk.BigEndianToUint64(vals[j])
	})
}
