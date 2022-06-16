package bls12381

import (
	"github.com/pkg/errors"
	blst "github.com/supranational/blst/bindings/go"
)

// GenKeyPair generates a random bls key pair based on a given seed
// the public key is compressed with 96 byte size
func GenKeyPair(seed []byte) (*blst.SecretKey, PublicKey) {
	sk := blst.KeyGen(seed)
	pk := new(BlsPubKey).From(sk)
	return sk, pk.Compress()
}

// Sign signs on a msg using a bls secret key
// the returned sig is compressed version with 48 byte size
func Sign(sk *blst.SecretKey, msg []byte) []byte {
	return new(BlsSig).Sign(sk, msg, DST).Compress()
}

// Verify verifies a bls sig over msg with a bls public key
// the sig and public key are all compressed
func Verify(sig Signature, pk PublicKey, msg []byte) (bool, error) {
	dummySig := new(BlsSig)
	return dummySig.VerifyCompressed(sig, false, pk, false, msg, DST), nil
}

// AggrSigs aggregates bls sigs into a single bls signature
func AggrSigs(sigs []Signature) (Signature, error) {
	aggSig := new(BlsMultiSig)
	sigBytes := make([][]byte, len(sigs))
	for i := 0; i < len(sigs); i++ {
		sigBytes[i] = sigs[i].Byte()
	}
	if !aggSig.AggregateCompressed(sigBytes, false) {
		return nil, errors.New("failed to aggregate bls signatures")
	}
	return aggSig.ToAffine().Compress(), nil
}

// AggrPKs aggregates bls public keys into a single bls public key
func AggrPKs(pks []PublicKey) ([]byte, error) {
	aggPk := new(BlsMultiPubKey)
	pkBytes := make([][]byte, len(pks))
	for i := 0; i < len(pks); i++ {
		pkBytes[i] = pks[i].Byte()
	}
	if !aggPk.AggregateCompressed(pkBytes, false) {
		return nil, errors.New("failed to aggregate bls public keys")
	}
	return aggPk.ToAffine().Compress(), nil
}

// VerifyMultiSig verifies a bls sig (compressed) over a message with
// a group of bls public keys (compressed)
func VerifyMultiSig(sig Signature, pks []PublicKey, msg []byte) (bool, error) {
	aggPk, err := AggrPKs(pks)
	if err != nil {
		return false, err
	}
	return Verify(sig, aggPk, msg)
}
