package datagen

import (
	"math/rand"

	"github.com/babylonchain/babylon/crypto/eots"
	bbn "github.com/babylonchain/babylon/types"
	ftypes "github.com/babylonchain/babylon/x/finality/types"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	msg.Sig = bbn.NewBIP340SignatureFromBTCSig(schnorrSig)
	return srList, msg, nil
}

func GenRandomEvidence(r *rand.Rand, sk *btcec.PrivateKey, height uint64) (*ftypes.Evidence, error) {
	pk := sk.PubKey()
	bip340PK := bbn.NewBIP340PubKeyFromBTCPK(pk)
	sr, pr, err := eots.RandGen(r)
	if err != nil {
		return nil, err
	}
	cAppHash := GenRandomByteArray(r, 32)
	cSig, err := eots.Sign(sk, sr, append(sdk.Uint64ToBigEndian(height), cAppHash...))
	if err != nil {
		return nil, err
	}
	fAppHash := GenRandomByteArray(r, 32)
	fSig, err := eots.Sign(sk, sr, append(sdk.Uint64ToBigEndian(height), fAppHash...))
	if err != nil {
		return nil, err
	}

	evidence := &ftypes.Evidence{
		ValBtcPk:             bip340PK,
		BlockHeight:          height,
		PubRand:              bbn.NewSchnorrPubRandFromFieldVal(pr),
		CanonicalAppHash:     cAppHash,
		ForkAppHash:          fAppHash,
		CanonicalFinalitySig: bbn.NewSchnorrEOTSSigFromModNScalar(cSig),
		ForkFinalitySig:      bbn.NewSchnorrEOTSSigFromModNScalar(fSig),
	}
	return evidence, nil
}
