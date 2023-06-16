package datagen

import (
	"math/rand"

	"github.com/btcsuite/btcd/btcec/v2"
)

func GenRandomBTCKeyPair(r *rand.Rand) (*btcec.PrivateKey, *btcec.PublicKey, error) {
	sk, err := btcec.NewPrivateKey()
	if err != nil {
		return nil, nil, err
	}
	return sk, sk.PubKey(), nil
}
