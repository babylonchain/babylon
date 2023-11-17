package schnorr_adaptor_signature

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// DecryptionKey is the decryption key in the adaptor
// signature scheme, noted by t in the paper
type DecryptionKey struct {
	btcec.ModNScalar
}

func NewDecyptionKeyFromModNScalar(scalar *btcec.ModNScalar) (*DecryptionKey, error) {
	if scalar.IsZero() {
		return nil, fmt.Errorf("the given scalar is zero")
	}

	return &DecryptionKey{*scalar}, nil
}

func NewDecyptionKeyFromBTCSK(btcSK *btcec.PrivateKey) (*DecryptionKey, error) {
	return NewDecyptionKeyFromModNScalar(&btcSK.Key)
}

func NewDecyptionKeyFromBytes(decKeyBytes []byte) (*DecryptionKey, error) {
	if len(decKeyBytes) != ModNScalarSize {
		return nil, fmt.Errorf(
			"the length of the given bytes for decryption key is incorrect (expected: %d, actual: %d)",
			ModNScalarSize,
			len(decKeyBytes),
		)
	}

	var decKeyScalar btcec.ModNScalar
	decKeyScalar.SetByteSlice(decKeyBytes) //nolint:errcheck

	return NewDecyptionKeyFromModNScalar(&decKeyScalar)
}

func (dk *DecryptionKey) GetEncKey() *EncryptionKey {
	var ekPoint btcec.JacobianPoint
	btcec.ScalarBaseMultNonConst(&dk.ModNScalar, &ekPoint)
	// NOTE: we convert ekPoint to affine coordinates for consistency
	ekPoint.ToAffine()
	return &EncryptionKey{ekPoint}
}

func (dk *DecryptionKey) ToBTCSK() *btcec.PrivateKey {
	return &btcec.PrivateKey{Key: dk.ModNScalar}
}

func (dk *DecryptionKey) ToBytes() []byte {
	scalarBytes := dk.ModNScalar.Bytes()
	return scalarBytes[:]
}

type EncryptionKey struct {
	btcec.JacobianPoint
}

func NewEncryptionKeyFromJacobianPoint(point *btcec.JacobianPoint) (*EncryptionKey, error) {
	// ensure the point is not at infinity
	if (point.X.IsZero() && point.Y.IsZero()) || point.Z.IsZero() {
		return nil, fmt.Errorf("the given Jacobian point is at infinity")
	}

	// convert point to affine coordinates if necessary
	affinePoint := *point
	if !affinePoint.Z.IsOne() {
		affinePoint.ToAffine()
	}

	return &EncryptionKey{affinePoint}, nil
}

func NewEncryptionKeyFromBTCPK(btcPK *btcec.PublicKey) (*EncryptionKey, error) {
	var btcPKPoint btcec.JacobianPoint
	btcPK.AsJacobian(&btcPKPoint)
	return NewEncryptionKeyFromJacobianPoint(&btcPKPoint)
}

func NewEncryptionKeyFromBytes(encKeyBytes []byte) (*EncryptionKey, error) {
	point, err := btcec.ParseJacobian(encKeyBytes)
	if err != nil {
		return nil, err
	}
	return NewEncryptionKeyFromJacobianPoint(&point)
}

func (ek *EncryptionKey) ToBTCPK() *btcec.PublicKey {
	affineEK := *ek
	return secp256k1.NewPublicKey(&affineEK.X, &affineEK.Y)
}

func (ek *EncryptionKey) ToBytes() []byte {
	return btcec.JacobianToByteSlice(ek.JacobianPoint)
}

func GenKeyPair() (*EncryptionKey, *DecryptionKey, error) {
	sk, err := btcec.NewPrivateKey()
	if err != nil {
		return nil, nil, err
	}
	dk, err := NewDecyptionKeyFromBTCSK(sk)
	if err != nil {
		return nil, nil, err
	}
	ek := dk.GetEncKey()
	return ek, dk, nil
}
