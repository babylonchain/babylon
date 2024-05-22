package eots

import (
	"crypto/sha256"
	"errors"
	"io"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	ecdsa_schnorr "github.com/decred/dcrd/dcrec/secp256k1/v4/schnorr"
)

type ModNScalar = btcec.ModNScalar
type PrivateKey = secp256k1.PrivateKey
type PublicKey = secp256k1.PublicKey
type PrivateRand = secp256k1.ModNScalar
type PublicRand = secp256k1.FieldVal

// The Signature is only the S part of the BEP-340 Schnorr signatures.
type Signature = ModNScalar

// KeyGen generates private key from a randomness source
func KeyGen(randSource io.Reader) (*PrivateKey, error) {
	return secp256k1.GeneratePrivateKeyFromRand(randSource)
}

// PubGen returns the associated public key from a private key.
func PubGen(k *PrivateKey) *PublicKey {
	return k.PubKey()
}

// RandGen returns the value to be used as random value when signing, and the associated public value.
func RandGen(randSource io.Reader) (*PrivateRand, *PublicRand, error) {
	sk, err := KeyGen(randSource)
	if err != nil {
		return nil, nil, err
	}
	var j secp256k1.JacobianPoint
	sk.PubKey().AsJacobian(&j)
	return &sk.Key, &j.X, nil
}

// hash function is used for hashing the message input for all functions of the library.
// Wrapper around sha256 in order to change only one function if the input hashing function is changed.
func hash(message []byte) [32]byte {
	return sha256.Sum256(message)
}

// Sign returns an extractable Schnorr signature for a message, signed with a private key and private randomness value.
// Note that the Signature is only the second (S) part of the typical bitcoin signature, the first (R) can be deduced from
// the public randomness value and the message.
func Sign(sk *PrivateKey, privateRand *PrivateRand, message []byte) (*Signature, error) {
	h := hash(message)
	return signHash(sk, privateRand, h)
}

// signHash returns an extractable Schnorr signature for a hashed message.
// The caller MUST ensure that hash is the output of a cryptographically secure hash function.
// Based on unexported schnorrSign of btcd.
func signHash(sk *PrivateKey, privateRand *PrivateRand, hash [32]byte) (*Signature, error) {
	if sk.Key.IsZero() {
		str := "private key is zero"
		return nil, signatureError(ecdsa_schnorr.ErrPrivateKeyIsZero, str)
	}

	// d' = int(d)
	var privKeyScalar ModNScalar
	privKeyScalar.Set(&sk.Key)

	pubKey := PubGen(sk)

	// Negate d if P.y is odd.
	pubKeyBytes := pubKey.SerializeCompressed()
	if pubKeyBytes[0] == secp256k1.PubKeyFormatCompressedOdd {
		privKeyScalar.Negate()
	}

	k := new(ModNScalar).Set(privateRand)

	// R = kG
	var R btcec.JacobianPoint
	btcec.ScalarBaseMultNonConst(k, &R)

	// Negate nonce k if R.y is odd (R.y is the y coordinate of the point R)
	//
	// Note that R must be in affine coordinates for this check.
	R.ToAffine()
	if R.Y.IsOdd() {
		k.Negate()
	}

	// e = tagged_hash("BIP0340/challenge", bytes(R) || bytes(P) || m) mod n
	var rBytes [32]byte
	r := &R.X
	r.PutBytesUnchecked(rBytes[:])
	pBytes := pubKey.SerializeCompressed()[1:]

	commitment := chainhash.TaggedHash(chainhash.TagBIP0340Challenge, rBytes[:], pBytes, hash[:])

	var e ModNScalar
	if overflow := e.SetBytes((*[32]byte)(commitment)); overflow != 0 {
		k.Zero()
		str := "hash of (r || P || m) too big"
		return nil, signatureError(ecdsa_schnorr.ErrSchnorrHashValue, str)
	}

	// s = k + e*d mod n
	sig := new(ModNScalar).Mul2(&e, &privKeyScalar).Add(k)

	// If Verify(bytes(P), m, sig) fails, abort.
	// optional

	// Return s
	return sig, nil
}

// Verify verifies that the signature is valid for this message, public key and random value.
func Verify(pubKey *PublicKey, r *PublicRand, message []byte, sig *Signature) error {
	h := hash(message)
	pubkeyBytes := schnorr.SerializePubKey(pubKey)
	return verifyHash(pubkeyBytes, r, h, sig)
}

