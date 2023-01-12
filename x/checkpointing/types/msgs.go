package types

import (
	"errors"

	ed255192 "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/babylonchain/babylon/crypto/bls12381"
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

func NewMsgWrappedCreateValidator(msgCreateVal *stakingtypes.MsgCreateValidator, blsPK *bls12381.PublicKey, pop *ProofOfPossession) (*MsgWrappedCreateValidator, error) {
	return &MsgWrappedCreateValidator{
		Key: &BlsKey{
			Pubkey: blsPK,
			Pop:    pop,
		},
		MsgCreateValidator: msgCreateVal,
	}, nil
}

// ValidateBasic validates stateless message elements
func (m *MsgAddBlsSig) ValidateBasic() error {
	// This function validates stateless message elements
	_, err := sdk.ValAddressFromBech32(m.BlsSig.SignerAddress)
	if err != nil {
		return err
	}

	err = m.BlsSig.BlsSig.ValidateBasic()
	if err != nil {
		return err
	}
	err = m.BlsSig.LastCommitHash.ValidateBasic()
	if err != nil {
		return err
	}

	return nil
}

func (m *MsgAddBlsSig) GetSigners() []sdk.AccAddress {
	signer, err := sdk.ValAddressFromBech32(m.BlsSig.SignerAddress)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{sdk.AccAddress(signer)}
}

func (m *MsgWrappedCreateValidator) VerifyPoP(valPubkey cryptotypes.PubKey) bool {
	return m.Key.Pop.IsValid(*m.Key.Pubkey, valPubkey)
}

// ValidateBasic validates statelesss message elements
func (m *MsgWrappedCreateValidator) ValidateBasic() error {
	if m.MsgCreateValidator == nil {
		return errors.New("MsgCreateValidator is nil")
	}
	err := m.MsgCreateValidator.ValidateBasic()
	if err != nil {
		return err
	}
	var pubKey ed255192.PubKey
	err = pubKey.Unmarshal(m.MsgCreateValidator.Pubkey.GetValue())
	if err != nil {
		return err
	}
	ok := m.VerifyPoP(&pubKey)
	if !ok {
		return errors.New("the proof-of-possession is not valid")
	}

	return nil
}

func (m *MsgWrappedCreateValidator) GetSigners() []sdk.AccAddress {
	return m.MsgCreateValidator.GetSigners()
}
