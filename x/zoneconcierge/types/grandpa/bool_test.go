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

func TestBool_EncodeDecode(t *testing.T) {
	assertRoundTripFuzz[Bool](t, 100)
	assertDecodeNilData[Bool](t)
	assertEncodeEmptyObj[Bool](t, 1)
}

func TestBool_EncodedLength(t *testing.T) {
	assertEncodedLength(t, []encodedLengthAssert{
		{NewBool(true), 1},
		{NewBool(false), 1},
	})
}

func TestBool_Encode(t *testing.T) {
	assertEncode(t, []encodingAssert{
		{NewBool(true), []byte{0x01}},
		{NewBool(false), []byte{0x00}},
	})
}

func TestBool_Decode(t *testing.T) {
	assertDecode(t, []decodingAssert{
		{[]byte{0x01}, NewBool(true)},
		{[]byte{0x00}, NewBool(false)},
	})
}

func TestBool_Hash(t *testing.T) {
	assertHash(t, []hashAssert{
		{NewBool(true), MustHexDecodeString("0xee155ace9c40292074cb6aff8c9ccdd273c81648ff1149ef36bcea6ebb8a3e25")},
		{NewBool(false), MustHexDecodeString("0x03170a2e7597b7b7e3d84c05391d139a62b157e78786d8c082f29dcf4c111314")},
	})
}

func TestBool_Hex(t *testing.T) {
	assertEncodeToHex(t, []encodeToHexAssert{
		{NewBool(true), "0x01"},
		{NewBool(false), "0x00"},
	})
}

func TestBool_String(t *testing.T) {
	assertString(t, []stringAssert{
		{NewBool(true), "true"},
		{NewBool(false), "false"},
	})
}

func TestBool_Eq(t *testing.T) {
	assertEq(t, []eqAssert{
		{NewBool(true), NewBool(true), true},
		{NewBool(false), NewBool(true), false},
		{NewBool(false), NewBool(false), true},
		{NewBool(true), NewBytes([]byte{0, 1, 2}), false},
		{NewBool(true), NewBytes([]byte{1}), false},
		{NewBool(false), NewBytes([]byte{0}), false},
	})
}
