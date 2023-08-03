package chain

import (
	"fmt"
	"net/url"

	"github.com/stretchr/testify/require"

	"github.com/babylonchain/babylon/test/e2e/util"
	bbn "github.com/babylonchain/babylon/types"
	bstypes "github.com/babylonchain/babylon/x/btcstaking/types"
	ftypes "github.com/babylonchain/babylon/x/finality/types"
)

func (n *NodeConfig) QueryBTCStakingParams() *bstypes.Params {
	bz, err := n.QueryGRPCGateway("/babylon/btcstaking/v1/params", url.Values{})
	require.NoError(n.t, err)

	var resp bstypes.QueryParamsResponse
	err = util.Cdc.UnmarshalJSON(bz, &resp)
	require.NoError(n.t, err)

	return &resp.Params
}

func (n *NodeConfig) QueryBTCValidators() []*bstypes.BTCValidator {
	bz, err := n.QueryGRPCGateway("/babylon/btcstaking/v1/btc_validators", url.Values{})
	require.NoError(n.t, err)

	var resp bstypes.QueryBTCValidatorsResponse
	err = util.Cdc.UnmarshalJSON(bz, &resp)
	require.NoError(n.t, err)

	return resp.BtcValidators
}

func (n *NodeConfig) QueryActiveBTCValidatorsAtHeight(height uint64) []*bstypes.BTCValidatorWithMeta {
	path := fmt.Sprintf("/babylon/btcstaking/v1/btc_validators/%d", height)
	bz, err := n.QueryGRPCGateway(path, url.Values{})
	require.NoError(n.t, err)

	var resp bstypes.QueryActiveBTCValidatorsAtHeightResponse
	err = util.Cdc.UnmarshalJSON(bz, &resp)
	require.NoError(n.t, err)

	return resp.BtcValidators
}

func (n *NodeConfig) QueryBTCValidatorDelegations(valBTCPK string) []*bstypes.BTCDelegations {
	path := fmt.Sprintf("/babylon/btcstaking/v1/btc_validators/%s/delegations", valBTCPK)
	bz, err := n.QueryGRPCGateway(path, url.Values{})
	require.NoError(n.t, err)

	var resp bstypes.QueryBTCValidatorDelegationsResponse
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
func (n *NodeConfig) QueryListPublicRandomness(valBTCPK *bbn.BIP340PubKey) map[uint64]*bbn.SchnorrPubRand {
	path := fmt.Sprintf("/babylon/finality/v1/btc_validators/%s/public_randomness_list", valBTCPK.MarshalHex())
	bz, err := n.QueryGRPCGateway(path, url.Values{})
	require.NoError(n.t, err)

	var resp ftypes.QueryListPublicRandomnessResponse
	err = util.Cdc.UnmarshalJSON(bz, &resp)
	require.NoError(n.t, err)

	return resp.PubRandMap
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
