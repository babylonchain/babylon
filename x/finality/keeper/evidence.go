package keeper

import (
	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/finality/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) SetEvidence(ctx sdk.Context, evidence *types.Evidence) {
	store := k.evidenceStore(ctx, evidence.ValBtcPk)
	store.Set(sdk.Uint64ToBigEndian(evidence.BlockHeight), k.cdc.MustMarshal(evidence))
}

func (k Keeper) HasEvidence(ctx sdk.Context, valBtcPK *bbn.BIP340PubKey, height uint64) bool {
	store := k.evidenceStore(ctx, valBtcPK)
	return store.Has(valBtcPK.MustMarshal())
}

func (k Keeper) GetEvidence(ctx sdk.Context, valBtcPK *bbn.BIP340PubKey, height uint64) (*types.Evidence, error) {
	if uint64(ctx.BlockHeight()) < height {
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
func (k Keeper) GetFirstSlashableEvidence(ctx sdk.Context, valBtcPK *bbn.BIP340PubKey) *types.Evidence {
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
func (k Keeper) evidenceStore(ctx sdk.Context, valBTCPK *bbn.BIP340PubKey) prefix.Store {
	store := ctx.KVStore(k.storeKey)
	eStore := prefix.NewStore(store, types.EvidenceKey)
	return prefix.NewStore(eStore, valBTCPK.MustMarshal())
}
