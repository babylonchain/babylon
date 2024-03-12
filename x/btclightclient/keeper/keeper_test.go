package keeper_test

import (
	"errors"
	"math/rand"
	"testing"

	sdkmath "cosmossdk.io/math"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/btcsuite/btcd/chaincfg"

	"github.com/babylonchain/babylon/testutil/datagen"
	keepertest "github.com/babylonchain/babylon/testutil/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func FuzzKeeperMainChainDepth(f *testing.F) {
	/*
		Checks:
		1. if the BTCHeaderBytes object is nil, an error is returned and the height is -1
		2. if the BTCHeaderBytes object does not exist in storage, (-1, error) is returned
		3. if the header is not on the main chain, (0, nil) is returned
		4. if the header exists and is on the mainchain, (depth, nil) is returned

		Data Generation:
		- Generate a random chain of headers.
		- Random generation of a header that is not inserted into storage.
		- Random selection of a header from the main chain and outside of it.
	*/
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		blcKeeper, ctx := keepertest.BTCLightClientKeeper(t)

		// Test nil input
		depth, err := blcKeeper.MainChainDepth(ctx, nil)
		if err == nil {
			t.Errorf("Nil input led to nil error")
		}
		if depth != 0 {
			t.Errorf("Nil input led to a result that is not -1")
		}

		// Test header not existing
		nonExistentHeader := datagen.GenRandomBTCHeaderBytes(r, nil, nil)
		depth, err = blcKeeper.MainChainDepth(ctx, nonExistentHeader.Hash())
		if err == nil {
			t.Errorf("Non existent header led to nil error")
		}
		if depth != 0 {
			t.Errorf("Non existing header led to a result that is not -1")
		}

		_, chain := datagen.GenRandBtcChainInsertingInKeeper(
			t,
			r,
			blcKeeper,
			ctx,
			0,
			datagen.RandomInt(r, 50)+10,
		)
		randomHeader := chain.GetRandomHeaderInfo(r)
		depth, err = blcKeeper.MainChainDepth(ctx, randomHeader.Hash)
		require.NoError(t, err)
		chainTip := chain.GetTipInfo()
		headerDepth := chainTip.Height - randomHeader.Height
		require.Equal(t, headerDepth, depth)
	})
}

func FuzzKeeperBlockHeight(f *testing.F) {
	/*
		Checks:
		1. if the BTCHeaderBytes object is nil, a (0, error) is returned
		2. if the BTCHeaderBytes object does not exist in storage, (0, error) is returned.
		3. if the BTCHeaderBytes object exists, (height, nil) is returned.

		Data Generation:
		- Generate a random chain of headers.
		- Random generation of a header that is not inserted into storage.
		- Random selection of a header from the main chain and outside of it.
	*/
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		blcKeeper, ctx := keepertest.BTCLightClientKeeper(t)

		// Test nil input
		height, err := blcKeeper.BlockHeight(ctx, nil)
		if err == nil {
			t.Errorf("Nil input led to nil error")
		}
		if height != 0 {
			t.Errorf("Nil input led to a result that is not -1")
		}

		// Test header not existing
		nonExistentHeader := datagen.GenRandomBTCHeaderBytes(r, nil, nil)
		height, err = blcKeeper.BlockHeight(ctx, nonExistentHeader.Hash())
		if err == nil {
			t.Errorf("Non existent header led to nil error")
		}
		if height != 0 {
			t.Errorf("Non existing header led to a result that is not -1")
		}

		_, chain := datagen.GenRandBtcChainInsertingInKeeper(
			t,
			r,
			blcKeeper,
			ctx,
			0,
			datagen.RandomInt(r, 50)+10,
		)

		header := chain.GetRandomHeaderInfo(r)
		height, err = blcKeeper.BlockHeight(ctx, header.Hash)
		if err != nil {
			t.Errorf("Existent header led to an error")
		}
		if height != header.Height {
			t.Errorf("BlockHeight returned %d, expected %d", height, header.Height)
		}
	})
}

