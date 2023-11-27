package schnorr_adaptor_signature

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

const (
	ModNScalarSize       = 32
	FieldValSize         = 32
	JacobianPointSize    = 33
	AdaptorSignatureSize = JacobianPointSize + ModNScalarSize + 1
)

func encSign(privKey, nonce *btcec.ModNScalar, pubKey *btcec.PublicKey, m []byte, T *btcec.JacobianPoint) (*AdaptorSignature, error) {
	// R' = kG
	var RHat btcec.JacobianPoint
	k := *nonce
	btcec.ScalarBaseMultNonConst(&k, &RHat)

	// get R = R'+T
	var R btcec.JacobianPoint
	btcec.AddNonConst(&RHat, T, &R)
	// negate k and R if R.y is odd
	affineRWithEvenY, needNegation := intoPointWithEvenY(&R)
	R = *affineRWithEvenY
	if needNegation {
		k.Negate()
	}

	// e = tagged_hash("BIP0340/challenge", bytes(R) || bytes(P) || m) mod n
	var rBytes [chainhash.HashSize]byte
	r := &R.X
	r.PutBytesUnchecked(rBytes[:])
	pBytes := schnorr.SerializePubKey(pubKey)
	commitment := chainhash.TaggedHash(
		chainhash.TagBIP0340Challenge, rBytes[:], pBytes, m,
	)
	var e btcec.ModNScalar
	e.SetBytes((*[ModNScalarSize]byte)(commitment))

	// s' = k + e*d mod n
	sHat := new(btcec.ModNScalar).Mul2(&e, privKey).Add(&k)

	// compose signature
	sig := newAdaptorSignature(&R, sHat, needNegation)

	// perform verification here. Failing to verify the generated signature
	// can only be because of bad nonces. The caller function `EncSign` will
	// keep trying `encSign` until finding a nonce that generates correct
	// signature
	if err := encVerify(sig, m, pBytes, T); err != nil {
		return nil, fmt.Errorf("the provided nonce does not work: %w", err)
	}

	// Return signature
	return sig, nil
}

func encVerify(sig *AdaptorSignature, m []byte, pubKeyBytes []byte, T *btcec.JacobianPoint) error {
	// Fail if m is not 32 bytes
	if len(m) != chainhash.HashSize {
		return fmt.Errorf("wrong size for message (got %v, want %v)",
			len(m), chainhash.HashSize)
	}

	// R' = R-T (or R+T if it needs negation)
	R := &sig.r // NOTE: R is an affine point
	var RHat btcec.JacobianPoint
	if sig.needNegation {
		btcec.AddNonConst(R, T, &RHat)
	} else {
		btcec.AddNonConst(R, negatePoint(T), &RHat)
	}

	RHat.ToAffine()

	// P = lift_x(int(pk))
	pubKey, err := schnorr.ParsePubKey(pubKeyBytes)
	if err != nil {
		return err
	}
	// Fail if P is not a point on the curve
	if !pubKey.IsOnCurve() {
		return fmt.Errorf("pubkey point is not on curve")
	}

	// e = int(tagged_hash("BIP0340/challenge", bytes(R) || bytes(P) || M)) mod n.
	var rBytes [chainhash.HashSize]byte
	R.X.PutBytesUnchecked(rBytes[:])
	pBytes := schnorr.SerializePubKey(pubKey)
	commitment := chainhash.TaggedHash(
		chainhash.TagBIP0340Challenge, rBytes[:], pBytes, m,
	)
	var e btcec.ModNScalar
	e.SetBytes((*[ModNScalarSize]byte)(commitment))

	// Negate e here so we can use AddNonConst below to subtract the s'*G
	// point from e*P.
	e.Negate()

	// expected R' = s'*G - e*P
	var P, expRHat, sHatG, eP btcec.JacobianPoint
	pubKey.AsJacobian(&P)
	btcec.ScalarBaseMultNonConst(&sig.sHat, &sHatG) // s'*G
	btcec.ScalarMultNonConst(&e, &P, &eP)           // -e*P
	btcec.AddNonConst(&sHatG, &eP, &expRHat)        // R' = s'*G-e*P

	// Fail if expected R' is the point at infinity
	if (expRHat.X.IsZero() && expRHat.Y.IsZero()) || expRHat.Z.IsZero() {
		return fmt.Errorf("expected R' point is at infinity")
	}

	expRHat.ToAffine()

	// fail if expected R'.y is odd
	if expRHat.Y.IsOdd() {
		return fmt.Errorf("expected R'.y is odd")
	}

	// ensure R' is same as the expected R' = s'*G - e*P
	if !expRHat.X.Equals(&RHat.X) {
		return fmt.Errorf("expected R' = s'*G - e*P is different from the actual R'")
	}

	return nil
}

// intoPointWithEvenY converts the given Jacobian point to an affine
// point with even y value, and returns a bool value on whether the
// negation is performed.
// The bool value will be used for decrypting an adaptor signature
// to a Schnorr signature.
func intoPointWithEvenY(point *btcec.JacobianPoint) (*btcec.JacobianPoint, bool) {
	affinePoint := point
	affinePoint.ToAffine()

	needNegation := affinePoint.Y.IsOdd()

	if needNegation {
		affinePoint = negatePoint(affinePoint)
	}

	return affinePoint, needNegation
}

// negatePoint negates a point (either Jacobian or affine)
func negatePoint(point *btcec.JacobianPoint) *btcec.JacobianPoint {
	nPoint := *point
	nPoint.Y.Negate(1).Normalize()
	return &nPoint
}

// unpackSchnorrSig
func unpackSchnorrSig(sig *schnorr.Signature) (*btcec.FieldVal, *btcec.ModNScalar) {
	sigBytes := sig.Serialize()
	var r btcec.FieldVal
	r.SetByteSlice(sigBytes[0:32])
	var s btcec.ModNScalar
	s.SetByteSlice(sigBytes[32:64])
	return &r, &s
}
