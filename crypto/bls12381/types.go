package bls12381

import (
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

const SignatureLen = 48
const PublicKeyLen = 96

func (sig Signature) Marshal() ([]byte, error) {
	return sig, nil
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
	if len(data) != SignatureLen {
		return errors.New("Invalid signature length")
	}

	*sig = data
	return nil
}

func (sig Signature) Byte() []byte {
	return sig
}

func (pk PublicKey) Marshal() ([]byte, error) {
	return pk, nil
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
	if len(data) != PublicKeyLen {
		return errors.New("Invalid public key length")
	}

	*pk = data
	return nil
}

func (pk PublicKey) Equals(k PublicKey) bool {
	return string(pk) == string(k)
}

func (pk PublicKey) Byte() []byte {
	return pk
}
