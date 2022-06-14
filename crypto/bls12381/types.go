package bls12381

import (
	blst "github.com/supranational/blst/bindings/go"
)

// For minimal-pubkey-size operations:
// type BlsPubKey = blst.P1Affine
// type BlsSig = blst.P2Affine
// type BlsMultiSig = blst.P2Aggregate
// type BlsMultiPubKey = blst.P1Aggregate

// For minimal-signature-size operations:
type BlsPubKey = blst.P2Affine
type BlsSig = blst.P1Affine
type BlsMultiSig = blst.P1Aggregate
type BlsMultiPubKey = blst.P2Aggregate

// default domain
var dst = []byte("BLS_SIG_BLS12381G1_XMD:SHA-256_SSWU_RO_NUL_")
