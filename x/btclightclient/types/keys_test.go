package types_test

import (
	"bytes"
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func FuzzHeadersObjectKey(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 100)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		hexHash := datagen.GenRandomHexStr(r, bbn.BTCHeaderHashLen)
		height := r.Uint64()
		// get chainhash and height
		heightBytes := sdk.Uint64ToBigEndian(height)
		headerHash, _ := bbn.NewBTCHeaderHashBytesFromHex(hexHash)

		// construct the expected key
		headerHashBytes := headerHash.MustMarshal()
		var expectedKey []byte
		expectedKey = append(expectedKey, heightBytes...)
		expectedKey = append(expectedKey, headerHashBytes...)

		gotKey := types.HeadersObjectKey(height, &headerHash)
		if !bytes.Equal(expectedKey, gotKey) {
			t.Errorf("Expected headers object key %s got %s", expectedKey, gotKey)
		}
	})
}

func FuzzHeadersObjectHeightAndWorkKey(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 100)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		hexHash := datagen.GenRandomHexStr(r, bbn.BTCHeaderHashLen)
		headerHash, _ := bbn.NewBTCHeaderHashBytesFromHex(hexHash)
		headerHashBytes := headerHash.MustMarshal()

		var expectedHeightKey []byte
		expectedHeightKey = append(expectedHeightKey, headerHashBytes...)
		gotHeightKey := types.HeadersObjectHeightKey(&headerHash)
		if !bytes.Equal(expectedHeightKey, gotHeightKey) {
			t.Errorf("Expected headers object height key %s got %s", expectedHeightKey, gotHeightKey)
		}

		var expectedWorkKey []byte
		expectedWorkKey = append(expectedWorkKey, headerHashBytes...)
		gotWorkKey := types.HeadersObjectWorkKey(&headerHash)
		if !bytes.Equal(expectedWorkKey, gotWorkKey) {
			t.Errorf("Expected headers object work key %s got %s", expectedWorkKey, gotWorkKey)
		}
	})
}

func TestTipKey(t *testing.T) {
	if !bytes.Equal(types.TipKey(), types.TipPrefix) {
		t.Errorf("Expected tip key %s got %s", types.TipKey(), types.TipPrefix)
	}
}
