package keeper

import (
	"fmt"
	"github.com/babylonchain/babylon/crypto/bls12381"
	"github.com/babylonchain/babylon/x/checkpointing/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"
)

type (
	Keeper struct {
		cdc            codec.BinaryCodec
		storeKey       sdk.StoreKey
		memKey         sdk.StoreKey
		epochingKeeper types.EpochingKeeper
		hooks          types.CheckpointingHooks
		paramstore     paramtypes.Subspace
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey sdk.StoreKey,
	ek types.EpochingKeeper,
	ps paramtypes.Subspace,
) Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:            cdc,
		storeKey:       storeKey,
		memKey:         memKey,
		epochingKeeper: ek,
		paramstore:     ps,
		hooks:          nil,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// SetHooks sets the validator hooks
func (k *Keeper) SetHooks(sh types.CheckpointingHooks) *Keeper {
	if k.hooks != nil {
		panic("cannot set validator hooks twice")
	}

	k.hooks = sh

	return k
}

// AddBlsSig add bls signatures into storage and generates a raw checkpoint
// if sufficient sigs are accumulated for a specific epoch
func (k Keeper) AddBlsSig(ctx sdk.Context, sig *types.BlsSig) error {
	// TODO: some checks: 1. duplication check 2. epoch check 3. raw ckpt existence check
	// TODO: aggregate bls sigs and try to build raw checkpoints
	k.BlsSigsState(ctx).CreateBlsSig(sig)
	return nil
}

// AddRawCheckpoint adds a raw checkpoint into the storage
// this API may not needed since checkpoints are generated internally
func (k Keeper) AddRawCheckpoint(ctx sdk.Context, ckpt *types.RawCheckpoint) error {
	// NOTE: may remove this API
	return k.CheckpointsState(ctx).CreateRawCkptWithMeta(ckpt)
}

// CheckpointEpoch verifies checkpoint from BTC and returns epoch number
func (k Keeper) CheckpointEpoch(ctx sdk.Context, rawCkptBytes []byte) (uint64, error) {
	ckpt, err := types.BytesToRawCkpt(k.cdc, rawCkptBytes)
	if err != nil {
		return 0, err
	}
	err = k.verifyRawCheckpoint(ckpt)
	if err != nil {
		return 0, err
	}
	return ckpt.EpochNum, nil
}

func (k Keeper) verifyRawCheckpoint(ckpt *types.RawCheckpoint) error {
	// TODO: verify checkpoint basic and bls multi-sig
	return nil
}

// UpdateCkptStatus updates the status of a raw checkpoint
func (k Keeper) UpdateCkptStatus(ctx sdk.Context, rawCkptBytes []byte, status types.CheckpointStatus) error {
	// TODO: some checks
	ckpt, err := types.BytesToRawCkpt(k.cdc, rawCkptBytes)
	if err != nil {
		return err
	}
	return k.CheckpointsState(ctx).UpdateCkptStatus(ckpt, status)
}

func (k Keeper) CreateRegistration(ctx sdk.Context, blsPubKey bls12381.PublicKey, valAddr types.ValidatorAddress) error {
	return k.RegistrationState(ctx).CreateRegistration(blsPubKey, valAddr)
}
