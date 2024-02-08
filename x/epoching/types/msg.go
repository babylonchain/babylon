package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// ensure that these message types implement the sdk.Msg interface
var (
	_ sdk.Msg = &MsgWrappedDelegate{}
	_ sdk.Msg = &MsgWrappedUndelegate{}
	_ sdk.Msg = &MsgWrappedBeginRedelegate{}
	_ sdk.Msg = &MsgWrappedCancelUnbondingDelegation{}
	_ sdk.Msg = &MsgUpdateParams{}
)

// NewMsgWrappedDelegate creates a new MsgWrappedDelegate instance.
func NewMsgWrappedDelegate(msg *stakingtypes.MsgDelegate) *MsgWrappedDelegate {
	return &MsgWrappedDelegate{
		Msg: msg,
	}
}

// NewMsgWrappedUndelegate creates a new MsgWrappedUndelegate instance.
func NewMsgWrappedUndelegate(msg *stakingtypes.MsgUndelegate) *MsgWrappedUndelegate {
	return &MsgWrappedUndelegate{
		Msg: msg,
	}
}

// NewMsgWrappedBeginRedelegate creates a new MsgWrappedBeginRedelegate instance.
func NewMsgWrappedBeginRedelegate(msg *stakingtypes.MsgBeginRedelegate) *MsgWrappedBeginRedelegate {
	return &MsgWrappedBeginRedelegate{
		Msg: msg,
	}
}

// NewMsgWrappedCancelUnbondingDelegation creates a new MsgWrappedCancelUnbondingDelegation instance.
func NewMsgWrappedCancelUnbondingDelegation(msg *stakingtypes.MsgCancelUnbondingDelegation) *MsgWrappedCancelUnbondingDelegation {
	return &MsgWrappedCancelUnbondingDelegation{
		Msg: msg,
	}
}
