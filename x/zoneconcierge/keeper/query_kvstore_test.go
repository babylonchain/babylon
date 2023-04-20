package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"
	zckeeper "github.com/babylonchain/babylon/x/zoneconcierge/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func FuzzQueryStore(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)

		_, babylonChain, _, babylonApp := SetupTest(t)
		zcKeeper := babylonApp.ZoneConciergeKeeper

		babylonChain.NextBlock()
		babylonChain.NextBlock()

		ctx := babylonChain.GetContext()

		epochQueryKey := append(epochingtypes.EpochInfoKey, sdk.Uint64ToBigEndian(1)...)
		// NOTE: QueryStore will use ctx.BlockHeight() - 1 as the height for ABCI query
		// This is because the ABCI query will return the inclusion proof corresponding
		// to the NEXT block rather than the given block.
		key, value, proof, err := zcKeeper.QueryStore(epochingtypes.StoreKey, epochQueryKey, ctx.BlockHeight())

		require.NoError(t, err)
		require.NotNil(t, proof)
		require.Equal(t, epochQueryKey, key)

		err = zckeeper.VerifyStore(ctx.BlockHeader().AppHash, epochingtypes.StoreKey, key, value, proof)
		require.NoError(t, err)
	})
}
