package privval

import (
	"github.com/babylonchain/babylon/crypto/bls12381"
	"github.com/babylonchain/babylon/x/checkpointing/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	tmcrypto "github.com/tendermint/tendermint/crypto"
)

type ValidatorKeys struct {
	ValPubkey tmcrypto.PubKey
	BlsPubkey bls12381.PublicKey
	PoP       *types.ProofOfPossession

	valPrivkey tmcrypto.PrivKey
	blsPrivkey bls12381.PrivateKey
}

func NewValidatorKeys(valPrivkey tmcrypto.PrivKey, blsPrivKey bls12381.PrivateKey, accPubkey cryptotypes.PubKey) (*ValidatorKeys, error) {
	pop, err := BuildPop(valPrivkey, blsPrivKey, accPubkey)
	if err != nil {
		return nil, err
	}
	return &ValidatorKeys{
		ValPubkey:  valPrivkey.PubKey(),
		BlsPubkey:  blsPrivKey.PubKey(),
		valPrivkey: valPrivkey,
		blsPrivkey: blsPrivKey,
		PoP:        pop,
	}, nil
}

// BuildPop builds a proof-of-possession by encrypt(key = BLS_sk, data = encrypt(key = Ed25519_sk, data = Secp256k1_pk))
// where valPrivKey is Ed25519_sk, accPubkey is Secp256k1_pk, blsPrivkey is BLS_sk
func BuildPop(valPrivKey tmcrypto.PrivKey, blsPrivkey bls12381.PrivateKey, accPubkey cryptotypes.PubKey) (*types.ProofOfPossession, error) {
	data, err := valPrivKey.Sign(accPubkey.Bytes())
	if err != nil {
		return nil, err
	}
	pop := bls12381.Sign(blsPrivkey, data)
	return &types.ProofOfPossession{BlsSig: &pop}, nil
}
