package types_test

import (
	"bytes"
	bbl "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"math/rand"
	"testing"
)

func FuzzHeadersObjectKey(f *testing.F) {
	f.Add(uint64(42), "00000000000000000002bf1c218853bc920f41f74491e6c92c6bc6fdc881ab47", int64(17))

	f.Fuzz(func(t *testing.T, height uint64, hexHash string, seed int64) {
		rand.Seed(seed)
		if !validHex(hexHash, bbl.BTCHeaderHashLen) {
			hexHash = genRandomHexStr(bbl.BTCHeaderHashLen)
		}
		// get chainhash and height
		chHash, _ := chainhash.NewHashFromStr(hexHash)
		heightBytes := sdk.Uint64ToBigEndian(height)

		// construct the expected key
		chHashBytes := chHash[:]
		expectedKey := append(types.HeadersObjectPrefix, heightBytes...)
		expectedKey = append(expectedKey, chHashBytes...)

		gotKey := types.HeadersObjectKey(height, chHash)
		if bytes.Compare(expectedKey, gotKey) != 0 {
			t.Errorf("Expected headers object key %s got %s", expectedKey, gotKey)
		}
	})
}

func FuzzHeadersObjectHeightKey(f *testing.F) {
	f.Add("00000000000000000002bf1c218853bc920f41f74491e6c92c6bc6fdc881ab47", int64(17))

	f.Fuzz(func(t *testing.T, hexHash string, seed int64) {
		rand.Seed(seed)
		if !validHex(hexHash, bbl.BTCHeaderHashLen) {
			hexHash = genRandomHexStr(bbl.BTCHeaderHashLen)
		}
		// Get the chainhash
		chHash, _ := chainhash.NewHashFromStr(hexHash)

		// Construct the expected key
		chHashBytes := chHash[:]

		expectedHeightKey := append(types.HashToHeightPrefix, chHashBytes...)
		gotHeightKey, err := types.HeadersObjectHeightKey(chHash)
		if err != nil {
			t.Errorf("Chainhash %s cannot be converted to a height key", chHash.String())
		}
		if bytes.Compare(expectedHeightKey, gotHeightKey) != 0 {
			t.Errorf("Expected headers object height key %s got %s", expectedHeightKey, gotHeightKey)
		}
	})
}

func TestTipKey(t *testing.T) {
	if bytes.Compare(types.TipKey(), types.TipPrefix) != 0 {
		t.Errorf("Expected tip key %s got %s", types.TipKey(), types.TipPrefix)
	}
}
