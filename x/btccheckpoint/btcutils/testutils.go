package btcutils

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

func HashFromString(s string) *chainhash.Hash {
	hash, e := chainhash.NewHashFromStr(s)
	if e != nil {
		panic("Invalid hex sting")
	}

	return hash
}
