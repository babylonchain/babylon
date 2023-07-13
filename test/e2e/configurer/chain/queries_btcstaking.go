package chain

import (
	"encoding/base64"
	"fmt"
	"net/url"

	"github.com/babylonchain/babylon/test/e2e/util"
	bbn "github.com/babylonchain/babylon/types"
	bstypes "github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/stretchr/testify/require"
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

func (n *NodeConfig) QueryBTCValidatorsAtHeight(height uint64) []*bstypes.BTCValidatorWithMeta {
	path := fmt.Sprintf("/babylonchain/babylon/btcstaking/v1/btc_validators/%d", height)
	bz, err := n.QueryGRPCGateway(path, url.Values{})
	require.NoError(n.t, err)

	var resp bstypes.QueryBTCValidatorsAtHeightResponse
	err = util.Cdc.UnmarshalJSON(bz, &resp)
	require.NoError(n.t, err)

	return resp.BtcValidators
}

func (n *NodeConfig) QueryBTCValidatorDelegationsAtHeight(valBTCPK *bbn.BIP340PubKey, height uint64) []*bstypes.BTCDelegationWithMeta {
	path := fmt.Sprintf("/babylonchain/babylon/btcstaking/v1/delegations/%d", height)
	valBTCPKStr := base64.URLEncoding.EncodeToString(valBTCPK.MustMarshal())
	bz, err := n.QueryGRPCGateway(path, url.Values{
		"val_btc_pk": []string{valBTCPKStr},
	})
	require.NoError(n.t, err)

	var resp bstypes.QueryBTCValidatorDelegationsAtHeightResponse
	err = util.Cdc.UnmarshalJSON(bz, &resp)
	require.NoError(n.t, err)

	return resp.BtcDelegations
}
