package query

import (
	"context"

	btcstakingtypes "github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/cosmos/cosmos-sdk/client"
	sdkquerytypes "github.com/cosmos/cosmos-sdk/types/query"
)

// QueryBTCStaking queries the BTCStaking module of the Babylon node according to the given function
func (c *QueryClient) QueryBTCStaking(f func(ctx context.Context, queryClient btcstakingtypes.QueryClient) error) error {
	ctx, cancel := c.getQueryContext()
	defer cancel()

	clientCtx := client.Context{Client: c.RPCClient}
	queryClient := btcstakingtypes.NewQueryClient(clientCtx)

	return f(ctx, queryClient)
}

// BTCStakingParams queries the BTC staking module parameters
func (c *QueryClient) BTCStakingParams() (*btcstakingtypes.QueryParamsResponse, error) {
	var resp *btcstakingtypes.QueryParamsResponse
	err := c.QueryBTCStaking(func(ctx context.Context, queryClient btcstakingtypes.QueryClient) error {
		var err error
		req := &btcstakingtypes.QueryParamsRequest{}
		resp, err = queryClient.Params(ctx, req)
		return err
	})

	return resp, err
}

// FinalityProvider queries the BTCStaking module for a given finlaity provider
func (c *QueryClient) FinalityProvider(fpBtcPkHex string) (*btcstakingtypes.QueryFinalityProviderResponse, error) {
	var resp *btcstakingtypes.QueryFinalityProviderResponse
	err := c.QueryBTCStaking(func(ctx context.Context, queryClient btcstakingtypes.QueryClient) error {
		var err error
		req := &btcstakingtypes.QueryFinalityProviderRequest{
			FpBtcPkHex: fpBtcPkHex,
		}
		resp, err = queryClient.FinalityProvider(ctx, req)
		return err
	})

	return resp, err
}

// FinalityProviders queries the BTCStaking module for all finality providers
func (c *QueryClient) FinalityProviders(pagination *sdkquerytypes.PageRequest) (*btcstakingtypes.QueryFinalityProvidersResponse, error) {
	var resp *btcstakingtypes.QueryFinalityProvidersResponse
	err := c.QueryBTCStaking(func(ctx context.Context, queryClient btcstakingtypes.QueryClient) error {
		var err error
		req := &btcstakingtypes.QueryFinalityProvidersRequest{
			Pagination: pagination,
		}
		resp, err = queryClient.FinalityProviders(ctx, req)
		return err
	})

	return resp, err
}

// FinalityProviderDelegations queries the BTCStaking module for all delegations of a finality provider
func (c *QueryClient) FinalityProviderDelegations(fpBtcPkHex string, pagination *sdkquerytypes.PageRequest) (*btcstakingtypes.QueryFinalityProviderDelegationsResponse, error) {
	var resp *btcstakingtypes.QueryFinalityProviderDelegationsResponse
	err := c.QueryBTCStaking(func(ctx context.Context, queryClient btcstakingtypes.QueryClient) error {
		var err error
		req := &btcstakingtypes.QueryFinalityProviderDelegationsRequest{
			FpBtcPkHex: fpBtcPkHex,
			Pagination: pagination,
		}
		resp, err = queryClient.FinalityProviderDelegations(ctx, req)
		return err
	})

	return resp, err
}

// BTCDelegations queries the BTCStaking module for all delegations under a given status
func (c *QueryClient) BTCDelegations(status btcstakingtypes.BTCDelegationStatus, pagination *sdkquerytypes.PageRequest) (*btcstakingtypes.QueryBTCDelegationsResponse, error) {
	var resp *btcstakingtypes.QueryBTCDelegationsResponse
	err := c.QueryBTCStaking(func(ctx context.Context, queryClient btcstakingtypes.QueryClient) error {
		var err error
		req := &btcstakingtypes.QueryBTCDelegationsRequest{
			Status:     status,
			Pagination: pagination,
		}
		resp, err = queryClient.BTCDelegations(ctx, req)
		return err
	})

	return resp, err
}

// BTCDelegation queries the BTCStaking module to retrieve delegation by corresponding staking tx hash
func (c *QueryClient) BTCDelegation(stakingTxHashHex string) (*btcstakingtypes.QueryBTCDelegationResponse, error) {
	var resp *btcstakingtypes.QueryBTCDelegationResponse
	err := c.QueryBTCStaking(func(ctx context.Context, queryClient btcstakingtypes.QueryClient) error {
		var err error
		req := &btcstakingtypes.QueryBTCDelegationRequest{
			StakingTxHashHex: stakingTxHashHex,
		}
		resp, err = queryClient.BTCDelegation(ctx, req)
		return err
	})

	return resp, err
}

// ActiveFinalityProvidersAtHeight queries the BTCStaking module for all finality providers
// with non-zero voting power at a given height
func (c *QueryClient) ActiveFinalityProvidersAtHeight(height uint64, pagination *sdkquerytypes.PageRequest) (*btcstakingtypes.QueryActiveFinalityProvidersAtHeightResponse, error) {
	var resp *btcstakingtypes.QueryActiveFinalityProvidersAtHeightResponse
	err := c.QueryBTCStaking(func(ctx context.Context, queryClient btcstakingtypes.QueryClient) error {
		var err error
		req := &btcstakingtypes.QueryActiveFinalityProvidersAtHeightRequest{
			Height:     height,
			Pagination: pagination,
		}
		resp, err = queryClient.ActiveFinalityProvidersAtHeight(ctx, req)
		return err
	})

	return resp, err
}

// FinalityProviderPowerAtHeight queries the BTCStaking module for the power of a finality provider at a given height
func (c *QueryClient) FinalityProviderPowerAtHeight(fpBtcPkHex string, height uint64) (*btcstakingtypes.QueryFinalityProviderPowerAtHeightResponse, error) {
	var resp *btcstakingtypes.QueryFinalityProviderPowerAtHeightResponse
	err := c.QueryBTCStaking(func(ctx context.Context, queryClient btcstakingtypes.QueryClient) error {
		var err error
		req := &btcstakingtypes.QueryFinalityProviderPowerAtHeightRequest{
			FpBtcPkHex: fpBtcPkHex,
			Height:     height,
		}
		resp, err = queryClient.FinalityProviderPowerAtHeight(ctx, req)
		return err
	})

	return resp, err
}

func (c *QueryClient) ActivatedHeight() (*btcstakingtypes.QueryActivatedHeightResponse, error) {
	var resp *btcstakingtypes.QueryActivatedHeightResponse
	err := c.QueryBTCStaking(func(ctx context.Context, queryClient btcstakingtypes.QueryClient) error {
		var err error
		req := &btcstakingtypes.QueryActivatedHeightRequest{}
		resp, err = queryClient.ActivatedHeight(ctx, req)
		return err
	})

	return resp, err
}
