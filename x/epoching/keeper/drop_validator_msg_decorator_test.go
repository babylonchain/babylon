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
		{&stakingtypes.MsgCreateValidator{}, false},
		{&stakingtypes.MsgDelegate{}, false},
		{&stakingtypes.MsgUndelegate{}, false},
		{&stakingtypes.MsgBeginRedelegate{}, false},
		{&stakingtypes.MsgEditValidator{}, true},
	}

	decorator := NewDropValidatorMsgDecorator()

	for _, tc := range testCases {
		err := decorator.IsValidatorRelatedMsg(tc.msg)
		if tc.expectPass {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
		}
	}
}
