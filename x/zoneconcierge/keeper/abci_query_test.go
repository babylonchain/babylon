package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	zckeeper "github.com/babylonchain/babylon/x/zoneconcierge/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"
)

func FuzzABCIQuery(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)

		_, babylonChain, _, zcKeeper := SetupTest(t)
		ctx := babylonChain.GetContext()
		val := babylonChain.Vals.Validators[0]

		babylonChain.NextBlock()

		key, value, proof, err := zcKeeper.QueryStore(ctx, banktypes.StoreKey, banktypes.CreateAccountBalancesPrefix(val.Address), ctx.BlockHeight())

		require.NoError(t, err)
		require.Equal(t, banktypes.CreateAccountBalancesPrefix(val.Address), key)

		err = zckeeper.VerifyStore(ctx.BlockHeader().AppHash, key, value, proof)
		require.NoError(t, err)
	})
}
