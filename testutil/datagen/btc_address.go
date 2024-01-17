package datagen

import (
	"io"
	"math/rand"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"golang.org/x/crypto/ripemd160" //nolint:all
)

func GenRandomPkHash(r *rand.Rand) []byte {
	md := ripemd160.New()
	io.WriteString(md, GenRandomHexStr(r, 20)) //nolint:errcheck
	return md.Sum(nil)
}

func GenRandomBTCAddress(r *rand.Rand, net *chaincfg.Params) (btcutil.Address, error) {
	var (
		addr btcutil.Address
		err  error
	)

	for {
		addr, err = btcutil.NewAddressPubKeyHash(GenRandomPkHash(r), net)
		if err != nil {
			// something is wrong in pkhash or net, return error
			return nil, err
		}
		if _, err := btcutil.DecodeAddress(addr.EncodeAddress(), net); err == nil {
			// this is a legit address, use it
			break
		}
	}

	return addr, nil
}

func GenRandomPubKeyHashScript(r *rand.Rand, net *chaincfg.Params) ([]byte, error) {
	addr, err := btcutil.NewAddressPubKeyHash(GenRandomPkHash(r), net)
	if err != nil {
		return nil, err
	}
	return txscript.PayToAddrScript(addr)
}
