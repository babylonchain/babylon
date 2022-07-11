package types_test

import (
	"bytes"
	"encoding/json"
	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/babylonchain/babylon/types"
	"math/rand"
	"testing"
)

func FuzzBTCHeaderBytesBytesOps(f *testing.F) {
	f.Add(int64(42))

	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)

		invalidHeader := false
		bz := datagen.GenRandomByteArray(types.BTCHeaderLen)
		if datagen.OneInN(10) {
			bz = datagen.GenRandomByteArray(datagen.RandomIntOtherThan(types.BTCHeaderLen))
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
		if bytes.Compare(m, bz) != 0 {
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
		if bytes.Compare(m, bz) != 0 {
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
	f.Add(int64(42))

	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)

		invalidHeader := false
		// 2 hex chars per byte
		hex := datagen.GenRandomHexStr(types.BTCHeaderLen)
		if datagen.OneInN(10) {
			if datagen.OneInN(2) {
				hex = datagen.GenRandomHexStr(datagen.RandomIntOtherThan(types.BTCHeaderLen))
			} else {
				hex = string(datagen.GenRandomByteArray(types.BTCHeaderLen * 2))
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
	f.Add(int64(42))

	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)

		invalidHeader := false
		// 2 hex chars per byte
		hex := datagen.GenRandomHexStr(types.BTCHeaderLen)
		if datagen.OneInN(10) {
			if datagen.OneInN(2) {
				hex = datagen.GenRandomHexStr(datagen.RandomIntOtherThan(types.BTCHeaderLen))
			} else {
				hex = string(datagen.GenRandomByteArray(types.BTCHeaderLen * 2))
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
		if bytes.Compare(h, jsonHex) != 0 {
			t.Errorf("Marshal returned %s while %s was expected", h, jsonHex)
		}
		if invalidHeader {
			t.Errorf("Invalid header passed all checks")
		}
	})
}

func FuzzBTCHeaderBytesBtcdBlockOps(f *testing.F) {
	defaultHeader, _ := types.NewBTCHeaderBytesFromHex("00006020c6c5a20e29da938a252c945411eba594cbeba021a1e20000000000000000000039e4bd0cd0b5232bb380a9576fcfe7d8fb043523f7a158187d9473e44c1740e6b4fa7c62ba01091789c24c22")
	defaultBtcdHeader := defaultHeader.ToBlockHeader()

	f.Add(
		defaultBtcdHeader.Version,
		defaultBtcdHeader.Bits,
		defaultBtcdHeader.Nonce,
		defaultBtcdHeader.Timestamp.Unix(),
		defaultBtcdHeader.PrevBlock.String(),
		defaultBtcdHeader.MerkleRoot.String(),
		int64(17))

	f.Fuzz(func(t *testing.T, version int32, bits uint32, nonce uint32,
		timeInt int64, prevBlockStr string, merkleRootStr string, seed int64) {

		rand.Seed(seed)
		btcdHeader := datagen.GenRandomBtcdHeader(version, bits, nonce, timeInt, prevBlockStr, merkleRootStr)

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
	defaultHeader, _ := types.NewBTCHeaderBytesFromHex("00006020c6c5a20e29da938a252c945411eba594cbeba021a1e20000000000000000000039e4bd0cd0b5232bb380a9576fcfe7d8fb043523f7a158187d9473e44c1740e6b4fa7c62ba01091789c24c22")
	defaultBtcdHeader := defaultHeader.ToBlockHeader()

	f.Add(
		defaultBtcdHeader.Version,
		defaultBtcdHeader.Bits,
		defaultBtcdHeader.Nonce,
		defaultBtcdHeader.Timestamp.Unix(),
		defaultBtcdHeader.PrevBlock.String(),
		defaultBtcdHeader.MerkleRoot.String(),
		int64(17))

	f.Fuzz(func(t *testing.T, version int32, bits uint32, nonce uint32,
		timeInt int64, prevBlockStr string, merkleRootStr string, seed int64) {

		rand.Seed(seed)
		btcdHeader := datagen.GenRandomBtcdHeader(version, bits, nonce, timeInt, prevBlockStr, merkleRootStr)

		btcdHeaderHash := btcdHeader.BlockHash()
		childPrevBlock := types.NewBTCHeaderHashBytesFromChainhash(&btcdHeaderHash)
		btcdHeaderChild := datagen.GenRandomBtcdHeader(version, bits, nonce, timeInt, childPrevBlock.MarshalHex(), merkleRootStr)

		var hb, hb2, hbChild types.BTCHeaderBytes
		hb.FromBlockHeader(btcdHeader)
		hb2.FromBlockHeader(btcdHeader)
		hbChild.FromBlockHeader(btcdHeaderChild)

		if !hb.Eq(&hb) {
			t.Errorf("BTCHeaderBytes object does not equal itself")
		}
		if !hb.Eq(&hb2) {
			t.Errorf("BTCHeaderBytes object does not equal a different object with the same bytes")
		}
		if hb.Eq(&hbChild) {
			t.Errorf("BTCHeaderBytes object equals a different object with different bytes")
		}

		if !hbChild.HasParent(&hb) {
			t.Errorf("HasParent method returns false with a correct parent")
		}
		if hbChild.HasParent(&hbChild) {
			t.Errorf("HasParent method returns true for the same object")
		}

		if !hbChild.ParentHash().Eq(&childPrevBlock) {
			t.Errorf("ParentHash did not return the parent hash")
		}

		if !hb.Hash().Eq(&childPrevBlock) {
			t.Errorf("Hash method does not return the correct hash")
		}

	})
}
