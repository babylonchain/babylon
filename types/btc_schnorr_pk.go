package types

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
)

type BIP340PubKey []byte

const BIP340PubKeyLen = schnorr.PubKeyBytesLen

func NewBIP340PubKey(data []byte) (*BIP340PubKey, error) {
	var pk BIP340PubKey
	err := pk.Unmarshal(data)
	return &pk, err
}

func NewBIP340PubKeyFromHex(hexStr string) (*BIP340PubKey, error) {
	var pk BIP340PubKey
	err := pk.UnmarshalHex(hexStr)
	return &pk, err
}

func NewBIP340PubKeyFromBTCPK(btcPK *btcec.PublicKey) *BIP340PubKey {
	pkBytes := schnorr.SerializePubKey(btcPK)
	pk := BIP340PubKey(pkBytes)
	return &pk
}

func (pk BIP340PubKey) ToBTCPK() (*btcec.PublicKey, error) {
	return schnorr.ParsePubKey(pk)
}

func (pk BIP340PubKey) MustToBTCPK() *btcec.PublicKey {
	btcPK, err := schnorr.ParsePubKey(pk)
	if err != nil {
		panic(err)
	}
	return btcPK
}

func (pk *BIP340PubKey) MarshalHex() string {
	return hex.EncodeToString(pk.MustMarshal())
}

func (pk *BIP340PubKey) UnmarshalHex(header string) error {
	// Decode the hash string from hex
	decoded, err := hex.DecodeString(header)
	if err != nil {
		return err
	}

	return pk.Unmarshal(decoded)
}

func (pk BIP340PubKey) Size() int {
	return len(pk.MustMarshal())
}

func (pk BIP340PubKey) Marshal() ([]byte, error) {
	return pk, nil
}

func (pk BIP340PubKey) MustMarshal() []byte {
	pkBytes, err := pk.Marshal()
	if err != nil {
		panic(err)
	}
	return pkBytes
}

func (pk BIP340PubKey) MarshalTo(data []byte) (int, error) {
	bz, err := pk.Marshal()
	if err != nil {
		return 0, err
	}
	copy(data, bz)
	return len(data), nil
}

func (pk *BIP340PubKey) Unmarshal(data []byte) error {
	if len(data) != BIP340PubKeyLen {
		return fmt.Errorf("malformed data for BIP340 public key")
	}

	*pk = data
	return nil
}

func (pk BIP340PubKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(pk.MarshalHex())
}

func (pk *BIP340PubKey) UnmarshalJSON(bz []byte) error {
	var pkHexString string
	err := json.Unmarshal(bz, &pkHexString)

	if err != nil {
		return err
	}

	return pk.UnmarshalHex(pkHexString)
}

func (pk *BIP340PubKey) Equals(pk2 *BIP340PubKey) bool {
	return bytes.Equal(*pk, *pk2)
}

func NewBTCPKsFromBIP340PKs(pks []BIP340PubKey) ([]*btcec.PublicKey, error) {
	btcPks := make([]*btcec.PublicKey, len(pks))
	for i, pk := range pks {
		btcPK, err := pk.ToBTCPK()
		if err != nil {
			return nil, err
		}
		btcPks[i] = btcPK
	}
	return btcPks, nil
}

func NewBIP340PKsFromBTCPKs(btcPKs []*btcec.PublicKey) []BIP340PubKey {
	pks := []BIP340PubKey{}
	for _, btcPK := range btcPKs {
		pks = append(pks, *NewBIP340PubKeyFromBTCPK(btcPK))
	}
	return pks
}

func SortBIP340PKs(keys []BIP340PubKey) []BIP340PubKey {
	sortedPKs := make([]BIP340PubKey, len(keys))
	copy(sortedPKs, keys)
	sort.SliceStable(sortedPKs, func(i, j int) bool {
		keyIBytes := sortedPKs[i].MustMarshal()
		keyJBytes := sortedPKs[j].MustMarshal()
		return bytes.Compare(keyIBytes, keyJBytes) == 1
	})
	return sortedPKs
}
