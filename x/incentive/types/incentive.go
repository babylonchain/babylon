package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewRewardGauge(coins ...sdk.Coin) *RewardGauge {
	return &RewardGauge{
		Coins:          sdk.NewCoins(coins...),
		WithdrawnCoins: sdk.NewCoins(coins...),
	}
}

// enum for stakeholder type, used as key prefix in KVStore
type StakeholderType byte

const (
	SubmitterType StakeholderType = iota
	ReporterType
	BTCValidatorType
	BTCDelegationType
)

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

func (st *StakeholderType) Bytes() []byte {
	return []byte{byte(*st)}
}

func (st *StakeholderType) String() string {
	if *st == SubmitterType {
		return "submitter"
	} else if *st == ReporterType {
		return "reporter"
	} else if *st == BTCValidatorType {
		return "btc_validator"
	} else if *st == BTCDelegationType {
		return "btc_delegation"
	}
	panic("invalid stakeholder type")
}
