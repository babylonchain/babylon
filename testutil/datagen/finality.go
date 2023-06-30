package datagen

import (
	"math/rand"

	bbn "github.com/babylonchain/babylon/types"
	ftypes "github.com/babylonchain/babylon/x/finality/types"
	"github.com/babylonchain/eots"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
)

func GenRandomMsgAddVote(r *rand.Rand, sk *btcec.PrivateKey)  (*ftypes.MsgAddVote, *bbn.SchnorrPubRand, error) {
	msg := &ftypes.MsgAddVote{
		Signer:      GenRandomAccount().Address,
		ValBtcPk:    bbn.NewBIP340PubKeyFromBTCPK(sk.PubKey()),
		BlockHeight: RandomInt(r, 100),
		BlockHash: GenRandomByteArray(r, 32),
	}
	msgToSign := msg.MsgToSign()
	sr, pr, err := eots.RandGen(r)
	if err != nil {
		return nil, nil, err
	}
	sig, err := eots.Sign(sk, sr, msgToSign)
	if err != nil {
		return nil, nil, err
	}
	msg.FinalitySig = bbn.NewSchnorrEOTSSigFromModNScalar(sig)

	return msg, bbn.NewSchnorrPubRandFromFieldVal(pr), nil
}

func GenRandomMsgCommitPubRand(r *rand.Rand, sk *btcec.PrivateKey) (*ftypes.MsgCommitPubRand, error) {
	prList := []bbn.SchnorrPubRand{}
	for i := 0; i < 1000; i++ {
		prBytes := GenRandomByteArray(r, bbn.SchnorrPubRandLen)
		pr, err := bbn.NewSchnorrPubRand(prBytes)
		if err != nil {
			return nil, err
		}
		prList = append(prList, *pr)
	}

	msg := &ftypes.MsgCommitPubRand{
		Signer:      GenRandomAccount().Address,
		ValBtcPk:    bbn.NewBIP340PubKeyFromBTCPK(sk.PubKey()),
		StartHeight: RandomInt(r, 100),
		PubRandList: prList,
	}
	hash, err := msg.HashToSign()
	if err != nil {
		return nil, err
	}
	schnorrSig, err := schnorr.Sign(sk, hash)
	if err != nil {
		return nil, err
	}
	sig := bbn.NewBIP340SignatureFromBTCSig(schnorrSig)
	msg.Sig = &sig
	return msg, nil
}