// Verify verifies that the signature is valid for this hashed message, public key and random value.
// Based on unexported schnorrVerify of btcd.
func verifyHash(pubKeyBytes []byte, r *PublicRand, hash [32]byte, sig *Signature) error {
	// Step 2.
	//
	// P = lift_x(int(pk))
	//
	// Fail if P is not a point on the curve
	pubKey, err := schnorr.ParsePubKey(pubKeyBytes)
	if err != nil {
		return err
	}

	// Fail if P is not a point on the curve
	if !pubKey.IsOnCurve() {
		str := "pubkey point is not on curve"
		return signatureError(ecdsa_schnorr.ErrPubKeyNotOnCurve, str)
	}

	// Fail if r >= p is already handled by the fact r is a field element.
	// Fail if s >= n is already handled by the fact s is a mod n scalar.

	// e = int(tagged_hash("BIP0340/challenge", bytes(r) || bytes(P) || M)) mod n.
	var rBytes [32]byte
	r.PutBytesUnchecked(rBytes[:])
	pBytes := pubKey.SerializeCompressed()[1:]

	commitment := chainhash.TaggedHash(chainhash.TagBIP0340Challenge, rBytes[:], pBytes, hash[:])

	var e ModNScalar
	if overflow := e.SetBytes((*[32]byte)(commitment)); overflow != 0 {
		str := "hash of (r || P || m) too big"
		return signatureError(ecdsa_schnorr.ErrSchnorrHashValue, str)
	}

	// Negate e here so we can use AddNonConst below to subtract the s*G
	// point from e*P.
	e.Negate()

	// R = s*G - e*P
	var P, R, sG, eP btcec.JacobianPoint
	pubKey.AsJacobian(&P)
	btcec.ScalarBaseMultNonConst(sig, &sG)
	btcec.ScalarMultNonConst(&e, &P, &eP)
	btcec.AddNonConst(&sG, &eP, &R)

	// Fail if R is the point at infinity
	if (R.X.IsZero() && R.Y.IsZero()) || R.Z.IsZero() {
		str := "calculated R point is the point at infinity"
		return signatureError(ecdsa_schnorr.ErrSigRNotOnCurve, str)
	}

	// Fail if R.y is odd
	//
	// Note that R must be in affine coordinates for this check.
	R.ToAffine()
	if R.Y.IsOdd() {
		str := "calculated R y-value is odd"
		return signatureError(ecdsa_schnorr.ErrSigRYIsOdd, str)
	}

	// verify signed with the right k random value
	if !r.Equals(&R.X) {
		str := "calculated R point was not given R"
		return signatureError(ecdsa_schnorr.ErrUnequalRValues, str)
	}

	return nil
}

// Extract extracts the private key from a public key and signatures for two distinct hashes messages.
func Extract(pubKey *PublicKey, r *PublicRand, message1 []byte, sig1 *Signature, message2 []byte, sig2 *Signature) (*PrivateKey, error) {
	h1 := hash(message1)
	h2 := hash(message2)
	return extractFromHashes(pubKey, r, h1, sig1, h2, sig2)
}

// extractFromHashes extracts the private key from hashes, instead of the non-hashed message directly as Extract does.
func extractFromHashes(pubKey *PublicKey, r *PublicRand, hash1 [32]byte, sig1 *Signature, hash2 [32]byte, sig2 *Signature) (*PrivateKey, error) {
	var rBytes [32]byte
	r.PutBytesUnchecked(rBytes[:])
	pBytes := pubKey.SerializeCompressed()[1:]

	if sig1.Equals(sig2) {
		return nil, errors.New("The two signatures need to be different in order to extract")
	}

	commitment1 := chainhash.TaggedHash(chainhash.TagBIP0340Challenge, rBytes[:], pBytes, hash1[:])
	var e1 ModNScalar
	if overflow := e1.SetBytes((*[32]byte)(commitment1)); overflow != 0 {
		str := "hash of (r || P || m1) too big"
		return nil, signatureError(ecdsa_schnorr.ErrSchnorrHashValue, str)
	}

	commitment2 := chainhash.TaggedHash(chainhash.TagBIP0340Challenge, rBytes[:], pBytes, hash2[:])
	var e2 ModNScalar
	if overflow := e2.SetBytes((*[32]byte)(commitment2)); overflow != 0 {
		str := "hash of (r || P || m2) too big"
		return nil, signatureError(ecdsa_schnorr.ErrSchnorrHashValue, str)
	}

	// x = (s1 - s2) / (e1 - e2)
	var x, denom ModNScalar
	denom.Add2(&e1, e2.Negate())
	x.Add2(sig1, sig2.Negate()).Mul(denom.InverseNonConst())

	pubKeyBytes := pubKey.SerializeCompressed()
	if pubKeyBytes[0] == secp256k1.PubKeyFormatCompressedOdd {
		x.Negate()
	}

	privKey := secp256k1.NewPrivateKey(&x)
	if privKey.PubKey().IsEqual(pubKey) {
		return privKey, nil
	} else {
		return privKey, errors.New("Extracted private key does not match public key")
	}
}
