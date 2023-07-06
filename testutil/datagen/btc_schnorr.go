package datagen

import (
	"math/rand"

	bbn "github.com/babylonchain/babylon/types"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

func GenRandomBTCKeyPair(r *rand.Rand) (*btcec.PrivateKey, *btcec.PublicKey, error) {
	sk, err := secp256k1.GeneratePrivateKeyFromRand(r)
	if err != nil {
		return nil, nil, err
	}
	return sk, sk.PubKey(), nil
}

func GenRandomBIP340PubKey(r *rand.Rand) (*bbn.BIP340PubKey, error) {
	sk, err := secp256k1.GeneratePrivateKeyFromRand(r)
	if err != nil {
		return nil, err
	}
	pk := sk.PubKey()
	btcPK := bbn.NewBIP340PubKeyFromBTCPK(pk)
	return btcPK, nil
}
