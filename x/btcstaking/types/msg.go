package types

import (
	"fmt"
	math "math"

	"github.com/babylonchain/babylon/btcstaking"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ensure that these message types implement the sdk.Msg interface
var (
	_ sdk.Msg = &MsgUpdateParams{}
	_ sdk.Msg = &MsgCreateFinalityProvider{}
	_ sdk.Msg = &MsgEditFinalityProvider{}
	_ sdk.Msg = &MsgCreateBTCDelegation{}
	_ sdk.Msg = &MsgAddCovenantSigs{}
	_ sdk.Msg = &MsgBTCUndelegate{}
)

func (m *MsgCreateFinalityProvider) ValidateBasic() error {
	if m.Commission == nil {
		return fmt.Errorf("empty commission")
	}
	if m.Description == nil {
		return fmt.Errorf("empty description")
	}
	if len(m.Description.Moniker) == 0 {
		return fmt.Errorf("empty moniker")
	}
	if _, err := m.Description.EnsureLength(); err != nil {
		return err
	}
	if m.BtcPk == nil {
		return fmt.Errorf("empty BTC public key")
	}
	if _, err := m.BtcPk.ToBTCPK(); err != nil {
		return fmt.Errorf("invalid BTC public key: %v", err)
	}
	if m.Pop == nil {
		return fmt.Errorf("empty proof of possession")
	}
	if _, err := sdk.AccAddressFromBech32(m.Addr); err != nil {
		return err
	}
	return m.Pop.ValidateBasic()
}

func (m *MsgEditFinalityProvider) ValidateBasic() error {
	if m.Commission == nil {
		return fmt.Errorf("empty commission")
	}
	if m.Description == nil {
		return fmt.Errorf("empty description")
	}
	if len(m.Description.Moniker) == 0 {
		return fmt.Errorf("empty moniker")
	}
	if _, err := m.Description.EnsureLength(); err != nil {
		return err
	}
	if len(m.BtcPk) != bbn.BIP340PubKeyLen {
		return fmt.Errorf("malformed BTC PK")
	}
	if _, err := bbn.NewBIP340PubKey(m.BtcPk); err != nil {
		return err
	}

	return nil
}

func (m *MsgCreateBTCDelegation) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.StakerAddr); err != nil {
		return fmt.Errorf("invalid staker addr %s: %w", m.StakerAddr, err)
	}
	if m.Pop == nil {
		return fmt.Errorf("empty proof of possession")
	}
	if m.BtcPk == nil {
		return fmt.Errorf("empty delegator BTC public key")
	}
	if _, err := m.BtcPk.ToBTCPK(); err != nil {
		return fmt.Errorf("invalid BTC public key: %v", err)
	}
	if m.StakingTx == nil {
		return fmt.Errorf("empty staking tx info")
	}
	if m.SlashingTx == nil {
		return fmt.Errorf("empty slashing tx")
	}

	if _, err := m.SlashingTx.ToMsgTx(); err != nil {
		return fmt.Errorf("invalid slashing tx: %w", err)
	}

	if m.DelegatorSlashingSig == nil {
		return fmt.Errorf("empty delegator signature")
	}

	if _, err := m.DelegatorSlashingSig.ToBTCSig(); err != nil {
		return fmt.Errorf("invalid delegator slashing signature: %w", err)
	}

	// Check staking time is at most uint16
	if m.StakingTime > math.MaxUint16 {
		return ErrInvalidStakingTx.Wrapf("invalid lock time: %d, max: %d", m.StakingTime, math.MaxUint16)
	}
	// Ensure list of finality provider BTC PKs is not empty
	if len(m.FpBtcPkList) == 0 {
		return ErrEmptyFpList
	}
	// Ensure list of finality provider BTC PKs is not duplicated
	if ExistsDup(m.FpBtcPkList) {
		return ErrDuplicatedFp
	}

	// staking tx should be correctly formatted
	if err := m.StakingTx.ValidateBasic(); err != nil {
		return err
	}
	if err := m.Pop.ValidateBasic(); err != nil {
		return err
	}

	// verifications about on-demand unbonding
	if m.UnbondingTx == nil {
		return fmt.Errorf("empty unbonding tx")
	}
	if m.UnbondingSlashingTx == nil {
		return fmt.Errorf("empty slashing tx")
	}
	if m.DelegatorUnbondingSlashingSig == nil {
		return fmt.Errorf("empty delegator signature")
	}

	if _, err := m.UnbondingSlashingTx.ToMsgTx(); err != nil {
		return fmt.Errorf("invalid unbonding slashing tx: %w", err)
	}

	if _, err := m.DelegatorUnbondingSlashingSig.ToBTCSig(); err != nil {
		return fmt.Errorf("invalid delegator unbonding slashing signature: %w", err)
	}

	unbondingTxMsg, err := bbn.NewBTCTxFromBytes(m.UnbondingTx)
	if err != nil {
		return err
	}
	if err := btcstaking.IsSimpleTransfer(unbondingTxMsg); err != nil {
		return err
	}

	// Check unbonding time is lower than max uint16
	if m.UnbondingTime > math.MaxUint16 {
		return ErrInvalidUnbondingTx.Wrapf("unbonding time %d must be lower than %d", m.UnbondingTime, math.MaxUint16)
	}

	return nil
}

func (m *MsgAddCovenantSigs) ValidateBasic() error {
	if m.Pk == nil {
		return fmt.Errorf("empty BTC covenant public key")
	}
	if _, err := m.Pk.ToBTCPK(); err != nil {
		return fmt.Errorf("invalid BTC public key: %v", err)
	}
	if m.SlashingTxSigs == nil {
		return fmt.Errorf("empty covenant signatures on slashing tx")
	}
	if len(m.StakingTxHash) != chainhash.MaxHashStringSize {
		return fmt.Errorf("staking tx hash is not %d", chainhash.MaxHashStringSize)
	}

	// verifications about on-demand unbonding
	if m.UnbondingTxSig == nil {
		return fmt.Errorf("empty covenant signature")
	}

	if _, err := m.UnbondingTxSig.ToBTCSig(); err != nil {
		return fmt.Errorf("invalid covenant unbonding signature: %w", err)
	}

	if m.SlashingUnbondingTxSigs == nil {
		return fmt.Errorf("empty covenant signature")
	}

	return nil
}

func (m *MsgBTCUndelegate) ValidateBasic() error {
	if len(m.StakingTxHash) != chainhash.MaxHashStringSize {
		return fmt.Errorf("staking tx hash is not %d", chainhash.MaxHashStringSize)
	}

	if m.UnbondingTxSig == nil {
		return fmt.Errorf("empty signature from the delegator")
	}

	if _, err := m.UnbondingTxSig.ToBTCSig(); err != nil {
		return fmt.Errorf("invalid delegator unbonding signature: %w", err)
	}

	return nil
}
