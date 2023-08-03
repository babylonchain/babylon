package types

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ensure that these message types implement the sdk.Msg interface
var (
	_ sdk.Msg = &MsgUpdateParams{}
	_ sdk.Msg = &MsgCreateBTCValidator{}
	_ sdk.Msg = &MsgCreateBTCDelegation{}
	_ sdk.Msg = &MsgAddJurySig{}
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

	// verify PoP
	if err := m.Pop.ValidateBasic(); err != nil {
		return err
	}
	stakingScriptData, err := m.StakingTx.GetStakingScriptData()
	if err != nil {
		return err
	}
	btcPK := bbn.NewBIP340PubKeyFromBTCPK(stakingScriptData.StakerKey)
	if err := m.Pop.Verify(m.BabylonPk, btcPK); err != nil {
		return err
	}

	return nil
}

func (m *MsgAddJurySig) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(m.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

func (m *MsgAddJurySig) ValidateBasic() error {
	if m.ValPk == nil {
		return fmt.Errorf("empty BTC validator public key")
	}
	if m.DelPk == nil {
		return fmt.Errorf("empty BTC delegation public key")
	}
	if m.Sig == nil {
		return fmt.Errorf("empty jury signature")
	}
	if len(m.StakingTxHash) != chainhash.MaxHashStringSize {
		return fmt.Errorf("staking tx hash is not %d", chainhash.MaxHashStringSize)
	}

	return nil
}
