package types

import (
	"encoding/json"
	"errors"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

type BTCHeaderHashBytes []byte

const BTCHeaderHashLen = 32

func NewBTCHeaderHashBytesFromHex(hex string) (BTCHeaderHashBytes, error) {
	var hashBytes BTCHeaderHashBytes
	err := hashBytes.UnmarshalHex(hex)
	if err != nil {
		return nil, err
	}
	return hashBytes, nil
}

func NewBTCHeaderHashBytesFromChainhash(chHash *chainhash.Hash) BTCHeaderHashBytes {
	var headerHashBytes BTCHeaderHashBytes
	headerHashBytes.FromChainhash(chHash)
	return headerHashBytes
}

func NewBTCHeaderHashBytesFromBytes(hash []byte) (BTCHeaderHashBytes, error) {
	var headerHashBytes BTCHeaderHashBytes
	err := headerHashBytes.Unmarshal(hash)
	if err != nil {
		return nil, err
	}
	return headerHashBytes, nil
}

func (m BTCHeaderHashBytes) MarshalJSON() ([]byte, error) {
	// Marshal the JSON from hex format
	return json.Marshal(m.MarshalHex())
}

func (m *BTCHeaderHashBytes) UnmarshalJSON(bz []byte) error {
	var headerHashStr string
	err := json.Unmarshal(bz, &headerHashStr)
	if err != nil {
		return err
	}

	return m.UnmarshalHex(headerHashStr)
}

func (m BTCHeaderHashBytes) Marshal() ([]byte, error) {
	// Just return the bytes
	return m, nil
}

func (m BTCHeaderHashBytes) MustMarshal() []byte {
	bz, err := m.Marshal()
	if err != nil {
		panic("Marshalling failed")
	}
	return bz
}

func (m *BTCHeaderHashBytes) Unmarshal(bz []byte) error {
	if len(bz) != BTCHeaderHashLen {
		return errors.New("invalid header hash length")
	}
	// Verify that the bytes can be transformed to a *chainhash.Hash object
	_, err := toChainhash(bz)
	if err != nil {
		return errors.New("bytes do not correspond to *chainhash.Hash object")
	}
	*m = bz
	return nil
}

func (m *BTCHeaderHashBytes) MarshalHex() string {
	return m.ToChainhash().String()
}

func (m *BTCHeaderHashBytes) UnmarshalHex(hash string) error {
	if len(hash) != BTCHeaderHashLen*2 {
		return errors.New("invalid hex length")
	}
	decoded, err := chainhash.NewHashFromStr(hash)
	if err != nil {
		return err
	}

	// Copy the bytes into the instance
	return m.Unmarshal(decoded[:])
}

func (m BTCHeaderHashBytes) MarshalTo(data []byte) (int, error) {
	bz, err := m.Marshal()
	if err != nil {
		return 0, err
	}
	copy(data, bz)
	return len(data), nil
}

func (m *BTCHeaderHashBytes) Size() int {
	bz, _ := m.Marshal()
	return len(bz)
}

func (m BTCHeaderHashBytes) ToChainhash() *chainhash.Hash {
	chHash, err := toChainhash(m)
	if err != nil {
		panic("BTCHeaderHashBytes cannot be converted to chainhash")
	}
	return chHash
}

func (m *BTCHeaderHashBytes) FromChainhash(hash *chainhash.Hash) {
	err := m.Unmarshal(hash[:])
	if err != nil {
		panic("*chainhash.Hash bytes cannot be unmarshalled")
	}
}

func (m *BTCHeaderHashBytes) String() string {
	return m.ToChainhash().String()
}

func (m *BTCHeaderHashBytes) Eq(hash *BTCHeaderHashBytes) bool {
	return m.String() == hash.String()
}

func toChainhash(data []byte) (*chainhash.Hash, error) {
	return chainhash.NewHash(data)
}
