package types

import "fmt"

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	p := DefaultParams()
	return &GenesisState{
		Params: []*Params{&p},
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if len(gs.Params) == 0 {
		return fmt.Errorf("params cannot be empty")
	}

	// TODO: add validation to other properties of genstate.
	for _, params := range gs.Params {
		if err := params.Validate(); err != nil {
			return err
		}
	}
	return nil
}
