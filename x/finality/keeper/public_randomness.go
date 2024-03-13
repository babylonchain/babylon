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

// SetPubRandList sets a list of public randomness starting from a given startHeight
// for a given finality provider
func (k Keeper) SetPubRandList(ctx context.Context, fpBtcPK *bbn.BIP340PubKey, startHeight uint64, pubRandList []bbn.SchnorrPubRand) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	cacheCtx, writeCache := sdkCtx.CacheContext()

	// write to a KV store cache
	store := k.pubRandFpStore(cacheCtx, fpBtcPK)
	for i, pr := range pubRandList {
		height := startHeight + uint64(i)
		store.Set(sdk.Uint64ToBigEndian(height), pr)
	}

	// atomically write the new public randomness back to KV store
	writeCache()
}

func (k Keeper) HasPubRand(ctx context.Context, fpBtcPK *bbn.BIP340PubKey, height uint64) bool {
	store := k.pubRandFpStore(ctx, fpBtcPK)
	return store.Has(sdk.Uint64ToBigEndian(height))
}

func (k Keeper) GetPubRand(ctx context.Context, fpBtcPK *bbn.BIP340PubKey, height uint64) (*bbn.SchnorrPubRand, error) {
	store := k.pubRandFpStore(ctx, fpBtcPK)
	prBytes := store.Get(sdk.Uint64ToBigEndian(height))
	if len(prBytes) == 0 {
		return nil, types.ErrPubRandNotFound
	}
	return bbn.NewSchnorrPubRand(prBytes)
}

func (k Keeper) IsFirstPubRand(ctx context.Context, fpBtcPK *bbn.BIP340PubKey) bool {
	store := k.pubRandFpStore(ctx, fpBtcPK)
	iter := store.ReverseIterator(nil, nil)

	// if the iterator is not valid, then this finality provider does not commit any randomness
	return !iter.Valid()
}

// GetLastPubRand retrieves the last public randomness committed by the given finality provider
func (k Keeper) GetLastPubRand(ctx context.Context, fpBtcPK *bbn.BIP340PubKey) (uint64, *bbn.SchnorrPubRand, error) {
	store := k.pubRandFpStore(ctx, fpBtcPK)
	iter := store.ReverseIterator(nil, nil)

	if !iter.Valid() {
		// this finality provider does not commit any randomness
		return 0, nil, types.ErrNoPubRandYet
	}

	height := sdk.BigEndianToUint64(iter.Key())
	pubRand, err := bbn.NewSchnorrPubRand(iter.Value())
	if err != nil {
		// failing to marshal public randomness in KVStore can only be a programming error
		panic(fmt.Errorf("failed to unmarshal public randomness in KVStore: %w", err))
	}
	return height, pubRand, nil
}

// commitedRandoms iterates over all commited randoms on the store, parses the finality provider public key
// and the height from the iterator key and the commited random from the iterator value.
func (k Keeper) commitedRandoms(ctx context.Context) ([]*types.PublicRandomness, error) {
	store := k.pubRandStore(ctx)
	iter := store.Iterator(nil, nil)
	defer iter.Close()

	commtRandoms := make([]*types.PublicRandomness, 0)
	for ; iter.Valid(); iter.Next() {
		// key contains the fp and the block height
		fpBTCPK, blkHeight, err := parsePubKeyAndBlkHeightFromStoreKey(iter.Key())
		if err != nil {
			return nil, err
		}
		pubRand, err := bbn.NewSchnorrPubRand(iter.Value())
		if err != nil {
			return nil, err
		}

		commtRandoms = append(commtRandoms, &types.PublicRandomness{
			BlockHeight: blkHeight,
			FpBtcPk:     fpBTCPK,
			PubRand:     pubRand,
		})
	}

	return commtRandoms, nil
}

// pubRandFpStore returns the KVStore of the public randomness
// prefix: PubRandKey
// key: (finality provider || PK block height)
// value: PublicRandomness
func (k Keeper) pubRandFpStore(ctx context.Context, fpBtcPK *bbn.BIP340PubKey) prefix.Store {
	prefixedStore := k.pubRandStore(ctx)
	return prefix.NewStore(prefixedStore, fpBtcPK.MustMarshal())
}

// pubRandStore returns the KVStore of the public randomness
// prefix: PubRandKey
// key: (prefix)
// value: PublicRandomness
func (k Keeper) pubRandStore(ctx context.Context) prefix.Store {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	return prefix.NewStore(storeAdapter, types.PubRandKey)
}

// parsePubKeyAndBlkHeightFromStoreKey expects to receive a key with
// BIP340PubKey(fpBTCPK) || BigEndianUint64(blkHeight)
func parsePubKeyAndBlkHeightFromStoreKey(key []byte) (fpBTCPK *bbn.BIP340PubKey, blkHeight uint64, err error) {
	sizeBigEndian := 8
	keyLen := len(key)
	if keyLen < sizeBigEndian+1 {
		return nil, 0, fmt.Errorf("key not long enough to parse BIP340PubKey and block height: %s", key)
	}

	startKeyHeight := keyLen - sizeBigEndian
	fpBTCPK, err = bbn.NewBIP340PubKey(key[:startKeyHeight])
	if err != nil {
		return nil, 0, fmt.Errorf("failed to parse pub key from key %w: %w", bbn.ErrUnmarshal, err)
	}

	blkHeight = sdk.BigEndianToUint64(key[startKeyHeight:])
	return fpBTCPK, blkHeight, nil
}
