package types

import (
	"github.com/babylonchain/babylon/crypto/bls12381"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
	"io/ioutil"
)

// DefaultIndex is the default capability global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {

	return gs.Params.Validate()
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
	genBlsJSONBytes, err := ioutil.ReadFile(filePath)
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
