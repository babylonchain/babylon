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
	"bytes"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestU8_EncodeDecode(t *testing.T) {
	assertRoundtrip(t, NewU8(0))
	assertRoundtrip(t, NewU8(12))
}

func TestU8_EncodedLength(t *testing.T) {
	assertEncodedLength(t, []encodedLengthAssert{{NewU8(13), 1}})
}

func TestU8_Encode(t *testing.T) {
	assertEncode(t, []encodingAssert{
		{NewU8(29), MustHexDecodeString("0x1d")},
	})
}

func TestU8_Hash(t *testing.T) {
	assertHash(t, []hashAssert{
		{NewU8(29), MustHexDecodeString("0x6a9843ae0195ae1e6f95c7fbd34a42414c77e243aa18a959b5912a1f0f391b54")},
	})
}

func TestU8_Hex(t *testing.T) {
	assertEncodeToHex(t, []encodeToHexAssert{
		{NewU8(29), "0x1d"},
	})
}

func TestU8_String(t *testing.T) {
	assertString(t, []stringAssert{
		{NewU8(29), "29"},
	})
}

func TestU8_Eq(t *testing.T) {
	assertEq(t, []eqAssert{
		{NewU8(23), NewU8(23), true},
		{NewU8(23), NewBool(false), false},
	})
}

func TestU8_MarshalUnmarshal(t *testing.T) {
	u := NewU8(3)

	assertJSONRoundTrip(t, &u)
}

func TestU16_EncodeDecode(t *testing.T) {
	assertRoundtrip(t, NewU16(0))
	assertRoundtrip(t, NewU16(12))
}

func TestU16_EncodedLength(t *testing.T) {
	assertEncodedLength(t, []encodedLengthAssert{{NewU16(13), 2}})
}

func TestU16_Encode(t *testing.T) {
	assertEncode(t, []encodingAssert{
		{NewU16(29), MustHexDecodeString("0x1d00")},
	})
}

func TestU16_Hash(t *testing.T) {
	assertHash(t, []hashAssert{
		{NewU16(29), MustHexDecodeString("0x4e59f743a8e19ecb3022652bdef4343e62793d1f7378a688a82741b5d029d3d5")},
	})
}

func TestU16_Hex(t *testing.T) {
	assertEncodeToHex(t, []encodeToHexAssert{
		{NewU16(29), "0x1d00"},
	})
}

func TestU16_String(t *testing.T) {
	assertString(t, []stringAssert{
		{NewU16(29), "29"},
	})
}

func TestU16_Eq(t *testing.T) {
	assertEq(t, []eqAssert{
		{NewU16(23), NewU16(23), true},
		{NewU16(23), NewBool(false), false},
	})
}

func TestU16_MarshalUnmarshal(t *testing.T) {
	u := NewU16(3)

	assertJSONRoundTrip(t, &u)
}

func TestU32_EncodeDecode(t *testing.T) {
	assertRoundtrip(t, NewU32(0))
	assertRoundtrip(t, NewU32(12))
}

func TestU32_EncodedLength(t *testing.T) {
	assertEncodedLength(t, []encodedLengthAssert{{NewU32(13), 4}})
}

func TestU32_Encode(t *testing.T) {
	assertEncode(t, []encodingAssert{
		{NewU32(29), MustHexDecodeString("0x1d000000")},
	})
}

func TestU32_Hash(t *testing.T) {
	assertHash(t, []hashAssert{
		{NewU32(29), MustHexDecodeString("0x60ebb66f09bc7fdd21772ab1ed0efb1fd1208e3f5cd20d2d9a29a2a79b6f953f")},
	})
}

func TestU32_Hex(t *testing.T) {
	assertEncodeToHex(t, []encodeToHexAssert{
		{NewU32(29), "0x1d000000"},
	})
}

func TestU32_String(t *testing.T) {
	assertString(t, []stringAssert{
		{NewU32(29), "29"},
	})
}

func TestU32_Eq(t *testing.T) {
	assertEq(t, []eqAssert{
		{NewU32(23), NewU32(23), true},
		{NewU32(23), NewBool(false), false},
	})
}

func TestU32_MarshalUnmarshal(t *testing.T) {
	u := NewU32(3)

	assertJSONRoundTrip(t, &u)
}

