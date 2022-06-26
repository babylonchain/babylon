package keeper

import (
	"errors"
	"github.com/babylonchain/babylon/x/checkpointing/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type CheckpointsState struct {
	cdc                codec.BinaryCodec
	checkpoints        sdk.KVStore
	lastConfirmedEpoch sdk.KVStore
	tipEpoch           sdk.KVStore
}

func (k Keeper) CheckpointsState(ctx sdk.Context) CheckpointsState {
	// Build the CheckpointsState storage
	store := ctx.KVStore(k.storeKey)
	return CheckpointsState{
		cdc:                k.cdc,
		checkpoints:        prefix.NewStore(store, types.CheckpointsPrefix),
		lastConfirmedEpoch: prefix.NewStore(store, types.CheckpointsPrefix),
		tipEpoch:           prefix.NewStore(store, types.CheckpointsPrefix),
	}
}

// CreateRawCkpt inserts the raw checkpoint into the storage by its epoch number
func (cs CheckpointsState) CreateRawCkpt(ckpt *types.RawCheckpoint) error {
	// save concrete ckpt object
	cs.checkpoints.Set(types.CkptsObjectKey(ckpt.EpochNum), cs.cdc.MustMarshal(ckpt))

	return cs.UpdateTipEpoch(ckpt.EpochNum)
}

// GetRawCkpt retrieves a raw checkpoint by its epoch number
func (cs CheckpointsState) GetRawCkpt(epoch uint64) (*types.RawCheckpoint, error) {
	ckptsKey := types.CkptsObjectKey(epoch)
	rawBytes := cs.checkpoints.Get(ckptsKey)
	if rawBytes == nil {
		return nil, types.ErrCkptDoesNotExist.Wrap("no raw checkpoint with provided epoch")
	}

	return cs.DeserializeCkpt(rawBytes), nil
}

// GetRawCkptsByStatus retrieves raw checkpoints by their status by the accending order of epoch
func (cs CheckpointsState) GetRawCkptsByStatus(status types.CkptStatus) ([]*types.RawCheckpoint, error) {
	var startEpoch, endEpoch uint64
	lce := cs.GetLastConfirmedEpoch()
	if status == types.Confirmed {
		endEpoch = cs.GetLastConfirmedEpoch()
		startEpoch = 0
	} else {
		startEpoch = lce + 1
		endEpoch = cs.GetTipEpoch()
	}
	if endEpoch <= startEpoch {
		return nil, types.ErrCkptsDoNotExist.Wrap("no raw checkpoints with provided status")
	}
	return cs.getRawCkptsByEpochRangeWithStatus(startEpoch, endEpoch, status)
}

func (cs CheckpointsState) getRawCkptsByEpochRangeWithStatus(start uint64, endEpoch uint64, status types.CkptStatus) ([]*types.RawCheckpoint, error) {
	ckptList := make([]*types.RawCheckpoint, endEpoch-start)
	for i := start; i <= endEpoch; i++ {
		ckpt, err := cs.GetRawCkpt(i)
		if err != nil {
			return nil, err
		}
		if status == ckpt.Status {
			ckptList = append(ckptList, ckpt)
		}
	}

	return ckptList, nil
}

// UpdateCkptStatus updates the checkpoint's status
func (cs CheckpointsState) UpdateCkptStatus(rawCkptBytes []byte, status types.CkptStatus) error {
	ckpt := cs.DeserializeCkpt(rawCkptBytes)
	c, err := cs.GetRawCkpt(ckpt.EpochNum)
	if err != nil {
		// the checkpoint should exist
		return err
	}
	if !c.Hash().Equals(ckpt.Hash()) {
		return errors.New("hash not the same with existing checkpoint")
	}
	ckpt.Status = status
	cs.checkpoints.Set(sdk.Uint64ToBigEndian(ckpt.EpochNum), cs.cdc.MustMarshal(ckpt))

	return nil
}

// GetLastConfirmedEpoch retrieves the last confirmed epoch
func (cs CheckpointsState) GetLastConfirmedEpoch() uint64 {
	if !cs.lastConfirmedEpoch.Has(types.LastConfirmedKey()) {
		return 0
	}
	bz := cs.lastConfirmedEpoch.Get(types.LastConfirmedKey())
	return sdk.BigEndianToUint64(bz)
}

func (cs CheckpointsState) UpdateLastConfirmedEpoch(epoch uint64) error {
	e := cs.GetLastConfirmedEpoch()
	if e >= epoch {
		return errors.New("failed to update last confirmed epoch")
	}
	epochKey := types.LastConfirmedKey()
	cs.lastConfirmedEpoch.Set(epochKey, sdk.Uint64ToBigEndian(epoch))
	return nil
}

// GetTipEpoch returns the highest epoch that has created a raw checkpoint
func (cs CheckpointsState) GetTipEpoch() uint64 {
	if !cs.tipEpoch.Has(types.TipKey()) {
		return 0
	}
	bz := cs.tipEpoch.Get(types.TipKey())
	return sdk.BigEndianToUint64(bz)
}

func (cs CheckpointsState) UpdateTipEpoch(epoch uint64) error {
	tipKey := types.TipKey()
	te := sdk.BigEndianToUint64(cs.tipEpoch.Get(tipKey))
	if te >= epoch {
		return errors.New("failed to update tip epoch")
	}
	cs.tipEpoch.Set(tipKey, sdk.Uint64ToBigEndian(epoch))
	return nil
}

func (cs CheckpointsState) DeserializeCkpt(rawCkptBytes []byte) *types.RawCheckpoint {
	ckpt := new(types.RawCheckpoint)
	cs.cdc.MustUnmarshal(rawCkptBytes, ckpt)
	return ckpt
}

func (cs CheckpointsState) SerializeCkpt(ckpt *types.RawCheckpoint) []byte {
	return cs.cdc.MustMarshal(ckpt)
}
