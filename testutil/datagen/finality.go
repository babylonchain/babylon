package datagen

import (
	"math/rand"

	"github.com/babylonchain/babylon/crypto/eots"
	bbn "github.com/babylonchain/babylon/types"
	ftypes "github.com/babylonchain/babylon/x/finality/types"
	"github.com/btcsuite/btcd/btcec/v2"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func GenRandomEvidence(r *rand.Rand, sk *btcec.PrivateKey, height uint64) (*ftypes.Evidence, error) {
	pk := sk.PubKey()
	bip340PK := bbn.NewBIP340PubKeyFromBTCPK(pk)
	msr, mpr, err := eots.NewMasterRandPair(r)
	if err != nil {
		return nil, err
	}
	sr, _, err := msr.DeriveRandPair(uint32(height))
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
		MasterPubRand:        mpr.MarshalBase58(),
		CanonicalAppHash:     cAppHash,
		ForkAppHash:          fAppHash,
		CanonicalFinalitySig: bbn.NewSchnorrEOTSSigFromModNScalar(cSig),
		ForkFinalitySig:      bbn.NewSchnorrEOTSSigFromModNScalar(fSig),
	}
	return evidence, nil
}
