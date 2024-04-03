package query

import (
	"context"

	finalitytypes "github.com/babylonchain/babylon/x/finality/types"
	"github.com/cosmos/cosmos-sdk/client"
	sdkquerytypes "github.com/cosmos/cosmos-sdk/types/query"
)

// QueryFinality queries the Finality module of the Babylon node according to the given function
func (c *QueryClient) QueryFinality(f func(ctx context.Context, queryClient finalitytypes.QueryClient) error) error {
	ctx, cancel := c.getQueryContext()
	defer cancel()

	clientCtx := client.Context{Client: c.RPCClient}
	queryClient := finalitytypes.NewQueryClient(clientCtx)

	return f(ctx, queryClient)
}

// FinalityParams queries the finality module parameters
func (c *QueryClient) FinalityParams() (*finalitytypes.Params, error) {
	var resp *finalitytypes.QueryParamsResponse
	err := c.QueryFinality(func(ctx context.Context, queryClient finalitytypes.QueryClient) error {
		var err error
		req := &finalitytypes.QueryParamsRequest{}
		resp, err = queryClient.Params(ctx, req)
		return err
	})

	return &resp.Params, err
}

// VotesAtHeight queries the Finality module to get signature set at a given babylon block height
func (c *QueryClient) VotesAtHeight(height uint64) (*finalitytypes.QueryVotesAtHeightResponse, error) {
	var resp *finalitytypes.QueryVotesAtHeightResponse
	err := c.QueryFinality(func(ctx context.Context, queryClient finalitytypes.QueryClient) error {
		var err error
		req := &finalitytypes.QueryVotesAtHeightRequest{
			Height: height,
		}
		resp, err = queryClient.VotesAtHeight(ctx, req)
		return err
	})

	return resp, err
}

// ListBlocks queries the Finality module to get blocks with a given status.
func (c *QueryClient) ListBlocks(status finalitytypes.QueriedBlockStatus, pagination *sdkquerytypes.PageRequest) (*finalitytypes.QueryListBlocksResponse, error) {
	var resp *finalitytypes.QueryListBlocksResponse
	err := c.QueryFinality(func(ctx context.Context, queryClient finalitytypes.QueryClient) error {
		var err error
		req := &finalitytypes.QueryListBlocksRequest{
			Status:     status,
			Pagination: pagination,
		}
		resp, err = queryClient.ListBlocks(ctx, req)
		return err
	})

	return resp, err
}

// Block queries a block at a given height.
func (c *QueryClient) Block(height uint64) (*finalitytypes.QueryBlockResponse, error) {
	var resp *finalitytypes.QueryBlockResponse
	err := c.QueryFinality(func(ctx context.Context, queryClient finalitytypes.QueryClient) error {
		var err error
		req := &finalitytypes.QueryBlockRequest{
			Height: height,
		}
		resp, err = queryClient.Block(ctx, req)
		return err
	})

	return resp, err
}

// ListEvidences queries the Finality module to get evidences after a given height.
func (c *QueryClient) ListEvidences(startHeight uint64, pagination *sdkquerytypes.PageRequest) (*finalitytypes.QueryListEvidencesResponse, error) {
	var resp *finalitytypes.QueryListEvidencesResponse
	err := c.QueryFinality(func(ctx context.Context, queryClient finalitytypes.QueryClient) error {
		var err error
		req := &finalitytypes.QueryListEvidencesRequest{
			StartHeight: startHeight,
			Pagination:  pagination,
		}
		resp, err = queryClient.ListEvidences(ctx, req)
		return err
	})

	return resp, err
}
