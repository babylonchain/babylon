package datagen

import (
	"math/rand"

	bbn "github.com/babylonchain/babylon/types"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

func GenRandomBTCKeyPair(r *rand.Rand) (*btcec.PrivateKey, *btcec.PublicKey, error) {
	sk, err := secp256k1.GeneratePrivateKeyFromRand(r)
	if err != nil {
		return nil, nil, err
	}
	return sk, sk.PubKey(), nil
}

func GenRandomBTCKeyPairs(r *rand.Rand, n int) ([]*btcec.PrivateKey, []*btcec.PublicKey, error) {
	sks, pks := []*btcec.PrivateKey{}, []*btcec.PublicKey{}
	for i := 0; i < n; i++ {
		sk, pk, err := GenRandomBTCKeyPair(r)
		if err != nil {
			return nil, nil, err
		}
		sks = append(sks, sk)
		pks = append(pks, pk)
	}

	return sks, pks, nil
}

func GenRandomBIP340PubKey(r *rand.Rand) (*bbn.BIP340PubKey, error) {
	sk, err := secp256k1.GeneratePrivateKeyFromRand(r)
	if err != nil {
		return nil, err
	}
	pk := sk.PubKey()
	btcPK := bbn.NewBIP340PubKeyFromBTCPK(pk)
	return btcPK, nil
}

// GenCovenantCommittee generates a covenant committee
// with random number of members and quorum size
func GenCovenantCommittee(r *rand.Rand) ([]*btcec.PrivateKey, []*btcec.PublicKey, uint32) {
	committeeSize := RandomInt(r, 5) + 5
	quorumSize := uint32(committeeSize/2 + 1)
	sks, pks := []*btcec.PrivateKey{}, []*btcec.PublicKey{}
	for i := uint64(0); i < committeeSize; i++ {
		skBytes := GenRandomByteArray(r, 32)
		sk, pk := btcec.PrivKeyFromBytes(skBytes)
		sks = append(sks, sk)
		pks = append(pks, pk)
	}
	return sks, pks, quorumSize
}
