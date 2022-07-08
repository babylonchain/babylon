package types_test

import (
	"bytes"
	"github.com/babylonchain/babylon/testutil/datagen"
	bbl "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/cosmos/cosmos-sdk/types/query"
	"math/rand"
	"testing"
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
	headerBytes := bbl.GetBaseHeaderBytes()
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
	f.Add("00000000000000000002bf1c218853bc920f41f74491e6c92c6bc6fdc881ab47", int64(17))
	f.Fuzz(func(t *testing.T, hexHash string, seed int64) {
		rand.Seed(seed)
		if !datagen.ValidHex(hexHash, bbl.BTCHeaderHashLen) {
			hexHash = datagen.GenRandomHexStr(bbl.BTCHeaderHashLen)
		}
		chHash, _ := chainhash.NewHashFromStr(hexHash)

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
		if bytes.Compare(*(queryContains.Hash), chHash[:]) != 0 {
			t.Errorf("expected hash bytes %s got %s", chHash[:], *(queryContains.Hash))
		}
	})
}

func TestNewQueryMainChainRequest(t *testing.T) {
	headerBytes := bbl.GetBaseHeaderBytes()
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
