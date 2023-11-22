package keeper_test

import (
	"context"
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/stretchr/testify/require"

	keepertest "github.com/babylonchain/babylon/testutil/keeper"
	"github.com/babylonchain/babylon/x/btclightclient/keeper"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func setupMsgServer(t testing.TB) (types.MsgServer, *keeper.Keeper, context.Context) {
	k, ctx := keepertest.BTCLightClientKeeper(t)
	return keeper.NewMsgServerImpl(*k), k, ctx
}

// Property: Inserting valid chain which has current tip as parent, should always update the chain
// and tip
func FuzzMsgServerInsertNewTip(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 5)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		srv, blcKeeper, sdkCtx := setupMsgServer(t)
		ctx := sdk.UnwrapSDKContext(sdkCtx)
		_, chain := genRandomChain(
			t,
			r,
			blcKeeper,
			ctx,
			datagen.RandomInt(r, 50)+10,
			datagen.RandomInt(r, 50)+10,
		)
		initTip := chain.GetTipInfo()

		checkTip(
			t,
			ctx,
			blcKeeper,
			*initTip.Work,
			initTip.Height,
			initTip.Header.ToBlockHeader(),
		)

		chainExenstionLength := uint32(r.Int31n(200) + 1)
		chainExtension := datagen.GenRandomValidChainStartingFrom(
			r,
			initTip.Height,
			initTip.Header.ToBlockHeader(),
			nil,
			chainExenstionLength,
		)
		chainExtensionWork := chainWork(chainExtension)

		msg := &types.MsgInsertHeaders{Headers: chainToChainBytes(chainExtension)}

		_, err := srv.InsertHeaders(sdkCtx, msg)
		require.NoError(t, err)

		extendedChainWork := initTip.Work.Add(*chainExtensionWork)
		extendedChainHeight := uint64(uint32(initTip.Height) + chainExenstionLength)

		checkTip(
			t,
			ctx,
			blcKeeper,
			extendedChainWork,
			extendedChainHeight,
			chainExtension[len(chainExtension)-1],
		)
	})
}

// Property: Inserting valid better chain should always update the chain and tip
func FuzzMsgServerReorgChain(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 5)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		srv, blcKeeper, sdkCtx := setupMsgServer(t)
		ctx := sdk.UnwrapSDKContext(sdkCtx)

		chainLength := datagen.RandomInt(r, 50) + 10
		_, chain := genRandomChain(
			t,
			r,
			blcKeeper,
			ctx,
			datagen.RandomInt(r, 50)+10,
			chainLength,
		)
		initTip := chain.GetTipInfo()

		checkTip(
			t,
			ctx,
			blcKeeper,
			*initTip.Work,
			initTip.Height,
			initTip.Header.ToBlockHeader(),
		)

		reorgDepth := r.Intn(int(chainLength-1)) + 1

		forkHeaderHeight := initTip.Height - uint64(reorgDepth)
		forkHeader := blcKeeper.GetHeaderByHeight(ctx, forkHeaderHeight)
		require.NotNil(t, forkHeader)

		// fork chain will always be longer that current c
		forkChainLen := reorgDepth + 10
		chainExtension := datagen.GenRandomValidChainStartingFrom(
			r,
			forkHeader.Height,
			forkHeader.Header.ToBlockHeader(),
			nil,
			uint32(forkChainLen),
		)
		chainExtensionWork := chainWork(chainExtension)
		msg := &types.MsgInsertHeaders{Headers: chainToChainBytes(chainExtension)}

		_, err := srv.InsertHeaders(sdkCtx, msg)
		require.NoError(t, err)

		extendedChainWork := forkHeader.Work.Add(*chainExtensionWork)
		extendedChainHeight := forkHeader.Height + uint64(forkChainLen)

		checkTip(
			t,
			ctx,
			blcKeeper,
			extendedChainWork,
			extendedChainHeight,
			chainExtension[len(chainExtension)-1],
		)
	})
}