func FuzzKeeperInsertValidChainExtension(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		blcKeeper, ctx := keepertest.BTCLightClientKeeper(t)

		_, chain := datagen.GenRandBtcChainInsertingInKeeper(
			t,
			r,
			blcKeeper,
			ctx,
			datagen.RandomInt(r, 50)+10,
			datagen.RandomInt(r, 50)+10,
		)

		mockHooks := NewMockHooks()
		blcKeeper.SetHooks(mockHooks)

		newChainLength := uint32(datagen.RandomInt(r, 10) + 1)

		chainToInsert := datagen.GenRandomValidChainStartingFrom(
			r,
			chain.GetTipInfo().Height,
			chain.GetTipInfo().Header.ToBlockHeader(),
			nil,
			newChainLength,
		)
		chainExtensionWork := chainWork(chainToInsert)

		ctx = ctx.WithEventManager(sdk.NewEventManager())
		oldTip := blcKeeper.HeadersState(ctx).GetTip()
		extendedChainWork := oldTip.Work.Add(*chainExtensionWork)
		extendedChainHeight := uint64(uint32(oldTip.Height) + newChainLength)

		err := blcKeeper.InsertHeaders(ctx, keepertest.ChainToChainBytes(chainToInsert))
		require.NoError(t, err)

		// updated tip
		newTip := blcKeeper.HeadersState(ctx).GetTip()
		require.False(t, newTip.Eq(oldTip))
		// check tip
		checkTip(
			t,
			ctx,
			blcKeeper,
			extendedChainWork,
			extendedChainHeight,
			chainToInsert[len(chainToInsert)-1],
		)
		// check all inserted headers
		for _, header := range chainToInsert {
			headerHash := header.BlockHash()
			hash := bbn.NewBTCHeaderHashBytesFromChainhash(&headerHash)
			headerInfoByHash := blcKeeper.GetHeaderByHash(ctx, &hash)
			require.NotNil(t, headerInfoByHash)
			headerInfoByHeight := blcKeeper.GetHeaderByHeight(ctx, headerInfoByHash.Height)
			require.NotNil(t, headerInfoByHeight)
			require.True(t, allFieldsEqual(headerInfoByHash, headerInfoByHeight))
		}

		// check events and hooks
		rollForwadType, _ := sdk.TypedEventToEvent(&types.EventBTCRollForward{})
		headerInsertedType, _ := sdk.TypedEventToEvent(&types.EventBTCHeaderInserted{})

		events := ctx.EventManager().Events()
		numEvents := len(events)
		require.Len(t, mockHooks.AfterBTCHeaderInsertedStore, len(chainToInsert))
		require.Len(t, mockHooks.AfterBTCRollForwardStore, len(chainToInsert))
		require.Len(t, mockHooks.AfterBTCRollBackStore, 0)
		require.Equal(t, numEvents, len(chainToInsert)*2)

		for i, header := range chainToInsert {
			headerHash := header.BlockHash()
			hash := bbn.NewBTCHeaderHashBytesFromChainhash(&headerHash)
			require.True(t, mockHooks.AfterBTCHeaderInsertedStore[i].Hash.Eq(&hash))
			require.True(t, mockHooks.AfterBTCRollForwardStore[i].Hash.Eq(&hash))
			// event should be in order inserted -> roll forward
			require.Equal(t, events[i*2].Type, headerInsertedType.Type)
			require.Equal(t, events[i*2+1].Type, rollForwadType.Type)
		}
	})
}

