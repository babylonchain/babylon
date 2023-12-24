package types

import (
	"errors"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	ed255192 "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/babylonchain/babylon/crypto/bls12381"
)

var (
	// Ensure that MsgInsertHeader implements all functions of the Msg interface
	_ sdk.Msg = (*MsgWrappedCreateValidator)(nil)
)

func NewMsgWrappedCreateValidator(msgCreateVal *stakingtypes.MsgCreateValidator, blsPK *bls12381.PublicKey, pop *ProofOfPossession) (*MsgWrappedCreateValidator, error) {
	return &MsgWrappedCreateValidator{
		Key: &BlsKey{
			Pubkey: blsPK,
			Pop:    pop,
		},
		MsgCreateValidator: msgCreateVal,
	}, nil
}

func (m *MsgWrappedCreateValidator) VerifyPoP(valPubkey cryptotypes.PubKey) bool {
	return m.Key.Pop.IsValid(*m.Key.Pubkey, valPubkey)
}

// ValidateBasic validates statelesss message elements
func (m *MsgWrappedCreateValidator) ValidateBasic() error {
	if m.MsgCreateValidator == nil {
		return errors.New("MsgCreateValidator is nil")
	}
	var pubKey ed255192.PubKey
	err := pubKey.Unmarshal(m.MsgCreateValidator.Pubkey.GetValue())
	if err != nil {
		return err
	}
	ok := m.VerifyPoP(&pubKey)
	if !ok {
		return errors.New("the proof-of-possession is not valid")
	}

	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
// Needed since msg.MsgCreateValidator.Pubkey is in type Any
func (msg MsgWrappedCreateValidator) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return msg.MsgCreateValidator.UnpackInterfaces(unpacker)
}
