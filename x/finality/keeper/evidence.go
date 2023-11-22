package keeper

import (
	"context"
	"cosmossdk.io/store/prefix"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/finality/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) SetEvidence(ctx context.Context, evidence *types.Evidence) {
	store := k.evidenceStore(ctx, evidence.ValBtcPk)
	store.Set(sdk.Uint64ToBigEndian(evidence.BlockHeight), k.cdc.MustMarshal(evidence))
}

func (k Keeper) HasEvidence(ctx context.Context, valBtcPK *bbn.BIP340PubKey, height uint64) bool {
	store := k.evidenceStore(ctx, valBtcPK)
	return store.Has(valBtcPK.MustMarshal())
}

func (k Keeper) GetEvidence(ctx context.Context, valBtcPK *bbn.BIP340PubKey, height uint64) (*types.Evidence, error) {
	if uint64(sdk.UnwrapSDKContext(ctx).BlockHeight()) < height {
		return nil, types.ErrHeightTooHigh
	}
	store := k.evidenceStore(ctx, valBtcPK)
	evidenceBytes := store.Get(sdk.Uint64ToBigEndian(height))
	if len(evidenceBytes) == 0 {
		return nil, types.ErrEvidenceNotFound
	}
	var evidence types.Evidence
	k.cdc.MustUnmarshal(evidenceBytes, &evidence)
	return &evidence, nil
}

// GetFirstSlashableEvidence gets the first evidence that is slashable,
// i.e., it contains all fields.
// NOTE: it's possible that the CanonicalFinalitySig field is empty for
// an evidence, which happens when the BTC validator signed a fork block
// but hasn't signed the canonical block yet.
func (k Keeper) GetFirstSlashableEvidence(ctx context.Context, valBtcPK *bbn.BIP340PubKey) *types.Evidence {
	store := k.evidenceStore(ctx, valBtcPK)
	iter := store.Iterator(nil, nil)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		evidenceBytes := iter.Value()
		var evidence types.Evidence
		k.cdc.MustUnmarshal(evidenceBytes, &evidence)
		if evidence.IsSlashable() {
			return &evidence
		}
	}
	return nil
}

// evidenceStore returns the KVStore of the evidences
// prefix: EvidenceKey
// key: (BTC validator PK || height)
// value: Evidence
func (k Keeper) evidenceStore(ctx context.Context, valBTCPK *bbn.BIP340PubKey) prefix.Store {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	eStore := prefix.NewStore(storeAdapter, types.EvidenceKey)
	return prefix.NewStore(eStore, valBTCPK.MustMarshal())
}
