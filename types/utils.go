package types

import (
	"fmt"
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func Reverse(s interface{}) {
	n := reflect.ValueOf(s).Len()
	swap := reflect.Swapper(s)
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}
}

func CheckForDuplicatesAndEmptyStrings(input []string) error {
	encountered := map[string]bool{}
	for _, str := range input {
		if len(str) == 0 {
			return fmt.Errorf("empty string is not allowed")
		}

		if encountered[str] {
			return fmt.Errorf("duplicate entry found: %s", str)
		}

		encountered[str] = true
	}

	return nil
}

// ParseBlkHeightAndPubKeyFromStoreKey expects to receive a key with
// BigEndianUint64(blkHeight) || BIP340PubKey(fpBTCPK)
func ParseBlkHeightAndPubKeyFromStoreKey(key []byte) (blkHeight uint64, fpBTCPK *BIP340PubKey, err error) {
	sizeBigEndian := 8
	if len(key) < sizeBigEndian+1 {
		return 0, nil, fmt.Errorf("key not long enough to parse block height and BIP340PubKey: %s", key)
	}

	fpBTCPK, err = NewBIP340PubKey(key[sizeBigEndian:])
	if err != nil {
		return 0, nil, fmt.Errorf("failed to parse pub key from key %w: %w", ErrUnmarshal, err)
	}

	blkHeight = sdk.BigEndianToUint64(key[:sizeBigEndian])
	return blkHeight, fpBTCPK, nil
}

// ParseUintsFromStoreKey expects to receive a key with
// BigEndianUint64(blkHeight) || BigEndianUint64(Idx)
func ParseUintsFromStoreKey(key []byte) (blkHeight, idx uint64, err error) {
	sizeBigEndian := 8
	if len(key) < sizeBigEndian*2 {
		return 0, 0, fmt.Errorf("key not long enough to parse two uint64: %s", key)
	}

	return sdk.BigEndianToUint64(key[:sizeBigEndian]), sdk.BigEndianToUint64(key[sizeBigEndian:]), nil
}

// ParseBIP340PubKeysFromStoreKey expects to receive a key with
// BIP340PubKey(fpBTCPK) || BIP340PubKey(delBTCPK)
func ParseBIP340PubKeysFromStoreKey(key []byte) (fpBTCPK, delBTCPK *BIP340PubKey, err error) {
	if len(key) < BIP340PubKeyLen*2 {
		return nil, nil, fmt.Errorf("key not long enough to parse two BIP340PubKey: %s", key)
	}

	fpBTCPK, err = NewBIP340PubKey(key[:BIP340PubKeyLen])
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse pub key from key %w: %w", ErrUnmarshal, err)
	}

	delBTCPK, err = NewBIP340PubKey(key[BIP340PubKeyLen:])
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse pub key from key %w: %w", ErrUnmarshal, err)
	}

	return fpBTCPK, delBTCPK, nil
}
