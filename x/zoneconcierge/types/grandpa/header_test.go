// Go Substrate RPC Client (GSRPC) provides APIs and types around Polkadot and any Substrate-based chain RPC calls
//
// Copyright 2019 Centrifuge GmbH
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package grandpa_types_test

import (
	"testing"

	. "github.com/babylonchain/babylon/x/zoneconcierge/types/grandpa"
)

var exampleHeader = Header{
	ParentHash:     Hash{1, 2, 3, 4, 5},
	Number:         42,
	StateRoot:      Hash{2, 3, 4, 5, 6},
	ExtrinsicsRoot: Hash{3, 4, 5, 6, 7},
	Digest: Digest{
		{IsOther: true, AsOther: Bytes{4, 5}},
		{IsChangesTrieRoot: true, AsChangesTrieRoot: Hash{6, 7}},
		{IsConsensus: true, AsConsensus: Consensus{ConsensusEngineID: [4]byte{9}, Bytes: Bytes{10, 11, 12}}},
		{IsSeal: true, AsSeal: Seal{ConsensusEngineID: [4]byte{11}, Bytes: Bytes{12, 13, 14}}},
		{IsPreRuntime: true, AsPreRuntime: PreRuntime{ConsensusEngineID: [4]byte{13}, Bytes: Bytes{14, 15, 16}}},
	},
}

var hash64 = []byte{
	1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0,
	1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0,
	1, 2, 3, 4,
}

func TestBlockNumber_EncodeDecode(t *testing.T) {
	assertRoundTripFuzz[BlockNumber](t, 100)
	assertEncodeEmptyObj[BlockNumber](t, 1)
}

func TestBlockNumber_JSONMarshalUnmarshal(t *testing.T) {
	b := BlockNumber(1)
	assertJSONRoundTrip(t, &b)
}

var (
	headerFuzzOpts = digestItemFuzzOpts
)

func TestHeader_EncodeDecode(t *testing.T) {
	assertRoundtrip(t, exampleHeader)
	assertRoundtripHeader(t, exampleHeader)
	//assertRoundTripFuzz[Header](t, 100, headerFuzzOpts...)
	assertDecodeNilData[Header](t)
	assertEncodeEmptyObj[Header](t, 98)
}

func TestHeader_EncodedLength(t *testing.T) {
	assertEncodedLength(t, []encodedLengthAssert{{exampleHeader, 162}})
}

func TestHeader_Encode(t *testing.T) {
	assertEncode(t, []encodingAssert{
		{exampleHeader, MustHexDecodeString("0x0102030405000000000000000000000000000000000000000000000000000000a802030405060000000000000000000000000000000000000000000000000000000304050607000000000000000000000000000000000000000000000000000000140008040502060700000000000000000000000000000000000000000000000000000000000004090000000c0a0b0c050b0000000c0c0d0e060d0000000c0e0f10")}, //nolint:lll
	})
}

func TestHeader_Hex(t *testing.T) {
	assertEncodeToHex(t, []encodeToHexAssert{
		{exampleHeader, "0x0102030405000000000000000000000000000000000000000000000000000000a802030405060000000000000000000000000000000000000000000000000000000304050607000000000000000000000000000000000000000000000000000000140008040502060700000000000000000000000000000000000000000000000000000000000004090000000c0a0b0c050b0000000c0c0d0e060d0000000c0e0f10"}, //nolint:lll
	})
}

func TestHeader_Eq(t *testing.T) {
	assertEq(t, []eqAssert{
		{exampleHeader, exampleHeader, true},
		{exampleHeader, NewBytes(hash64), false},
		{exampleHeader, NewBool(false), false},
	})
}
