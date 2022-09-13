package types_test

import (
	"bytes"
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/cosmos/cosmos-sdk/types/query"
)

func TestNewQueryParamsRequest(t *testing.T) {
	newQueryParams := types.NewQueryParamsRequest()
	if newQueryParams == nil {
		t.Errorf("A nil object was returned")
	}

	emptyQueryParams := types.QueryParamsRequest{}
	if *newQueryParams != emptyQueryParams {
		t.Errorf("expected an empty QueryParamsRequest")
	}
}

func TestNewQueryHashesRequest(t *testing.T) {
	headerBytes := bbn.GetBaseBTCHeaderBytes()
	headerHashBytes := headerBytes.Hash()
	req := query.PageRequest{
		Key: headerHashBytes.MustMarshal(),
	}
	newQueryHashes := types.NewQueryHashesRequest(&req)
	if newQueryHashes == nil {
		t.Errorf("A nil object was returned")
	}

	expectedQueryHashes := types.QueryHashesRequest{
		Pagination: &req,
	}
	if *newQueryHashes != expectedQueryHashes {
		t.Errorf("expected a QueryHashesRequest %s", expectedQueryHashes)
	}
}

func FuzzNewQueryContainsRequest(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 100)
	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)
		hexHash := datagen.GenRandomHexStr(bbn.BTCHeaderHashLen)

		btcHeaderHashBytes, _ := bbn.NewBTCHeaderHashBytesFromHex(hexHash)

		queryContains, err := types.NewQueryContainsRequest(hexHash)
		if err != nil {
			t.Errorf("returned error for valid hex %s", hexHash)
		}
		if queryContains == nil {
			t.Errorf("returned a nil reference to a query")
		}
		if queryContains.Hash == nil {
			t.Errorf("has an empty hash attribute")
		}
		if bytes.Compare(queryContains.Hash, btcHeaderHashBytes.MustMarshal()) != 0 {
			t.Errorf("expected hash bytes %s got %s", btcHeaderHashBytes.MustMarshal(), queryContains.Hash)
		}
	})
}

func TestNewQueryMainChainRequest(t *testing.T) {
	headerBytes := bbn.GetBaseBTCHeaderBytes()
	req := query.PageRequest{
		Key: headerBytes.MustMarshal(),
	}
	newQueryMainChain := types.NewQueryMainChainRequest(&req)
	if newQueryMainChain == nil {
		t.Errorf("A nil object was returned")
	}

	expectedQueryMainChain := types.QueryMainChainRequest{
		Pagination: &req,
	}
	if *newQueryMainChain != expectedQueryMainChain {
		t.Errorf("expected a QueryMainChainRequest %s", expectedQueryMainChain)
	}
}
