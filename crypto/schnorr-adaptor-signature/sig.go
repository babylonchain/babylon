package schnorr_adaptor_signature

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	secp "github.com/decred/dcrd/dcrec/secp256k1/v4"
)

var (
	// rfc6979ExtraDataV0 is the extra data to feed to RFC6979 when
	// generating the deterministic nonce for the BIP-340 scheme.  This
	// ensures the same nonce is not generated for the same message and key
	// as for other signing algorithms such as ECDSA.
	//
	// It is equal to SHA-256([]byte("BIP-340")).
	rfc6979ExtraDataV0 = [chainhash.HashSize]uint8{
		0xa3, 0xeb, 0x4c, 0x18, 0x2f, 0xae, 0x7e, 0xf4,
		0xe8, 0x10, 0xc6, 0xee, 0x13, 0xb0, 0xe9, 0x26,
		0x68, 0x6d, 0x71, 0xe8, 0x7f, 0x39, 0x4f, 0x79,
		0x9c, 0x00, 0xa5, 0x21, 0x03, 0xcb, 0x4e, 0x17,
	}
)

// AdaptorSignature is the structure for an adaptor signature
// the adaptor signature is a triple (R, s', need_negation) where
//   - `R` is the tweaked public randomness, which is derived from
//     offsetting public randomness R' sampled by the signer by
//     using encryption key T
//   - `sHat` is the secret s' in the adaptor signature
//   - `needNegation` is a bool value indicating whether decryption
//     key needs to be negated when decrypting a Schnorr signature
//     It is needed since (R, s') does not tell whether R'+T has odd
//     or even y index, thus does not tell whether decryption key needs
//     to be negated upon decryption.
type AdaptorSignature struct {
	r            btcec.JacobianPoint
	sHat         btcec.ModNScalar
	needNegation bool
}

func newAdaptorSignature(r *btcec.JacobianPoint, sHat *btcec.ModNScalar, needNegation bool) *AdaptorSignature {
	var sig AdaptorSignature
	sig.r.Set(r)
	sig.sHat.Set(sHat)
	sig.needNegation = needNegation
	return &sig
}

// EncVerify verifies that the adaptor signature is valid w.r.t. the given
// public key, encryption key and message hash
func (sig *AdaptorSignature) EncVerify(pk *btcec.PublicKey, encKey *EncryptionKey, msgHash []byte) error {
	pkBytes := schnorr.SerializePubKey(pk)
	return encVerify(sig, msgHash, pkBytes, &encKey.JacobianPoint)
}

// Decrypt decrypts the adaptor signature to a Schnorr signature by
// using the decryption key `decKey`, noted by `t` in the paper
func (sig *AdaptorSignature) Decrypt(decKey *DecryptionKey) *schnorr.Signature {
	R := sig.r

	t := decKey.ModNScalar
	if sig.needNegation {
		t.Negate()
	}
	// s = s' + t (or s'-t if negation is needed)
	s := sig.sHat
	s.Add(&t)

	return schnorr.NewSignature(&R.X, &s)
}

// Recover recovers the decryption key by using the adaptor signature
// and the Schnorr signature decrypted from it
func (sig *AdaptorSignature) Recover(decryptedSchnorrSig *schnorr.Signature) *DecryptionKey {
	// unpack s and R from Schnorr signature
	_, s := unpackSchnorrSig(decryptedSchnorrSig)
	sHat := sig.sHat

	// extract encryption key t = s - s'
	sHat.Negate()
	t := s.Add(&sHat)

	if sig.needNegation {
		t.Negate()
	}

	return &DecryptionKey{*t}
}

// Marshal is to implement proto interface
func (sig *AdaptorSignature) Marshal() ([]byte, error) {
	if sig == nil {
		return nil, nil
	}
	var asigBytes []byte
	// append r
	rBytes := btcec.JacobianToByteSlice(sig.r)
	asigBytes = append(asigBytes, rBytes...)
	// append sHat
	sHatBytes := sig.sHat.Bytes()
	asigBytes = append(asigBytes, sHatBytes[:]...)
	// append needNegation
	if sig.needNegation {
		asigBytes = append(asigBytes, 0x01)
	} else {
		asigBytes = append(asigBytes, 0x00)
	}
	return asigBytes, nil
}

