package bls12381

import (
	"encoding/hex"
	"errors"
	blst "github.com/supranational/blst/bindings/go"
)

// For minimal-pubkey-size operations:
// type BlsPubKey = blst.P1Affine
// type BlsSig = blst.P2Affine
// type BlsMultiSig = blst.P2Aggregate
// type BlsMultiPubKey = blst.P1Aggregate

// Domain Separation Tag for signatures on G2 (minimal-pubkey-size)
// const DST = []byte("BLS_SIG_BLS12381G1_XMD:SHA-256_SSWU_RO_NUL_")

// For minimal-signature-size operations:
type BlsPubKey = blst.P2Affine
type BlsSig = blst.P1Affine
type BlsMultiSig = blst.P1Aggregate
type BlsMultiPubKey = blst.P2Aggregate

// Domain Separation Tag for signatures on G1 (minimal-signature-size)
var DST = []byte("BLS_SIG_BLS12381G1_XMD:SHA-256_SSWU_RO_NUL_")

type Signature []byte
type PublicKey []byte
type PrivateKey []byte

const (
	// SignatureSize is the size, in bytes, of a compressed BLS signature
	SignatureSize = 48
	// PubKeySize is the size, in bytes, of a compressed BLS public key
	PubKeySize = 96
	// SeedSize is the size, in bytes, of private key seeds
	SeedSize = 32
)

func (sig Signature) ValidateBasic() error {
	if sig == nil {
		return errors.New("invalid BLS signature")
	}
	if len(sig) != SignatureSize {
		return errors.New("invalid BLS signature")
	}

	return nil
}

func (sig Signature) Marshal() ([]byte, error) {
	return sig, nil
}

func (sig Signature) MustMarshal() []byte {
	bz, err := sig.Marshal()
	if err != nil {
		panic(err)
	}

	return bz
}

func (sig Signature) MarshalTo(data []byte) (int, error) {
	copy(data, sig)
	return len(data), nil
}

func (sig Signature) Size() int {
	bz, _ := sig.Marshal()
	return len(bz)
}

func (sig *Signature) Unmarshal(data []byte) error {
	if len(data) != SignatureSize {
		return errors.New("Invalid signature length")
	}

	*sig = data
	return nil
}

func (sig Signature) Bytes() []byte {
	return sig
}

func (sig Signature) Equal(s Signature) bool {
	return string(sig) == string(s)
}

func NewBLSSigFromHex(s string) (Signature, error) {
	bz, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}

	var sig Signature
	err = sig.Unmarshal(bz)
	if err != nil {
		return nil, err
	}

	return sig, nil
}

func (sig Signature) String() string {
	bz := sig.MustMarshal()

	return hex.EncodeToString(bz)
}

func (pk PublicKey) Marshal() ([]byte, error) {
	return pk, nil
}

func (pk PublicKey) MustMarshal() []byte {
	bz, err := pk.Marshal()
	if err != nil {
		panic(err)
	}

	return bz
}

func (pk PublicKey) MarshalTo(data []byte) (int, error) {
	copy(data, pk)
	return len(data), nil
}

func (pk PublicKey) Size() int {
	bz, _ := pk.Marshal()
	return len(bz)
}

func (pk *PublicKey) Unmarshal(data []byte) error {
	if len(data) != PubKeySize {
		return errors.New("Invalid public key length")
	}

	*pk = data
	return nil
}

func (pk PublicKey) Equal(k PublicKey) bool {
	return string(pk) == string(k)
}

func (pk PublicKey) Bytes() []byte {
	return pk
}

func (sk PrivateKey) PubKey() PublicKey {
	secretKey := new(blst.SecretKey)
	secretKey.Deserialize(sk)
	pk := new(BlsPubKey).From(secretKey)
	return pk.Compress()
}
