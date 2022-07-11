package types_test

import (
	"bytes"
	"encoding/json"
	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/babylonchain/babylon/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"math/rand"
	"testing"
)

func FuzzBTCHeaderHashBytesBytesOps(f *testing.F) {
	f.Add(int64(42))

	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)

		invalidHash := false
		bz := datagen.GenRandomByteArray(types.BTCHeaderHashLen)
		// 1/10 times generate an invalid size
		if datagen.OneInN(10) {
			bz = datagen.GenRandomByteArray(datagen.RandomInt(types.BTCHeaderHashLen * 10))
			invalidHash = true
		}

		hbb, err := types.NewBTCHeaderHashBytesFromBytes(bz)
		if err != nil {
			if !invalidHash {
				t.Fatalf("Valid hash %s led to a NewBTCHeaderHashBytesFromBytes error %s", bz, err)
			}
			t.Skip()
		}

		err = hbb.Unmarshal(bz)
		if err != nil {
			if !invalidHash {
				t.Fatalf("Valid hash %s led to an unmarshal error %s", bz, err)
			}
			t.Skip()
		}

		m, err := hbb.Marshal()
		if err != nil {
			if !invalidHash {
				t.Fatalf("Marshaling of bytes slice led to error %s", err)
			}
			t.Skip()
		}
		if bytes.Compare(m, bz) != 0 {
			t.Errorf("Marshal returned %s while %s was expected", m, bz)
		}

		m = make([]byte, len(bz))
		sz, err := hbb.MarshalTo(m)
		if err != nil {
			if !invalidHash {
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

		if invalidHash {
			t.Errorf("Invalid hash succeeded in all operations")
		}

		sz = hbb.Size()
		if sz != len(bz) {
			t.Errorf("Size returned %d while %d was expected", sz, len(bz))
		}
	})
}

func FuzzBTCHeaderHashBytesHexOps(f *testing.F) {
	f.Add(int64(42))

	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)

		invalidHash := false
		hex := datagen.GenRandomHexStr(types.BTCHeaderHashLen)
		// 1/4 times generate an invalid hash
		if datagen.OneInN(10) {
			if datagen.OneInN(2) {
				// 1/4 times generate an invalid hash size
				hex = datagen.GenRandomHexStr(datagen.RandomInt(types.BTCHeaderHashLen * 20))
			} else {
				// 1/4 times generate an invalid hex
				hex = string(datagen.GenRandomByteArray(types.BTCHeaderHashLen * 2))
			}
			invalidHash = true
		}
		hbb, err := types.NewBTCHeaderHashBytesFromHex(hex)
		if err != nil {
			if !invalidHash {
				t.Fatalf("Valid header %s %d led to a NewBTCHeaderHashBytesFromHex error %s", hex, len(hex), err)
			}
			t.Skip()
		}

		err = hbb.UnmarshalHex(hex)
		if err != nil {
			if !invalidHash {
				t.Fatalf("Valid hash %s led to an unmarshal error %s", hex, err)
			}
			t.Skip()
		}

		h := hbb.MarshalHex()
		if h != hex {
			t.Errorf("Marshal returned %s while %s was expected", h, hex)
		}
		if invalidHash {
			t.Errorf("Invalid hash passed all checks")
		}
	})
}

func FuzzBTCHeaderHashBytesJSONOps(f *testing.F) {
	f.Add(int64(42))

	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)

		invalidHash := false
		hex := datagen.GenRandomHexStr(types.BTCHeaderHashLen)
		// 1/4 times generate an invalid hash
		if datagen.OneInN(10) {
			if datagen.OneInN(2) {
				// 1/4 times generate an invalid hash size
				hex = datagen.GenRandomHexStr(datagen.RandomInt(types.BTCHeaderHashLen * 20))
			} else {
				// 1/4 times generate an invalid hex
				hex = string(datagen.GenRandomByteArray(types.BTCHeaderHashLen * 2))
			}
			invalidHash = true
		}
		hbb, err := types.NewBTCHeaderHashBytesFromHex(hex)
		if err != nil {
			if !invalidHash {
				t.Fatalf("Valid hash %s %d led to a NewBTCHeaderHashBytesFromHex error %s", hex, len(hex), err)
			}
			t.Skip()
		}

		jsonHex, _ := json.Marshal(hex)

		err = hbb.UnmarshalJSON(jsonHex)
		if err != nil {
			if !invalidHash {
				t.Fatalf("Valid hash %s led to an unmarshal error %s", hex, err)
			}
			t.Skip()
		}

		h, err := hbb.MarshalJSON()
		if err != nil {
			if !invalidHash {
				t.Fatalf("Valid hash %s led to an marshal error %s", hex, err)
			}
			t.Skip()
		}
		if bytes.Compare(h, jsonHex) != 0 {
			t.Errorf("Marshal returned %s while %s was expected", h, jsonHex)
		}

		if invalidHash {
			t.Errorf("Invalid hash passed all checks")
		}
	})
}

func FuzzHeaderHashBytesChainhashOps(f *testing.F) {
	f.Add(int64(42))
	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)

		hexHash := datagen.GenRandomHexStr(types.BTCHeaderHashLen)
		chHash, _ := chainhash.NewHashFromStr(hexHash)

		var hbb types.BTCHeaderHashBytes
		hbb.FromChainhash(chHash)
		gotHash := hbb.ToChainhash()
		if chHash.String() != gotHash.String() {
			t.Errorf("Got hash %s while %s was expected", gotHash, chHash)
		}

		hbb = types.NewBTCHeaderHashBytesFromChainhash(chHash)
		gotHash = hbb.ToChainhash()
		if chHash.String() != gotHash.String() {
			t.Errorf("Got hash %s while %s was expected", gotHash, chHash)
		}
	})
}
