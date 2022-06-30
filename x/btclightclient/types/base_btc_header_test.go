package types_test

import (
	"bytes"
	bbl "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	"testing"
)

func FuzzNewBaseBTCHeader(f *testing.F) {
	defaultHeader, _ := bbl.NewBTCHeaderBytesFromHex(types.DefaultBaseHeaderHex)
	defaultHeaderBytes, _ := defaultHeader.Marshal()
	f.Add(defaultHeaderBytes, uint64(42))

	f.Fuzz(func(t *testing.T, headerBytes []byte, height uint64) {
		headerBytesObj, err := bbl.NewBTCHeaderBytesFromBytes(headerBytes)
		if err != nil {
			// Invalid header bytes
			t.Skip()
		}

		gotObject := types.NewBaseBTCHeader(headerBytesObj, height)

		gotHeaderBytes, err := gotObject.Header.Marshal()
		if err != nil {
			t.Errorf("Header returned cannot be marshalled")
		}
		gotHeight := gotObject.Height

		if bytes.Compare(headerBytes, gotHeaderBytes) != 0 {
			t.Errorf("Header attribute is different")
		}
		if height != gotHeight {
			t.Errorf("Height attribute is different")
		}
	})
}

func FuzzBaseBTCHeader_Validate(f *testing.F) {
	defaultHeader, _ := bbl.NewBTCHeaderBytesFromHex(types.DefaultBaseHeaderHex)
	defaultHeaderBytes, _ := defaultHeader.Marshal()
	f.Add(defaultHeaderBytes, uint64(42))

	f.Fuzz(func(t *testing.T, headerBytes []byte, height uint64) {
		headerBytesObj, err := bbl.NewBTCHeaderBytesFromBytes(headerBytes)
		if err != nil {
			// Invalid header bytes
			t.Skip()
		}

		baseBTCHeader := types.NewBaseBTCHeader(headerBytesObj, height)

		err = baseBTCHeader.Validate()
		if err != nil {
			t.Errorf("Got error %s when validating", err)
		}
	})
}
