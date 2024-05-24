package chain

import (
	"fmt"
	"net/url"

	"github.com/babylonchain/babylon/test/e2e/util"
	bbn "github.com/babylonchain/babylon/types"
	bstypes "github.com/babylonchain/babylon/x/btcstaking/types"
	ftypes "github.com/babylonchain/babylon/x/finality/types"
	"github.com/stretchr/testify/require"
)

func (n *NodeConfig) QueryBTCStakingParams() *bstypes.Params {
	bz, err := n.QueryGRPCGateway("/babylon/btcstaking/v1/params", url.Values{})
	require.NoError(n.t, err)

	var resp bstypes.QueryParamsResponse
	err = util.Cdc.UnmarshalJSON(bz, &resp)
	require.NoError(n.t, err)

	return &resp.Params
}

func (n *NodeConfig) QueryFinalityProviders() []*bstypes.FinalityProviderResponse {
	bz, err := n.QueryGRPCGateway("/babylon/btcstaking/v1/finality_providers", url.Values{})
	require.NoError(n.t, err)

	var resp bstypes.QueryFinalityProvidersResponse
	err = util.Cdc.UnmarshalJSON(bz, &resp)
	require.NoError(n.t, err)

	return resp.FinalityProviders
}

func (n *NodeConfig) QueryActiveFinalityProvidersAtHeight(height uint64) []*bstypes.FinalityProviderWithMeta {
	path := fmt.Sprintf("/babylon/btcstaking/v1/finality_providers/%d", height)
	bz, err := n.QueryGRPCGateway(path, url.Values{})
	require.NoError(n.t, err)

	var resp bstypes.QueryActiveFinalityProvidersAtHeightResponse
	err = util.Cdc.UnmarshalJSON(bz, &resp)
	require.NoError(n.t, err)

	return resp.FinalityProviders
}

func (n *NodeConfig) QueryFinalityProviderDelegations(fpBTCPK string) []*bstypes.BTCDelegatorDelegationsResponse {
	path := fmt.Sprintf("/babylon/btcstaking/v1/finality_providers/%s/delegations", fpBTCPK)
	bz, err := n.QueryGRPCGateway(path, url.Values{})
	require.NoError(n.t, err)

	var resp bstypes.QueryFinalityProviderDelegationsResponse
	err = util.Cdc.UnmarshalJSON(bz, &resp)
	require.NoError(n.t, err)

	return resp.BtcDelegatorDelegations
}

func (n *NodeConfig) QueryBtcDelegation(stakingTxHash string) *bstypes.QueryBTCDelegationResponse {
	path := fmt.Sprintf("/babylon/btcstaking/v1/btc_delegations/%s", stakingTxHash)
	bz, err := n.QueryGRPCGateway(path, url.Values{})
	require.NoError(n.t, err)

	var resp bstypes.QueryBTCDelegationResponse
	err = util.Cdc.UnmarshalJSON(bz, &resp)
	require.NoError(n.t, err)

	return &resp
}

func (n *NodeConfig) QueryUnbondedDelegations() []*bstypes.BTCDelegationResponse {
	queryParams := url.Values{}
	queryParams.Add("status", fmt.Sprintf("%d", bstypes.BTCDelegationStatus_UNBONDED))
	bz, err := n.QueryGRPCGateway("/babylon/btcstaking/v1/btc_delegations", queryParams)
	require.NoError(n.t, err)

	var resp bstypes.QueryBTCDelegationsResponse
	err = util.Cdc.UnmarshalJSON(bz, &resp)
	require.NoError(n.t, err)

	return resp.BtcDelegations
}

func (n *NodeConfig) QueryActivatedHeight() uint64 {
	bz, err := n.QueryGRPCGateway("/babylon/btcstaking/v1/activated_height", url.Values{})
	require.NoError(n.t, err)

	var resp bstypes.QueryActivatedHeightResponse
	err = util.Cdc.UnmarshalJSON(bz, &resp)
	require.NoError(n.t, err)

	return resp.Height
}

// TODO: pagination support
// TODO: remove public randomness storage?
func (n *NodeConfig) QueryListPublicRandomness(fpBTCPK *bbn.BIP340PubKey) map[uint64]*bbn.SchnorrPubRand {
	path := fmt.Sprintf("/babylon/finality/v1/finality_providers/%s/public_randomness_list", fpBTCPK.MarshalHex())
	bz, err := n.QueryGRPCGateway(path, url.Values{})
	require.NoError(n.t, err)

	var resp ftypes.QueryListPublicRandomnessResponse
	err = util.Cdc.UnmarshalJSON(bz, &resp)
	require.NoError(n.t, err)

	return resp.PubRandMap
}

// TODO: pagination support
func (n *NodeConfig) QueryListPubRandCommit(fpBTCPK *bbn.BIP340PubKey) map[uint64]*ftypes.PubRandCommitResponse {
	path := fmt.Sprintf("/babylon/finality/v1/finality_providers/%s/pub_rand_commit_list", fpBTCPK.MarshalHex())
	bz, err := n.QueryGRPCGateway(path, url.Values{})
	require.NoError(n.t, err)

	var resp ftypes.QueryListPubRandCommitResponse
	err = util.Cdc.UnmarshalJSON(bz, &resp)
	require.NoError(n.t, err)

	return resp.PubRandCommitMap
}

func (n *NodeConfig) QueryVotesAtHeight(height uint64) []bbn.BIP340PubKey {
	path := fmt.Sprintf("/babylon/finality/v1/votes/%d", height)
	bz, err := n.QueryGRPCGateway(path, url.Values{})
	require.NoError(n.t, err)

	var resp ftypes.QueryVotesAtHeightResponse
	err = util.Cdc.UnmarshalJSON(bz, &resp)
	require.NoError(n.t, err)

	return resp.BtcPks
}

// TODO: pagination support
func (n *NodeConfig) QueryListBlocks(status ftypes.QueriedBlockStatus) []*ftypes.IndexedBlock {
	values := url.Values{}
	values.Set("status", fmt.Sprintf("%d", status))
	bz, err := n.QueryGRPCGateway("/babylon/finality/v1/blocks", values)
	require.NoError(n.t, err)

	var resp ftypes.QueryListBlocksResponse
	err = util.Cdc.UnmarshalJSON(bz, &resp)
	require.NoError(n.t, err)

	return resp.Blocks
}

func (n *NodeConfig) QueryIndexedBlock(height uint64) *ftypes.IndexedBlock {
	path := fmt.Sprintf("/babylon/finality/v1/blocks/%d", height)
	bz, err := n.QueryGRPCGateway(path, url.Values{})
	require.NoError(n.t, err)

	var resp ftypes.QueryBlockResponse
	err = util.Cdc.UnmarshalJSON(bz, &resp)
	require.NoError(n.t, err)

	return resp.Block
}
