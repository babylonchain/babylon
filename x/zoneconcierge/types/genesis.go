package types

import (
	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"
)

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		PortId: PortID,
		Params: DefaultParams(),
	}
}

// NewGenesis creates a new GenesisState instance
func NewGenesis(params Params) *GenesisState {
	return &GenesisState{
		PortId: PortID,
		Params: params,
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if err := host.PortIdentifierValidator(gs.PortId); err != nil {
		return err
	}
	if err := gs.Params.Validate(); err != nil {
		return err
	}
	return nil
}
