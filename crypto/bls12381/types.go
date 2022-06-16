package bls12381

import (
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

func (sig Signature) Byte() []byte {
	return sig
}

func (pk PublicKey) Byte() []byte {
	return pk
}
