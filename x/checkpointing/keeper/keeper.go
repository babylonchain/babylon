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

// addBlsSig add bls signatures into storage and generates a raw checkpoint
// if sufficient sigs are accumulated for a specific epoch
func (k Keeper) addBlsSig(ctx sdk.Context, sig *types.BlsSig) error {
	// assuming stateless checks have done in Antehandler

	// get raw checkpoint
	ckptWithMeta, err := k.GetRawCheckpoint(ctx, sig.GetEpochNum())
	if err != nil {
		return err
	}
	vals, err
	if ckptWithMeta.Status != types.Accumulating {
		return types.ErrCkptNotAccumulating.Wrapf("raw checkpoint at epoch %v is not accumulating BLS sigs", sig.GetEpochNum())
	}

	aggSig, err := bls12381.AggrSig(*ckpt.Ckpt.BlsMultiSig, *sig.BlsSig)
	if err != nil {
		return err
	}
	// 2. aggregate bls sigs

	// 3. change status if needed
	// 4. store modified raw checkpoint
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

func (k Keeper) CreateRegistration(ctx sdk.Context, blsPubKey bls12381.PublicKey, valAddr sdk.ValAddress) error {
	return k.RegistrationState(ctx).CreateRegistration(blsPubKey, valAddr)
}

func (k Keeper) GetBlsPubKey(ctx sdk.Context, address sdk.ValAddress) (bls12381.PublicKey, error) {
	return k.RegistrationState(ctx).GetBlsPubKey(address)
}

func (k Keeper) GetEpochBoundary(ctx sdk.Context) sdk.Uint {
	return k.epochingKeeper.GetEpochBoundary(ctx)
}

func (k Keeper) GetEpochNumber(ctx sdk.Context) sdk.Uint {
	return k.epochingKeeper.GetEpochNumber(ctx)
}

func (k Keeper) GetValidatorSet(ctx sdk.Context, epoch sdk.Uint) map[string]int64 {
	return k.GetValidatorSet(ctx, epoch)
}
