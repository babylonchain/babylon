package keeper

import (
	"encoding/hex"
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	tmcrypto "github.com/tendermint/tendermint/proto/tendermint/crypto"
)

// QueryStore queries a KV pair in the KVStore, where
// - moduleStoreKey is the store key of a module, e.g., zctypes.StoreKey
// - key is the key of the queried KV pair, including the prefix, e.g., zctypes.EpochChainInfoKey || chainID in the chain info store
// and returns
// - key of this KV pair
// - value of this KV pair
// - Merkle proof of this KV pair
// - error
// (adapted from https://github.com/cosmos/cosmos-sdk/blob/v0.46.6/client/query.go)
func (k Keeper) QueryStore(ctx sdk.Context, moduleStoreKey string, key []byte, queryHeight int64) ([]byte, []byte, *tmcrypto.ProofOps, error) {
	// construct the query path for ABCI query
	// path := fmt.Sprintf("/store/%s/key", moduleStoreKey) // e.g., "/store/epoching/key"
	path := fmt.Sprintf("/%s/key", moduleStoreKey) // e.g., "/store/epoching/key"

	// query the KV with Merkle proof
	resp := k.storeQuerier.Query(abci.RequestQuery{
		Path:   path,
		Data:   key,
		Height: queryHeight,
		Prove:  true,
	})
	if resp.Code != 0 {
		return nil, nil, nil, fmt.Errorf("query (with path %s) failed with response: %v", path, resp)
	}

	return resp.Key, resp.Value, resp.ProofOps, nil
}

// VerifyStore verifies whether a KV pair is committed to the Merkle root, with the assistance of a Merkle proof
// (adapted from https://github.com/cosmos/cosmos-sdk/blob/v0.46.6/store/rootmulti/proof_test.go)
func VerifyStore(root []byte, key []byte, value []byte, proof *tmcrypto.ProofOps) error {
	prt := rootmulti.DefaultProofRuntime()
	return prt.VerifyValue(proof, root, "/"+hex.EncodeToString(key), value)
}
