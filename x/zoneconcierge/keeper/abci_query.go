package keeper

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	tmcrypto "github.com/tendermint/tendermint/proto/tendermint/crypto"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
)

// queryStore queries a KV pair in the KVStore, where
// - moduleStoreKey is the store key of a module, e.g., zctypes.StoreKey
// - key is the key of the queried KV pair, including the prefix, e.g., zctypes.EpochChainInfoKey || chainID in the chain info store
// and returns
// - key of this KV pair
// - value of this KV pair
// - Merkle proof of this KV pair
// - error
// (adapted from https://github.com/cosmos/cosmos-sdk/blob/v0.46.6/client/query.go)
func (k Keeper) queryStore(moduleStoreKey string, key []byte, queryHeight int64) ([]byte, []byte, *tmcrypto.ProofOps, error) {
	prefix := fmt.Sprintf("/store/%s/key", moduleStoreKey) // path of the entry in KVStore
	opts := rpcclient.ABCIQueryOptions{
		Height: queryHeight,
		Prove:  true,
	}
	resp, err := k.tmClient.ABCIQueryWithOptions(context.Background(), prefix, key, opts)
	return resp.Response.Key, resp.Response.Value, resp.Response.ProofOps, err
}

// verifyStore verifies whether a KV pair is committed to the Merkle root, with the assistance of a Merkle proof
// (adapted from https://github.com/cosmos/cosmos-sdk/blob/v0.46.6/store/rootmulti/proof_test.go)
func verifyStore(root []byte, keyWithPrefix []byte, value []byte, proof *tmcrypto.ProofOps) error {
	prt := rootmulti.DefaultProofRuntime()
	return prt.VerifyValue(proof, root, string(keyWithPrefix), value)
}