func FuzzKeeperInsertValidBetterChain(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		blcKeeper, ctx := keepertest.BTCLightClientKeeper(t)
		_, chain := datagen.GenRandBtcChainInsertingInKeeper(
			t,
			r,
			blcKeeper,
			ctx,
			datagen.RandomInt(r, 50)+10,
			datagen.RandomInt(r, 50)+10,
		)

		mockHooks := NewMockHooks()
		blcKeeper.SetHooks(mockHooks)

		forkHeaderParent := chain.GetRandomHeaderInfoNoTip(r)
		// new chain will always be better that existing one
		newChainLength := uint32(chain.ChainLength() + 1)
		chainToInsert := datagen.GenRandomValidChainStartingFrom(
			r,
			forkHeaderParent.Height,
			forkHeaderParent.Header.ToBlockHeader(),
			nil,
			newChainLength,
		)
		chainExtensionWork := chainWork(chainToInsert)
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		extendedChainWork := forkHeaderParent.Work.Add(*chainExtensionWork)
		extendedChainHeight := uint64(uint32(forkHeaderParent.Height) + newChainLength)

		oldTip := blcKeeper.HeadersState(ctx).GetTip()
		removedBranch := blcKeeper.GetMainChainFrom(ctx, forkHeaderParent.Height+1)

		require.True(t, len(removedBranch) > 0)

		err := blcKeeper.InsertHeaders(ctx, keepertest.ChainToChainBytes(chainToInsert))
		require.NoError(t, err)

		// updated tip
		newTip := blcKeeper.HeadersState(ctx).GetTip()
		require.False(t, newTip.Eq(oldTip))
		// check tip
		checkTip(
			t,
			ctx,
			blcKeeper,
			extendedChainWork,
			extendedChainHeight,
			chainToInsert[len(chainToInsert)-1],
		)

		// check all headers from removed branch were removed
		for _, headerInfo := range removedBranch {
			headerInfoByHash := blcKeeper.GetHeaderByHash(ctx, headerInfo.Hash)
			require.Nil(t, headerInfoByHash)
		}

		// check all inserted headers
		for _, header := range chainToInsert {
			headerHash := header.BlockHash()
			hash := bbn.NewBTCHeaderHashBytesFromChainhash(&headerHash)
			headerInfoByHash := blcKeeper.GetHeaderByHash(ctx, &hash)
			require.NotNil(t, headerInfoByHash)
			headerInfoByHeight := blcKeeper.GetHeaderByHeight(ctx, headerInfoByHash.Height)
			require.NotNil(t, headerInfoByHeight)
			require.True(t, allFieldsEqual(headerInfoByHash, headerInfoByHeight))
		}

		// check events and hooks
		rollBackType, _ := sdk.TypedEventToEvent(&types.EventBTCRollBack{})
		rollForwadType, _ := sdk.TypedEventToEvent(&types.EventBTCRollForward{})
		headerInsertedType, _ := sdk.TypedEventToEvent(&types.EventBTCHeaderInserted{})

		events := ctx.EventManager().Events()
		numEvents := len(events)
		require.Len(t, mockHooks.AfterBTCHeaderInsertedStore, len(chainToInsert))
		require.Len(t, mockHooks.AfterBTCRollForwardStore, len(chainToInsert))
		// there is one roll back event
		require.Len(t, mockHooks.AfterBTCRollBackStore, 1)
		require.Equal(t, numEvents, len(chainToInsert)*2+1)

		// Events should be ordered:
		// Rollback, Insert, RollForward, Insert, RollForward, ...
		for i, header := range chainToInsert {
			if i == 0 {
				// rollback event
				require.Equal(t, events[0].Type, rollBackType.Type)
				// rollback should indicate highest common ancestor i.e fork header parent
				require.True(t, mockHooks.AfterBTCRollBackStore[0].Hash.Eq(forkHeaderParent.Hash))
				continue
			}

			headerHash := header.BlockHash()
			hash := bbn.NewBTCHeaderHashBytesFromChainhash(&headerHash)
			require.True(t, mockHooks.AfterBTCHeaderInsertedStore[i].Hash.Eq(&hash))
			require.True(t, mockHooks.AfterBTCRollForwardStore[i].Hash.Eq(&hash))

			require.Equal(t, events[i*2+1].Type, headerInsertedType.Type)
			require.Equal(t, events[i*2].Type, rollForwadType.Type)
		}
	})
}

func FuzzKeeperInsertInvalidChain(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		blcKeeper, ctx := keepertest.BTCLightClientKeeper(t)
		_, _ = datagen.GenRandBtcChainInsertingInKeeper(
			t,
			r,
			blcKeeper,
			ctx,
			0,
			datagen.RandomInt(r, 50)+10,
		)
		currentTip := blcKeeper.GetTipInfo(ctx)
		require.NotNil(t, currentTip)

		// Inserting nil headers should result with error
		errNil := blcKeeper.InsertHeaders(ctx, nil)
		require.Error(t, errNil)

		// Inserting empty headers should result with error
		errEmpty := blcKeeper.InsertHeaders(ctx, []bbn.BTCHeaderBytes{})
		require.Error(t, errEmpty)

		// Inserting header without existing parent should result with error
		chain := datagen.NewBTCHeaderChainWithLength(r, 0, 0, 10)
		errNoParent := blcKeeper.InsertHeaders(ctx, chain.ChainToBytes()[1:])
		require.Error(t, errNoParent)
		require.True(t, errors.Is(errNoParent, types.ErrHeaderParentDoesNotExist))

		// Inserting header chain with invalid header should result in error
		newChainLength := uint32(datagen.RandomInt(r, 10) + 5)
		// valid chain with at least 5 headers
		chainToInsert := datagen.GenRandomValidChainStartingFrom(
			r,
			chain.GetTipInfo().Height,
			chain.GetTipInfo().Header.ToBlockHeader(),
			nil,
			newChainLength,
		)

		// bump the nonce, it should fail validation and tip should not change
		chainToInsert[3].Nonce = chainToInsert[3].Nonce + 1
		errInvalidHeader := blcKeeper.InsertHeaders(ctx, keepertest.ChainToChainBytes(chainToInsert))
		require.Error(t, errInvalidHeader)
		newTip := blcKeeper.GetTipInfo(ctx)
		// tip did not change
		require.True(t, allFieldsEqual(currentTip, newTip))

		// Inserting header chain with less work than current chain work should result in error
		headerBeforeTip := blcKeeper.GetHeaderByHeight(ctx, currentTip.Height-1)
		require.NotNil(t, headerBeforeTip)
		worseChain := datagen.GenRandomValidChainStartingFrom(
			r,
			headerBeforeTip.Height,
			headerBeforeTip.Header.ToBlockHeader(),
			nil,
			1,
		)
		errWorseChain := blcKeeper.InsertHeaders(ctx, keepertest.ChainToChainBytes(worseChain))
		require.Error(t, errWorseChain)
		require.True(t, errors.Is(errWorseChain, types.ErrChainWithNotEnoughWork))
	})
}

