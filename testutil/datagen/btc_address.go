package datagen

import (
	"math/rand"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
)

func GenRandomBTCAddress(r *rand.Rand, net *chaincfg.Params) (string, error) {
	pkHash := GenRandomByteArray(r, 20)
	addr, err := btcutil.NewAddressPubKeyHash(pkHash, net)
	if err != nil {
		return "", err
	}
	return addr.EncodeAddress(), nil
}
