package helper

import (
	"github.com/cometbft/cometbft/crypto/merkle"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// CalculateValHash calculate validator hash and new header
// (adapted from https://github.com/cosmos/cosmos-sdk/blob/v0.45.5/simapp/test_helpers.go#L156-L163)
func CalculateValHash(valSet []stakingtypes.Validator) []byte {
	bzs := make([][]byte, len(valSet))
	for i, val := range valSet {
		consAddr, _ := val.GetConsAddr()
		bzs[i] = consAddr
	}
	return merkle.HashFromByteSlices(bzs)
}
