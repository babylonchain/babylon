package types

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	// TODO: add validate to IndexedBlocks, Evidences, VoteSigs, PublicRandomness
	return gs.Params.Validate()
}
