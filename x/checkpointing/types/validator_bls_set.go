package types

import (
	"fmt"

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

	// ensure bitmap is big enough to contain ks
	if bm.Len() < len(ks) {
		return valSet, sum, fmt.Errorf("bitmap (with %d bits) is not large enough to contain the validator set with size %d", bm.Len(), len(ks))
	}

	// NOTE: we cannot use bm.Len() to iterate over the bitmap
	// Our bm has 13 bytes = 104 bits, while the validator set has 100 validators
	// If iterating over the 104 bits, then the last 4 bits will trigger the our of range error in ks
	for i := 0; i < len(ks); i++ {
		if bm.Get(i) {
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
