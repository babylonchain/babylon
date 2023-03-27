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
	"fmt"
	"reflect"
	"testing"

	fuzz "github.com/google/gofuzz"
	"github.com/stretchr/testify/assert"


	. "github.com/babylonchain/babylon/types/grandpa"
)

type fuzzOpt func(f *fuzz.Fuzzer)

func withNilChance(p float64) fuzzOpt {
	return func(f *fuzz.Fuzzer) {
		f.NilChance(p)
	}
}

func withFuzzFuncs(fuzzFns ...any) fuzzOpt {
	return func(f *fuzz.Fuzzer) {
		f.Funcs(fuzzFns...)
	}
}

func withNumElements(atLeast, atMost int) fuzzOpt {
	return func(f *fuzz.Fuzzer) {
		f.NumElements(atLeast, atMost)
	}
}

func withMaxDepth(depth int) fuzzOpt {
	return func(f *fuzz.Fuzzer) {
		f.MaxDepth(depth)
	}
}

func combineFuzzOpts(optsSlice ...[]fuzzOpt) []fuzzOpt {
	var o []fuzzOpt

	for _, opts := range optsSlice {
		o = append(o, opts...)
	}

	return o
}

func assertRoundTripFuzz[T any](t *testing.T, fuzzCount int, fuzzOpts ...fuzzOpt) {
	f := fuzz.New().NilChance(0)

	for _, opt := range fuzzOpts {
		opt(f)
	}

	for i := 0; i < fuzzCount; i++ {
		var fuzzData T

		f.Fuzz(&fuzzData)

		assertRoundtrip(t, fuzzData)
	}
}

func assertDecodeNilData[T any](t *testing.T) {
	var obj T

	err := NewDecoder(bytes.NewReader(nil)).Decode(&obj)
	assert.NotNil(t, err)
}

func assertEncodeEmptyObj[T any](t *testing.T, expectedByteLen int) {
	var obj T

	var buffer = bytes.Buffer{}

	err := NewEncoder(&buffer).Encode(obj)
	assert.Nil(t, err)
	assert.Len(t, buffer.Bytes(), expectedByteLen)
}

type encodedLengthAssert struct {
	input    interface{}
	expected int
}

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	if reflect.DeepEqual(a, b) {
		return
	}
	t.Errorf("Received %#v (type %v), expected %#v (type %v)", a, reflect.TypeOf(a), b, reflect.TypeOf(b))
}

func assertRoundtrip(t *testing.T, value interface{}) {
	var buffer = bytes.Buffer{}
	err := NewEncoder(&buffer).Encode(value)
	assert.NoError(t, err)
	target := reflect.New(reflect.TypeOf(value))
	err = NewDecoder(&buffer).Decode(target.Interface())
	assert.NoError(t, err)
	assertEqual(t, target.Elem().Interface(), value)
}

func assertRoundtripHeader(t *testing.T, value Header) {
	var buffer = bytes.Buffer{}
	err := NewEncoder(&buffer).Encode(value)
	assert.NoError(t, err)
	target := reflect.New(reflect.TypeOf(value))
	fmt.Println(target.Interface())
	header, err := DecodeHeader(NewDecoder(&buffer))
	// err = NewDecoder(&buffer).Decode(target.Interface())
	assert.NoError(t, err)
	assertEqual(t, *(header), value)
}

func assertEncodedLength(t *testing.T, encodedLengthAsserts []encodedLengthAssert) {
	for _, test := range encodedLengthAsserts {
		result, err := EncodedLength(test.input)
		if err != nil {
			t.Errorf("Encoded length error for input %v: %v\n", test.input, err)
		}
		if result != test.expected {
			t.Errorf("Fail, input %v, expected %v, result %v\n", test.input, test.expected, result)
		}
	}
}

type encodingAssert struct {
	input    interface{}
	expected []byte
}

func assertEncode(t *testing.T, encodingAsserts []encodingAssert) {
	for _, test := range encodingAsserts {
		result, err := Encode(test.input)
		if err != nil {
			t.Errorf("Encoding error for input %v: %v\n", test.input, err)
		}

		if !bytes.Equal(result, test.expected) {
			t.Errorf("Fail, input %v, expected %#x, result %#x\n", test.input, test.expected, result)
		}
	}
}

type decodingAssert struct {
	input    []byte
	expected interface{}
}

func assertDecode(t *testing.T, decodingAsserts []decodingAssert) {
	for _, test := range decodingAsserts {
		target := reflect.New(reflect.TypeOf(test.expected))
		err := Decode(test.input, target.Interface())
		if err != nil {
			t.Errorf("Encoding error for input %v: %v\n", test.input, err)
		}
		assertEqual(t, target.Elem().Interface(), test.expected)
	}
}

type hashAssert struct {
	input    interface{}
	expected []byte
}

func assertHash(t *testing.T, hashAsserts []hashAssert) {
	for _, test := range hashAsserts {
		result, err := GetHash(test.input)
		if err != nil {
			t.Errorf("Hash error for input %v: %v\n", test.input, err)
		}
		if !bytes.Equal(result[:], test.expected) {
			t.Errorf("Fail, input %v, expected %#x, result %#x\n", test.input, test.expected, result)
		}
	}
}

type encodeToHexAssert struct {
	input    interface{}
	expected string
}

func assertEncodeToHex(t *testing.T, encodeToHexAsserts []encodeToHexAssert) {
	for _, test := range encodeToHexAsserts {
		result, err := EncodeToHex(test.input)
		if err != nil {
			t.Errorf("Hex error for input %v: %v\n", test.input, err)
		}
		if result != test.expected {
			t.Errorf("Fail, input %v, expected %v, result %v\n", test.input, test.expected, result)
		}
	}
}

type stringAssert struct {
	input    interface{}
	expected string
}

func assertString(t *testing.T, stringAsserts []stringAssert) {
	for _, test := range stringAsserts {
		result := fmt.Sprintf("%v", test.input)
		if result != test.expected {
			t.Errorf("Fail, input %v, expected %v, result %v\n", test.input, test.expected, result)
		}
	}
}

type eqAssert struct {
	input    interface{}
	other    interface{}
	expected bool
}

func assertEq(t *testing.T, eqAsserts []eqAssert) {
	for _, test := range eqAsserts {
		result := Eq(test.input, test.other)
		if result != test.expected {
			t.Errorf("Fail, input %v, other %v, expected %v, result %v\n", test.input, test.other, test.expected, result)
		}
	}
}

type jsonMarshalerUnmarshaler[T any] interface {
	UnmarshalJSON([]byte) error
	MarshalJSON() ([]byte, error)

	*T // helper type that allows us to instantiate a non-nil T
}

func assertJSONRoundTrip[T any, U jsonMarshalerUnmarshaler[T]](t *testing.T, value U) {
	b, err := value.MarshalJSON()
	assert.NoError(t, err)

	tt := U(new(T))
	err = tt.UnmarshalJSON(b)
	assert.NoError(t, err)

	assertEqual(t, value, tt)
}

