package datagen

import (
	"math/rand"

	"github.com/babylonchain/babylon/crypto/eots"
	bbn "github.com/babylonchain/babylon/types"
	ftypes "github.com/babylonchain/babylon/x/finality/types"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/cometbft/cometbft/crypto/merkle"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type RandListInfo struct {
	SRList     []*eots.PrivateRand
	PRList     []bbn.SchnorrPubRand
	Commitment []byte
	ProofList  []*merkle.Proof
}

func GenRandomPubRandList(r *rand.Rand, numPubRand uint64) (*RandListInfo, error) {
	// generate a list of secret/public randomness
	srList := []*eots.PrivateRand{}
	prList := []bbn.SchnorrPubRand{}
	for i := uint64(0); i < numPubRand; i++ {
		eotsSR, eotsPR, err := eots.RandGen(r)
		if err != nil {
			return nil, err
		}
		pr := bbn.NewSchnorrPubRandFromFieldVal(eotsPR)
		srList = append(srList, eotsSR)
		prList = append(prList, *pr)
	}

	prByteList := [][]byte{}
	for i := range prList {
		prByteList = append(prByteList, prList[i])
	}

	// generate the commitment to these public randomness
	commitment, proofList := merkle.ProofsFromByteSlices(prByteList)

	return &RandListInfo{srList, prList, commitment, proofList}, nil
}

func GenRandomMsgCommitPubRandList(r *rand.Rand, sk *btcec.PrivateKey, startHeight uint64, numPubRand uint64) (*RandListInfo, *ftypes.MsgCommitPubRandList, error) {
	randListInfo, err := GenRandomPubRandList(r, numPubRand)
	if err != nil {
		return nil, nil, err
	}

	msg := &ftypes.MsgCommitPubRandList{
		Signer:      GenRandomAccount().Address,
		FpBtcPk:     bbn.NewBIP340PubKeyFromBTCPK(sk.PubKey()),
		StartHeight: startHeight,
		NumPubRand:  numPubRand,
		Commitment:  randListInfo.Commitment,
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
	return randListInfo, msg, nil
}

func NewMsgAddFinalitySig(
	signer string,
	sk *btcec.PrivateKey,
	startHeight uint64,
	blockHeight uint64,
	randListInfo *RandListInfo,
	blockAppHash []byte,
) (*ftypes.MsgAddFinalitySig, error) {
	idx := blockHeight - startHeight

	msg := &ftypes.MsgAddFinalitySig{
		Signer:       signer,
		FpBtcPk:      bbn.NewBIP340PubKeyFromBTCPK(sk.PubKey()),
		PubRand:      &randListInfo.PRList[idx],
		Proof:        randListInfo.ProofList[idx].ToProto(),
		BlockHeight:  blockHeight,
		BlockAppHash: blockAppHash,
		FinalitySig:  nil,
	}
	msgToSign := msg.MsgToSign()
	sig, err := eots.Sign(sk, randListInfo.SRList[idx], msgToSign)
	if err != nil {
		return nil, err
	}
	msg.FinalitySig = bbn.NewSchnorrEOTSSigFromModNScalar(sig)

	return msg, nil
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
		FpBtcPk:              bip340PK,
		BlockHeight:          height,
		PubRand:              bbn.NewSchnorrPubRandFromFieldVal(pr),
		CanonicalAppHash:     cAppHash,
		ForkAppHash:          fAppHash,
		CanonicalFinalitySig: bbn.NewSchnorrEOTSSigFromModNScalar(cSig),
		ForkFinalitySig:      bbn.NewSchnorrEOTSSigFromModNScalar(fSig),
	}
	return evidence, nil
}