func (sig *AdaptorSignature) MustMarshal() []byte {
	if sig == nil {
		return nil
	}
	bz, err := sig.Marshal()
	if err != nil {
		panic(err)
	}

	return bz
}

func (sig *AdaptorSignature) MarshalHex() string {
	return hex.EncodeToString(sig.MustMarshal())
}

// Size is to implement proto interface
func (sig *AdaptorSignature) Size() int {
	return AdaptorSignatureSize
}

// MarshalTo is to implement proto interface
func (sig *AdaptorSignature) MarshalTo(data []byte) (int, error) {
	bz, err := sig.Marshal()
	if err != nil {
		return 0, err
	}
	copy(data, bz)
	return len(data), nil
}

// Unmarshal is to implement proto interface
func (sig *AdaptorSignature) Unmarshal(data []byte) error {
	adaptorSig, err := NewAdaptorSignatureFromBytes(data)
	if err != nil {
		return err
	}

	*sig = *adaptorSig

	return nil
}

func (sig *AdaptorSignature) Equals(sig2 AdaptorSignature) bool {
	return bytes.Equal(sig.MustMarshal(), sig2.MustMarshal())
}

// EncSign generates an adaptor signature by using the given secret key,
// encryption key (noted by `T` in the paper) and message hash
func EncSign(sk *btcec.PrivateKey, encKey *EncryptionKey, msgHash []byte) (*AdaptorSignature, error) {
	// d' = int(d)
	var skScalar btcec.ModNScalar
	skScalar.Set(&sk.Key)

	// Fail if msgHash is not 32 bytes
	if len(msgHash) != chainhash.HashSize {
		return nil, fmt.Errorf("wrong size for message hash (got %v, want %v)", len(msgHash), chainhash.HashSize)
	}

	// Fail if d = 0 or d >= n
	if skScalar.IsZero() {
		return nil, fmt.Errorf("private key is zero")
	}

	// P = 'd*G
	pk := sk.PubKey()

	// Negate d if P.y is odd.
	pubKeyBytes := pk.SerializeCompressed()
	if pubKeyBytes[0] == secp.PubKeyFormatCompressedOdd {
		skScalar.Negate()
	}

	var privKeyBytes [chainhash.HashSize]byte
	skScalar.PutBytes(&privKeyBytes)
	for iteration := uint32(0); ; iteration++ {
		// Use RFC6979 to generate a deterministic nonce in [1, n-1]
		// parameterized by the private key, message being signed, extra data
		// that identifies the scheme, and an iteration count
		nonce := btcec.NonceRFC6979(
			privKeyBytes[:], msgHash, rfc6979ExtraDataV0[:], nil, iteration,
		)

		// try to generate adaptor signature
		sig, err := encSign(&skScalar, nonce, pk, msgHash, &encKey.JacobianPoint)
		if err != nil {
			// Try again with a new nonce.
			continue
		}

		return sig, nil
	}
}

// NewAdaptorSignatureFromBytes parses the given byte array to an adaptor signature
func NewAdaptorSignatureFromBytes(asigBytes []byte) (*AdaptorSignature, error) {
	if len(asigBytes) != AdaptorSignatureSize {
		return nil, fmt.Errorf(
			"the length of the given bytes for adaptor signature is incorrect (expected: %d, actual: %d)",
			AdaptorSignatureSize,
			len(asigBytes),
		)
	}

	// extract r
	r, err := btcec.ParseJacobian(asigBytes[0:JacobianPointSize])
	if err != nil {
		return nil, err
	}
	// extract sHat
	var sHat btcec.ModNScalar
	sHat.SetByteSlice(asigBytes[JacobianPointSize : JacobianPointSize+ModNScalarSize]) //nolint:errcheck
	// extract needNegation
	needNegation := asigBytes[AdaptorSignatureSize-1] != 0x00

	return newAdaptorSignature(&r, &sHat, needNegation), nil
}

// NewAdaptorSignatureFromHex parses the given hex string to an adaptor signature
func NewAdaptorSignatureFromHex(asigHex string) (*AdaptorSignature, error) {
	asigBytes, err := hex.DecodeString(asigHex)
	if err != nil {
		return nil, err
	}
	return NewAdaptorSignatureFromBytes(asigBytes)
}
