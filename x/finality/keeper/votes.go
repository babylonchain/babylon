package keeper

import (
	"fmt"

	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/finality/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

//nolint:unused
func (k Keeper) setSig(ctx sdk.Context, height uint64, valBtcPK *bbn.BIP340PubKey, sig *bbn.SchnorrEOTSSig) {
	store := k.voteStore(ctx, height)
	store.Set(valBtcPK.MustMarshal(), sig.MustMarshal())
}

func (k Keeper) HasSig(ctx sdk.Context, height uint64, valBtcPK *bbn.BIP340PubKey) bool {
	store := k.voteStore(ctx, height)
	return store.Has(valBtcPK.MustMarshal())
}

func (k Keeper) GetSig(ctx sdk.Context, height uint64, valBtcPK *bbn.BIP340PubKey) (*bbn.SchnorrEOTSSig, error) {
	if uint64(ctx.BlockHeight()) < height {
		return nil, types.ErrHeightTooHigh
	}
	store := k.voteStore(ctx, height)
	sigBytes := store.Get(valBtcPK.MustMarshal())
	if len(sigBytes) == 0 {
		return nil, types.ErrVoteNotFound
	}
	sig, err := bbn.NewSchnorrEOTSSig(sigBytes)
	if err != nil {
		panic(fmt.Errorf("failed to unmarshal EOTS signature: %w", err))
	}
	return sig, nil
}

// voteStore returns the KVStore of the votes
// prefix: VoteKey
// key: (block height || BTC validator PK)
// value: EOTS sig
func (k Keeper) voteStore(ctx sdk.Context, height uint64) prefix.Store {
	store := ctx.KVStore(k.storeKey)
	prefixedStore := prefix.NewStore(store, types.VoteKey)
	return prefix.NewStore(prefixedStore, sdk.Uint64ToBigEndian(height))
}
