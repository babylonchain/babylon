package keeper

import (
	storetypes "cosmossdk.io/store/types"
	"fmt"

	"cosmossdk.io/store/rootmulti"
	"github.com/cometbft/cometbft/crypto/merkle"
	tmcrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
)

// QueryStore queries a KV pair in the KVStore, where
// - moduleStoreKey is the store key of a module, e.g., zctypes.StoreKey
// - key is the key of the queried KV pair, including the prefix, e.g., zctypes.EpochChainInfoKey || chainID in the chain info store
// and returns
// - key of this KV pair
// - value of this KV pair
// - Merkle proof of this KV pair
// - error
// (adapted from https://github.com/cosmos/cosmos-sdk/blob/v0.46.6/baseapp/abci.go#L774-L795)
func (k Keeper) QueryStore(moduleStoreKey string, key []byte, queryHeight int64) ([]byte, []byte, *tmcrypto.ProofOps, error) {
	// construct the query path for ABCI query
	// since we are querying the DB directly, the path will not need prefix "/store" as done in ABCIQuery
	// Instead, it will be formed as "/<moduleStoreKey>/key", e.g., "/epoching/key"
	path := fmt.Sprintf("/%s/key", moduleStoreKey)

	// query the KV with Merkle proof
	resp, err := k.storeQuerier.Query(&storetypes.RequestQuery{
		Path:   path,
		Data:   key,
		Height: queryHeight - 1, // NOTE: the inclusion proof corresponds to the NEXT header
		Prove:  true,
	})
	if err != nil {
		return nil, nil, nil, err
	}
	if resp.Code != 0 {
		return nil, nil, nil, fmt.Errorf("query (with path %s) failed with response: %v", path, resp)
	}

	return resp.Key, resp.Value, resp.ProofOps, nil
}

// VerifyStore verifies whether a KV pair is committed to the Merkle root, with the assistance of a Merkle proof
// (adapted from https://github.com/cosmos/cosmos-sdk/blob/v0.46.6/store/rootmulti/proof_test.go)
func VerifyStore(root []byte, moduleStoreKey string, key []byte, value []byte, proof *tmcrypto.ProofOps) error {
	prt := rootmulti.DefaultProofRuntime()

	keypath := merkle.KeyPath{}
	keypath = keypath.AppendKey([]byte(moduleStoreKey), merkle.KeyEncodingURL)
	keypath = keypath.AppendKey(key, merkle.KeyEncodingURL)
	keypathStr := keypath.String()

	return prt.VerifyAbsence(proof, root, keypathStr) // TODO: verify value rather than just existence
}
