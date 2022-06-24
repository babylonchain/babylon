package types

import (
	"encoding/hex"
	"encoding/json"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

type BTCHeaderHashBytes []byte
type BTCHeaderHashesBytes []BTCHeaderHashBytes

func (m BTCHeaderHashBytes) MarshalJSON() ([]byte, error) {
	// Get the chainhash representation
	chHash, err := m.MarshalChainhash()
	if err != nil {
		return nil, err
	}
	// Marshal the JSON from hex format
	return json.Marshal(chHash.String())
}

func (m *BTCHeaderHashBytes) UnmarshalJSON(bz []byte) error {
	var headerHashStr string
	err := json.Unmarshal(bz, &headerHashStr)
	if err != nil {
		return err
	}

	decoded, err := chainhash.NewHashFromStr(headerHashStr)
	if err != nil {
		return err
	}
	*m = decoded[:]
	return nil
}

func (m BTCHeaderHashBytes) Marshal() ([]byte, error) {
	// Just return the bytes
	return m, nil
}

func (m *BTCHeaderHashBytes) Unmarshal(bz []byte) error {
	// The size of BTCHeaderHashBytes should be chainhash.HashSize
	buf := make([]byte, chainhash.HashSize)
	copy(buf, bz)
	*m = buf
	return nil
}

func (m *BTCHeaderHashBytes) MarshalHex() (string, error) {
	chHash, err := m.MarshalChainhash()
	if err != nil {
		return "", err
	}

	return chHash.String(), nil
}

func (m *BTCHeaderHashBytes) UnmarshalHex(hash string) error {
	// Decode the hash string from hex
	decoded, err := hex.DecodeString(hash)
	if err != nil {
		return err
	}

	// Copy the bytes into the instance
	err = m.Unmarshal(decoded)
	if err != nil {
		return err
	}
	// Our internal representation of bytes involves a reverse
	// form from the bytes represented by hex
	// This is also the internal representation used by chainhash.Hash
	m.reverse()
	return nil
}

func (m BTCHeaderHashBytes) MarshalTo(data []byte) (int, error) {
	copy(data, m)
	return len(data), nil
}

func (m *BTCHeaderHashBytes) Size() int {
	bz, _ := m.Marshal()
	return len(bz)
}

func (m BTCHeaderHashBytes) MarshalChainhash() (*chainhash.Hash, error) {
	return chainhash.NewHash(m)
}

func (m *BTCHeaderHashBytes) UnmarshalChainhash(hash *chainhash.Hash) {
	var headerHashBytes BTCHeaderHashBytes
	headerHashBytes.Unmarshal(hash[:])
	*m = headerHashBytes
}

func (m BTCHeaderHashBytes) reverse() {
	for i := 0; i < chainhash.HashSize/2; i++ {
		m[i], m[chainhash.HashSize-1-i] = m[chainhash.HashSize-1-i], m[i]
	}
}
