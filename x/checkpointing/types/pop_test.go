package types_test

import (
	"github.com/babylonchain/babylon/crypto/bls12381"
	"github.com/babylonchain/babylon/privval"
	"github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"testing"
)

func TestProofOfPossession_IsValid(t *testing.T) {
	valPrivKey := ed25519.GenPrivKey()
	blsPrivKey := bls12381.GenPrivKey()
	pop, err := privval.BuildPoP(valPrivKey, blsPrivKey)
	require.NoError(t, err)
	valpk, err := codec.FromTmPubKeyInterface(valPrivKey.PubKey())
	require.NoError(t, err)
	require.True(t, pop.IsValid(blsPrivKey.PubKey(), valpk))
}
