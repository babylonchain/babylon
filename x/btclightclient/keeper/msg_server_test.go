package keeper_test

import (
	"context"
	"github.com/babylonchain/babylon/testutil/datagen"
	"math/big"
	"math/rand"
	"testing"

	keepertest "github.com/babylonchain/babylon/testutil/keeper"
	"github.com/babylonchain/babylon/x/btclightclient/keeper"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func setupMsgServer(t testing.TB) (types.MsgServer, *keeper.Keeper, context.Context) {
	k, ctx := keepertest.BTCLightClientKeeper(t)
	return keeper.NewMsgServerImpl(*k), k, sdk.WrapSDKContext(ctx)
}

func FuzzMsgServerInsertHeader(f *testing.F) {
	/*
		Test that:
		1. if the input message is nil, (nil, error) is returned
		2. if the msg does not contain a header, (nil, error) is returned
		3. if the parent of the header does not exist, (nil, error) is returned
		4. if the work of the header is not within the limits of the new header, (nil, error) is returned
		5. if all checks pass, the header is inserted into storage and an (empty MsgInsertHeaderResponse, nil) is returned
		   - we do not need to perform insertion checks since those are performed on FuzzKeeperInsertHeader
		Building:
		- Construct a random tree and insert into storage
		- Generate a random header for which its parent does not exist
		- Select a random header from the tree and construct BTCHeaderBytes objects on top of it with different work
			 1. 4 times the work of parent
			 2. 1 < work < 4 times the work of parent
			 3. work > 4 times the work of the parent
			 4. parent 4 times the work of the header
			 5. parent 1 < work < 4 times the work of the header
			 6. parent > 4 times the work of the header
	*/
	f.Add(int64(42))
	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)
		msgServer, blcKeeper, sdkCtx := setupMsgServer(t)

		// If the input message is nil, (nil, error) is returned
		var msg *types.MsgInsertHeader = nil
		resp, err := msgServer.InsertHeader(sdkCtx, msg)
		if resp != nil {
			t.Errorf("Nil message returned a response")
		}
		if err == nil {
			t.Errorf("Nil message did not return an error")
		}

		// If the message does not contain a header, (nil, error) is returned.
		msg = &types.MsgInsertHeader{}
		resp, err = msgServer.InsertHeader(sdkCtx, msg)
		if resp != nil {
			t.Errorf("Message without a header returned a response")
		}
		if err == nil {
			t.Errorf("Message without a header did not return an error")
		}

		// If the header has a parent that does not exist, (nil, error) is returned
		headerParentNotExists := datagen.GenRandomBTCHeaderInfo().Header
		msg = &types.MsgInsertHeader{Header: headerParentNotExists}
		resp, err = msgServer.InsertHeader(sdkCtx, msg)
		if resp != nil {
			t.Errorf("Message with header with non-existent parent returned a response")
		}
		if err == nil {
			t.Errorf("Message with header with non-existent parent did not return an error")
		}

		// Construct a tree and insert it into storage
		headersMap := datagen.GenRandomHeaderInfoTree()
		ctx := sdk.UnwrapSDKContext(sdkCtx)
		for _, headerInfo := range headersMap {
			blcKeeper.HeadersState(ctx).CreateHeader(headerInfo)
		}
		// Select a random header
		parentHeaderIdx := datagen.RandomInt(len(headersMap))
		parentHeader := selectRandomHeaders(headersMap, []uint64{parentHeaderIdx})[0]
		// Do not work with different cases. Select a random integer between 1-BTCDifficultyMultiplier+1
		// 1/BTCDifficultyMultiplier times, the work is going to be invalid
		parentHeaderDifficulty := parentHeader.Header.Difficulty()
		// Avoid BTCDifficultyMultiplier itself, since the many conversions might lead to inconsistencies
		mul := datagen.RandomInt(keeper.BTCDifficultyMultiplier-1) + 1
		if datagen.OneInN(10) { // Give an invalid mul sometimes
			mul = keeper.BTCDifficultyMultiplier + 1
		}
		headerDifficultyMul := sdk.NewUintFromBigInt(new(big.Int).Mul(parentHeaderDifficulty, big.NewInt(int64(mul))))
		headerDifficultyDiv := sdk.NewUintFromBigInt(new(big.Int).Div(parentHeaderDifficulty, big.NewInt(int64(mul))))

		// Do tests
		headerMoreWork := datagen.GenRandomBTCHeaderInfoWithParentAndBits(parentHeader, &headerDifficultyMul)
		msg = &types.MsgInsertHeader{Header: headerMoreWork.Header}
		resp, err = msgServer.InsertHeader(sdkCtx, msg)
		if mul > keeper.BTCDifficultyMultiplier && resp != nil {
			t.Errorf("Invalid header work led to a response getting returned")
		}
		if mul > keeper.BTCDifficultyMultiplier && err == nil {
			t.Errorf("Invalid header work did not lead to an error %d %s %s %s", mul, headerDifficultyMul, headerDifficultyDiv, parentHeaderDifficulty)
		}
		if mul <= keeper.BTCDifficultyMultiplier && err != nil {
			t.Errorf("Valid header work led to an error")
		}

		headerLessWork := datagen.GenRandomBTCHeaderInfoWithParentAndBits(parentHeader, &headerDifficultyDiv)
		msg = &types.MsgInsertHeader{Header: headerLessWork.Header}
		resp, err = msgServer.InsertHeader(sdkCtx, msg)
		if mul > keeper.BTCDifficultyMultiplier && resp != nil {
			t.Errorf("Invalid header work led to a response getting returned")
		}
		if mul > keeper.BTCDifficultyMultiplier && err == nil {
			t.Errorf("Invalid header work did not lead to an error")
		}
		if mul <= keeper.BTCDifficultyMultiplier && err != nil {
			t.Errorf("Valid header work led to an error %d %s", mul, err)
		}
	})
}
