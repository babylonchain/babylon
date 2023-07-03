package types

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"github.com/babylonchain/babylon/crypto/eots"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/cometbft/cometbft/crypto/tmhash"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ensure that these message types implement the sdk.Msg interface
var (
	_ sdk.Msg = &MsgUpdateParams{}
	_ sdk.Msg = &MsgAddVote{}
	_ sdk.Msg = &MsgCommitPubRand{}
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

// GetSigners returns the expected signers for a MsgAddVote message.
func (m *MsgAddVote) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(m.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

// ValidateBasic does a sanity check on the provided data.
func (m *MsgAddVote) ValidateBasic() error {
	if m.ValBtcPk == nil {
		return fmt.Errorf("empty validator BTC PK")
	}
	if len(m.BlockHash) != tmhash.Size {
		return fmt.Errorf("malformed block hash")
	}
	if m.FinalitySig == nil {
		return fmt.Errorf("empty finality signature")
	}

	return nil
}

// MsgToSign returns (block_height || block_hash)
// The EOTS signature in MsgAddVote will be on this msg
func (m *MsgAddVote) MsgToSign() []byte {
	msgToSign := []byte{}
	msgToSign = append(msgToSign, sdk.Uint64ToBigEndian(m.BlockHeight)...)
	msgToSign = append(msgToSign, m.BlockHash...)
	return msgToSign
}

func (m *MsgAddVote) VerifyEOTSSig(pubRand *bbn.SchnorrPubRand) error {
	msgToSign := m.MsgToSign()
	pk, err := m.ValBtcPk.ToBTCPK()
	if err != nil {
		return err
	}

	return eots.Verify(pk, pubRand.ToFieldVal(), msgToSign, m.FinalitySig.ToModNScalar())
}

// GetSigners returns the expected signers for a MsgCommitPubRand message.
func (m *MsgCommitPubRand) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(m.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

// ValidateBasic does a sanity check on the provided data.
func (m *MsgCommitPubRand) ValidateBasic() error {
	if m.ValBtcPk == nil {
		return fmt.Errorf("empty validator BTC PK")
	}
	if len(m.PubRandList) == 0 {
		return fmt.Errorf("empty list of public randomness")
	}
	if m.Sig == nil {
		return fmt.Errorf("empty signature")
	}
	return m.verifySig()
}

// HashToSign returns a 32-byte hash of (start_height || pub_rand_list)
// The signature in MsgCommitPubRand will be on this hash
func (m *MsgCommitPubRand) HashToSign() ([]byte, error) {
	hasher := tmhash.New()
	if _, err := hasher.Write(sdk.Uint64ToBigEndian(m.StartHeight)); err != nil {
		return nil, err
	}
	for _, pr := range m.PubRandList {
		if _, err := hasher.Write(pr.MustMarshal()); err != nil {
			return nil, err
		}
	}
	return hasher.Sum(nil), nil
}

func (m *MsgCommitPubRand) verifySig() error {
	msgHash, err := m.HashToSign()
	if err != nil {
		return err
	}
	pk, err := m.ValBtcPk.ToBTCPK()
	if err != nil {
		return err
	}
	schnorrSig, err := m.Sig.ToBTCSig()
	if err != nil {
		return err
	}
	if !schnorrSig.Verify(msgHash, pk) {
		return fmt.Errorf("failed to verify signature")
	}
	return nil
}
