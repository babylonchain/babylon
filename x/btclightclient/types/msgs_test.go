package types_test

import (
	"bytes"
	"math/rand"
	"testing"

	sdkmath "cosmossdk.io/math"

	"github.com/babylonchain/babylon/testutil/datagen"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func FuzzMsgInsertHeader(f *testing.F) {
	maxDifficulty := bbn.GetMaxDifficulty()

	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		errorKind := 0

		addressBytes := datagen.GenRandomByteArray(r, 1+uint64(r.Intn(255)))
		headerBytes := datagen.GenRandomBTCHeaderInfo(r).Header
		headerHex := headerBytes.MarshalHex()

		// Get the signer structure
		var signer sdk.AccAddress
		err := signer.Unmarshal(addressBytes)
		require.NoError(t, err)

		// Perform modifications on the header
		errorKind = r.Intn(2)
		var bitsBig sdkmath.Uint
		switch errorKind {
		case 0:
			// Valid input
			// Set the work bits to the pow limit
			bitsBig = sdkmath.NewUintFromBigInt(&maxDifficulty)
		case 1:
			// Zero PoW
			bitsBig = sdkmath.NewUint(0)
		default:
			bitsBig = sdkmath.NewUintFromBigInt(&maxDifficulty)
		}

		// Generate a header with the provided modifications
		newHeader := datagen.GenRandomBTCHeaderInfoWithBits(r, &bitsBig).Header
		newHeaderHex := newHeader.MarshalHex()

		// Check whether the hash is still bigger than the maximum allowed
		// This happens because even though we pass a series of "f"s as an input
		// the maximum that the bits field can contain is 2^23-1, meaning
		// that there is still space for block hashes that are less than that
		headerDifficulty := types.CalcWork(newHeader)
		if headerDifficulty.GT(sdkmath.NewUintFromBigInt(&maxDifficulty)) {
			t.Skip()
		}

		numHeaders := r.Intn(20) + 1
		headersHex := ""
		for i := 0; i < numHeaders; i++ {
			headersHex += newHeaderHex
		}

		// empty string
		_, err = types.NewMsgInsertHeaders(signer, "")
		require.NotNil(t, err)

		// hex string with invalid length
		invalidLength := uint64(r.Int31n(79) + 1)
		_, err = types.NewMsgInsertHeaders(signer, headersHex+datagen.GenRandomHexStr(r, invalidLength))
		require.NotNil(t, err)

		// Check the message creation
		msgInsertHeader, err := types.NewMsgInsertHeaders(signer, headersHex)
		if err != nil {
			t.Errorf("Valid parameters led to error")
		}
		if msgInsertHeader == nil {
			t.Fatalf("nil returned")
		}
		if msgInsertHeader.Headers == nil || len(msgInsertHeader.Headers) != numHeaders {
			t.Errorf("invalid number of headers")
		}

		for _, header := range msgInsertHeader.Headers {
			if !bytes.Equal(newHeader.MustMarshal(), header.MustMarshal()) {
				t.Errorf("Expected header bytes %s got %s", newHeader.MustMarshal(), header.MustMarshal())
			}
		}

		// Validate the message
		err = msgInsertHeader.ValidateHeaders(&maxDifficulty)
		if err != nil && errorKind == 0 {
			t.Errorf("Valid message %s failed with %s", headerHex, err)
		}
		if err == nil && errorKind != 0 {
			t.Errorf("Invalid message did not fail")
		}
	})
}
