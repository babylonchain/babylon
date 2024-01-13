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
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		height := r.Uint64()
		// get chainhash and height
		heightBytes := sdk.Uint64ToBigEndian(height)
		// construct the expected key
		var expectedKey []byte
		expectedKey = append(expectedKey, heightBytes...)

		gotKey := types.HeadersObjectKey(height)
		if !bytes.Equal(expectedKey, gotKey) {
			t.Errorf("Expected headers object key %s got %s", expectedKey, gotKey)
		}
	})
}

func FuzzHeadersObjectHeightAndWorkKey(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

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
	})
}
