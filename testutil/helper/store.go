package helper

import (
	"bytes"
	"fmt"
	"sync"

	corestore "cosmossdk.io/core/store"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/types/kv"
)

// DiffKVStores compares two KVstores and returns all the key/value pairs
// that differ from one another. It also skips value comparison for a set of provided prefixes.
func DiffKVStores(a, b corestore.KVStore, prefixesToSkip [][]byte) (diffA, diffB []kv.Pair) {
	iterA, err := a.Iterator(nil, nil)
	if err != nil {
		panic(err)
	}
	defer iterA.Close()

	iterB, err := b.Iterator(nil, nil)
	if err != nil {
		panic(err)
	}
	defer iterB.Close()

	var wg sync.WaitGroup

	wg.Add(1)
	kvAs := make([]kv.Pair, 0)
	go func() {
		defer wg.Done()
		kvAs = getKVPairs(iterA, prefixesToSkip)
	}()

	wg.Add(1)
	kvBs := make([]kv.Pair, 0)
	go func() {
		defer wg.Done()
		kvBs = getKVPairs(iterB, prefixesToSkip)
	}()

	wg.Wait()

	if len(kvAs) != len(kvBs) {
		fmt.Printf("KV stores are different: %d key/value pairs in store A and %d key/value pairs in store B\n", len(kvAs), len(kvBs))
	}

	return getDiffFromKVPair(kvAs, kvBs)
}

// getDiffFromKVPair compares two KVstores and returns all the key/value pairs
func getDiffFromKVPair(kvAs, kvBs []kv.Pair) (diffA, diffB []kv.Pair) {
	// we assume that kvBs is equal or larger than kvAs
	// if not, we swap the two
	if len(kvAs) > len(kvBs) {
		kvAs, kvBs = kvBs, kvAs
		// we need to swap the diffA and diffB as well
		defer func() {
			diffA, diffB = diffB, diffA
		}()
	}

	// in case kvAs is empty we can return early
	// since there is nothing to compare
	// if kvAs == kvBs, then diffA and diffB will be empty
	if len(kvAs) == 0 {
		return []kv.Pair{}, kvBs
	}

	index := make(map[string][]byte, len(kvBs))
	for _, kv := range kvBs {
		index[string(kv.Key)] = kv.Value
	}

	for _, kvA := range kvAs {
		if kvBValue, ok := index[string(kvA.Key)]; !ok {
			diffA = append(diffA, kvA)
			diffB = append(diffB, kv.Pair{Key: kvA.Key}) // the key is missing from kvB so we append a pair with an empty value
		} else if !bytes.Equal(kvA.Value, kvBValue) {
			diffA = append(diffA, kvA)
			diffB = append(diffB, kv.Pair{Key: kvA.Key, Value: kvBValue})
		} else {
			// values are equal, so we remove the key from the index
			delete(index, string(kvA.Key))
		}
	}

	// add the remaining keys from kvBs
	for key, value := range index {
		diffA = append(diffA, kv.Pair{Key: []byte(key)}) // the key is missing from kvA so we append a pair with an empty value
		diffB = append(diffB, kv.Pair{Key: []byte(key), Value: value})
	}

	return diffA, diffB
}

func getKVPairs(iter dbm.Iterator, prefixesToSkip [][]byte) (kvs []kv.Pair) {
	for iter.Valid() {
		key, value := iter.Key(), iter.Value()

		// do not add the KV pair if the key is prefixed to be skipped.
		skip := false
		for _, prefix := range prefixesToSkip {
			if bytes.HasPrefix(key, prefix) {
				skip = true
				break
			}
		}

		if !skip {
			kvs = append(kvs, kv.Pair{Key: key, Value: value})
		}

		iter.Next()
	}

	return kvs
}
