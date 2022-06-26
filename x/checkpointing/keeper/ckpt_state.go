package keeper

import (
	"errors"
	"github.com/babylonchain/babylon/x/checkpointing/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type CheckpointsState struct {
	cdc         codec.BinaryCodec
	checkpoints sdk.KVStore
}

func (k Keeper) CheckpointsState(ctx sdk.Context) CheckpointsState {
	// Build the CheckpointsState storage
	store := ctx.KVStore(k.storeKey)
	return CheckpointsState{
		cdc:         k.cdc,
		checkpoints: prefix.NewStore(store, types.CheckpointsPrefix),
	}
}

// CreateRawCkpt inserts the raw checkpoint into the storage by its epoch number
func (cs CheckpointsState) CreateRawCkpt(ckpt *types.RawCheckpoint) {
	// save concrete ckpt object
	cs.checkpoints.Set(types.CkptsObjectKey(ckpt.EpochNum), cs.SerializeCkpt(ckpt))
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

// GetRawCkptsByStatus retrieves raw checkpoints by their status by the descending order of epoch
func (cs CheckpointsState) GetRawCkptsByStatus(status types.CkptStatus) []*types.RawCheckpoint {
	var ckpts []*types.RawCheckpoint

	store := prefix.NewStore(cs.checkpoints, types.CkptsObjectPrefix)
	iter := store.ReverseIterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		ckptBytes := iter.Value()
		ckpt := cs.DeserializeCkpt(ckptBytes)
		// the loop can end if the current status is CONFIRMED but the requested status is not CONFIRMED
		if status != types.Confirmed && ckpt.Status == types.Confirmed {
			return ckpts
		}
		if ckpt.Status == status {
			ckpts = append(ckpts, ckpt)
		}
	}
	return ckpts
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

func (cs CheckpointsState) DeserializeCkpt(rawCkptBytes []byte) *types.RawCheckpoint {
	ckpt := new(types.RawCheckpoint)
	cs.cdc.MustUnmarshal(rawCkptBytes, ckpt)
	return ckpt
}

func (cs CheckpointsState) SerializeCkpt(ckpt *types.RawCheckpoint) []byte {
	return cs.cdc.MustMarshal(ckpt)
}
