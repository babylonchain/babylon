package testepoching

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/merkle"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// NewValidator is a testing helper method to create validators in tests
func NewValidator(t testing.TB, operator sdk.ValAddress, pubKey cryptotypes.PubKey) stakingtypes.Validator {
	v, err := stakingtypes.NewValidator(operator, pubKey, stakingtypes.Description{})
	require.NoError(t, err)
	return v
}

// calculate validator hash and new header
// (adapted from https://github.com/cosmos/cosmos-sdk/blob/v0.45.5/simapp/test_helpers.go#L156-L163)
func CalculateValHash(valSet []stakingtypes.Validator) []byte {
	bzs := make([][]byte, len(valSet))
	for i, val := range valSet {
		consAddr, _ := val.GetConsAddr()
		bzs[i] = consAddr
	}
	return merkle.HashFromByteSlices(bzs)
}
