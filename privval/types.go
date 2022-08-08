package privval

import (
	"github.com/babylonchain/babylon/crypto/bls12381"
	"github.com/babylonchain/babylon/x/checkpointing/types"
	tmcrypto "github.com/tendermint/tendermint/crypto"
)

type ValidatorKeys struct {
	ValPubkey tmcrypto.PubKey
	BlsPubkey bls12381.PublicKey
	Pop       *types.ProofOfPossession

	valPrivkey tmcrypto.PrivKey
	blsPrivkey bls12381.PrivateKey
}

func NewValidatorKeys(valPrivkey tmcrypto.PrivKey, blsPrivKey bls12381.PrivateKey, msg []byte) (*ValidatorKeys, error) {
	pop, err := BuildPop(valPrivkey, blsPrivKey, msg)
	if err != nil {
		return nil, err
	}
	return &ValidatorKeys{
		ValPubkey:  valPrivkey.PubKey(),
		BlsPubkey:  blsPrivKey.PubKey(),
		valPrivkey: valPrivkey,
		blsPrivkey: blsPrivKey,
		Pop:        pop,
	}, nil
}

// BuildPop builds a proof-of-possession by encrypt(key = BLS_sk, data = encrypt(key = Ed25519_sk, data = Secp256k1_pk))
// where valPrivKey is Ed25519_sk, accPubkey is Secp256k1_pk, blsPrivkey is BLS_sk
func BuildPop(valPrivKey tmcrypto.PrivKey, blsPrivkey bls12381.PrivateKey, msg []byte) (*types.ProofOfPossession, error) {
	data, err := valPrivKey.Sign(msg)
	if err != nil {
		return nil, err
	}
	pop := bls12381.Sign(blsPrivkey, data)
	return &types.ProofOfPossession{BlsSig: &pop}, nil
}
