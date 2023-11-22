package keeper

import (
	"context"
	corestoretypes "cosmossdk.io/core/store"
	"fmt"

	ckpttypes "github.com/babylonchain/babylon/x/checkpointing/types"

	"cosmossdk.io/log"
	"github.com/babylonchain/babylon/x/monitor/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type (
	Keeper struct {
		cdc                  codec.BinaryCodec
		storeService         corestoretypes.KVStoreService
		btcLightClientKeeper types.BTCLightClientKeeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService corestoretypes.KVStoreService,
	bk types.BTCLightClientKeeper,
) Keeper {
	return Keeper{
		cdc:                  cdc,
		storeService:         storeService,
		btcLightClientKeeper: bk,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func bytesToUint64(bytes []byte) (uint64, error) {
	if len(bytes) != 8 {
		return 0, fmt.Errorf("epoch bytes must have exactly 8 bytes")
	}

	return sdk.BigEndianToUint64(bytes), nil
}

func (k Keeper) updateBtcLightClientHeightForEpoch(ctx context.Context, epoch uint64) {
	store := k.storeService.OpenKVStore(ctx)
	currentTipHeight := k.btcLightClientKeeper.GetTipInfo(ctx).Height
	if err := store.Set(types.GetEpochEndLightClientHeightKey(epoch), sdk.Uint64ToBigEndian(currentTipHeight)); err != nil {
		panic(err)
	}
}

func (k Keeper) updateBtcLightClientHeightForCheckpoint(ctx context.Context, ckpt *ckpttypes.RawCheckpoint) error {
	store := k.storeService.OpenKVStore(ctx)
	ckptHashStr := ckpt.HashStr()

	storeKey, err := types.GetCheckpointReportedLightClientHeightKey(ckptHashStr)
	if err != nil {
		return err
	}

	// if the checkpoint exists, meaning an earlier checkpoint with a lower BTC height is already recorded
	// we should keep the lower BTC height in the store
	has, err := store.Has(storeKey)
	if err != nil {
		panic(err)
	}
	if has {
		k.Logger(sdk.UnwrapSDKContext(ctx)).With("module", fmt.Sprintf("checkpoint %s is already recorded", ckptHashStr))
		return nil
	}

	currentTipHeight := k.btcLightClientKeeper.GetTipInfo(ctx).Height
	return store.Set(storeKey, sdk.Uint64ToBigEndian(currentTipHeight))
}

func (k Keeper) removeCheckpointRecord(ctx context.Context, ckpt *ckpttypes.RawCheckpoint) error {
	store := k.storeService.OpenKVStore(ctx)
	ckptHashStr := ckpt.HashStr()

	storeKey, err := types.GetCheckpointReportedLightClientHeightKey(ckptHashStr)
	if err != nil {
		return err
	}

	if err := store.Delete(storeKey); err != nil {
		panic(err)
	}
	return nil
}

func (k Keeper) LightclientHeightAtEpochEnd(ctx context.Context, epoch uint64) (uint64, error) {
	if epoch == 0 {
		return k.btcLightClientKeeper.GetBaseBTCHeader(ctx).Height, nil
	}

	store := k.storeService.OpenKVStore(ctx)

	btcHeightBytes, err := store.Get(types.GetEpochEndLightClientHeightKey(epoch))
	if err != nil {
		panic(err)
	}
	// nil would be returned if key does not exist
	if btcHeightBytes == nil {
		// we do not have any key under given epoch, most probably epoch did not finish
		// yet
		return 0, types.ErrEpochNotEnded.Wrapf("epoch %d", epoch)
	}

	btcHeight, err := bytesToUint64(btcHeightBytes)

	if err != nil {
		panic("Invalid data in database")
	}

	return btcHeight, nil
}

func (k Keeper) LightclientHeightAtCheckpointReported(ctx context.Context, hashString string) (uint64, error) {
	store := k.storeService.OpenKVStore(ctx)

	storeKey, err := types.GetCheckpointReportedLightClientHeightKey(hashString)
	if err != nil {
		return 0, err
	}

	btcHeightBytes, err := store.Get(storeKey)
	if err != nil {
		panic(err)
	}
	// nil would be returned if key does not exist
	if btcHeightBytes == nil {
		return 0, types.ErrCheckpointNotReported.Wrapf("checkpoint hash: %s", hashString)
	}

	btcHeight, err := bytesToUint64(btcHeightBytes)
	if err != nil {
		panic("invalid data in database")
	}

	return btcHeight, nil
}
