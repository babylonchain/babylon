package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Validator struct {
	Addr  sdk.ValAddress
	Power int64
}

type Validators []Validator

func (v Validators) Len() int {
	return len(v)
}

func (v Validators) Less(i, j int) bool {
	return v[i].Power < v[j].Power || (v[i].Power == v[j].Power && sdk.BigEndianToUint64(v[i].Addr) < sdk.BigEndianToUint64(v[j].Addr))
}

func (v Validators) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}
