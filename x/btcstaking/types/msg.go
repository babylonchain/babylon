package types

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"github.com/babylonchain/babylon/btcstaking"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ensure that these message types implement the sdk.Msg interface
var (
	_ sdk.Msg = &MsgUpdateParams{}
	_ sdk.Msg = &MsgCreateBTCValidator{}
	_ sdk.Msg = &MsgCreateBTCDelegation{}
	_ sdk.Msg = &MsgAddCovenantSig{}
	_ sdk.Msg = &MsgBTCUndelegate{}
	_ sdk.Msg = &MsgAddCovenantUnbondingSigs{}
	_ sdk.Msg = &MsgAddValidatorUnbondingSig{}
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
	if m.Commission == nil {
		return fmt.Errorf("empty commission")
	}
	if m.Description == nil {
		return fmt.Errorf("empty description")
	}
	if _, err := m.Description.EnsureLength(); err != nil {
		return err
	}
	if m.BabylonPk == nil {
		return fmt.Errorf("empty Babylon public key")
	}
	if m.BtcPk == nil {
		return fmt.Errorf("empty BTC public key")
	}
	if m.Pop == nil {
		return fmt.Errorf("empty proof of possession")
	}
	if _, err := sdk.AccAddressFromBech32(m.Signer); err != nil {
		return err
	}
	if err := m.Pop.ValidateBasic(); err != nil {
		return err
	}

	return nil
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
		return fmt.Errorf("empty Babylon public key")
	}
	if m.Pop == nil {
		return fmt.Errorf("empty proof of possession")
	}
	if m.StakingTx == nil {
		return fmt.Errorf("empty staking tx")
	}
	if m.StakingTxInfo == nil {
		return fmt.Errorf("empty staking tx info")
	}
	if m.SlashingTx == nil {
		return fmt.Errorf("empty slashing tx")
	}
	if m.DelegatorSig == nil {
		return fmt.Errorf("empty delegator signature")
	}
	if _, err := sdk.AccAddressFromBech32(m.Signer); err != nil {
		return err
	}

	// staking tx should be correctly formatted
	if err := m.StakingTx.ValidateBasic(); err != nil {
		return err
	}
	if err := m.Pop.ValidateBasic(); err != nil {
		return err
	}

	return nil
}

func (m *MsgAddCovenantSig) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(m.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

func (m *MsgAddCovenantSig) ValidateBasic() error {
	if m.ValPk == nil {
		return fmt.Errorf("empty BTC validator public key")
	}
	if m.DelPk == nil {
		return fmt.Errorf("empty BTC delegation public key")
	}
	if m.Sig == nil {
		return fmt.Errorf("empty covenant signature")
	}
	if len(m.StakingTxHash) != chainhash.MaxHashStringSize {
		return fmt.Errorf("staking tx hash is not %d", chainhash.MaxHashStringSize)
	}

	return nil
}

func (m *MsgBTCUndelegate) ValidateBasic() error {
	if m.UnbondingTx == nil {
		return fmt.Errorf("empty unbodning tx")
	}
	if m.SlashingTx == nil {
		return fmt.Errorf("empty slashing tx")
	}
	if m.DelegatorSlashingSig == nil {
		return fmt.Errorf("empty delegator signature")
	}
	if _, err := sdk.AccAddressFromBech32(m.Signer); err != nil {
		return err
	}

	unbondingTxMsg, err := m.UnbondingTx.ToMsgTx()

	if err != nil {
		return err
	}

	if err := btcstaking.IsSimpleTransfer(unbondingTxMsg); err != nil {
		return err
	}

	// unbodning tx should be correctly formatted - valid transaction and valid script
	if err := m.UnbondingTx.ValidateBasic(); err != nil {
		return err
	}

	return nil
}

func (m *MsgBTCUndelegate) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(m.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

func (m *MsgAddCovenantUnbondingSigs) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(m.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

func (m *MsgAddCovenantUnbondingSigs) ValidateBasic() error {
	if m.ValPk == nil {
		return fmt.Errorf("empty BTC validator public key")
	}
	if m.DelPk == nil {
		return fmt.Errorf("empty BTC delegation public key")
	}
	if m.UnbondingTxSig == nil {
		return fmt.Errorf("empty covenant signature")
	}
	if m.SlashingUnbondingTxSig == nil {
		return fmt.Errorf("empty covenant signature")
	}
	if len(m.StakingTxHash) != chainhash.MaxHashStringSize {
		return fmt.Errorf("staking tx hash is not %d", chainhash.MaxHashStringSize)
	}
	return nil
}

func (m *MsgAddValidatorUnbondingSig) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(m.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

func (m *MsgAddValidatorUnbondingSig) ValidateBasic() error {
	if m.ValPk == nil {
		return fmt.Errorf("empty BTC validator public key")
	}
	if m.DelPk == nil {
		return fmt.Errorf("empty BTC delegation public key")
	}
	if m.UnbondingTxSig == nil {
		return fmt.Errorf("empty covenant signature")
	}
	if len(m.StakingTxHash) != chainhash.MaxHashStringSize {
		return fmt.Errorf("staking tx hash is not %d", chainhash.MaxHashStringSize)
	}

	return nil
}
