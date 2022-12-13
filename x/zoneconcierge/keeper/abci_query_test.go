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
		babylonChain.NextBlock()

		ctx := babylonChain.GetContext()
		val := babylonChain.Vals.Validators[0]

		// NOTE: the queryHeight has to be the previous block because
		// NextBlock() only invokes BeginBlock(), but not EndBlock(), for the new block
		key, value, proof, err := zcKeeper.QueryStore(banktypes.StoreKey, banktypes.CreateAccountBalancesPrefix(val.Address), ctx.BlockHeight()-1)

		require.NoError(t, err)
		require.NotNil(t, proof)
		require.Equal(t, banktypes.CreateAccountBalancesPrefix(val.Address), key)

		err = zckeeper.VerifyStore(ctx.BlockHeader().AppHash, banktypes.StoreKey, key, value, proof)
		require.NoError(t, err)
	})
}
