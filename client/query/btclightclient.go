package query

import (
	"context"

	btclctypes "github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/cosmos/cosmos-sdk/client"
	sdkquerytypes "github.com/cosmos/cosmos-sdk/types/query"
)

// QueryBTCLightclient queries the BTCLightclient module of the Babylon node
// according to the given function
func (c *QueryClient) QueryBTCLightclient(f func(ctx context.Context, queryClient btclctypes.QueryClient) error) error {
	ctx, cancel := c.getQueryContext()
	defer cancel()

	clientCtx := client.Context{Client: c.RPCClient}
	queryClient := btclctypes.NewQueryClient(clientCtx)

	return f(ctx, queryClient)
}

// BTCHeaderChainTip queries hash/height of the latest BTC block in the btclightclient module
func (c *QueryClient) BTCHeaderChainTip() (*btclctypes.QueryTipResponse, error) {
	var resp *btclctypes.QueryTipResponse
	err := c.QueryBTCLightclient(func(ctx context.Context, queryClient btclctypes.QueryClient) error {
		var err error
		req := &btclctypes.QueryTipRequest{}
		resp, err = queryClient.Tip(ctx, req)
		return err
	})

	return resp, err
}

// BTCBaseHeader queries the base BTC header of the btclightclient module
func (c *QueryClient) BTCBaseHeader() (*btclctypes.QueryBaseHeaderResponse, error) {
	var resp *btclctypes.QueryBaseHeaderResponse
	err := c.QueryBTCLightclient(func(ctx context.Context, queryClient btclctypes.QueryClient) error {
		var err error
		req := &btclctypes.QueryBaseHeaderRequest{}
		resp, err = queryClient.BaseHeader(ctx, req)
		return err
	})

	return resp, err
}

// ContainsBTCBlock queries the btclightclient module for the existence of a block hash
func (c *QueryClient) ContainsBTCBlock(blockHash *chainhash.Hash) (*btclctypes.QueryContainsBytesResponse, error) {
	var resp *btclctypes.QueryContainsBytesResponse
	err := c.QueryBTCLightclient(func(ctx context.Context, queryClient btclctypes.QueryClient) error {
		var err error
		req := &btclctypes.QueryContainsBytesRequest{
			Hash: blockHash.CloneBytes(),
		}
		resp, err = queryClient.ContainsBytes(ctx, req)
		return err
	})

	return resp, err
}

// BTCMainChain queries the btclightclient module for the BTC canonical chain
func (c *QueryClient) BTCMainChain(pagination *sdkquerytypes.PageRequest) (*btclctypes.QueryMainChainResponse, error) {
	var resp *btclctypes.QueryMainChainResponse
	err := c.QueryBTCLightclient(func(ctx context.Context, queryClient btclctypes.QueryClient) error {
		var err error
		req := &btclctypes.QueryMainChainRequest{
			Pagination: pagination,
		}
		resp, err = queryClient.MainChain(ctx, req)
		return err
	})

	return resp, err
}
