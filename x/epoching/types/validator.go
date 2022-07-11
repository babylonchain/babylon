package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Validator struct {
	Addr  sdk.ValAddress
	Power int64
}

type ValidatorSet []Validator

func (v ValidatorSet) Len() int {
	return len(v)
}

func (v ValidatorSet) Less(i, j int) bool {
	return v[i].Power < v[j].Power || (v[i].Power == v[j].Power && sdk.BigEndianToUint64(v[i].Addr) < sdk.BigEndianToUint64(v[j].Addr))
}

func (v ValidatorSet) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}
