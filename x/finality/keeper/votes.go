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

func (k Keeper) SetSig(ctx context.Context, height uint64, fpBtcPK *bbn.BIP340PubKey, sig *bbn.SchnorrEOTSSig) {
	store := k.voteHeightStore(ctx, height)
	store.Set(fpBtcPK.MustMarshal(), sig.MustMarshal())
}

func (k Keeper) HasSig(ctx context.Context, height uint64, fpBtcPK *bbn.BIP340PubKey) bool {
	store := k.voteHeightStore(ctx, height)
	return store.Has(fpBtcPK.MustMarshal())
}

func (k Keeper) GetSig(ctx context.Context, height uint64, fpBtcPK *bbn.BIP340PubKey) (*bbn.SchnorrEOTSSig, error) {
	if uint64(sdk.UnwrapSDKContext(ctx).HeaderInfo().Height) < height {
		return nil, types.ErrHeightTooHigh
	}
	store := k.voteHeightStore(ctx, height)
	sigBytes := store.Get(fpBtcPK.MustMarshal())
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
	store := k.voteHeightStore(ctx, height)
	iter := store.Iterator(nil, nil)
	defer iter.Close()

	// if there is no vote on this height, return nil
	if !iter.Valid() {
		return nil
	}

	sigs := map[string]*bbn.SchnorrEOTSSig{}
	for ; iter.Valid(); iter.Next() {
		fpBTCPK, err := bbn.NewBIP340PubKey(iter.Key())
		if err != nil {
			// failing to unmarshal finality provider's BTC PK in KVStore is a programming error
			panic(fmt.Errorf("%w: %w", bbn.ErrUnmarshal, err))
		}
		sig, err := bbn.NewSchnorrEOTSSig(iter.Value())
		if err != nil {
			// failing to unmarshal EOTS sig in KVStore is a programming error
			panic(fmt.Errorf("failed to unmarshal EOTS signature: %w", err))
		}
		sigs[fpBTCPK.MarshalHex()] = sig
	}
	return sigs
}

// GetVoters gets returns a map of voters' BTC PKs to the given height
func (k Keeper) GetVoters(ctx context.Context, height uint64) map[string]struct{} {
	store := k.voteHeightStore(ctx, height)
	iter := store.Iterator(nil, nil)
	defer iter.Close()

	// if there is no vote on this height, return nil
	if !iter.Valid() {
		return nil
	}

	voterBTCPKs := map[string]struct{}{}
	for ; iter.Valid(); iter.Next() {
		// accumulate voterBTCPKs
		fpBTCPK, err := bbn.NewBIP340PubKey(iter.Key())
		if err != nil {
			// failing to unmarshal finality provider's BTC PK in KVStore is a programming error
			panic(fmt.Errorf("%w: %w", bbn.ErrUnmarshal, err))
		}
		voterBTCPKs[fpBTCPK.MarshalHex()] = struct{}{}
	}
	return voterBTCPKs
}

// voteHeightStore returns the KVStore of the votes
// prefix: VoteKey
// key: (block height || finality provider PK)
// value: EOTS sig
func (k Keeper) voteHeightStore(ctx context.Context, height uint64) prefix.Store {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	prefixedStore := prefix.NewStore(storeAdapter, types.VoteKey)
	return prefix.NewStore(prefixedStore, sdk.Uint64ToBigEndian(height))
}

// // voteHeightStore returns the KVStore of the votes
// // prefix: VoteKey
// // key: (block height || finality provider PK)
// // value: EOTS sig
// func (k Keeper) voteHeightStore(ctx context.Context, height uint64) prefix.Store {
// 	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
// 	prefixedStore := prefix.NewStore(storeAdapter, types.VoteKey)
// 	return prefix.NewStore(prefixedStore, sdk.Uint64ToBigEndian(height))
// }
