package types

import (
	"errors"
	"github.com/babylonchain/babylon/crypto/bls12381"
	"github.com/boljen/go-bitmap"
)

type ValidatorWithBLSSet []*ValidatorWithBlsKey

// FindSubsetWithPowerSum returns a subset and the sum of the voting Power
// based on the given bitmap
func (ks ValidatorWithBLSSet) FindSubsetWithPowerSum(bm bitmap.Bitmap) (ValidatorWithBLSSet, uint64, error) {
	var (
		sum    uint64
		valSet ValidatorWithBLSSet
	)

	for i := 0; i < bm.Len(); i++ {
		if bm.Get(i) {
			if i >= len(ks) {
				return valSet, sum, errors.New("invalid validator index")
			}
			valSet = append(valSet, ks[i])
			sum += ks[i].VotingPower
		}
	}
	return valSet, sum, nil
}

func (ks ValidatorWithBLSSet) GetBLSKeySet() []bls12381.PublicKey {
	var blsKeySet []bls12381.PublicKey
	for _, val := range ks {
		blsKeySet = append(blsKeySet, val.BlsPubKey)
	}
	return blsKeySet
}

func (ks ValidatorWithBLSSet) GetTotalPower() uint64 {
	var total uint64
	for _, key := range ks {
		total += key.VotingPower
	}

	return total
}
