package types

import (
	"github.com/babylonchain/babylon/crypto/eots"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/btcsuite/btcd/btcec/v2"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ensure that these message types implement the sdk.Msg interface
var (
	_ sdk.Msg = &MsgUpdateParams{}
	_ sdk.Msg = &MsgAddFinalitySig{}
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

func (m *MsgAddFinalitySig) VerifyEOTSSig(mpr *eots.MasterPublicRand) error {
	msgToSign := m.MsgToSign()
	pk, err := m.FpBtcPk.ToBTCPK()
	if err != nil {
		return err
	}
	pubRand, err := mpr.DerivePubRand(uint32(m.BlockHeight))
	if err != nil {
		return err
	}

	return eots.Verify(pk, pubRand, msgToSign, m.FinalitySig.ToModNScalar())
}