func TestU64_EncodeDecode(t *testing.T) {
	assertRoundtrip(t, NewU64(0))
	assertRoundtrip(t, NewU64(12))
}

func TestU64_EncodedLength(t *testing.T) {
	assertEncodedLength(t, []encodedLengthAssert{{NewU64(13), 8}})
}

func TestU64_Encode(t *testing.T) {
	assertEncode(t, []encodingAssert{
		{NewU64(29), MustHexDecodeString("0x1d00000000000000")},
	})
}

func TestU64_Hash(t *testing.T) {
	assertHash(t, []hashAssert{
		{NewU64(29), MustHexDecodeString("0x83e168a13a013e6d47b0778f046aaa05d6c01d6857d044d9e9b658a6d85eb865")},
	})
}

func TestU64_Hex(t *testing.T) {
	assertEncodeToHex(t, []encodeToHexAssert{
		{NewU64(29), "0x1d00000000000000"},
	})
}

func TestU64_String(t *testing.T) {
	assertString(t, []stringAssert{
		{NewU64(29), "29"},
	})
}

func TestU64_Eq(t *testing.T) {
	assertEq(t, []eqAssert{
		{NewU64(23), NewU64(23), true},
		{NewU64(23), NewBool(false), false},
	})
}

func TestU64_MarshalUnmarshal(t *testing.T) {
	u := NewU64(3)

	assertJSONRoundTrip(t, &u)
}

func TestUCompact_EncodeDecode(t *testing.T) {
	bn := MustHexDecodeString("0x5C2D3BE75CEF559F050") //27205758526767196926032
	uc := NewUCompact(big.NewInt(0).SetBytes(bn))

	// Encode
	var buffer = bytes.Buffer{}
	err := NewEncoder(&buffer).Encode(uc)
	assert.NoError(t, err)
	assert.Equal(t, buffer.Bytes(), MustHexDecodeString("0x1b50f059f5ce75bed3c205")) // Encoded number above

	// Decode
	dec := NewDecoder(bytes.NewReader(buffer.Bytes()))
	var res UCompact
	err = dec.Decode(&res)
	assert.NoError(t, err)
	assert.Equal(t, uc, res)
}

func TestUCompact_EncodeDecode_MaxValue(t *testing.T) {
	// Valid Max Number Encode/Decode [ 67 bytes -> MAX = (2**536)-1 ]
	bigNumber := []byte{
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff,
	}
	maxValue := NewUCompact(new(big.Int).SetBytes(bigNumber))
	expectedEncoded := MustHexDecodeString("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")

	var buffer = bytes.Buffer{}
	err := NewEncoder(&buffer).Encode(maxValue)
	assert.NoError(t, err)
	assert.Equal(t, buffer.Bytes(), expectedEncoded)

	dec := NewDecoder(bytes.NewReader(buffer.Bytes()))
	var res UCompact
	err = dec.Decode(&res)
	assert.NoError(t, err)
	assert.Equal(t, maxValue, res)

	// Invalid max number Encode/Decode [ 68 bytes ]
	bigNumber = append(bigNumber, []byte{0xff}...)
	maxValue = NewUCompact(new(big.Int).SetBytes(bigNumber))

	buffer = bytes.Buffer{}
	err = NewEncoder(&buffer).Encode(maxValue)
	assert.Error(t, err)

	// Invalid big number max length field max 256 - 272
	reallyBigNumber := append(bigNumber, append(bigNumber, append(bigNumber, bigNumber...)...)...)
	maxValue = NewUCompact(new(big.Int).SetBytes(reallyBigNumber))

	buffer = bytes.Buffer{}
	err = NewEncoder(&buffer).Encode(maxValue)
	assert.Error(t, err)

	// Decoding truncates at max length
	expectedEncoded = append(expectedEncoded, []byte{0xab, 0xff, 0x34}...)
	dec = NewDecoder(bytes.NewReader(expectedEncoded))
	var res1 UCompact
	err = dec.Decode(&res1)
	assert.NoError(t, err)
	assert.Equal(t, res, res1)
}

func TestUCompact_EncodeNegative(t *testing.T) {
	negNumber := NewUCompact(big.NewInt(-100))
	var buffer = bytes.Buffer{}
	err := NewEncoder(&buffer).Encode(negNumber)
	assert.Error(t, err)
}

