package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
)

type Validator struct {
	Addr  sdk.ValAddress
	Power int64
}

type ValidatorSet struct {
	vals []*Validator
}

func (vs ValidatorSet) Len() int {
	return len(vs.vals)
}

func (vs ValidatorSet) Less(i, j int) bool {
	return vs.vals[i].Power < vs.vals[j].Power || (vs.vals[i].Power == vs.vals[j].Power && sdk.BigEndianToUint64(vs.vals[i].Addr) < sdk.BigEndianToUint64(vs.vals[j].Addr))
}

func (vs ValidatorSet) Swap(i, j int) {
	vs.vals[i], vs.vals[j] = vs.vals[j], vs.vals[i]
}

func (vs ValidatorSet) FindValidatorWithIndex(valAddr sdk.ValAddress) (*Validator, int, error) {
	for i, v := range vs.vals {
		if v.Addr.Equals(valAddr) {
			return v, i, nil
		}
	}
	return nil, 0, errors.New("validator address does not exist in the validator set")
}

func (vs *ValidatorSet) AppendValidator(validator *Validator) {
	vs.vals = append(vs.vals, validator)
}
