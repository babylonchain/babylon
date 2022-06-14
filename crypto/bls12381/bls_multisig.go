package bls12381

import blst "github.com/supranational/blst/bindings/go"

// GenerateBlsKeyPair generates a random bls key pair based on a given seed
// the public key is compressed with 96 byte size
func GenerateBlsKeyPair(seed []byte) (*blst.SecretKey, []byte) {
	sk := blst.KeyGen(seed)
	pk := new(BlsPubKey).From(sk)
	return sk, pk.Compress()
}

// SignMsg signs on a msg using a bls secret key
// the returned sig is compressed version with 48 byte size
func SignMsg(sk *blst.SecretKey, msg []byte) []byte {
	return new(BlsSig).Sign(sk, msg, dst).Compress()
}

// VerifyBlsSig verifies a bls sig over msg with a bls public key
// the sig and public key are all compressed
func VerifyBlsSig(sig []byte, pk []byte, msg []byte) bool {
	dummySig := new(BlsSig)
	return dummySig.VerifyCompressed(sig, false, pk, false, msg, dst)
}

// AggregateBlsSigs aggregates bls sigs into a single bls signature
func AggregateBlsSigs(sigs [][]byte) ([]byte, bool) {
	aggSig := new(BlsMultiSig)
	if !aggSig.AggregateCompressed(sigs, false) {
		return nil, false
	}
	return aggSig.ToAffine().Compress(), true
}

// AggregateBlsPubKeys aggregates bls public keys into a single bls public key
func AggregateBlsPubKeys(pks [][]byte) ([]byte, bool) {
	aggPk := new(BlsMultiPubKey)
	if !aggPk.AggregateCompressed(pks, false) {
		return nil, false
	}
	return aggPk.ToAffine().Compress(), true
}

// VerifyBlsMultiSig verifies a bls sig (compressed) over a message with
// a group of bls public keys (compressed)
func VerifyBlsMultiSig(sig []byte, pks [][]byte, msg []byte) bool {
	aggPk, ok := AggregateBlsPubKeys(pks)
	if !ok {
		return false
	}
	return VerifyBlsSig(sig, aggPk, msg)
}
