package types

import (
	fmt "fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewGauge(coins ...sdk.Coin) *Gauge {
	return &Gauge{
		Coins: coins,
	}
}

func (g *Gauge) GetCoinsPortion(portion math.LegacyDec) sdk.Coins {
	return GetCoinsPortion(g.Coins, portion)
}

func NewRewardGauge(coins ...sdk.Coin) *RewardGauge {
	return &RewardGauge{
		Coins:          coins,
		WithdrawnCoins: sdk.NewCoins(),
	}
}

// GetWithdrawableCoins returns withdrawable coins in this reward gauge
func (rg *RewardGauge) GetWithdrawableCoins() sdk.Coins {
	return rg.Coins.Sub(rg.WithdrawnCoins...)
}

// SetFullyWithdrawn makes the reward gauge to have no withdrawable coins
// typically called after the stakeholder withdraws its reward
func (rg *RewardGauge) SetFullyWithdrawn() {
	rg.WithdrawnCoins = sdk.NewCoins(rg.Coins...)
}

// IsFullyWithdrawn returns whether the reward gauge has nothing to withdraw
func (rg *RewardGauge) IsFullyWithdrawn() bool {
	return rg.Coins.IsEqual(rg.WithdrawnCoins)
}

func (rg *RewardGauge) Add(coins sdk.Coins) {
	rg.Coins = rg.Coins.Add(coins...)
}

func GetCoinsPortion(coinsInt sdk.Coins, portion math.LegacyDec) sdk.Coins {
	// coins with decimal value
	coins := sdk.NewDecCoinsFromCoins(coinsInt...)
	// portion of coins with decimal values
	portionCoins := coins.MulDecTruncate(portion)
	// truncate back
	// TODO: how to deal with changes?
	portionCoinsInt, _ := portionCoins.TruncateDecimal()
	return portionCoinsInt
}

// enum for stakeholder type, used as key prefix in KVStore
type StakeholderType byte

const (
	SubmitterType StakeholderType = iota
	ReporterType
	BTCValidatorType
	BTCDelegationType
)

func GetAllStakeholderTypes() []StakeholderType {
	return []StakeholderType{SubmitterType, ReporterType, BTCValidatorType, BTCDelegationType}
}

func NewStakeHolderType(stBytes []byte) (StakeholderType, error) {
	if len(stBytes) != 1 {
		return SubmitterType, fmt.Errorf("invalid format for stBytes")
	}
	if stBytes[0] == byte(SubmitterType) {
		return SubmitterType, nil
	} else if stBytes[0] == byte(ReporterType) {
		return ReporterType, nil
	} else if stBytes[0] == byte(BTCValidatorType) {
		return BTCValidatorType, nil
	} else if stBytes[0] == byte(BTCDelegationType) {
		return BTCDelegationType, nil
	} else {
		return SubmitterType, fmt.Errorf("invalid stBytes")
	}
}

func NewStakeHolderTypeFromString(stStr string) (StakeholderType, error) {
	if stStr == "submitter" {
		return SubmitterType, nil
	} else if stStr == "reporter" {
		return ReporterType, nil
	} else if stStr == "btc_validator" {
		return BTCValidatorType, nil
	} else if stStr == "btc_delegation" {
		return BTCDelegationType, nil
	} else {
		return SubmitterType, fmt.Errorf("invalid stStr")
	}
}

func (st StakeholderType) Bytes() []byte {
	return []byte{byte(st)}
}

func (st StakeholderType) String() string {
	if st == SubmitterType {
		return "submitter"
	} else if st == ReporterType {
		return "reporter"
	} else if st == BTCValidatorType {
		return "btc_validator"
	} else if st == BTCDelegationType {
		return "btc_delegation"
	}
	panic("invalid stakeholder type")
}
