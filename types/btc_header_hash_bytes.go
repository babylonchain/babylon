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

func NewBTCHeaderHashBytesFromChainhash(chHash *chainhash.Hash) (BTCHeaderHashBytes, error) {
	var headerHashBytes BTCHeaderHashBytes
	err := headerHashBytes.Unmarshal(chHash[:])
	if err != nil {
		return nil, err
	}
	return headerHashBytes, nil
}

func NewBTCHeaderHashBytesFromBytes(hash []byte) BTCHeaderHashBytes {
	return hash
}

func (m BTCHeaderHashBytes) MarshalJSON() ([]byte, error) {
	hex, err := m.MarshalHex()
	if err != nil {
		return nil, err
	}
	// Marshal the JSON from hex format
	return json.Marshal(hex)
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

func (m *BTCHeaderHashBytes) Unmarshal(bz []byte) error {
	if len(bz) != BTCHeaderHashLen {
		return errors.New("invalid header hash length")
	}
	*m = bz
	return nil
}

func (m *BTCHeaderHashBytes) MarshalHex() (string, error) {
	chHash, err := m.ToChainhash()
	if err != nil {
		return "", err
	}

	return chHash.String(), nil
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
	copy(data, m)
	return len(data), nil
}

func (m *BTCHeaderHashBytes) Size() int {
	bz, _ := m.Marshal()
	return len(bz)
}

func (m BTCHeaderHashBytes) ToChainhash() (*chainhash.Hash, error) {
	return chainhash.NewHash(m)
}

func (m *BTCHeaderHashBytes) FromChainhash(hash *chainhash.Hash) error {
	return m.Unmarshal(hash[:])
}
