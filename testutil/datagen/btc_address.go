package datagen

import (
	"math/rand"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
)

func GenRandomPkScript(r *rand.Rand) []byte {
	// NOTE: this generates non-standard pkscript
	return GenRandomByteArray(r, 20)
}

func GenRandomBTCAddress(r *rand.Rand, net *chaincfg.Params) (btcutil.Address, error) {
	addr, err := btcutil.NewAddressPubKeyHash(GenRandomByteArray(r, 20), net)
	if err != nil {
		return nil, err
	}
	return addr, nil
}

func GenRandomPubKeyHashScript(r *rand.Rand, net *chaincfg.Params) ([]byte, error) {
	addr, err := btcutil.NewAddressPubKeyHash(GenRandomByteArray(r, 20), net)
	if err != nil {
		return nil, err
	}
	return txscript.PayToAddrScript(addr)
}
