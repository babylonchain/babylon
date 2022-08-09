package types

import (
	"github.com/babylonchain/babylon/crypto/bls12381"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var (
	// Ensure that MsgInsertHeader implements all functions of the Msg interface
	_ sdk.Msg = (*MsgAddBlsSig)(nil)
)

func NewMsgAddBlsSig(epochNum uint64, lch LastCommitHash, sig bls12381.Signature, addr sdk.ValAddress) *MsgAddBlsSig {
	return &MsgAddBlsSig{BlsSig: &BlsSig{
		EpochNum:       epochNum,
		LastCommitHash: &lch,
		BlsSig:         &sig,
		SignerAddress:  addr.String(),
	}}
}

func NewMsgWrappedCreateValidator(msgCreateVal *stakingtypes.MsgCreateValidator) *MsgWrappedCreateValidator {
	return &MsgWrappedCreateValidator{

		MsgCreateValidator: msgCreateVal,
	}

	return &MsgWrappedCreateValidator{
		Key: &BlsKey{
			Pubkey: nil,
			Pop:    nil,
		},
		MsgCreateValidator: nil,
	}
}

func (m *MsgAddBlsSig) ValidateBasic() error {
	// This function validates stateless message elements
	_, err := sdk.AccAddressFromBech32(m.BlsSig.SignerAddress)
	if err != nil {
		return err
	}

	// TODO: verify bls sig

	return nil
}

func (m *MsgAddBlsSig) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(m.BlsSig.SignerAddress)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{signer}
}

func (m *MsgWrappedCreateValidator) ValidateBasic() error {
	// This function validates stateless message elements
	// TODO: verify bls sig

	return m.MsgCreateValidator.ValidateBasic()
}

func (m *MsgWrappedCreateValidator) GetSigners() []sdk.AccAddress {
	return m.MsgCreateValidator.GetSigners()
}
