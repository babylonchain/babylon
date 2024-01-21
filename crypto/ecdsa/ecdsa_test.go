package ecdsa_test

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"testing"

	"github.com/babylonchain/babylon/crypto/ecdsa"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/stretchr/testify/require"
)

const (
	// test vector from https://github.com/okx/js-wallet-sdk/blob/a57c2acbe6ce917c0aa4e951d96c4e562ad58444/packages/coin-bitcoin/tests/btc.test.ts#L113-L126
	skHex         = "adce25dc25ef89f06a722abdc4b601d706c9efc6bc84075355e6b96ca3871621"
	testMsg       = "hello world"
	testSigBase64 = "IDtG3XPLpiKOp4PjTzCo/ng8gm4MFTTyHeh/DaPC1XYsYaj5Jr4h8dnxmwuJtNkPkH40rEfnrrO8fgZKNOIF5iM="
)

func TestECDSA(t *testing.T) {
	// decode SK and PK
	skBytes, err := hex.DecodeString(skHex)
	require.NoError(t, err)
	sk, pk := btcec.PrivKeyFromBytes(skBytes)
	require.NotNil(t, sk)
	require.NotNil(t, pk)
	// sign
	sig, err := ecdsa.Sign(sk, testMsg)
	require.NoError(t, err)
	testSigBytes, err := base64.StdEncoding.DecodeString(testSigBase64)
	require.NoError(t, err)
	// ensure sig is same as that in test vector
	require.True(t, bytes.Equal(sig, testSigBytes))
	// verify
	err = ecdsa.Verify(pk, testMsg, sig)
	require.NoError(t, err)
}
