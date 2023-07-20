package chain

import (
	"fmt"
	"github.com/babylonchain/babylon/test/e2e/util"
	bstypes "github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/stretchr/testify/require"
	"net/url"
)

func (n *NodeConfig) QueryBTCStakingParams() *bstypes.Params {
	bz, err := n.QueryGRPCGateway("/babylonchain/babylon/btcstaking/v1/params", url.Values{})
	require.NoError(n.t, err)

	var resp bstypes.QueryParamsResponse
	err = util.Cdc.UnmarshalJSON(bz, &resp)
	require.NoError(n.t, err)

	return &resp.Params
}

func (n *NodeConfig) QueryBTCValidators() []*bstypes.BTCValidator {
	bz, err := n.QueryGRPCGateway("/babylonchain/babylon/btcstaking/v1/btc_validators", url.Values{})
	require.NoError(n.t, err)

	var resp bstypes.QueryBTCValidatorsResponse
	err = util.Cdc.UnmarshalJSON(bz, &resp)
	require.NoError(n.t, err)

	return resp.BtcValidators
}

func (n *NodeConfig) QueryActiveBTCValidatorsAtHeight(height uint64) []*bstypes.BTCValidatorWithMeta {
	path := fmt.Sprintf("/babylonchain/babylon/btcstaking/v1/btc_validators/%d", height)
	bz, err := n.QueryGRPCGateway(path, url.Values{})
	require.NoError(n.t, err)

	var resp bstypes.QueryActiveBTCValidatorsAtHeightResponse
	err = util.Cdc.UnmarshalJSON(bz, &resp)
	require.NoError(n.t, err)

	return resp.BtcValidators
}

func (n *NodeConfig) QueryBTCValidatorDelegations(valBTCPK string, status bstypes.BTCDelegationStatus) []*bstypes.BTCDelegation {
	path := fmt.Sprintf("/babylonchain/babylon/btcstaking/v1/btc_validators/%s/delegations", valBTCPK)
	values := url.Values{}
	values.Set("del_status", fmt.Sprintf("%d", status))
	bz, err := n.QueryGRPCGateway(path, values)
	require.NoError(n.t, err)

	var resp bstypes.QueryBTCValidatorDelegationsResponse
	err = util.Cdc.UnmarshalJSON(bz, &resp)
	require.NoError(n.t, err)

	return resp.BtcDelegations
}
