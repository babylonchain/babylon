package bls12381

import (
	"crypto/rand"
	"io"

	cmtcrypto "github.com/cometbft/cometbft/crypto"
	"github.com/pkg/errors"
	blst "github.com/supranational/blst/bindings/go"
)

// GenKeyPair generates a random BLS key pair based on a given seed
// the public key is compressed with 96 byte size
func GenKeyPair() (PrivateKey, PublicKey) {
	skSerialized := GenPrivKey()
	sk := new(blst.SecretKey)
	sk.Deserialize(skSerialized)
	pk := new(BlsPubKey).From(sk)
	return skSerialized, pk.Compress()
}

func GenPrivKey() PrivateKey {
	return genPrivKey(rand.Reader)
}

func GenPrivKeyFromSecret(secret []byte) PrivateKey {
	seed := cmtcrypto.Sha256(secret)

	return genPrivKeyFromSeed(seed)
}

func genPrivKey(rand io.Reader) PrivateKey {
	seed := make([]byte, SeedSize)

	_, err := io.ReadFull(rand, seed)
	if err != nil {
		panic(err)
	}

	return genPrivKeyFromSeed(seed)
}

func genPrivKeyFromSeed(seed []byte) PrivateKey {
	return blst.KeyGen(seed).Serialize()
}

// Sign signs on a msg using a BLS secret key
// the returned sig is compressed version with 48 byte size
func Sign(sk PrivateKey, msg []byte) Signature {
	secretKey := new(blst.SecretKey)
	secretKey.Deserialize(sk)
	return new(BlsSig).Sign(secretKey, msg, DST).Compress()
}

// Verify verifies a BLS sig over msg with a BLS public key
// the sig and public key are all compressed
func Verify(sig Signature, pk PublicKey, msg []byte) (bool, error) {
	dummySig := new(BlsSig)
	return dummySig.VerifyCompressed(sig, false, pk, false, msg, DST), nil
}

// AggrSig aggregates BLS signatures in an accumulative manner
func AggrSig(existingSig Signature, newSig Signature) (Signature, error) {
	if existingSig == nil {
		return newSig, nil
	}
	sigs := []Signature{existingSig, newSig}
	return AggrSigList(sigs)
}

// AggrSigList aggregates BLS sigs into a single BLS signature
func AggrSigList(sigs []Signature) (Signature, error) {
	aggSig := new(BlsMultiSig)
	sigBytes := make([][]byte, len(sigs))
	for i := 0; i < len(sigs); i++ {
		sigBytes[i] = sigs[i].Bytes()
	}
	if !aggSig.AggregateCompressed(sigBytes, false) {
		return nil, errors.New("failed to aggregate bls signatures")
	}
	return aggSig.ToAffine().Compress(), nil
}

// AggrPK aggregates BLS public keys in an accumulative manner
func AggrPK(existingPK PublicKey, newPK PublicKey) (PublicKey, error) {
	if existingPK == nil {
		return newPK, nil
	}
	pks := []PublicKey{existingPK, newPK}
	return AggrPKList(pks)
}

// AggrPKList aggregates BLS public keys into a single BLS public key
func AggrPKList(pks []PublicKey) (PublicKey, error) {
	aggPk := new(BlsMultiPubKey)
	pkBytes := make([][]byte, len(pks))
	for i := 0; i < len(pks); i++ {
		pkBytes[i] = pks[i].Bytes()
	}
	if !aggPk.AggregateCompressed(pkBytes, false) {
		return nil, errors.New("failed to aggregate bls public keys")
	}
	return aggPk.ToAffine().Compress(), nil
}

// VerifyMultiSig verifies a BLS sig (compressed) over a message with
// a group of BLS public keys (compressed)
func VerifyMultiSig(sig Signature, pks []PublicKey, msg []byte) (bool, error) {
	aggPk, err := AggrPKList(pks)
	if err != nil {
		return false, err
	}
	return Verify(sig, aggPk, msg)
}
