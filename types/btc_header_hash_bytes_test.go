package types_test

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/babylonchain/babylon/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

func FuzzBTCHeaderHashBytesBytesOps(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		invalidHash := false
		bz := datagen.GenRandomByteArray(r, types.BTCHeaderHashLen)
		// 1/10 times generate an invalid size
		if datagen.OneInN(r, 10) {
			bzSz := datagen.RandomIntOtherThan(r, types.BTCHeaderHashLen, types.BTCHeaderHashLen*10)
			bz = datagen.GenRandomByteArray(r, bzSz)
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
		if !bytes.Equal(m, bz) {
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
			t.Errorf("MarshalTo marshalled %d bytes instead of %d", sz, len(bz))
		}
		if !bytes.Equal(m, bz) {
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
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		invalidHash := false
		hex := datagen.GenRandomHexStr(r, types.BTCHeaderHashLen)
		// 1/4 times generate an invalid hash
		if datagen.OneInN(r, 10) {
			if datagen.OneInN(r, 2) {
				// 1/4 times generate an invalid hash size
				bzSz := datagen.RandomIntOtherThan(r, types.BTCHeaderHashLen, types.BTCHeaderHashLen*20)
				hex = datagen.GenRandomHexStr(r, bzSz)
			} else {
				// 1/4 times generate an invalid hex
				hex = string(datagen.GenRandomByteArray(r, types.BTCHeaderHashLen*2))
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
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		invalidHash := false
		hex := datagen.GenRandomHexStr(r, types.BTCHeaderHashLen)
		// 1/4 times generate an invalid hash
		if datagen.OneInN(r, 10) {
			if datagen.OneInN(r, 2) {
				// 1/4 times generate an invalid hash size
				bzSz := datagen.RandomIntOtherThan(r, types.BTCHeaderHashLen, types.BTCHeaderHashLen*20)
				hex = datagen.GenRandomHexStr(r, bzSz)
			} else {
				// 1/4 times generate an invalid hex
				hex = string(datagen.GenRandomByteArray(r, types.BTCHeaderHashLen*2))
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
		if !bytes.Equal(h, jsonHex) {
			t.Errorf("Marshal returned %s while %s was expected", h, jsonHex)
		}

		if invalidHash {
			t.Errorf("Invalid hash passed all checks")
		}
	})
}

func FuzzHeaderHashBytesChainhashOps(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		hexHash := datagen.GenRandomHexStr(r, types.BTCHeaderHashLen)
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

func FuzzHeaderHashBytesOperators(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		hexHash := datagen.GenRandomHexStr(r, types.BTCHeaderHashLen)
		hexHash2 := datagen.GenRandomHexStr(r, types.BTCHeaderHashLen)
		chHash, _ := chainhash.NewHashFromStr(hexHash)

		var hbb, hbb2 types.BTCHeaderHashBytes
		hbb.FromChainhash(chHash)
		hbb2, _ = types.NewBTCHeaderHashBytesFromHex(hexHash2)

		if hbb.String() != chHash.String() {
			t.Errorf("String() method returned %s while %s was expected", hbb, chHash)
		}

		if !hbb.Eq(&hbb) {
			t.Errorf("Object does not equal itself")
		}
		if hbb.Eq(&hbb2) {
			t.Errorf("Object %s equals %s", hbb, hbb2)
		}
	})
}
