package keeper

import (
	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/finality/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) AddVote(ctx sdk.Context, vote *types.Vote) error {
	// TODO verification rules of vote
	k.setVote(ctx, vote)
	return nil
}

func (k Keeper) setVote(ctx sdk.Context, vote *types.Vote) {
	store := k.voteStore(ctx, vote.BlockHeight)
	voteBytes := k.cdc.MustMarshal(vote)
	store.Set(sdk.Uint64ToBigEndian(vote.BlockHeight), voteBytes)
}

func (k Keeper) HasVote(ctx sdk.Context, height uint64, valBtcPK *bbn.BIP340PubKey) bool {
	store := k.voteStore(ctx, height)
	return store.Has(*valBtcPK)
}

func (k Keeper) GetVote(ctx sdk.Context, height uint64, valBtcPK *bbn.BIP340PubKey) (*types.Vote, error) {
	if uint64(ctx.BlockHeight()) < height {
		return nil, types.ErrHeightTooHigh
	}
	store := k.voteStore(ctx, height)
	voteBytes := store.Get(*valBtcPK)
	if len(voteBytes) == 0 {
		return nil, types.ErrVoteNotFound
	}
	var vote types.Vote
	k.cdc.MustUnmarshal(voteBytes, &vote)
	return &vote, nil
}

// voteStore returns the KVStore of the votes
// prefix: VoteKey
// key: (block height || BTC validator PK)
// value: Vote
func (k Keeper) voteStore(ctx sdk.Context, height uint64) prefix.Store {
	store := ctx.KVStore(k.storeKey)
	prefixedStore := prefix.NewStore(store, types.VoteKey)
	return prefix.NewStore(prefixedStore, sdk.Uint64ToBigEndian(height))
}
