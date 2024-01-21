package keeper_test

import (
	"testing"

	"github.com/babylonchain/babylon/x/incentive/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestEmptyRewardGauge(t *testing.T) {
	emptyRewardGauge := &types.RewardGauge{
		Coins:          sdk.NewCoins(),
		WithdrawnCoins: sdk.NewCoins(),
	}
	rgBytes, err := emptyRewardGauge.Marshal()
	require.NoError(t, err)
	require.NotNil(t, rgBytes)         // the marshaled empty reward gauge is not nil
	require.True(t, len(rgBytes) == 0) // the marshalled empty reward gauge has 0 bytes
}
