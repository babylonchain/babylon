package types

import (
	"bytes"
	"encoding/json"
	fmt "fmt"
	"sort"

	"github.com/boljen/go-bitmap"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
)

func (v *Validator) GetValAddress() sdk.ValAddress {
	return sdk.ValAddress(v.Addr)
}

func (v *Validator) GetValAddressStr() string {
	return v.GetValAddress().String()
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

func (vs ValidatorSet) FindSubset(bm bitmap.Bitmap) (ValidatorSet, error) {
	valSet := make([]Validator, 0)

	// ensure bitmap is big enough to contain vs
	if bm.Len() < len(vs) {
		return valSet, fmt.Errorf("bitmap (with %d bits) is not large enough to contain the validator set with size %d", bm.Len(), len(vs))
	}

	// NOTE: we cannot use bm.Len() to iterate over the bitmap
	// Our bm has 13 bytes = 104 bits, while the validator set has 100 validators
	// If iterating over the 104 bits, then the last 4 bits will trigger the our of range error in ks
	for i := 0; i < len(vs); i++ {
		if bm.Get(i) {
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

		if bytes.Equal(midAddr, targetAddr) {
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
