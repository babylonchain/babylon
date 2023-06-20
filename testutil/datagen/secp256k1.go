package datagen

import (
	"math/rand"

	secp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

func GenRandomSecp256k1KeyPair(r *rand.Rand) (cryptotypes.PrivKey, cryptotypes.PubKey, error) {
	randBytes := GenRandomByteArray(r, 10)
	sk := secp256k1.GenPrivKeyFromSecret(randBytes)
	pk := sk.PubKey()
	return sk, pk, nil
}
