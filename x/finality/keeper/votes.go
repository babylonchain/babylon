package keeper

import (
	"context"
	"fmt"
	"github.com/cosmos/cosmos-sdk/runtime"

	"cosmossdk.io/store/prefix"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/finality/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) SetSig(ctx context.Context, height uint64, valBtcPK *bbn.BIP340PubKey, sig *bbn.SchnorrEOTSSig) {
	store := k.voteStore(ctx, height)
	store.Set(valBtcPK.MustMarshal(), sig.MustMarshal())
}

func (k Keeper) HasSig(ctx context.Context, height uint64, valBtcPK *bbn.BIP340PubKey) bool {
	store := k.voteStore(ctx, height)
	return store.Has(valBtcPK.MustMarshal())
}

func (k Keeper) GetSig(ctx context.Context, height uint64, valBtcPK *bbn.BIP340PubKey) (*bbn.SchnorrEOTSSig, error) {
	if uint64(sdk.UnwrapSDKContext(ctx).BlockHeight()) < height {
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

// GetSigSet gets all EOTS signatures at a given height
func (k Keeper) GetSigSet(ctx context.Context, height uint64) map[string]*bbn.SchnorrEOTSSig {
	store := k.voteStore(ctx, height)
	iter := store.Iterator(nil, nil)
	defer iter.Close()

	// if there is no vote on this height, return nil
	if !iter.Valid() {
		return nil
	}

	sigs := map[string]*bbn.SchnorrEOTSSig{}
	for ; iter.Valid(); iter.Next() {
		valBTCPK, err := bbn.NewBIP340PubKey(iter.Key())
		if err != nil {
			// failing to unmarshal validator BTC PK in KVStore is a programming error
			panic(fmt.Errorf("%w: %w", bbn.ErrUnmarshal, err))
		}
		sig, err := bbn.NewSchnorrEOTSSig(iter.Value())
		if err != nil {
			// failing to unmarshal EOTS sig in KVStore is a programming error
			panic(fmt.Errorf("failed to unmarshal EOTS signature: %w", err))
		}
		sigs[valBTCPK.MarshalHex()] = sig
	}
	return sigs
}

// GetVoters gets returns a map of voters' BTC PKs to the given height
func (k Keeper) GetVoters(ctx context.Context, height uint64) map[string]struct{} {
	store := k.voteStore(ctx, height)
	iter := store.Iterator(nil, nil)
	defer iter.Close()

	// if there is no vote on this height, return nil
	if !iter.Valid() {
		return nil
	}

	voterBTCPKs := map[string]struct{}{}
	for ; iter.Valid(); iter.Next() {
		// accumulate voterBTCPKs
		valBTCPK, err := bbn.NewBIP340PubKey(iter.Key())
		if err != nil {
			// failing to unmarshal validator BTC PK in KVStore is a programming error
			panic(fmt.Errorf("%w: %w", bbn.ErrUnmarshal, err))
		}
		voterBTCPKs[valBTCPK.MarshalHex()] = struct{}{}
	}
	return voterBTCPKs
}

// voteStore returns the KVStore of the votes
// prefix: VoteKey
// key: (block height || BTC validator PK)
// value: EOTS sig
func (k Keeper) voteStore(ctx context.Context, height uint64) prefix.Store {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	prefixedStore := prefix.NewStore(storeAdapter, types.VoteKey)
	return prefix.NewStore(prefixedStore, sdk.Uint64ToBigEndian(height))
}
