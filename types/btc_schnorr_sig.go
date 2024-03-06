package types

import (
	"encoding/hex"
	"errors"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"
)

type BIP340Signature []byte

const BIP340SignatureLen = schnorr.SignatureSize

func NewBIP340Signature(data []byte) (*BIP340Signature, error) {

	var sig BIP340Signature
	err := sig.Unmarshal(data)

	if _, err := sig.ToBTCSig(); err != nil {
		return nil, errors.New("bytes cannot be converted to a *schnorr.Signature object")
	}

	return &sig, err
}

func NewBIP340SignatureFromHex(sigHex string) (*BIP340Signature, error) {
	sigBytes, err := hex.DecodeString(sigHex)
	if err != nil {
		return nil, err
	}
	return NewBIP340Signature(sigBytes)
}

func NewBIP340SignatureFromBTCSig(btcSig *schnorr.Signature) *BIP340Signature {
	sigBytes := btcSig.Serialize()
	sig := BIP340Signature(sigBytes)
	return &sig
}

func (sig BIP340Signature) ToBTCSig() (*schnorr.Signature, error) {
	return schnorr.ParseSignature(sig)
}

func (sig BIP340Signature) MustToBTCSig() *schnorr.Signature {
	btcSig, err := schnorr.ParseSignature(sig)
	if err != nil {
		panic(err)
	}
	return btcSig
}

func (sig BIP340Signature) Size() int {
	return len(sig.MustMarshal())
}

func (sig BIP340Signature) Marshal() ([]byte, error) {
	return sig, nil
}

func (sig BIP340Signature) MustMarshal() []byte {
	sigBytes, err := sig.Marshal()
	if err != nil {
		panic(err)
	}
	return sigBytes
}

func (sig BIP340Signature) MarshalTo(data []byte) (int, error) {
	bz, err := sig.Marshal()
	if err != nil {
		return 0, err
	}
	copy(data, bz)
	return len(data), nil
}

func (sig *BIP340Signature) Unmarshal(data []byte) error {
	*sig = data
	return nil
}

func (sig *BIP340Signature) ToHexStr() string {
	sigBytes := sig.MustMarshal()
	return hex.EncodeToString(sigBytes)
}
