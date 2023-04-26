package types_test

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/babylonchain/babylon/types"
)

func FuzzBTCHeaderBytesBytesOps(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 100)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		invalidHeader := false
		bz := datagen.GenRandomByteArray(r, types.BTCHeaderLen)
		if datagen.OneInN(r, 10) {
			bz = datagen.GenRandomByteArray(r, datagen.RandomIntOtherThan(r, types.BTCHeaderLen, 10*types.BTCHeaderLen))
			invalidHeader = true
		}

		hb, err := types.NewBTCHeaderBytesFromBytes(bz)
		if err != nil {
			if !invalidHeader {
				t.Fatalf("Valid header %s led to a NewBTCHeaderBytesFromBytes error %s", bz, err)
			}
			t.Skip()
		}

		err = hb.Unmarshal(bz)
		if err != nil {
			if !invalidHeader {
				t.Fatalf("Valid header %s led to an unmarshal error %s", bz, err)
			}
			t.Skip()
		}

		m, err := hb.Marshal()
		if err != nil {
			if !invalidHeader {
				t.Fatalf("Marshaling of bytes slice led to error %s", err)
			}
			t.Skip()
		}
		if !bytes.Equal(m, bz) {
			t.Errorf("Marshal returned %s while %s was expected", m, bz)
		}

		m = make([]byte, len(bz))
		sz, err := hb.MarshalTo(m)
		if err != nil {
			if !invalidHeader {
				t.Fatalf("MarshalTo led to error %s", err)
			}
			t.Skip()
		}
		if sz != len(bz) {
			t.Errorf("MarhslTo marshalled %d bytes instead of %d", sz, len(bz))
		}
		if !bytes.Equal(m, bz) {
			t.Errorf("MarshalTo copied %s while %s was expected", m, bz)
		}

		if invalidHeader {
			t.Errorf("Invalid header succeeded in all operations")
		}

		sz = hb.Size()
		if sz != len(bz) {
			t.Errorf("Size returned %d while %d was expected", sz, len(bz))
		}
	})
}

func FuzzBTCHeaderBytesHexOps(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 100)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		invalidHeader := false
		// 2 hex chars per byte
		hex := datagen.GenRandomHexStr(r, types.BTCHeaderLen)
		if datagen.OneInN(r, 10) {
			if datagen.OneInN(r, 2) {
				hex = datagen.GenRandomHexStr(r, datagen.RandomIntOtherThan(r, types.BTCHeaderLen, 10*types.BTCHeaderLen))
			} else {
				hex = string(datagen.GenRandomByteArray(r, types.BTCHeaderLen*2))
			}
			invalidHeader = true
		}
		hb, err := types.NewBTCHeaderBytesFromHex(hex)
		if err != nil {
			if !invalidHeader {
				t.Fatalf("Valid header %s %d led to a NewBTCHeaderBytesFromHex error %s", hex, len(hex), err)
			}
			t.Skip()
		}

		err = hb.UnmarshalHex(hex)
		if err != nil {
			if !invalidHeader {
				t.Fatalf("Valid header %s led to an unmarshal error %s", hex, err)
			}
			t.Skip()
		}

		h := hb.MarshalHex()
		if h != hex {
			t.Errorf("Marshal returned %s while %s was expected", h, hex)
		}
		if invalidHeader {
			t.Errorf("Invalid header passed all checks")
		}
	})
}

func FuzzBTCHeaderBytesJSONOps(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 100)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		invalidHeader := false
		// 2 hex chars per byte
		hex := datagen.GenRandomHexStr(r, types.BTCHeaderLen)
		if datagen.OneInN(r, 10) {
			if datagen.OneInN(r, 2) {
				hex = datagen.GenRandomHexStr(r, datagen.RandomIntOtherThan(r, types.BTCHeaderLen, 10*types.BTCHeaderLen))
			} else {
				hex = string(datagen.GenRandomByteArray(r, types.BTCHeaderLen*2))
			}
			invalidHeader = true
		}
		hb, err := types.NewBTCHeaderBytesFromHex(hex)
		if err != nil {
			if !invalidHeader {
				t.Fatalf("Valid header %s %d led to a NewBTCHeaderBytesFromHex error %s", hex, len(hex), err)
			}
			t.Skip()
		}

		jsonHex, _ := json.Marshal(hex)

		err = hb.UnmarshalJSON(jsonHex)
		if err != nil {
			if !invalidHeader {
				t.Fatalf("Valid header %s led to an unmarshal error %s", hex, err)
			}
			t.Skip()
		}

		h, err := hb.MarshalJSON()
		if err != nil {
			if !invalidHeader {
				t.Fatalf("Valid header %s led to an marshal error %s", hex, err)
			}
			t.Skip()
		}
		if !bytes.Equal(h, jsonHex) {
			t.Errorf("Marshal returned %s while %s was expected", h, jsonHex)
		}
		if invalidHeader {
			t.Errorf("Invalid header passed all checks")
		}
	})
}

func FuzzBTCHeaderBytesBtcdBlockOps(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 100)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		btcdHeader := datagen.GenRandomBtcdHeader(r)

		var hb types.BTCHeaderBytes
		hb.FromBlockHeader(btcdHeader)
		hbBlockHeader := hb.ToBlockHeader()

		if btcdHeader.BlockHash() != hbBlockHeader.BlockHash() {
			t.Errorf("Expected block hash %s got block hash %s", btcdHeader.BlockHash(), hbBlockHeader.BlockHash())
		}

		hb = types.NewBTCHeaderBytesFromBlockHeader(btcdHeader)
		if btcdHeader.BlockHash() != hbBlockHeader.BlockHash() {
			t.Errorf("Expected block hash %s got block hash %s", btcdHeader.BlockHash(), hbBlockHeader.BlockHash())
		}
	})
}

func FuzzBTCHeaderBytesOperators(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 100)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		parent := datagen.GenRandomBTCHeaderInfo(r)
		hb := parent.Header
		hb2 := types.NewBTCHeaderBytesFromBlockHeader(hb.ToBlockHeader())
		hbChild := datagen.GenRandomBTCHeaderBytes(r, parent, nil)

		if !hb.Eq(hb) {
			t.Errorf("BTCHeaderBytes object does not equal itself")
		}
		if !hb.Eq(&hb2) {
			t.Errorf("BTCHeaderBytes object does not equal a different object with the same bytes")
		}
		if hb.Eq(&hbChild) {
			t.Errorf("BTCHeaderBytes object equals a different object with different bytes")
		}

		if !hbChild.HasParent(hb) {
			t.Errorf("HasParent method returns false with a correct parent")
		}
		if hbChild.HasParent(&hbChild) {
			t.Errorf("HasParent method returns true for the same object")
		}

		if !hbChild.ParentHash().Eq(hb.Hash()) {
			t.Errorf("ParentHash did not return the parent hash")
		}
	})
}
