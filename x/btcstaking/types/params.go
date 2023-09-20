package types

import (
	"cosmossdk.io/math"
	"fmt"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

var _ paramtypes.ParamSet = (*Params)(nil)

func defaultJuryPk() *bbn.BIP340PubKey {
	// 32 bytes
	skBytes := []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	_, defaultPK := btcec.PrivKeyFromBytes(skBytes)
	return bbn.NewBIP340PubKeyFromBTCPK(defaultPK)
}

func defaultSlashingAddress() string {
	// 20 bytes
	pkHash := []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	addr, err := btcutil.NewAddressPubKeyHash(pkHash, &chaincfg.SimNetParams)
	if err != nil {
		panic(err)
	}
	return addr.EncodeAddress()
}

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return Params{
		JuryPk:              defaultJuryPk(),
		SlashingAddress:     defaultSlashingAddress(),
		MinSlashingTxFeeSat: 1000,
		MinCommissionRate:   math.LegacyZeroDec(),
	}
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{}
}

func validateMinSlashingTxFeeSat(fee int64) error {
	if fee <= 0 {
		return fmt.Errorf("minimum slashing tx fee has to be positive")
	}
	return nil
}

func validateMinCommissionRate(rate sdk.Dec) error {
	if rate.IsNil() {
		return fmt.Errorf("minimum commission rate cannot be nil")
	}

	if rate.IsNegative() {
		return fmt.Errorf("minimum commission rate cannot be negative")
	}

	if rate.GT(math.LegacyOneDec()) {
		return fmt.Errorf("minimum commission rate cannot be greater than 100%%")
	}
	return nil
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateMinSlashingTxFeeSat(p.MinSlashingTxFeeSat); err != nil {
		return err
	}
	if err := validateMinCommissionRate(p.MinCommissionRate); err != nil {
		return err
	}
	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}
