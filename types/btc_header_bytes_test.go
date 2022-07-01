package types_test

import (
	"encoding/hex"
	"encoding/json"
	"github.com/babylonchain/babylon/types"
	"github.com/stretchr/testify/suite"
	"strings"
	"testing"
)

type headerBytesTestSuite struct {
	suite.Suite
	valid, validHeaderHash, invalid, invalidHex, tooLong, tooShort string
}

func TestHeaderBytesTestSuite(t *testing.T) {
	suite.Run(t, new(headerBytesTestSuite))
}

func (s *headerBytesTestSuite) SetupSuite() {
	s.T().Parallel()
	s.valid = "00006020c6c5a20e29da938a252c945411eba594cbeba021a1e20000000000000000000039e4bd0cd0b5232bb380a9576fcfe7d8fb043523f7a158187d9473e44c1740e6b4fa7c62ba01091789c24c22"
	s.validHeaderHash = "00000000000000000002bf1c218853bc920f41f74491e6c92c6bc6fdc881ab47"
	s.invalid = "notvalidhex"
	s.tooLong = "00006020c6c5a20e29da938a252c945411eba594cbeba021a1e20000000000000000000039e4bd0cd0b5232bb380a9576fcfe7d8fb043523f7a158187d9473e44c1740e6b4fa7c62ba01091789c24c22222"
	s.tooShort = "00006020c6c5a20e29da938a252c945411eba594cbeba021a1e20000000000000000000039e4bd0cd0b5232bb380a9576fcfe7d8fb043523f7a158187d9473e44c1740e6b4fa7c62ba01091789c24c"
}

func (s *headerBytesTestSuite) TestBTCHeaderBytes_Marshal() {
	// Marshal should just return the bytes themselves
	data := []struct {
		name string
		b    []byte
	}{
		{"one length", []byte("a")},
		{"long length", []byte("aaaa")},
		{"zero length", []byte("")},
	}
	for _, d := range data {
		hb := types.BTCHeaderBytes(d.b)
		m, err := hb.Marshal()
		s.Require().NoError(err, d.name)
		s.Require().Equal(d.b, m, d.name)
	}
}

func (s *headerBytesTestSuite) TestBTCHeaderBytes_MarshalHex() {
	data := []struct {
		name string
		hex  string
	}{
		{"valid", s.valid},
	}
	for _, d := range data {
		var hb types.BTCHeaderBytes
		hb.UnmarshalHex(d.hex)

		h, err := hb.MarshalHex()
		s.Require().Equal(d.hex, h, d.name)
		s.Require().NoError(err, d.name)
	}
}

func (s *headerBytesTestSuite) TestBTCHeaderBytes_MarshalJSON() {
	data := []struct {
		name string
		hex  string
	}{
		{"valid", s.valid},
	}
	for _, d := range data {
		jme, _ := json.Marshal(d.hex)

		var hb types.BTCHeaderBytes
		hb.UnmarshalHex(d.hex)

		jm, err := hb.MarshalJSON()
		s.Require().Equal(jme, jm, d.name)
		s.Require().NoError(err, d.name)
	}
}

func (s *headerBytesTestSuite) TestBTCHeaderBytes_MarshalTo() {
	// MarshalTo should just copy the bytes
	data := []struct {
		name string
		b    []byte
	}{
		{"one length", []byte("a")},
		{"long length", []byte("aaaa")},
		{"zero length", []byte("")},
	}
	for _, d := range data {
		hb := types.BTCHeaderBytes(d.b)
		bz := make([]byte, len(d.b))

		size, err := hb.MarshalTo(bz)
		s.Require().NoError(err, d.name)
		s.Require().Equal(d.b, bz, d.name)
		s.Require().Equal(len(d.b), size, d.name)
	}
}

func (s *headerBytesTestSuite) TestBTCHeaderBytes_Size() {
	data := []struct {
		name string
		b    []byte
	}{
		{"one length", []byte("a")},
		{"long length", []byte("aaaa")},
		{"zero length", []byte("")},
	}
	for _, d := range data {
		hb := types.BTCHeaderBytes(d.b)
		s.Require().Equal(len(d.b), hb.Size(), d.name)
	}
}

func (s *headerBytesTestSuite) TestBTCHeaderBytes_Unmarshal() {
	// Unmarshal should check whether the data has a length of 80 bytes
	// and then copy the data into the header
	bz := []byte(strings.Repeat("a", types.HeaderLen))
	data := []struct {
		name   string
		bytes  []byte
		hasErr bool
	}{
		{"valid", bz, false},
		{"too long", append(bz, byte('a')), true},
		{"too short", bz[:types.HeaderLen-1], true},
	}
	for _, d := range data {
		var hb types.BTCHeaderBytes
		err := hb.Unmarshal(d.bytes)
		if d.hasErr {
			s.Require().Error(err, d.name)
		} else {
			s.Require().NoError(err, d.name)
			s.Require().Equal(types.BTCHeaderBytes(d.bytes), hb, d.name)
		}
	}
}

func (s *headerBytesTestSuite) TestBTCHeaderBytes_UnmarshalHex() {
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
		var hb types.BTCHeaderBytes
		err := hb.UnmarshalHex(d.hex)
		if d.hasErr {
			s.Require().Error(err, d.name)
		} else {
			s.Require().NoError(err, d.name)
			decoded, _ := hex.DecodeString(d.hex)
			s.Require().Equal(types.BTCHeaderBytes(decoded), hb, d.name)
		}
	}
}

func (s *headerBytesTestSuite) TestBTCHeaderBytes_UnmarshalJSON() {
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
		var hb types.BTCHeaderBytes
		err := hb.UnmarshalJSON(d.jms)
		if d.hasErr {
			s.Require().Error(err, d.name)
		} else {
			s.Require().NoError(err, d.name)
			var bz types.BTCHeaderBytes
			json.Unmarshal(d.jms, &bz)
			s.Require().Equal(bz, hb, d.name)
		}
	}
}

func (s *headerBytesTestSuite) TestBTCHeaderBytes_ToBlockHeader() {
	data := []struct {
		name       string
		header     string
		headerHash string
	}{{"valid", s.valid, s.validHeaderHash}}

	for _, d := range data {
		var hb types.BTCHeaderBytes
		hb.UnmarshalHex(d.header)

		btcdBlock, err := hb.ToBlockHeader()
		s.Require().NoError(err)
		s.Require().Equal(d.headerHash, btcdBlock.BlockHash().String(), d.name)
	}
}

func (s *headerBytesTestSuite) TestBTCHeaderBytes_FromBlockHeader() {
	data := []struct {
		name string
		hex  string
	}{{"valid", s.valid}}

	for _, d := range data {
		var hb types.BTCHeaderBytes
		hb.UnmarshalHex(d.hex)
		btcdBlock, _ := hb.ToBlockHeader()

		var hb2 types.BTCHeaderBytes
		err := hb2.FromBlockHeader(btcdBlock)

		s.Require().NoError(err)

		bz1, _ := hb.Marshal()
		bz2, _ := hb2.Marshal()
		s.Require().Equal(bz1, bz2, d.name)
	}
}
