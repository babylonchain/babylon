package keeper_test

import (
	"context"
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	keepertest "github.com/babylonchain/babylon/testutil/keeper"
	"github.com/babylonchain/babylon/x/btcstaking/keeper"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func setupMsgServer(t testing.TB) (*keeper.Keeper, types.MsgServer, context.Context) {
	k, ctx := keepertest.BTCStakingKeeper(t)
	return k, keeper.NewMsgServerImpl(*k), sdk.WrapSDKContext(ctx)
}

func TestMsgServer(t *testing.T) {
	_, ms, ctx := setupMsgServer(t)
	require.NotNil(t, ms)
	require.NotNil(t, ctx)
}

func FuzzMsgCreateBTCValidator(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		keeper, ms, goCtx := setupMsgServer(t)
		ctx := sdk.UnwrapSDKContext(goCtx)

		// generate new BTC validators
		btcVals := []*types.BTCValidator{}
		for i := 0; i < int(datagen.RandomInt(r, 10)); i++ {
			btcVal, err := datagen.GenRandomBTCValidator(r)
			require.NoError(t, err)
			msg := &types.MsgCreateBTCValidator{
				Signer:    datagen.GenRandomAccount().Address,
				BabylonPk: btcVal.BabylonPk,
				BtcPk:     btcVal.BtcPk,
				Pop:       btcVal.Pop,
			}
			_, err = ms.CreateBTCValidator(goCtx, msg)
			require.NoError(t, err)

			btcVals = append(btcVals, btcVal)
		}
		// assert these validators exist in KVStore
		for _, btcVal := range btcVals {
			btcPK := *btcVal.BtcPk
			require.True(t, keeper.HasBTCValidator(ctx, btcPK))
		}

		// duplicated BTC validators should not pass
		for _, btcVal2 := range btcVals {
			msg := &types.MsgCreateBTCValidator{
				Signer:    datagen.GenRandomAccount().Address,
				BabylonPk: btcVal2.BabylonPk,
				BtcPk:     btcVal2.BtcPk,
				Pop:       btcVal2.Pop,
			}
			_, err := ms.CreateBTCValidator(goCtx, msg)
			require.Error(t, err)
		}
	})
}
