package types

import (
	"encoding/json"
	"sort"

	"github.com/boljen/go-bitmap"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
)

type ValState int

const (
	ValStateCreateRequestSubmitted = iota
	ValStateCreated
	ValStateBondingRequestSubmitted
	ValStateBonded
	ValStateUnbondingRequestSubmitted
	ValStateUnbonding
	ValStateUnbonded
)

type Validator struct {
	Addr  sdk.ValAddress `json:"addr"`
	Power int64          `json:"power"`
}

type ValidatorSet []Validator

// NewSortedValidatorSet returns a sorted ValidatorSet by validator's address in the ascending order
func NewSortedValidatorSet(vals []Validator) ValidatorSet {
	sort.Slice(vals, func(i, j int) bool {
		return sdk.BigEndianToUint64(vals[i].Addr) < sdk.BigEndianToUint64(vals[j].Addr)
	})
	return vals
}

func NewValidatorSetFromBytes(vsBytes []byte) (ValidatorSet, error) {
	var vs ValidatorSet
	err := json.Unmarshal(vsBytes, &vs)
	return vs, err
}

// FindValidatorWithIndex returns the validator and its index
// an error is returned if the validator does not exist in the set
func (vs ValidatorSet) FindValidatorWithIndex(valAddr sdk.ValAddress) (*Validator, int, error) {
	index := vs.binarySearch(valAddr)
	if index == -1 {
		return nil, 0, errors.New("validator address does not exist in the validator set")
	}
	return &vs[index], index, nil
}

func (vs ValidatorSet) FindSubset(bitmap bitmap.Bitmap) (ValidatorSet, error) {
	valSet := make([]Validator, 0)
	for i := 0; i < bitmap.Len(); i++ {
		if bitmap.Get(i) {
			if i >= len(vs) {
				return nil, errors.New("invalid validator index")
			}
			valSet = append(valSet, vs[i])
		}
	}
	return valSet, nil
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

func (vs ValidatorSet) Marshal() ([]byte, error) {
	return json.Marshal(vs)
}

func (vs ValidatorSet) MustMarshal() []byte {
	vsBytes, err := vs.Marshal()
	if err != nil {
		panic(err)
	}
	return vsBytes
}
