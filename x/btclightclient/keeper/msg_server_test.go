package keeper_test

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/stretchr/testify/require"

	keepertest "github.com/babylonchain/babylon/testutil/keeper"
	"github.com/babylonchain/babylon/x/btclightclient/keeper"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func setupMsgServer(t testing.TB) (types.MsgServer, *keeper.Keeper, context.Context) {
	k, ctx := keepertest.BTCLightClientKeeper(t)
	return keeper.NewMsgServerImpl(*k), k, ctx
}

func setupMsgServerWithCustomParams(t testing.TB, p types.Params) (types.MsgServer, *keeper.Keeper, context.Context) {
	k, ctx := keepertest.BTCLightClientKeeperWithCustomParams(t, p)
	return keeper.NewMsgServerImpl(*k), k, ctx
}

// Property: Inserting valid chain which has current tip as parent, should always update the chain
// and tip
func FuzzMsgServerInsertNewTip(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 5)
	senderPrivKey := secp256k1.GenPrivKey()
	address, err := sdk.AccAddressFromHexUnsafe(senderPrivKey.PubKey().Address().String())
	require.NoError(f, err)

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

		msg := &types.MsgInsertHeaders{Signer: address.String(), Headers: chainToChainBytes(chainExtension)}

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
	senderPrivKey := secp256k1.GenPrivKey()
	address, err := sdk.AccAddressFromHexUnsafe(senderPrivKey.PubKey().Address().String())
	require.NoError(f, err)

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
		msg := &types.MsgInsertHeaders{Signer: address.String(), Headers: chainToChainBytes(chainExtension)}

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

func TestAllowUpdatesOnlyFromReportesInTheList(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	sender1 := secp256k1.GenPrivKey()
	address1, err := sdk.AccAddressFromHexUnsafe(sender1.PubKey().Address().String())
	require.NoError(t, err)
	sender2 := secp256k1.GenPrivKey()
	address2, err := sdk.AccAddressFromHexUnsafe(sender2.PubKey().Address().String())
	require.NoError(t, err)
	sender3 := secp256k1.GenPrivKey()
	address3, err := sdk.AccAddressFromHexUnsafe(sender3.PubKey().Address().String())
	require.NoError(t, err)

	params := types.NewParams(
		// only sender1 and sender2 are allowed to update
		[]string{address1.String(), address2.String()},
	)

	srv, blcKeeper, sdkCtx := setupMsgServerWithCustomParams(t, params)
	ctx := sdk.UnwrapSDKContext(sdkCtx)

	_, chain := genRandomChain(
		t,
		r,
		blcKeeper,
		ctx,
		0,
		10,
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

	chainExtension := datagen.GenRandomValidChainStartingFrom(
		r,
		initTip.Height,
		initTip.Header.ToBlockHeader(),
		nil,
		10,
	)

	// sender 1 is allowed to update, it should succeed
	msg := &types.MsgInsertHeaders{Signer: address1.String(), Headers: chainToChainBytes(chainExtension)}
	_, err = srv.InsertHeaders(sdkCtx, msg)
	require.NoError(t, err)

	newTip := blcKeeper.GetTipInfo(ctx)
	require.NotNil(t, newTip)

	newChainExt := datagen.GenRandomValidChainStartingFrom(
		r,
		newTip.Height,
		newTip.Header.ToBlockHeader(),
		nil,
		10,
	)

	// sender 3 is not allowed to update, it should fail
	msg1 := &types.MsgInsertHeaders{Signer: address3.String(), Headers: chainToChainBytes(newChainExt)}
	_, err = srv.InsertHeaders(sdkCtx, msg1)
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrUnauthorizedReporter)

	// sender 2 is allowed to update, it should succeed
	msg1 = &types.MsgInsertHeaders{Signer: address2.String(), Headers: chainToChainBytes(newChainExt)}
	_, err = srv.InsertHeaders(sdkCtx, msg1)
	require.NoError(t, err)
}
