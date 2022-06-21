package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// staking message types
const (
	TypeMsgCreateValidatorBLS     = "create_validator_bls"
	TypeMsgWrappedDelegate        = "wrapped_delegate"
	TypeMsgWrappedUndelegate      = "wrapped_begin_unbonding"
	TypeMsgWrappedBeginRedelegate = "wrapped_begin_redelegate"
)

var (
	_ sdk.Msg = &MsgCreateValidatorBLS{}
	_ sdk.Msg = &MsgWrappedDelegate{}
	_ sdk.Msg = &MsgWrappedUndelegate{}
	_ sdk.Msg = &MsgWrappedBeginRedelegate{}
)

// NewMsgCreateValidatorBLS creates a new MsgCreateValidatorBLS instance.
func NewMsgCreateValidatorBLS(
	msg *stakingtypes.MsgCreateValidator,
) (*MsgCreateValidatorBLS, error) {
	return &MsgCreateValidatorBLS{
		Msg: msg,
	}, nil
}

// Route implements the sdk.Msg interface.
func (msg MsgCreateValidatorBLS) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
func (msg MsgCreateValidatorBLS) Type() string { return TypeMsgCreateValidatorBLS }

// GetSigners implements the sdk.Msg interface. It returns the address(es) that
// must sign over msg.GetSignBytes().
// If the validator address is not same as delegator's, then the validator must
// sign the msg as well.
func (msg MsgCreateValidatorBLS) GetSigners() []sdk.AccAddress {
	return msg.Msg.GetSigners()
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgCreateValidatorBLS) GetSignBytes() []byte {
	return msg.Msg.GetSignBytes()
}

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgCreateValidatorBLS) ValidateBasic() error {
	return msg.Msg.ValidateBasic()
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg MsgCreateValidatorBLS) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return msg.Msg.UnpackInterfaces(unpacker)
}

// NewMsgWrappedDelegate creates a new MsgWrappedDelegate instance.
func NewMsgWrappedDelegate(
	msg *stakingtypes.MsgDelegate,
) (*MsgWrappedDelegate, error) {
	return &MsgWrappedDelegate{
		Msg: msg,
	}, nil
}

// Route implements the sdk.Msg interface.
func (msg MsgWrappedDelegate) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
func (msg MsgWrappedDelegate) Type() string { return TypeMsgWrappedDelegate }

// GetSigners implements the sdk.Msg interface. It returns the address(es) that
// must sign over msg.GetSignBytes().
// If the validator address is not same as delegator's, then the validator must
// sign the msg as well.
func (msg MsgWrappedDelegate) GetSigners() []sdk.AccAddress {
	return msg.Msg.GetSigners()
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgWrappedDelegate) GetSignBytes() []byte {
	return msg.Msg.GetSignBytes()
}

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgWrappedDelegate) ValidateBasic() error {
	return msg.Msg.ValidateBasic()
}

// NewMsgWrappedUndelegate creates a new MsgWrappedUndelegate instance.
func NewMsgWrappedUndelegate(
	msg *stakingtypes.MsgUndelegate,
) (*MsgWrappedUndelegate, error) {
	return &MsgWrappedUndelegate{
		Msg: msg,
	}, nil
}

// Route implements the sdk.Msg interface.
func (msg MsgWrappedUndelegate) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
func (msg MsgWrappedUndelegate) Type() string { return TypeMsgWrappedUndelegate }

// GetSigners implements the sdk.Msg interface. It returns the address(es) that
// must sign over msg.GetSignBytes().
// If the validator address is not same as delegator's, then the validator must
// sign the msg as well.
func (msg MsgWrappedUndelegate) GetSigners() []sdk.AccAddress {
	return msg.Msg.GetSigners()
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgWrappedUndelegate) GetSignBytes() []byte {
	return msg.Msg.GetSignBytes()
}

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgWrappedUndelegate) ValidateBasic() error {
	return msg.Msg.ValidateBasic()
}

// NewMsgWrappedBeginRedelegate creates a new MsgWrappedBeginRedelegate instance.
func NewMsgWrappedBeginRedelegate(
	msg *stakingtypes.MsgBeginRedelegate,
) (*MsgWrappedBeginRedelegate, error) {
	return &MsgWrappedBeginRedelegate{
		Msg: msg,
	}, nil
}

// Route implements the sdk.Msg interface.
func (msg MsgWrappedBeginRedelegate) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
func (msg MsgWrappedBeginRedelegate) Type() string { return TypeMsgWrappedBeginRedelegate }

// GetSigners implements the sdk.Msg interface. It returns the address(es) that
// must sign over msg.GetSignBytes().
// If the validator address is not same as delegator's, then the validator must
// sign the msg as well.
func (msg MsgWrappedBeginRedelegate) GetSigners() []sdk.AccAddress {
	return msg.Msg.GetSigners()
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgWrappedBeginRedelegate) GetSignBytes() []byte {
	return msg.Msg.GetSignBytes()
}

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgWrappedBeginRedelegate) ValidateBasic() error {
	return msg.Msg.ValidateBasic()
}
