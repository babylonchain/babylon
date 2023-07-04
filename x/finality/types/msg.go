package types

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"github.com/babylonchain/babylon/crypto/eots"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/cometbft/cometbft/crypto/tmhash"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ensure that these message types implement the sdk.Msg interface
var (
	_ sdk.Msg = &MsgUpdateParams{}
	_ sdk.Msg = &MsgAddFinalitySig{}
	_ sdk.Msg = &MsgCommitPubRandList{}
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

func NewMsgAddFinalitySig(signer string, sk *btcec.PrivateKey, sr *eots.PrivateRand, blockHeight uint64, blockHash []byte) (*MsgAddFinalitySig, error) {
	msg := &MsgAddFinalitySig{
		Signer:              signer,
		ValBtcPk:            bbn.NewBIP340PubKeyFromBTCPK(sk.PubKey()),
		BlockHeight:         blockHeight,
		BlockLastCommitHash: blockHash,
	}
	msgToSign := msg.MsgToSign()
	sig, err := eots.Sign(sk, sr, msgToSign)
	if err != nil {
		return nil, err
	}
	msg.FinalitySig = bbn.NewSchnorrEOTSSigFromModNScalar(sig)

	return msg, nil
}

// GetSigners returns the expected signers for a MsgAddFinalitySig message.
func (m *MsgAddFinalitySig) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(m.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

// ValidateBasic does a sanity check on the provided data.
func (m *MsgAddFinalitySig) ValidateBasic() error {
	if m.ValBtcPk == nil {
		return fmt.Errorf("empty validator BTC PK")
	}
	if len(m.BlockLastCommitHash) != tmhash.Size {
		return fmt.Errorf("malformed block hash")
	}
	if m.FinalitySig == nil {
		return fmt.Errorf("empty finality signature")
	}

	return nil
}

func (m *MsgAddFinalitySig) MsgToSign() []byte {
	return MsgToSignForVote(m.BlockHeight, m.BlockLastCommitHash)
}

func (m *MsgAddFinalitySig) VerifyEOTSSig(pubRand *bbn.SchnorrPubRand) error {
	msgToSign := m.MsgToSign()
	pk, err := m.ValBtcPk.ToBTCPK()
	if err != nil {
		return err
	}

	return eots.Verify(pk, pubRand.ToFieldVal(), msgToSign, m.FinalitySig.ToModNScalar())
}

// GetSigners returns the expected signers for a MsgCommitPubRandList message.
func (m *MsgCommitPubRandList) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(m.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

// ValidateBasic does a sanity check on the provided data.
func (m *MsgCommitPubRandList) ValidateBasic() error {
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
// The signature in MsgCommitPubRandList will be on this hash
func (m *MsgCommitPubRandList) HashToSign() ([]byte, error) {
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

func (m *MsgCommitPubRandList) verifySig() error {
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
