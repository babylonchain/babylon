package types

import (
	"encoding/json"
	"errors"
	"os"

	tmjson "github.com/cometbft/cometbft/libs/json"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/babylonchain/babylon/crypto/bls12381"
)

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	addresses := make(map[string]struct{}, 0)
	for _, gk := range gs.GenesisKeys {
		if _, exists := addresses[gk.ValidatorAddress]; exists {
			return errors.New("duplicate genesis key")
		}
		addresses[gk.ValidatorAddress] = struct{}{}
		err := gk.Validate()
		if err != nil {
			return err
		}
	}

	return nil
}

func NewGenesisKey(delAddr sdk.ValAddress, blsPubKey *bls12381.PublicKey, pop *ProofOfPossession, pubkey cryptotypes.PubKey) (*GenesisKey, error) {
	if !pop.IsValid(*blsPubKey, pubkey) {
		return nil, ErrInvalidPoP
	}
	gk := &GenesisKey{
		ValidatorAddress: delAddr.String(),
		BlsKey: &BlsKey{
			Pubkey: blsPubKey,
			Pop:    pop,
		},
		ValPubkey: pubkey.(*ed25519.PubKey),
	}

	return gk, nil
}

func LoadGenesisKeyFromFile(filePath string) (*GenesisKey, error) {
	genBlsJSONBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	genBls := new(GenesisKey)
	err = tmjson.Unmarshal(genBlsJSONBytes, genBls)
	if err != nil {
		return nil, err
	}
	err = genBls.Validate()
	if err != nil {
		return nil, err
	}
	return genBls, nil
}

func (gk *GenesisKey) Validate() error {
	if !gk.BlsKey.Pop.IsValid(*gk.BlsKey.Pubkey, gk.ValPubkey) {
		return ErrInvalidPoP
	}
	return nil
}

// GetGenesisStateFromAppState returns x/Checkpointing GenesisState given raw application
// genesis state.
func GetGenesisStateFromAppState(cdc codec.Codec, appState map[string]json.RawMessage) GenesisState {
	var genesisState GenesisState

	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}

	return genesisState
}
