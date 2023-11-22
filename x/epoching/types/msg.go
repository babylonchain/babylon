package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// staking message types
const (
	TypeMsgWrappedDelegate        = "wrapped_delegate"
	TypeMsgWrappedUndelegate      = "wrapped_begin_unbonding"
	TypeMsgWrappedBeginRedelegate = "wrapped_begin_redelegate"
)

// ensure that these message types implement the sdk.Msg interface
var (
	_ sdk.Msg = &MsgWrappedDelegate{}
	_ sdk.Msg = &MsgWrappedUndelegate{}
	_ sdk.Msg = &MsgWrappedBeginRedelegate{}
	_ sdk.Msg = &MsgUpdateParams{}
)

// NewMsgWrappedDelegate creates a new MsgWrappedDelegate instance.
func NewMsgWrappedDelegate(msg *stakingtypes.MsgDelegate) *MsgWrappedDelegate {
	return &MsgWrappedDelegate{
		Msg: msg,
	}
}

// Route implements the sdk.Msg interface.
func (msg MsgWrappedDelegate) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
func (msg MsgWrappedDelegate) Type() string { return TypeMsgWrappedDelegate }

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgWrappedDelegate) ValidateBasic() error {
	if msg.Msg == nil {
		return ErrNoWrappedMsg
	}
	return nil
}

// NewMsgWrappedUndelegate creates a new MsgWrappedUndelegate instance.
func NewMsgWrappedUndelegate(msg *stakingtypes.MsgUndelegate) *MsgWrappedUndelegate {
	return &MsgWrappedUndelegate{
		Msg: msg,
	}
}

// Route implements the sdk.Msg interface.
func (msg MsgWrappedUndelegate) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
func (msg MsgWrappedUndelegate) Type() string { return TypeMsgWrappedUndelegate }

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgWrappedUndelegate) ValidateBasic() error {
	if msg.Msg == nil {
		return ErrNoWrappedMsg
	}
	return nil
}

// NewMsgWrappedBeginRedelegate creates a new MsgWrappedBeginRedelegate instance.
func NewMsgWrappedBeginRedelegate(msg *stakingtypes.MsgBeginRedelegate) *MsgWrappedBeginRedelegate {
	return &MsgWrappedBeginRedelegate{
		Msg: msg,
	}
}

// Route implements the sdk.Msg interface.
func (msg MsgWrappedBeginRedelegate) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
func (msg MsgWrappedBeginRedelegate) Type() string { return TypeMsgWrappedBeginRedelegate }

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgWrappedBeginRedelegate) ValidateBasic() error {
	if msg.Msg == nil {
		return ErrNoWrappedMsg
	}
	return nil
}

// GetSigners returns the expected signers for a MsgUpdateParams message.
func (m *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Authority)
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check on the provided data.
func (m *MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return errorsmod.Wrap(err, "invalid authority address")
	}

	if err := m.Params.Validate(); err != nil {
		return err
	}

	return nil
}
