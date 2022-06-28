package types

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"github.com/btcsuite/btcd/wire"
)

type BTCHeaderBytes []byte

func NewBTCHeaderBytesFromHex(hex string) (BTCHeaderBytes, error) {
	var headerBytes BTCHeaderBytes
	err := headerBytes.UnmarshalHex(hex)
	if err != nil {
		return nil, err
	}
	return headerBytes, nil
}

func NewBTCHeaderBytesFromBlockHeader(header *wire.BlockHeader) BTCHeaderBytes {
	var headerBytes BTCHeaderBytes
	headerBytes.FromBlockHeader(header)
	return headerBytes
}

func NewBTCHeaderBytesFromBytes(header []byte) BTCHeaderBytes {
	return header
}

func (m BTCHeaderBytes) MarshalJSON() ([]byte, error) {
	hex, err := m.MarshalHex()
	if err != nil {
		return nil, err
	}
	return json.Marshal(hex)
}

func (m *BTCHeaderBytes) UnmarshalJSON(bz []byte) error {
	var headerHexStr string
	err := json.Unmarshal(bz, &headerHexStr)

	if err != nil {
		return err
	}

	return m.UnmarshalHex(headerHexStr)
}

func (m BTCHeaderBytes) Marshal() ([]byte, error) {
	return m, nil
}

func (m *BTCHeaderBytes) Unmarshal(data []byte) error {
	*m = data
	return nil
}

func (m BTCHeaderBytes) MarshalHex() (string, error) {
	btcdHeader, err := m.ToBlockHeader()
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	btcdHeader.Serialize(&buf)
	return hex.EncodeToString(buf.Bytes()), nil
}

func (m *BTCHeaderBytes) UnmarshalHex(header string) error {
	// Decode the hash string from hex
	decoded, err := hex.DecodeString(header)
	if err != nil {
		return err
	}

	return m.Unmarshal(decoded)
}

func (m BTCHeaderBytes) MarshalTo(data []byte) (int, error) {
	copy(data, m)
	return len(data), nil
}

func (m *BTCHeaderBytes) Size() int {
	bz, _ := m.Marshal()
	return len(bz)
}

func (m BTCHeaderBytes) ToBlockHeader() (*wire.BlockHeader, error) {
	// Create an empty header
	header := &wire.BlockHeader{}

	// The Deserialize method expects an io.Reader instance
	reader := bytes.NewReader(m)
	// Decode the header bytes
	err := header.Deserialize(reader)
	// There was a parsing error
	if err != nil {
		return nil, err
	}
	return header, nil
}

func (m *BTCHeaderBytes) FromBlockHeader(header *wire.BlockHeader) {
	var buf bytes.Buffer
	header.Serialize(&buf)

	*m = buf.Bytes()
}
