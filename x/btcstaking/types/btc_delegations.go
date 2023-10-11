package types

import (
	"bytes"
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

func NewBTCDelegatorDelegationIndex() *BTCDelegatorDelegationIndex {
	return &BTCDelegatorDelegationIndex{
		StakingTxHashList: [][]byte{},
	}
}

func (i *BTCDelegatorDelegationIndex) Has(stakingTxHash chainhash.Hash) bool {
	for _, hash := range i.StakingTxHashList {
		if bytes.Equal(stakingTxHash[:], hash) {
			return true
		}
	}
	return false
}

func (i *BTCDelegatorDelegationIndex) Add(stakingTxHash chainhash.Hash) error {
	// ensure staking tx hash is not duplicated
	for _, hash := range i.StakingTxHashList {
		if bytes.Equal(stakingTxHash[:], hash) {
			return fmt.Errorf("the given stakingTxHash %s is duplicated", stakingTxHash.String())
		}
	}
	// add
	i.StakingTxHashList = append(i.StakingTxHashList, stakingTxHash[:])

	return nil
}

// VotingPower calculates the total voting power of all BTC delegations
func (dels *BTCDelegatorDelegations) VotingPower(btcHeight uint64, w uint64) uint64 {
	power := uint64(0)
	for _, del := range dels.Dels {
		power += del.VotingPower(btcHeight, w)
	}
	return power
}
