package types

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ensure that these message types implement the sdk.Msg interface
var (
	_ sdk.Msg = &MsgUpdateParams{}
	_ sdk.Msg = &MsgCreateBTCValidator{}
	_ sdk.Msg = &MsgCreateBTCDelegation{}
)

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

func (m *MsgCreateBTCValidator) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(m.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

func (m *MsgCreateBTCValidator) ValidateBasic() error {
	if m.BabylonPk == nil {
		return fmt.Errorf("empty BabylonPk")
	}
	if m.BtcPk == nil {
		return fmt.Errorf("empty BtcPk")
	}
	if m.Pop == nil {
		return fmt.Errorf("empty Pop")
	}
	if _, err := sdk.AccAddressFromBech32(m.Signer); err != nil {
		return err
	}
	if err := m.Pop.ValidateBasic(); err != nil {
		return err
	}
	return m.Pop.Verify(m.BabylonPk, m.BtcPk)
}

func (m *MsgCreateBTCDelegation) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(m.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

func (m *MsgCreateBTCDelegation) ValidateBasic() error {
	if m.BabylonPk == nil {
		return fmt.Errorf("empty BabylonPk")
	}
	if m.Pop == nil {
		return fmt.Errorf("empty Pop")
	}
	if m.StakingTx == nil {
		return fmt.Errorf("empty StakingTx")
	}
	if m.StakingTxSig == nil {
		return fmt.Errorf("empty StakingTxSig")
	}
	if m.StakingTxInfo == nil {
		return fmt.Errorf("empty StakingTxInfo")
	}
	if m.SlashingTx == nil {
		return fmt.Errorf("empty SlashingTx")
	}
	if m.SlashingTxSig == nil {
		return fmt.Errorf("empty SlashingTxSig")
	}
	if _, err := sdk.AccAddressFromBech32(m.Signer); err != nil {
		return err
	}

	// TODO: verification rules

	if err := m.Pop.ValidateBasic(); err != nil {
		return err
	}
	// TODO: extract BTC PK and verify PoP
	return nil
}
