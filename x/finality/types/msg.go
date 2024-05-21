package types

import (
	fmt "fmt"

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

func NewMsgAddFinalitySig(signer string, sk *btcec.PrivateKey, sr *eots.PrivateRand, blockHeight uint64, blockHash []byte) (*MsgAddFinalitySig, error) {
	msg := &MsgAddFinalitySig{
		Signer:       signer,
		FpBtcPk:      bbn.NewBIP340PubKeyFromBTCPK(sk.PubKey()),
		BlockHeight:  blockHeight,
		BlockAppHash: blockHash,
	}
	msgToSign := msg.MsgToSign()
	sig, err := eots.Sign(sk, sr, msgToSign)
	if err != nil {
		return nil, err
	}
	msg.FinalitySig = bbn.NewSchnorrEOTSSigFromModNScalar(sig)

	return msg, nil
}

func (m *MsgAddFinalitySig) MsgToSign() []byte {
	return msgToSignForVote(m.BlockHeight, m.BlockAppHash)
}

func (m *MsgAddFinalitySig) VerifyEOTSSig(pubRand *bbn.SchnorrPubRand) error {
	msgToSign := m.MsgToSign()
	pk, err := m.FpBtcPk.ToBTCPK()
	if err != nil {
		return err
	}

	return eots.Verify(pk, pubRand.ToFieldVal(), msgToSign, m.FinalitySig.ToModNScalar())
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

func (m *MsgCommitPubRandList) VerifySig() error {
	msgHash, err := m.HashToSign()
	if err != nil {
		return err
	}
	pk, err := m.FpBtcPk.ToBTCPK()
	if err != nil {
		return err
	}
	if m.Sig == nil {
		return fmt.Errorf("empty signature")
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
