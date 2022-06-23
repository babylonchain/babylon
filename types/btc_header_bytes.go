package types

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"github.com/btcsuite/btcd/wire"
)

type BTCHeaderBytes []byte
type BTCHeadersBytes []BTCHeaderBytes

func (m BTCHeaderBytes) MarshalJSON() ([]byte, error) {
	btcdHeader, err := m.ToBtcdHeader()
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	btcdHeader.Serialize(&buf)
	str := hex.EncodeToString(buf.Bytes())
	return json.Marshal(str)
}

func (m *BTCHeaderBytes) UnmarshalJSON(bz []byte) error {
	var headerHexStr string
	err := json.Unmarshal(bz, &headerHexStr)

	if err != nil {
		return err
	}

	decoded, err := hex.DecodeString(headerHexStr)
	if err != nil {
		return err
	}
	*m = decoded
	return nil
}

func (m BTCHeaderBytes) Marshal() ([]byte, error) {
	return m, nil
}

func (m *BTCHeaderBytes) Unmarshal(data []byte) error {
	buf := make([]byte, len(data))
	copy(buf, data)
	*m = buf
	return nil
}

func (m *BTCHeaderBytes) MarshalHex() (string, error) {
	btcdHeader, err := m.ToBtcdHeader()
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

	// Copy the bytes into the instance
	err = m.Unmarshal(decoded)
	if err != nil {
		return err
	}
	return nil
}

func (m BTCHeaderBytes) MarshalTo(data []byte) (int, error) {
	copy(data, m)
	return len(data), nil
}

func (m *BTCHeaderBytes) Size() int {
	bz, _ := m.Marshal()
	return len(bz)
}

// ToBtcdHeader parse header bytes into a BlockHeader instance
func (m BTCHeaderBytes) ToBtcdHeader() (*wire.BlockHeader, error) {
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

// BtcdHeaderToHeaderBytes gets a BlockHeader instance and returns the header bytes
func BtcdHeaderToHeaderBytes(header *wire.BlockHeader) BTCHeaderBytes {
	var buf bytes.Buffer
	header.Serialize(&buf)

	headerBytes := buf.Bytes()
	return headerBytes
}
