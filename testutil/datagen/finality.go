package datagen

import (
	"math/rand"

	"github.com/babylonchain/babylon/crypto/eots"
	bbn "github.com/babylonchain/babylon/types"
	ftypes "github.com/babylonchain/babylon/x/finality/types"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
)

func GenRandomPubRandList(r *rand.Rand, numPubRand uint64) ([]*eots.PrivateRand, []bbn.SchnorrPubRand, error) {
	srList := []*eots.PrivateRand{}
	prList := []bbn.SchnorrPubRand{}
	for i := uint64(0); i < numPubRand; i++ {
		eotsSR, eotsPR, err := eots.RandGen(r)
		if err != nil {
			return nil, nil, err
		}
		pr := bbn.NewSchnorrPubRandFromFieldVal(eotsPR)
		srList = append(srList, eotsSR)
		prList = append(prList, *pr)
	}
	return srList, prList, nil
}

func GenRandomMsgCommitPubRandList(r *rand.Rand, sk *btcec.PrivateKey, startHeight uint64, numPubRand uint64) ([]*eots.PrivateRand, *ftypes.MsgCommitPubRandList, error) {
	srList, prList, err := GenRandomPubRandList(r, numPubRand)
	if err != nil {
		return nil, nil, err
	}

	msg := &ftypes.MsgCommitPubRandList{
		Signer:      GenRandomAccount().Address,
		ValBtcPk:    bbn.NewBIP340PubKeyFromBTCPK(sk.PubKey()),
		StartHeight: startHeight,
		PubRandList: prList,
	}
	hash, err := msg.HashToSign()
	if err != nil {
		return nil, nil, err
	}
	schnorrSig, err := schnorr.Sign(sk, hash)
	if err != nil {
		return nil, nil, err
	}
	sig := bbn.NewBIP340SignatureFromBTCSig(schnorrSig)
	msg.Sig = &sig
	return srList, msg, nil
}
