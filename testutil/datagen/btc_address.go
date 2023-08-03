package datagen

import (
	"math/rand"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
)

func GenRandomBTCAddress(r *rand.Rand, net *chaincfg.Params) (string, error) {
	addr, err := btcutil.NewAddressWitnessPubKeyHash(GenRandomByteArray(r, 20), net)
	if err != nil {
		return "", err
	}
	return addr.EncodeAddress(), nil
}
