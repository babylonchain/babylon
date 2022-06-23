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
	hashToEpoch        sdk.KVStore
	hashToStatus       sdk.KVStore
	lastConfirmedEpoch sdk.KVStore
	tipEpoch           sdk.KVStore
}

func (k Keeper) CheckpointsState(ctx sdk.Context) CheckpointsState {
	// Build the CheckpointsState storage
	store := ctx.KVStore(k.storeKey)
	return CheckpointsState{
		cdc:                k.cdc,
		checkpoints:        prefix.NewStore(store, types.CheckpointsPrefix),
		hashToEpoch:        prefix.NewStore(store, types.CkptsHashToEpochPrefix),
		hashToStatus:       prefix.NewStore(store, types.CkptsHashToStatusPrefix),
		lastConfirmedEpoch: prefix.NewStore(store, types.CheckpointsPrefix),
		tipEpoch:           prefix.NewStore(store, types.CheckpointsPrefix),
	}
}

// CreateRawCkpt inserts the raw checkpoint into the hash->epoch, hash->status, and (epoch, status, hash)->ckpt storage
func (cs CheckpointsState) CreateRawCkpt(ckpt *types.RawCheckpoint) error {
	ckptHash := ckpt.Hash()
	ckptsKey := types.CkptsObjectKey(ckpt.EpochNum, ckpt.Status, ckptHash)
	epochKey := types.CkptsEpochKey(ckptHash)
	statusKey := types.CkptsStatusKey(ckptHash)

	// save concrete ckpt object
	cs.checkpoints.Set(ckptsKey, cs.cdc.MustMarshal(ckpt))
	// map ckpt to epoch
	cs.hashToEpoch.Set(epochKey, sdk.Uint64ToBigEndian(ckpt.EpochNum))
	// map ckpt to status
	cs.hashToEpoch.Set(statusKey, types.Uint32ToBitEndian(ckpt.Status))

	return cs.UpdateTipEpoch(ckpt.EpochNum)
}

// GetRawCkpt retrieves a raw checkpoint by its epoch, status, and hash
func (cs CheckpointsState) GetRawCkpt(epoch uint64, status uint32, hash types.RawCkptHash) (*types.RawCheckpoint, error) {
	ckptsKey := types.CkptsObjectKey(epoch, status, hash)
	rawBytes := cs.checkpoints.Get(ckptsKey)
	if rawBytes == nil {
		return nil, types.ErrCkptDoesNotExist.Wrap("no raw checkpoint with provided epoch and hash")
	}

	ckpt := new(types.RawCheckpoint)
	cs.cdc.MustUnmarshal(rawBytes, ckpt)

	return ckpt, nil
}

// GetRawCkptEpoch retrieves the epoch of a raw checkpoint
func (cs CheckpointsState) GetRawCkptEpoch(hash types.RawCkptHash) (uint64, error) {
	epochKey := types.CkptsEpochKey(hash)
	bz := cs.hashToEpoch.Get(epochKey)
	if bz == nil {
		return 0, types.ErrCkptsEpochDoesNotExist.Wrap("no checkpoint epoch with provided hash")
	}
	return sdk.BigEndianToUint64(bz), nil
}

// GetRawCkptStatus retrieves the status of a raw checkpoint
func (cs CheckpointsState) GetRawCkptStatus(hash types.RawCkptHash) (uint32, error) {
	statusKey := types.CkptsStatusKey(hash)
	bz := cs.hashToStatus.Get(statusKey)
	if bz == nil {
		return 0, types.ErrCkptsStatusDoesNotExist.Wrap("no checkpoint status with provided hash")
	}
	return types.BigEndianToUint32(bz), nil
}

// GetRawCkptByHash retrieves a raw checkpoint by its hash
func (cs CheckpointsState) GetRawCkptByHash(hash types.RawCkptHash) (*types.RawCheckpoint, error) {
	epoch, err := cs.GetRawCkptEpoch(hash)
	if err != nil {
		return nil, err
	}
	status, err := cs.GetRawCkptStatus(hash)
	if err != nil {
		return nil, err
	}
	return cs.GetRawCkpt(epoch, status, hash)
}

