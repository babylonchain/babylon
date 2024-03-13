package types

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"
	"time"

	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/wire"
)

type BTCHeaderBytes []byte

const BTCHeaderLen = 80

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

func NewBTCHeaderBytesFromBytes(header []byte) (BTCHeaderBytes, error) {
	var headerBytes BTCHeaderBytes
	err := headerBytes.Unmarshal(header)
	if err != nil {
		return nil, err
	}
	return headerBytes, nil
}

func (m BTCHeaderBytes) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.MarshalHex())
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

func (m BTCHeaderBytes) MustMarshal() []byte {
	bz, err := m.Marshal()
	if err != nil {
		panic("Marshalling failed")
	}
	return bz
}

func (m *BTCHeaderBytes) Unmarshal(data []byte) error {
	if len(data) != BTCHeaderLen {
		return errors.New("invalid header length")
	}
	// Verify that the bytes can be transformed to a *wire.BlockHeader object
	_, err := NewBlockHeader(data)
	if err != nil {
		return errors.New("bytes do not correspond to a *wire.BlockHeader object")
	}

	*m = data
	return nil
}

func (m BTCHeaderBytes) MarshalHex() string {
	btcdHeader := m.ToBlockHeader()

	var buf bytes.Buffer
	err := btcdHeader.Serialize(&buf)
	if err != nil {
		panic("Block header object cannot be converted to hex")
	}
	return hex.EncodeToString(buf.Bytes())
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
	bz, err := m.Marshal()
	if err != nil {
		return 0, err
	}
	copy(data, bz)
	return len(data), nil
}

func (m *BTCHeaderBytes) Size() int {
	bz, _ := m.Marshal()
	return len(bz)
}

func (m BTCHeaderBytes) ToBlockHeader() *wire.BlockHeader {
	header, err := NewBlockHeader(m)
	// There was a parsing error
	if err != nil {
		panic("BTCHeaderBytes cannot be converted to a block header object")
	}
	return header
}

func (m *BTCHeaderBytes) FromBlockHeader(header *wire.BlockHeader) {
	var buf bytes.Buffer
	err := header.Serialize(&buf)
	if err != nil {
		panic("*wire.BlockHeader cannot be serialized")
	}

	err = m.Unmarshal(buf.Bytes())
	if err != nil {
		panic("*wire.BlockHeader serialized bytes cannot be unmarshalled")
	}
}

func (m *BTCHeaderBytes) HasParent(header *BTCHeaderBytes) bool {
	current := m.ToBlockHeader()
	parent := header.ToBlockHeader()

	return current.PrevBlock.String() == parent.BlockHash().String()
}

func (m *BTCHeaderBytes) Eq(other *BTCHeaderBytes) bool {
	return m.Hash().Eq(other.Hash())
}

func (m *BTCHeaderBytes) Hash() *BTCHeaderHashBytes {
	blockHash := m.ToBlockHeader().BlockHash()
	hashBytes := NewBTCHeaderHashBytesFromChainhash(&blockHash)
	return &hashBytes
}

func (m *BTCHeaderBytes) ParentHash() *BTCHeaderHashBytes {
	parentHash := m.ToBlockHeader().PrevBlock
	hashBytes := NewBTCHeaderHashBytesFromChainhash(&parentHash)
	return &hashBytes
}

func (m *BTCHeaderBytes) Bits() uint32 {
	return m.ToBlockHeader().Bits
}

func (m *BTCHeaderBytes) Time() time.Time {
	return m.ToBlockHeader().Timestamp
}

func (m *BTCHeaderBytes) Difficulty() *big.Int {
	return blockchain.CompactToBig(m.Bits())
}

// NewBlockHeader creates a block header from bytes.
func NewBlockHeader(data []byte) (*wire.BlockHeader, error) {
	// Create an empty header
	header := &wire.BlockHeader{}

	// The Deserialize method expects an io.Reader instance
	reader := bytes.NewReader(data)
	// Decode the header bytes
	err := header.Deserialize(reader)
	return header, err
}
