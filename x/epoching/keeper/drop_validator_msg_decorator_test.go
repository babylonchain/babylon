package keeper

import (
	"testing"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestDropValidatorMsgDecorator(t *testing.T) {
	testCases := []struct {
		msg        sdk.Msg
		expectPass bool
	}{
		// wrapped message types that should be rejected
		{&stakingtypes.MsgCreateValidator{}, true},
		{&stakingtypes.MsgDelegate{}, true},
		{&stakingtypes.MsgUndelegate{}, true},
		{&stakingtypes.MsgBeginRedelegate{}, true},
		// allowed message types
		{&stakingtypes.MsgEditValidator{}, false},
	}

	decorator := NewDropValidatorMsgDecorator()

	for _, tc := range testCases {
		res := decorator.IsValidatorRelatedMsg(tc.msg)
		if tc.expectPass {
			require.True(t, res)
		} else {
			require.False(t, res)
		}
	}
}