// GetRawCkptsByEpoch retrieves raw checkpoints by their epoch
func (cs CheckpointsState) GetRawCkptsByEpoch(epoch uint64) ([]*types.RawCheckpoint, error) {
	ckpts := make([]*types.RawCheckpoint, 0)
	for _, s := range types.RAW_CKPT_STATUS {
		pf := append(sdk.Uint64ToBigEndian(epoch), types.Uint32ToBitEndian(s)...)
		store := prefix.NewStore(cs.checkpoints, pf)
		func() {
			iter := store.Iterator(nil, nil)
			defer iter.Close()

			for ; iter.Valid(); iter.Next() {
				rawBytes := iter.Value()
				ckpt := new(types.RawCheckpoint)
				cs.cdc.MustUnmarshal(rawBytes, ckpt)
				ckpts = append(ckpts, ckpt)
			}
		}()
	}
	if len(ckpts) == 0 {
		return nil, types.ErrCkptsDoNotExist.Wrap("no raw checkpoints with provided epoch")
	}
	return ckpts, nil
}

// GetRawCkptByStatus retrieves raw checkpoints by their status by the accending order of epoch
func (cs CheckpointsState) GetRawCkptByStatus(status uint32) ([]*types.RawCheckpoint, error) {
	lce, err := cs.GetLastConfirmedEpoch()
	if err != nil && status == types.CONFIRMED {
		return nil, err
	}
	var startEpoch, endEpoch uint64
	if status == types.CONFIRMED {
		// start from the beginning
		startEpoch = 0
		endEpoch = lce
	} else {
		// start from lce + 1
		startEpoch = lce + 1
		endEpoch = cs.GetTipEpoch()
	}
	if endEpoch <= startEpoch {
		return nil, types.ErrCkptsDoNotExist.Wrap("no raw checkpoints with provided status")
	}
	ckpts := make([]*types.RawCheckpoint, 0)
	for e := startEpoch; e <= endEpoch; e++ {
		pf := append(sdk.Uint64ToBigEndian(e), types.Uint32ToBitEndian(status)...)
		store := prefix.NewStore(cs.checkpoints, pf)
		func() {
			iter := store.Iterator(nil, nil)
			defer iter.Close()

			for ; iter.Valid(); iter.Next() {
				rawBytes := iter.Value()
				ckpt := new(types.RawCheckpoint)
				cs.cdc.MustUnmarshal(rawBytes, ckpt)
				ckpts = append(ckpts, ckpt)
			}
		}()
	}
	if len(ckpts) == 0 {
		return nil, types.ErrCkptsDoNotExist.Wrap("no raw checkpoints with provided status")
	}
	return ckpts, nil
}

// UpdateCkptStatus updates the checkpoint's status
func (cs CheckpointsState) UpdateCkptStatus(hash types.RawCkptHash, status uint32) error {
	ckpt, err := cs.GetRawCkptByHash(hash)
	if err != nil {
		return err
	}
	ckpt.Status = status
	epoch, err := cs.GetRawCkptEpoch(hash)
	if err != nil {
		return err
	}
	oldStatus, err := cs.GetRawCkptStatus(hash)
	if err != nil {
		return err
	}
	statusKey := types.CkptsStatusKey(hash)
	ckptKey := types.CkptsObjectKey(epoch, oldStatus, hash)
	cs.checkpoints.Set(ckptKey, cs.cdc.MustMarshal(ckpt))
	cs.hashToStatus.Set(statusKey, types.Uint32ToBitEndian(status))

	if status == types.CONFIRMED {
		err := cs.UpdateLastConfirmedEpoch(ckpt.EpochNum)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetLastConfirmedEpoch retrieves the last confirmed epoch
func (cs CheckpointsState) GetLastConfirmedEpoch() (uint64, error) {
	if !cs.lastConfirmedEpoch.Has(types.LastConfirmedKey()) {
		return 0, types.ErrLastConfirmedEpochDoesNotExist.Wrap("no last confirmed epoch found")
	}
	bz := cs.lastConfirmedEpoch.Get(types.LastConfirmedKey())
	return sdk.BigEndianToUint64(bz), nil
}

func (cs CheckpointsState) UpdateLastConfirmedEpoch(epoch uint64) error {
	e, err := cs.GetLastConfirmedEpoch()
	if err != nil {
		return err
	}
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
