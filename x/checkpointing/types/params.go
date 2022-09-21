package types

import (
	"fmt"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
	"time"
)

const (
	defaultRetrySleepTime    string = "1s"
	defaultMaxRetrySleepTime string = "1m"
)

var (
	KeyRetrySleepTime    = []byte("RetrySleepTime")
	KeyMaxRetrySleepTime = []byte("MaxRetrySleepTime")
)

var _ paramtypes.ParamSet = (*Params)(nil)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(retrySleepTime string, maxRetrySleepTime string) Params {
	return Params{
		RetrySleepTime:    retrySleepTime,
		MaxRetrySleepTime: maxRetrySleepTime,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(defaultRetrySleepTime, defaultMaxRetrySleepTime)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyRetrySleepTime, &p.RetrySleepTime, validateRetrySleepTime),
		paramtypes.NewParamSetPair(KeyMaxRetrySleepTime, &p.MaxRetrySleepTime, validateRetrySleepTime),
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

func validateRetrySleepTime(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	_, err := time.ParseDuration(v)
	if err != nil {
		return fmt.Errorf("can not parse retry sleep time %v", v)
	}

	return nil
}
