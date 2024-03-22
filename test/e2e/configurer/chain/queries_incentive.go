package chain

import (
	"fmt"
	"net/url"

	"github.com/babylonchain/babylon/test/e2e/util"
	incentivetypes "github.com/babylonchain/babylon/x/incentive/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func (n *NodeConfig) QueryBTCStakingGauge(height uint64) (*incentivetypes.Gauge, error) {
	path := fmt.Sprintf("/babylon/incentive/btc_staking_gauge/%d", height)
	bz, err := n.QueryGRPCGateway(path, url.Values{})
	if err != nil {
		return nil, err
	}

	var resp incentivetypes.QueryBTCStakingGaugeResponse
	if err := util.Cdc.UnmarshalJSON(bz, &resp); err != nil {
		return nil, err
	}

	return resp.Gauge, nil
}

func (n *NodeConfig) QueryIncentiveParams() (*incentivetypes.Params, error) {
	path := "/babylon/incentive/params"
	bz, err := n.QueryGRPCGateway(path, url.Values{})
	require.NoError(n.t, err)

	var resp incentivetypes.QueryParamsResponse
	if err := util.Cdc.UnmarshalJSON(bz, &resp); err != nil {
		return nil, err
	}

	return &resp.Params, nil
}

func (n *NodeConfig) QueryRewardGauge(sAddr sdk.AccAddress) (map[string]*incentivetypes.RewardGauge, error) {
	path := fmt.Sprintf("/babylon/incentive/address/%s/reward_gauge", sAddr.String())
	bz, err := n.QueryGRPCGateway(path, url.Values{})
	if err != nil {
		return nil, err
	}
	var resp incentivetypes.QueryRewardGaugesResponse
	if err := util.Cdc.UnmarshalJSON(bz, &resp); err != nil {
		return nil, err
	}

	return resp.RewardGauges, nil
}

func (n *NodeConfig) QueryBTCTimestampingGauge(epoch uint64) (*incentivetypes.Gauge, error) {
	path := fmt.Sprintf("/babylon/incentive/btc_timestamping_gauge/%d", epoch)
	bz, err := n.QueryGRPCGateway(path, url.Values{})
	if err != nil {
		return nil, err
	}

	var resp incentivetypes.QueryBTCTimestampingGaugeResponse
	if err := util.Cdc.UnmarshalJSON(bz, &resp); err != nil {
		return nil, err
	}

	return resp.Gauge, nil
}
