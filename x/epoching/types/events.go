package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewEventSlashThreshold(slashedVotingPower int64, totalVotingPower int64, slashedVals []sdk.ValAddress) EventSlashThreshold {
	slashedValBytes := [][]byte{}
	for _, slashedVal := range slashedVals {
		slashedValBytes = append(slashedValBytes, slashedVal)
	}
	return EventSlashThreshold{
		SlashedVotingPower: slashedVotingPower,
		TotalVotingPower:   totalVotingPower,
		SlashedValidators:  slashedValBytes,
	}
}
