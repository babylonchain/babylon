package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewParams creates a new Params instance
func NewParams(allowedAddresses []string) Params {
	return Params{
		InsertHeadersAllowList: allowedAddresses,
	}
}

func NewParamsValidate(allowedAddresses []string) (Params, error) {
	p := NewParams(allowedAddresses)
	if err := p.Validate(); err != nil {
		return Params{}, err
	}
	return p, nil
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(
		[]string{},
	)
}

func ValidateAddressList(i interface{}) error {
	allowList, ok := i.([]string)

	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	for _, a := range allowList {
		if _, err := sdk.AccAddressFromBech32(a); err != nil {
			return fmt.Errorf("invalid address")
		}
	}

	return nil
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := ValidateAddressList(p.InsertHeadersAllowList); err != nil {
		return err
	}

	return nil
}

func (p *Params) AllowAllReporters() bool {
	return len(p.InsertHeadersAllowList) == 0
}
