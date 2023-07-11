package keeper

import (
	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/finality/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) SetEvidence(ctx sdk.Context, evidence *types.Evidence) {
	store := k.evidenceStore(ctx, evidence.BlockHeight)
	store.Set(evidence.ValBtcPk.MustMarshal(), k.cdc.MustMarshal(evidence))
}

func (k Keeper) HasEvidence(ctx sdk.Context, height uint64, valBtcPK *bbn.BIP340PubKey) bool {
	store := k.evidenceStore(ctx, height)
	return store.Has(valBtcPK.MustMarshal())
}

func (k Keeper) GetEvidence(ctx sdk.Context, height uint64, valBtcPK *bbn.BIP340PubKey) (*types.Evidence, error) {
	if uint64(ctx.BlockHeight()) < height {
		return nil, types.ErrHeightTooHigh
	}
	store := k.evidenceStore(ctx, height)
	evidenceBytes := store.Get(valBtcPK.MustMarshal())
	if len(evidenceBytes) == 0 {
		return nil, types.ErrEvidenceNotFound
	}
	var evidence types.Evidence
	k.cdc.MustUnmarshal(evidenceBytes, &evidence)
	return &evidence, nil
}

// evidenceStore returns the KVStore of the evidences
// prefix: EvidenceKey
// key: (block height || BTC validator PK)
// value: Evidence
func (k Keeper) evidenceStore(ctx sdk.Context, height uint64) prefix.Store {
	store := ctx.KVStore(k.storeKey)
	prefixedStore := prefix.NewStore(store, types.EvidenceKey)
	return prefix.NewStore(prefixedStore, sdk.Uint64ToBigEndian(height))
}