func TestU128_EncodeDecode(t *testing.T) {
	assertRoundtrip(t, NewU128(*big.NewInt(0)))
	assertRoundtrip(t, NewU128(*big.NewInt(12)))

	bigPos := big.NewInt(0)
	bigPos.SetBytes([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16})
	assertRoundtrip(t, NewU128(*bigPos))

	assertDecodeNilData[U128](t)
}

func TestU128_EncodedLength(t *testing.T) {
	assertEncodedLength(t, []encodedLengthAssert{{NewU128(*big.NewInt(13)), 16}})
}

func TestU128_Encode(t *testing.T) {
	a := big.NewInt(0).SetBytes([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 18, 52})
	b := big.NewInt(0).SetBytes([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16})
	c := big.NewInt(0).SetBytes([]byte{255, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16})

	assertEncode(t, []encodingAssert{
		{NewU128(*big.NewInt(0)), MustHexDecodeString("0x00000000000000000000000000000000")},
		{NewU128(*big.NewInt(29)), MustHexDecodeString("0x1d000000000000000000000000000000")},
		{NewU128(*a), MustHexDecodeString("0x34120000000000000000000000000000")},
		{NewU128(*b), MustHexDecodeString("0x100f0e0d0c0b0a090807060504030201")},
		{NewU128(*c), MustHexDecodeString("0x100f0e0d0c0b0a0908070605040302ff")},
	})
}

func TestU128_Decode(t *testing.T) {
	a := big.NewInt(0).SetBytes([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 18, 52})
	b := big.NewInt(0).SetBytes([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16})
	c := big.NewInt(0).SetBytes([]byte{255, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16})

	assertDecode(t, []decodingAssert{
		{MustHexDecodeString("0x00000000000000000000000000000000"), NewU128(*big.NewInt(0))},
		{MustHexDecodeString("0x1d000000000000000000000000000000"), NewU128(*big.NewInt(29))},
		{MustHexDecodeString("0x34120000000000000000000000000000"), NewU128(*a)},
		{MustHexDecodeString("0x100f0e0d0c0b0a090807060504030201"), NewU128(*b)},
		{MustHexDecodeString("0x100f0e0d0c0b0a0908070605040302ff"), NewU128(*c)},
	})
}

func TestU128_Hash(t *testing.T) {
	assertHash(t, []hashAssert{
		{NewU128(*big.NewInt(29)), MustHexDecodeString(
			"0x139bd9153bbc4913d4161f7a5dd39912b5d22b57a8b557f0a24078a11f943174")},
	})
}

func TestU128_Hex(t *testing.T) {
	assertEncodeToHex(t, []encodeToHexAssert{
		{NewU128(*big.NewInt(29)), "0x1d000000000000000000000000000000"},
	})
}

func TestU128_String(t *testing.T) {
	assertString(t, []stringAssert{
		{NewU128(*big.NewInt(29)), "29"},
	})
}

func TestU128_Eq(t *testing.T) {
	assertEq(t, []eqAssert{
		{NewU128(*big.NewInt(23)), NewU128(*big.NewInt(23)), true},
		{NewU128(*big.NewInt(23)), NewU64(23), false},
		{NewU128(*big.NewInt(23)), NewBool(false), false},
	})
}

func TestU128_GobEncodeDecode(t *testing.T) {
	u := NewU128(*big.NewInt(123))
	b, err := u.GobEncode()
	assert.NoError(t, err)

	target := new(U128)

	err = target.GobDecode(b)
	assert.NoError(t, err)

	assertEqual(t, u, *target)
}

func TestU256_EncodeDecode(t *testing.T) {
	assertRoundtrip(t, NewU256(*big.NewInt(0)))
	assertRoundtrip(t, NewU256(*big.NewInt(12)))

	bigPos := big.NewInt(0)
	bigPos.SetBytes([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
		17, 18, 19, 20, 21, 22, 23, 24, 26, 27, 28, 29, 30, 31, 32})
	assertRoundtrip(t, NewU256(*bigPos))

	assertDecodeNilData[U256](t)
}

func TestU256_EncodedLength(t *testing.T) {
	assertEncodedLength(t, []encodedLengthAssert{{NewU256(*big.NewInt(13)), 32}})
}

