package types

import (
	"fmt"

	"cosmossdk.io/math"
	"github.com/babylonchain/babylon/btcstaking"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

const (
	defaultMaxActiveBtcValidators uint32 = 100
)

var _ paramtypes.ParamSet = (*Params)(nil)

func defaultCovenantPk() *bbn.BIP340PubKey {
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
		CovenantPk:          defaultCovenantPk(),
		SlashingAddress:     defaultSlashingAddress(),
		MinSlashingTxFeeSat: 1000,
		MinCommissionRate:   math.LegacyZeroDec(),
		// The Default slashing rate is 0.1 i.e., 10% of the total staked BTC will be burned.
		SlashingRate:           math.LegacyNewDecWithPrec(1, 1), // 1 * 10^{-1} = 0.1
		MaxActiveBtcValidators: defaultMaxActiveBtcValidators,
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

// validateMaxActiveBTCValidators checks if the maximum number of
// active BTC validators is at least the default value
func validateMaxActiveBTCValidators(maxActiveBtcValidators uint32) error {
	if maxActiveBtcValidators == 0 {
		return fmt.Errorf("max validators must be positive")
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

	if !btcstaking.IsSlashingRateValid(p.SlashingRate) {
		return btcstaking.ErrInvalidSlashingRate
	}

	if err := validateMaxActiveBTCValidators(p.MaxActiveBtcValidators); err != nil {
		return err
	}
	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}
