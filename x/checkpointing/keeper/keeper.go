package keeper

import (
	"fmt"

	"github.com/babylonchain/babylon/crypto/bls12381"
	"github.com/babylonchain/babylon/x/checkpointing/types"
	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"
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

// addBlsSig adds a BLS signature to the raw checkpoint and updates the status
// if sufficient signatures are accumulated for the epoch.
func (k Keeper) addBlsSig(ctx sdk.Context, sig *types.BlsSig) error {
	// assuming stateless checks have done in Antehandler

	// get raw checkpoint
	ckptWithMeta, err := k.GetRawCheckpoint(ctx, sig.GetEpochNum())
	if err != nil {
		return err
	}

	// the checkpoint is not accumulating
	if ckptWithMeta.Status != types.Accumulating {
		return nil
	}

	// get signer's address
	signerAddr, err := sdk.ValAddressFromBech32(sig.SignerAddress)
	if err != nil {
		return err
	}

	// get validators for the epoch
	vals := k.GetValidatorSet(ctx, sig.GetEpochNum())
	signerBlsKey, err := k.GetBlsPubKey(ctx, signerAddr)
	if err != nil {
		return err
	}

	// accumulate BLS signatures
	updated, err := ckptWithMeta.Accumulate(
		vals, signerAddr, signerBlsKey, *sig.BlsSig, k.GetTotalVotingPower(ctx, sig.GetEpochNum()))
	if err != nil {
		return err
	}

	if updated {
		err = k.UpdateCheckpoint(ctx, ckptWithMeta)
	}
	if err != nil {
		return err
	}

	return nil
}

func (k Keeper) GetRawCheckpoint(ctx sdk.Context, epochNum uint64) (*types.RawCheckpointWithMeta, error) {
	return k.CheckpointsState(ctx).GetRawCkptWithMeta(epochNum)
}

// AddRawCheckpoint adds a raw checkpoint into the storage
func (k Keeper) AddRawCheckpoint(ctx sdk.Context, ckptWithMeta *types.RawCheckpointWithMeta) error {
	return k.CheckpointsState(ctx).CreateRawCkptWithMeta(ckptWithMeta)
}

func (k Keeper) BuildRawCheckpoint(ctx sdk.Context, epochNum uint64, lch types.LastCommitHash) error {
	ckptWithMeta := types.NewCheckpointWithMeta(types.NewCheckpoint(epochNum, lch), types.Accumulating)

	return k.AddRawCheckpoint(ctx, ckptWithMeta)
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

func (k Keeper) UpdateCheckpoint(ctx sdk.Context, ckptWithMeta *types.RawCheckpointWithMeta) error {
	return k.CheckpointsState(ctx).UpdateCheckpoint(ckptWithMeta)
}

func (k Keeper) CreateRegistration(ctx sdk.Context, blsPubKey bls12381.PublicKey, valAddr sdk.ValAddress) error {
	return k.RegistrationState(ctx).CreateRegistration(blsPubKey, valAddr)
}

func (k Keeper) GetBlsPubKey(ctx sdk.Context, address sdk.ValAddress) (bls12381.PublicKey, error) {
	return k.RegistrationState(ctx).GetBlsPubKey(address)
}

func (k Keeper) GetEpoch(ctx sdk.Context) epochingtypes.Epoch {
	return k.epochingKeeper.GetEpoch(ctx)
}

func (k Keeper) GetValidatorSet(ctx sdk.Context, epochNumber uint64) epochingtypes.ValidatorSet {
	return k.epochingKeeper.GetValidatorSet(ctx, epochNumber)
}

func (k Keeper) GetTotalVotingPower(ctx sdk.Context, epochNumber uint64) int64 {
	return k.epochingKeeper.GetTotalVotingPower(ctx, epochNumber)
}
