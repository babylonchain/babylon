package keeper_test

import (
	"math/rand"
	"testing"

	testkeeper "github.com/babylonchain/babylon/testutil/keeper"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestParamsQuery(t *testing.T) {
	keeper, ctx := testkeeper.BTCLightClientKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	params := types.DefaultParams()
	keeper.SetParams(ctx, params)

	response, err := keeper.Params(wctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, &types.QueryParamsResponse{Params: params}, response)
}

func FuzzHashesQuery(f *testing.F) {
	/*
		Test that:
		1. If the request is nil, an error is returned
		2. If the pagination key has not been set,
		   `limit` number of hashes are returned and the pagination key
			has been set to the next hash.
		3. If the pagination key has been set,
		   the `limit` number of hashes after the key are returned.
		4. End of pagination: the last hashes are returned properly.
		5. If the pagination key is not a valid hash, an error is returned.
		Building:
		- Generate a random tree of headers and insert their hashes
		  into the hashToHeight storage.
	*/
	f.Add(int64(42))
	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)
		t.Skip()
	})
}

func FuzzContainsQuery(f *testing.F) {
	/*
		Test that:
		1. If the request is nil, an error is returned
		2. The query returns true or false depending on whether the type has been built.
		Building:
		- Generate a random tree of headers and insert the first half of the tree
		  into the headers storage.
		  Use the first half for returning true and the second half for returning false.
	*/
	f.Add(int64(42))
	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)
		t.Skip()
	})
}

func FuzzMainChainQuery(f *testing.F) {
	/*
		 Test that:
		 1. If the request is nil, an error is returned
		 2. If the pagination key is not a valid hash, an error is returned.
		 3. If the pagination key has not been set,
			the first `limit` items of the main chain are returned
		 4. If the pagination key has been set, the `limit` items after it are returned.
		 5. End of pagination: the last elements are returned properly and the next_key is set to nil.
		 Building:
		 - Generate a random tree of headers with different PoW and insert them into the headers storage.
		 - Calculate the main chain using the `HeadersState().MainChain()` function (here we only test the query)
	*/
	f.Add(int64(42))
	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)
		t.Skip()
	})
}
