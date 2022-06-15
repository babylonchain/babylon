package bls12381

import blst "github.com/supranational/blst/bindings/go"

// GeneKeyPair generates a random bls key pair based on a given seed
// the public key is compressed with 96 byte size
func GeneKeyPair(seed []byte) (*blst.SecretKey, PublicKey) {
	sk := blst.KeyGen(seed)
	pk := new(BlsPubKey).From(sk)
	return sk, pk.Compress()
}

// Sign signs on a msg using a bls secret key
// the returned sig is compressed version with 48 byte size
func Sign(sk *blst.SecretKey, msg []byte) []byte {
	return new(BlsSig).Sign(sk, msg, dst).Compress()
}

// Verify verifies a bls sig over msg with a bls public key
// the sig and public key are all compressed
func Verify(sig Signature, pk PublicKey, msg []byte) bool {
	dummySig := new(BlsSig)
	return dummySig.VerifyCompressed(sig, false, pk, false, msg, dst)
}

// AggrSigs aggregates bls sigs into a single bls signature
func AggrSigs(sigs []Signature) (Signature, bool) {
	aggSig := new(BlsMultiSig)
	sigBytes := make([][]byte, len(sigs))
	for i := 0; i < len(sigs); i++ {
		sigBytes[i] = sigs[i].ToByte()
	}
	if !aggSig.AggregateCompressed(sigBytes, false) {
		return nil, false
	}
	return aggSig.ToAffine().Compress(), true
}

// AggrPKs aggregates bls public keys into a single bls public key
func AggrPKs(pks []PublicKey) ([]byte, bool) {
	aggPk := new(BlsMultiPubKey)
	pkBytes := make([][]byte, len(pks))
	for i := 0; i < len(pks); i++ {
		pkBytes[i] = pks[i].ToByte()
	}
	if !aggPk.AggregateCompressed(pkBytes, false) {
		return nil, false
	}
	return aggPk.ToAffine().Compress(), true
}

// VerifyMultiSig verifies a bls sig (compressed) over a message with
// a group of bls public keys (compressed)
func VerifyMultiSig(sig Signature, pks []PublicKey, msg []byte) bool {
	aggPk, ok := AggrPKs(pks)
	if !ok {
		return false
	}
	return Verify(sig, aggPk, msg)
}
