package keeper

import (
	"context"
	"encoding/binary"
	"fmt"

	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/cosmos/cosmos-sdk/runtime"
)

// cosmos-sdk does not have utils for uint32
func uint32ToBytes(v uint32) []byte {
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], v)
	return buf[:]
}
func uint32FromBytes(b []byte) (uint32, error) {
	if len(b) != 4 {
		return 0, fmt.Errorf("invalid uint32 bytes length: %d", len(b))
	}

	return binary.BigEndian.Uint32(b), nil
}

func mustUint32FromBytes(b []byte) uint32 {
	v, err := uint32FromBytes(b)
	if err != nil {
		panic(err)
	}

	return v
}

func (k Keeper) paramsStore(ctx context.Context) prefix.Store {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	return prefix.NewStore(storeAdapter, types.ParamsKey)
}

func (k Keeper) nextParamsVersion(ctx context.Context) uint32 {
	paramsStore := k.paramsStore(ctx)
	it := paramsStore.ReverseIterator(nil, nil)
	defer it.Close()

	if !it.Valid() {
		return 0
	}

	return mustUint32FromBytes(it.Key()) + 1
}

func (k Keeper) getLastParams(ctx context.Context) *types.StoredParams {
	paramsStore := k.paramsStore(ctx)
	it := paramsStore.ReverseIterator(nil, nil)
	defer it.Close()

	if !it.Valid() {
		return nil
	}
	var sp types.StoredParams
	k.cdc.MustUnmarshal(it.Value(), &sp)
	return &sp
}

// SetParams sets the x/btcstaking module parameters.
func (k Keeper) SetParams(ctx context.Context, p types.Params) error {
	if err := p.Validate(); err != nil {
		return err
	}

	nextVersion := k.nextParamsVersion(ctx)
	paramsStore := k.paramsStore(ctx)

	sp := types.StoredParams{
		Params:  p,
		Version: nextVersion,
	}

	paramsStore.Set(uint32ToBytes(nextVersion), k.cdc.MustMarshal(&sp))
	return nil
}

func (k Keeper) GetAllParams(ctx context.Context) []*types.Params {
	paramsStore := k.paramsStore(ctx)
	it := paramsStore.Iterator(nil, nil)
	defer it.Close()

	var p []*types.Params
	for ; it.Valid(); it.Next() {
		var sp types.StoredParams
		k.cdc.MustUnmarshal(it.Value(), &sp)
		p = append(p, &sp.Params)
	}

	return p
}

func (k Keeper) GetParamsByVersion(ctx context.Context, v uint32) *types.Params {
	paramsStore := k.paramsStore(ctx)
	spBytes := paramsStore.Get(uint32ToBytes(v))
	if len(spBytes) == 0 {
		return nil
	}

	var sp types.StoredParams
	k.cdc.MustUnmarshal(spBytes, &sp)
	return &sp.Params
}

func mustGetLastParams(ctx context.Context, k Keeper) types.StoredParams {
	sp := k.getLastParams(ctx)
	if sp == nil {
		panic("last params not found")
	}

	return *sp
}

// GetParams returns the latest x/btcstaking module parameters.
func (k Keeper) GetParams(ctx context.Context) types.Params {
	return mustGetLastParams(ctx, k).Params
}

func (k Keeper) GetParamsWithVersion(ctx context.Context) types.StoredParams {
	return mustGetLastParams(ctx, k)
}

// MinCommissionRate returns the minimal commission rate of finality providers
func (k Keeper) MinCommissionRate(ctx context.Context) math.LegacyDec {
	return k.GetParams(ctx).MinCommissionRate
}
