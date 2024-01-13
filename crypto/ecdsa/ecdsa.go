package ecdsa

import (
	"bytes"
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

const (
	MAGIC_MESSAGE_PREFIX = "Bitcoin Signed Message:\n"
)

// magicHash encodes the given msg into byte array, then calculates its sha256d hash
// ref: https://github.com/okx/js-wallet-sdk/blob/a57c2acbe6ce917c0aa4e951d96c4e562ad58444/packages/coin-bitcoin/src/message.ts#L28-L34
func magicHash(msg string) chainhash.Hash {
	buf := bytes.NewBuffer(nil)
	// we have to use wire.WriteVarString which encodes the string length into the byte array in Bitcoin's own way
	// message prefix
	// NOTE: we have control over the buffer so no need to check error
	wire.WriteVarString(buf, 0, MAGIC_MESSAGE_PREFIX) //nolint:errcheck
	// message
	wire.WriteVarString(buf, 0, msg) //nolint:errcheck
	bytes := buf.Bytes()

	return chainhash.DoubleHashH(bytes)
}

func Sign(sk *btcec.PrivateKey, msg string) ([]byte, error) {
	msgHash := magicHash(msg)
	return ecdsa.SignCompact(sk, msgHash[:], true)
}

func Verify(pk *btcec.PublicKey, msg string, sigBytes []byte) error {
	msgHash := magicHash(msg)
	recoveredPK, _, err := ecdsa.RecoverCompact(sigBytes, msgHash[:])
	if err != nil {
		return err
	}
	pkBytes := schnorr.SerializePubKey(pk)
	recoveredPKBytes := schnorr.SerializePubKey(recoveredPK)
	if !bytes.Equal(pkBytes, recoveredPKBytes) {
		return fmt.Errorf("the recovered PK does not match the given PK")
	}
	return nil
}