func FuzzKeeperValdateHeaderAtDifficultyAdjustmentBoundaries(f *testing.F) {
	// less seeds as we generate longer chains
	datagen.AddRandomSeedsToFuzzer(f, 3)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		numBlockPerRetarget := types.BlocksPerRetarget(&chaincfg.SimNetParams)
		blcKeeper, ctx := keepertest.BTCLightClientKeeper(t)

		genesisHeader := bbn.NewBTCHeaderBytesFromBlockHeader(&chaincfg.SimNetParams.GenesisBlock.Header)
		genesisHash := bbn.NewBTCHeaderHashBytesFromChainhash(chaincfg.SimNetParams.GenesisHash)
		genesisWork := sdkmath.NewUint(0)

		genesisInfo := types.NewBTCHeaderInfo(
			&genesisHeader,
			&genesisHash,
			0,
			&genesisWork,
		)

		require.True(t, types.IsRetargetBlock(genesisInfo, &chaincfg.SimNetParams))
		blcKeeper.SetBaseBTCHeader(ctx, *genesisInfo)
		randomChain := datagen.NewBTCHeaderChainFromParentInfo(
			r,
			genesisInfo,
			uint32(numBlockPerRetarget),
		)

		// this will always fail as last header is at adjustment boundary, but we created
		// it without adjustment
		err := blcKeeper.InsertHeaders(ctx, randomChain.ChainToBytes())
		require.Error(t, err)

		randomChainWithoutLastHeader := randomChain.Headers[:len(randomChain.Headers)-1]
		chain := keepertest.ChainToChainBytes(randomChainWithoutLastHeader)
		// now all headers are valid, and we are below adjustment boundary
		err = blcKeeper.InsertHeaders(ctx, chain)
		require.NoError(t, err)

		currentTip := blcKeeper.GetTipInfo(ctx)
		require.NotNil(t, currentTip)
		require.Equal(t, currentTip.Height, uint64(numBlockPerRetarget)-1)

		invalidAdjustedHeader := datagen.GenRandomBtcdValidHeader(
			r,
			currentTip.Header.ToBlockHeader(),
			nil,
			nil,
		)
		// try to insert header at adjustment boundary without adjustment should fail
		err = blcKeeper.InsertHeaders(ctx, []bbn.BTCHeaderBytes{bbn.NewBTCHeaderBytesFromBlockHeader(invalidAdjustedHeader)})
		require.Error(t, err)

		// Inserting valid adjusted header should succeed
		rt := datagen.RetargetInfo{
			LastRetargetHeader: genesisHeader.ToBlockHeader(),
			Params:             &chaincfg.SimNetParams,
		}
		validAdjustedHeader := datagen.GenRandomBtcdValidHeader(
			r,
			// current tip heigh is 2015
			currentTip.Header.ToBlockHeader(),
			nil,
			&rt,
		)
		validAdjustedHeaderBytes := bbn.NewBTCHeaderBytesFromBlockHeader(validAdjustedHeader)

		err = blcKeeper.InsertHeaders(ctx, []bbn.BTCHeaderBytes{bbn.NewBTCHeaderBytesFromBlockHeader(validAdjustedHeader)})
		require.NoError(t, err)

		newTip := blcKeeper.GetTipInfo(ctx)
		require.NotNil(t, newTip)
		// tip should be at adjustment boundary now
		require.Equal(t, newTip.Height, uint64(numBlockPerRetarget))
		require.True(t, newTip.Header.Eq(&validAdjustedHeaderBytes))
		require.True(t, types.IsRetargetBlock(newTip, &chaincfg.SimNetParams))
	})
}
