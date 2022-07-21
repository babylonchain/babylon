package types

func NewEventSlashThreshold(slashedVotingPower int64, totalVotingPower int64, slashedValSet ValidatorSet) EventSlashThreshold {
	slashedValBytes := [][]byte{}
	for _, slashedVal := range slashedValSet {
		slashedValBytes = append(slashedValBytes, slashedVal.Addr)
	}
	return EventSlashThreshold{
		SlashedVotingPower: slashedVotingPower,
		TotalVotingPower:   totalVotingPower,
		SlashedValidators:  slashedValBytes,
	}
}
