package btcstaking

import (
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
)

// XonlyPubKey is a wrapper around btcec.PublicKey that represents BTC public
// key deserialized from a 32-byte array i.e with implicit assumption that Y coordinate
// is even.
type XonlyPubKey struct {
	PubKey *btcec.PublicKey
}

func XOnlyPublicKeyFromBytes(pkBytes []byte) (*XonlyPubKey, error) {
	pk, err := schnorr.ParsePubKey(pkBytes)

	if err != nil {
		return nil, err
	}

	return &XonlyPubKey{pk}, nil
}

func (p *XonlyPubKey) Marshall() []byte {
	return schnorr.SerializePubKey(p.PubKey)
}
