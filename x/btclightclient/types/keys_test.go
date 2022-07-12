package types_test

import (
	"bytes"
	"github.com/babylonchain/babylon/testutil/datagen"
	bbl "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"math/rand"
	"testing"
)

func FuzzHeadersObjectKey(f *testing.F) {
	f.Add(uint64(42), "00000000000000000002bf1c218853bc920f41f74491e6c92c6bc6fdc881ab47", int64(17))

	f.Fuzz(func(t *testing.T, height uint64, hexHash string, seed int64) {
		rand.Seed(seed)
		if !datagen.ValidHex(hexHash, bbl.BTCHeaderHashLen) {
			hexHash = datagen.GenRandomHexStr(bbl.BTCHeaderHashLen)
		}
		// get chainhash and height
		heightBytes := sdk.Uint64ToBigEndian(height)
		headerHash, _ := bbl.NewBTCHeaderHashBytesFromHex(hexHash)

		// construct the expected key
		headerHashBytes := headerHash.MustMarshal()
		var expectedKey []byte
		expectedKey = append(expectedKey, heightBytes...)
		expectedKey = append(expectedKey, headerHashBytes...)

		gotKey := types.HeadersObjectKey(height, &headerHash)
		if bytes.Compare(expectedKey, gotKey) != 0 {
			t.Errorf("Expected headers object key %s got %s", expectedKey, gotKey)
		}
	})
}

func FuzzHeadersObjectHeightAndWorkKey(f *testing.F) {
	f.Add("00000000000000000002bf1c218853bc920f41f74491e6c92c6bc6fdc881ab47", int64(17))

	f.Fuzz(func(t *testing.T, hexHash string, seed int64) {
		rand.Seed(seed)
		if !datagen.ValidHex(hexHash, bbl.BTCHeaderHashLen) {
			hexHash = datagen.GenRandomHexStr(bbl.BTCHeaderHashLen)
		}
		headerHash, _ := bbl.NewBTCHeaderHashBytesFromHex(hexHash)
		headerHashBytes := headerHash.MustMarshal()

		var expectedHeightKey []byte
		expectedHeightKey = append(expectedHeightKey, headerHashBytes...)
		gotHeightKey := types.HeadersObjectHeightKey(&headerHash)
		if bytes.Compare(expectedHeightKey, gotHeightKey) != 0 {
			t.Errorf("Expected headers object height key %s got %s", expectedHeightKey, gotHeightKey)
		}

		var expectedWorkKey []byte
		expectedWorkKey = append(expectedWorkKey, headerHashBytes...)
		gotWorkKey := types.HeadersObjectWorkKey(&headerHash)
		if bytes.Compare(expectedWorkKey, gotWorkKey) != 0 {
			t.Errorf("Expected headers object work key %s got %s", expectedWorkKey, gotWorkKey)
		}
	})
}

func TestTipKey(t *testing.T) {
	if bytes.Compare(types.TipKey(), types.TipPrefix) != 0 {
		t.Errorf("Expected tip key %s got %s", types.TipKey(), types.TipPrefix)
	}
}
