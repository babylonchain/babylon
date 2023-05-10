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
		signer.Unmarshal(addressBytes) //nolint:errcheck // this is a test

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
			bitsBig = sdk.NewUint(0)
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

		// Check the message creation
		msgInsertHeader, err := types.NewMsgInsertHeader(signer, newHeaderHex)
		if err != nil {
			t.Errorf("Valid parameters led to error")
		}
		if msgInsertHeader == nil {
			t.Fatalf("nil returned")
		}
		if msgInsertHeader.Header == nil {
			t.Errorf("nil header")
		}
		if !bytes.Equal(newHeader.MustMarshal(), msgInsertHeader.Header.MustMarshal()) {
			t.Errorf("Expected header bytes %s got %s", newHeader.MustMarshal(), msgInsertHeader.Header.MustMarshal())
		}

		// Validate the message
		err = msgInsertHeader.ValidateHeader(&maxDifficulty)
		if err != nil && errorKind == 0 {
			t.Errorf("Valid message %s failed with %s", headerHex, err)
		}
		if err == nil && errorKind != 0 {
			t.Errorf("Invalid message did not fail")
		}
	})
}