func TestU256_Encode(t *testing.T) {
	a := big.NewInt(0).SetBytes([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 18, 52})
	b := big.NewInt(0).SetBytes([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
		17, 18, 19, 20, 21, 22, 23, 24, 26, 27, 28, 29, 30, 31, 32})

	assertEncode(t, []encodingAssert{
		{NewU256(*big.NewInt(0)), MustHexDecodeString(
			"0x0000000000000000000000000000000000000000000000000000000000000000")},
		{NewU256(*big.NewInt(29)), MustHexDecodeString(
			"0x1d00000000000000000000000000000000000000000000000000000000000000")},
		{NewU256(*a), MustHexDecodeString("0x3412000000000000000000000000000000000000000000000000000000000000")},
		{NewU256(*b), MustHexDecodeString("0x201f1e1d1c1b1a1817161514131211100f0e0d0c0b0a09080706050403020100")},
	})
}

func TestU256_Decode(t *testing.T) {
	a := big.NewInt(0).SetBytes([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 18, 52})
	b := big.NewInt(0).SetBytes([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
		17, 18, 19, 20, 21, 22, 23, 24, 26, 27, 28, 29, 30, 31, 32})

	assertDecode(t, []decodingAssert{
		{MustHexDecodeString("0x0000000000000000000000000000000000000000000000000000000000000000"),
			NewU256(*big.NewInt(0))},
		{MustHexDecodeString("0x1d00000000000000000000000000000000000000000000000000000000000000"),
			NewU256(*big.NewInt(29))},
		{MustHexDecodeString("0x3412000000000000000000000000000000000000000000000000000000000000"), NewU256(*a)},
		{MustHexDecodeString("0x201f1e1d1c1b1a1817161514131211100f0e0d0c0b0a09080706050403020100"), NewU256(*b)},
	})
}

func TestU256_Hash(t *testing.T) {
	assertHash(t, []hashAssert{
		{NewU256(*big.NewInt(29)), MustHexDecodeString(
			"0x92d6618c3e5941a74d1e805e1a8485469229f9a9c58145761bd9209bc2f4360d")},
	})
}

func TestU256_Hex(t *testing.T) {
	assertEncodeToHex(t, []encodeToHexAssert{
		{NewU256(*big.NewInt(29)), "0x1d00000000000000000000000000000000000000000000000000000000000000"},
	})
}

func TestU256_String(t *testing.T) {
	assertString(t, []stringAssert{
		{NewU256(*big.NewInt(29)), "29"},
	})
}

func TestU256_Eq(t *testing.T) {
	assertEq(t, []eqAssert{
		{NewU256(*big.NewInt(23)), NewU256(*big.NewInt(23)), true},
		{NewU256(*big.NewInt(23)), NewU128(*big.NewInt(23)), false},
		{NewU256(*big.NewInt(23)), NewI256(*big.NewInt(23)), false},
		{NewU256(*big.NewInt(23)), NewU64(23), false},
		{NewU256(*big.NewInt(23)), NewBool(false), false},
	})
}

func TestBigIntToUintBytes(t *testing.T) {
	res, err := BigIntToUintBytes(big.NewInt(4), 2)
	assert.NoError(t, err)
	assert.Equal(t, MustHexDecodeString("0x0004"), res)

	_, err = BigIntToUintBytes(big.NewInt(0).Neg(big.NewInt(4)), 2)
	assert.EqualError(t, err, "cannot encode a negative big.Int into an unsigned integer")

	_, err = BigIntToUintBytes(big.NewInt(266), 1)
	assert.EqualError(t, err, "cannot encode big.Int to []byte: given big.Int exceeds highest number 256 for an "+
		"uint with 8 bits")
}

func TestUintBytesToBigInt(t *testing.T) {
	res, err := UintBytesToBigInt(MustHexDecodeString("0x0004"))
	assert.NoError(t, err)
	assert.Equal(t, big.NewInt(4), res)

	res, err = UintBytesToBigInt(MustHexDecodeString("0xfffc"))
	assert.NoError(t, err)
	assert.Equal(t, big.NewInt(65532), res)

	_, err = UintBytesToBigInt([]byte{})
	assert.EqualError(t, err, "cannot decode an empty byte slice")
}

func TestBigIntToUintBytes_128(t *testing.T) {
	b := big.NewInt(0)
	b.SetBytes([]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x12, 0x34})

	res, err := BigIntToUintBytes(b, 16)
	assert.NoError(t, err)
	assert.Equal(t, MustHexDecodeString("0x00000000000000000000000000001234"), res)
}
