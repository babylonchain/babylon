package types_test

import (
	"bytes"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"testing"
)

func FuzzHeadersObjectKey(f *testing.F) {
	f.Add(uint64(42), "00000000000000000002bf1c218853bc920f41f74491e6c92c6bc6fdc881ab47")

	f.Fuzz(func(t *testing.T, height uint64, hexHash string) {
		chHash, err := chainhash.NewHashFromStr(hexHash)
		if err != nil {
			// the hexHash is an invalid one
			t.Skip()
		}

		heightBytes := sdk.Uint64ToBigEndian(height)
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
	f.Add("00000000000000000002bf1c218853bc920f41f74491e6c92c6bc6fdc881ab47")

	f.Fuzz(func(t *testing.T, hexHash string) {
		chHash, err := chainhash.NewHashFromStr(hexHash)
		if err != nil {
			// the hexHash is not a valid one
			t.Skip()
		}

		chHashBytes := chHash[:]
		expectedKey := append(types.HashToHeightPrefix, chHashBytes...)

		gotKey := types.HeadersObjectHeightKey(chHash)

		if bytes.Compare(expectedKey, gotKey) != 0 {
			t.Errorf("Expected headers object height key %s got %s", expectedKey, gotKey)
		}
	})
}

func TestTipKey(t *testing.T) {
	if bytes.Compare(types.TipKey(), types.TipPrefix) != 0 {
		t.Errorf("Expected tip key %s got %s", types.TipKey(), types.TipPrefix)
	}
}
