package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"sort"
)

type Validator struct {
	Addr  sdk.ValAddress
	Power int64
}

type ValidatorSet []*Validator

func NewValidatorSet(vals []*Validator) ValidatorSet {
	sort.Slice(vals, func(i, j int) bool {
		return sdk.BigEndianToUint64(vals[i].Addr) < sdk.BigEndianToUint64(vals[j].Addr)
	})
	return vals
}

func (vs ValidatorSet) FindValidatorWithIndex(valAddr sdk.ValAddress) (*Validator, int, error) {
	index := vs.binarySearch(valAddr)
	if index == -1 {
		return nil, 0, errors.New("validator address does not exist in the validator set")
	}
	return vs[index], index, nil
}

func (vs ValidatorSet) binarySearch(targetAddr sdk.ValAddress) int {
	var lo = 0
	var hi = len(vs) - 1

	for lo <= hi {
		var mid = lo + (hi-lo)/2
		midAddr := vs[mid].Addr

		if midAddr.Equals(targetAddr) {
			return mid
		} else if sdk.BigEndianToUint64(midAddr) > sdk.BigEndianToUint64(targetAddr) {
			hi = mid - 1
		} else {
			lo = mid + 1
		}
	}

	return -1
}
