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

		_, ctx, keeper, _, _, valSet := setupTestKeeperWithValSet(t)
		getValSet := keeper.GetValidatorSet(ctx, 0)
		require.Equal(t, len(valSet.Validators), len(getValSet))
		for i := range getValSet {
			require.Equal(t, sdk.ValAddress(valSet.Validators[i].Address), getValSet[i].Addr)
		}

		// TODO (stateful tests): randomly add/remove validators, then verify whether the actual validator set is expected or not
	})
}
