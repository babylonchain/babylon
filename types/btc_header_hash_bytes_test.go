package types_test

import (
	"encoding/json"
	"github.com/babylonchain/babylon/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/stretchr/testify/suite"
	"strings"
	"testing"
)

type headerHashBytesTestSuite struct {
	suite.Suite
	valid, invalid, invalidHex, tooLong, tooShort string
}

func TestHeaderHashBytesTestSuite(t *testing.T) {
	suite.Run(t, new(headerHashBytesTestSuite))
}

func (s *headerHashBytesTestSuite) SetupSuite() {
	s.T().Parallel()
	s.valid = "00000000000000000002bf1c218853bc920f41f74491e6c92c6bc6fdc881ab47"
	s.invalid = "notvalidhex"
	s.tooLong = "00000000000000000002bf1c218853bc920f41f74491e6c92c6bc6fdc881ab47aa"
	s.tooShort = "00000000000000000002bf1c218853bc920f41f74491e6c92c6bc6fdc881"
}

func (s *headerHashBytesTestSuite) TestBTCHeaderHashBytes_Marshal() {
	// Marshal should just return the bytes themselves
	data := []struct {
		name string
		b    []byte
		size int
	}{
		{"one length", []byte("a"), 1},
		{"long length", []byte("aaaa"), 4},
		{"zero length", []byte(""), 0},
	}
	for _, d := range data {
		hhb := types.BTCHeaderHashBytes(d.b)
		m, err := hhb.Marshal()
		s.Require().NoError(err, d.name)
		s.Require().Equal(d.b, m, d.name)
	}
}

func (s *headerHashBytesTestSuite) TestBTCHeaderHashBytes_MarshalHex() {
	data := []struct {
		name string
		hex  string
	}{
		{"valid", s.valid},
	}
	for _, d := range data {
		var hhb types.BTCHeaderHashBytes
		hhb.UnmarshalHex(d.hex)

		h, err := hhb.MarshalHex()
		s.Require().Equal(d.hex, h, d.name)
		s.Require().NoError(err, d.name)
	}
}

func (s *headerHashBytesTestSuite) TestBTCHeaderHashBytes_MarshalJSON() {
	data := []struct {
		name string
		hex  string
	}{
		{"valid", s.valid},
	}
	for _, d := range data {
		jme, _ := json.Marshal(d.hex)

		var hb types.BTCHeaderHashBytes
		hb.UnmarshalHex(d.hex)

		jm, err := hb.MarshalJSON()
		s.Require().Equal(jme, jm, d.name)
		s.Require().NoError(err, d.name)
	}
}

func (s *headerHashBytesTestSuite) TestBTCHeaderHashBytes_MarshalTo() {
	// MarshalTo should just copy the bytes
	data := []struct {
		name string
		b    []byte
		size int
	}{
		{"one length", []byte("a"), 1},
		{"long length", []byte("aaaa"), 4},
		{"zero length", []byte(""), 0},
	}
	for _, d := range data {
		hb := types.BTCHeaderHashBytes(d.b)
		bz := make([]byte, len(d.b))

		size, err := hb.MarshalTo(bz)
		s.Require().NoError(err, d.name)
		s.Require().Equal(d.b, bz, d.name)
		s.Require().Equal(len(d.b), size, d.name)
	}
}

func (s *headerHashBytesTestSuite) TestBTCHeaderHashBytes_Size() {
	data := []struct {
		name string
		b    []byte
		size int
	}{
		{"one length", []byte("a"), 1},
		{"long length", []byte("aaaa"), 4},
		{"zero length", []byte(""), 0},
	}
	for _, d := range data {
		hb := types.BTCHeaderHashBytes(d.b)
		s.Require().Equal(d.size, hb.Size(), d.name)
	}
}

func (s *headerHashBytesTestSuite) TestBTCHeaderHashBytes_Unmarshal() {
	// Unmarshal should check whether the data has a length of 80 bytes
	// and then copy the data into the header
	bz := []byte(strings.Repeat("a", types.HeaderHashLen))
	data := []struct {
		name   string
		bytes  []byte
		hasErr bool
	}{
		{"valid", bz, false},
		{"too long", append(bz, byte('a')), true},
		{"too short", bz[:types.HeaderHashLen-1], true},
	}
	for _, d := range data {
		var hb types.BTCHeaderHashBytes
		err := hb.Unmarshal(d.bytes)
		if d.hasErr {
			s.Require().Error(err, d.name)
		} else {
			s.Require().NoError(err, d.name)
			s.Require().Equal(types.BTCHeaderHashBytes(d.bytes), hb, d.name)
		}
	}
}

func (s *headerHashBytesTestSuite) TestBTCHeaderHashBytes_UnmarshalHex() {
	data := []struct {
		name   string
		hex    string
		hasErr bool
	}{
		{"valid", s.valid, false},
		{"invalid", s.invalid, true},
		{"too long", s.tooLong, true},
		{"too short", s.tooShort, true},
	}
	for _, d := range data {
		var hb types.BTCHeaderHashBytes
		err := hb.UnmarshalHex(d.hex)
		if d.hasErr {
			s.Require().Error(err, d.name)
		} else {
			s.Require().NoError(err, d.name)
			decoded, _ := chainhash.NewHashFromStr(d.hex)
			s.Require().Equal(types.BTCHeaderHashBytes(decoded[:]), hb, d.name)
		}
	}
}

func (s *headerHashBytesTestSuite) TestBTCHeaderHashBytes_UnmarshalJSON() {
	validJm, _ := json.Marshal(s.valid)
	invalidJm, _ := json.Marshal(s.invalid)
	notJsonJm := []byte("smth")
	data := []struct {
		name   string
		jms    []byte
		hasErr bool
	}{
		{"valid", validJm, false},
		{"invalid", invalidJm, true},
		{"not json", notJsonJm, true},
	}
	for _, d := range data {
		var hb types.BTCHeaderHashBytes
		err := hb.UnmarshalJSON(d.jms)
		if d.hasErr {
			s.Require().Error(err, d.name)
		} else {
			s.Require().NoError(err, d.name)
			var bz types.BTCHeaderHashBytes
			json.Unmarshal(d.jms, &bz)
			s.Require().Equal(bz, hb, d.name)
		}
	}
}

func (s *headerHashBytesTestSuite) TestBTCHeaderHashBytes_ToChainhash() {
	data := []struct {
		name       string
		headerHash string
	}{{"valid", s.valid}}

	for _, d := range data {
		var hhb types.BTCHeaderHashBytes
		hhb.UnmarshalHex(d.headerHash)

		chHash, err := hhb.ToChainhash()
		s.Require().NoError(err)
		s.Require().Equal(d.headerHash, chHash.String(), d.name)
	}
}

func (s *headerHashBytesTestSuite) TestBTCHeaderHashBytes_FromChainhash() {
	data := []struct {
		name string
		hex  string
	}{{"valid", s.valid}}

	for _, d := range data {
		var hhb types.BTCHeaderHashBytes
		hhb.UnmarshalHex(d.hex)
		chHash, _ := hhb.ToChainhash()

		var hhb2 types.BTCHeaderHashBytes
		err := hhb2.FromChainhash(chHash)

		s.Require().NoError(err)

		bz1, _ := hhb.Marshal()
		bz2, _ := hhb2.Marshal()
		s.Require().Equal(bz1, bz2, d.name)
	}
}
