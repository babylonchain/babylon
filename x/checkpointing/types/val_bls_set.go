package types

import (
	"github.com/babylonchain/babylon/crypto/bls12381"
	"github.com/boljen/go-bitmap"
	"github.com/cosmos/cosmos-sdk/codec"
)

func ValidatorBlsKeySetToBytes(cdc codec.BinaryCodec, valBlsSet *ValidatorWithBlsKeySet) []byte {
	return cdc.MustMarshal(valBlsSet)
}

func BytesToValidatorBlsKeySet(cdc codec.BinaryCodec, bz []byte) (*ValidatorWithBlsKeySet, error) {
	valBlsSet := new(ValidatorWithBlsKeySet)
	err := cdc.Unmarshal(bz, valBlsSet)
	return valBlsSet, err
}

// FindSubsetWithPowerSum returns a subset and the sum of the voting Power
// based on the given bitmap
func (ks *ValidatorWithBlsKeySet) FindSubsetWithPowerSum(bm bitmap.Bitmap) (*ValidatorWithBlsKeySet, uint64) {
	var sum uint64
	valSet := &ValidatorWithBlsKeySet{
		ValSet: make([]*ValidatorWithBlsKey, 0),
	}
	for i := 0; i < len(ks.ValSet); i++ {
		if bm.Get(i) {
			valSet.ValSet = append(valSet.ValSet, ks.ValSet[i])
			sum += ks.ValSet[i].VotingPower
		}
	}
	return valSet, sum
}

func (ks *ValidatorWithBlsKeySet) GetBLSKeySet() []bls12381.PublicKey {
	var blsKeySet []bls12381.PublicKey
	for _, val := range ks.ValSet {
		blsKeySet = append(blsKeySet, val.BlsPubKey)
	}
	return blsKeySet
}

func (ks *ValidatorWithBlsKeySet) GetTotalPower() uint64 {
	var total uint64
	for _, val := range ks.ValSet {
		total += val.VotingPower
	}

	return total
}
