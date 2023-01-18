package keeper

import (
	"fmt"
	ckpttypes "github.com/babylonchain/babylon/x/checkpointing/types"

	"github.com/babylonchain/babylon/x/monitor/types"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"
)

type (
	Keeper struct {
		cdc                  codec.BinaryCodec
		storeKey             storetypes.StoreKey
		memKey               storetypes.StoreKey
		paramstore           paramtypes.Subspace
		btcLightClientKeeper types.BTCLightClientKeeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	bk types.BTCLightClientKeeper,
) Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:                  cdc,
		storeKey:             storeKey,
		memKey:               memKey,
		paramstore:           ps,
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

func (k Keeper) updateBtcLightClientHeightForEpoch(ctx sdk.Context, epoch uint64) {
	store := ctx.KVStore(k.storeKey)
	currentTipHeight := k.btcLightClientKeeper.GetTipInfo(ctx).Height
	store.Set(types.GetEpochEndLightClientHeightKey(epoch), sdk.Uint64ToBigEndian(currentTipHeight))
}

func (k Keeper) updateBtcLightClientHeightForCheckpoint(ctx sdk.Context, ckpt *ckpttypes.RawCheckpoint) {
	store := ctx.KVStore(k.storeKey)
	ckptHash := ckpt.Hash()

	// if the checkpoint exists, meaning an earlier checkpoint with a lower btc height is already recorded
	// we should keep the lower btc height in the store
	if store.Has(ckptHash) {
		return
	}

	currentTipHeight := k.btcLightClientKeeper.GetTipInfo(ctx).Height
	store.Set(types.GetCheckpointReportedLightClientHeightKey(ckptHash), sdk.Uint64ToBigEndian(currentTipHeight))
}

func (k Keeper) LightclientHeightAtEpochEnd(ctx sdk.Context, epoch uint64) (uint64, error) {
	store := ctx.KVStore(k.storeKey)

	btcHeightBytes := store.Get(types.GetEpochEndLightClientHeightKey(epoch))

	if len(btcHeightBytes) == 0 {
		// we do not have any key under given epoch, most probably epoch did not finish
		// yet
		return 0, types.ErrEpochNotFinishedYet
	}

	btcHeight, err := bytesToUint64(btcHeightBytes)

	if err != nil {
		panic("Invalid data in database")
	}

	return btcHeight, nil
}

func (k Keeper) LightclientHeightAtCheckpointReported(ctx sdk.Context, hash []byte) (uint64, error) {
	store := ctx.KVStore(k.storeKey)

	btcHeightBytes := store.Get(types.GetCheckpointReportedLightClientHeightKey(hash))

	if len(btcHeightBytes) == 0 {
		return 0, types.ErrCheckpointNotReported.Wrapf("checkpoint hash: %x", hash)
	}

	btcHeight, err := bytesToUint64(btcHeightBytes)
	if err != nil {
		panic("invalid data in database")
	}

	return btcHeight, nil
}
