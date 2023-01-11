package keeper

import (
	"errors"
	"fmt"

	"github.com/babylonchain/babylon/crypto/bls12381"
	"github.com/babylonchain/babylon/x/checkpointing/types"
	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"
)

type (
	Keeper struct {
		cdc            codec.BinaryCodec
		storeKey       storetypes.StoreKey
		memKey         storetypes.StoreKey
		blsSigner      BlsSigner
		epochingKeeper types.EpochingKeeper
		hooks          types.CheckpointingHooks
		paramstore     paramtypes.Subspace
		clientCtx      client.Context
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	signer BlsSigner,
	ek types.EpochingKeeper,
	ps paramtypes.Subspace,
	clientCtx client.Context,
) Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:            cdc,
		storeKey:       storeKey,
		memKey:         memKey,
		blsSigner:      signer,
		epochingKeeper: ek,
		paramstore:     ps,
		hooks:          nil,
		clientCtx:      clientCtx,
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

	// verify BLS sig
	msgBytes := append(sdk.Uint64ToBigEndian(sig.GetEpochNum()), *sig.LastCommitHash...)
	ok, err := bls12381.Verify(*sig.BlsSig, signerBlsKey, msgBytes)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("BLS signature does not match the public key")
	}

	// accumulate BLS signatures
	err = ckptWithMeta.Accumulate(vals, signerAddr, signerBlsKey, *sig.BlsSig, k.GetTotalVotingPower(ctx, sig.GetEpochNum()))
	if err != nil {
		return err
	}

	if ckptWithMeta.Status == types.Sealed {
		// emit event
		err = ctx.EventManager().EmitTypedEvent(
			&types.EventCheckpointSealed{Checkpoint: ckptWithMeta},
		)
		if err != nil {
			ctx.Logger().Error("failed to emit checkpoint sealed event for epoch %v", ckptWithMeta.Ckpt.EpochNum)
		}
		// record state update of Sealed
		ckptWithMeta.RecordStateUpdate(ctx, types.Sealed)
		// log in console
		ctx.Logger().Info(fmt.Sprintf("Checkpointing: checkpoint for epoch %v is Sealed", ckptWithMeta.Ckpt.EpochNum))
	}

	// if reaching this line, it means ckptWithMeta is updated,
	// and we need to write the updated ckptWithMeta back to KVStore
	if err := k.UpdateCheckpoint(ctx, ckptWithMeta); err != nil {
		return err
	}

	return nil
}

func (k Keeper) GetRawCheckpoint(ctx sdk.Context, epochNum uint64) (*types.RawCheckpointWithMeta, error) {
	return k.CheckpointsState(ctx).GetRawCkptWithMeta(epochNum)
}

func (k Keeper) GetStatus(ctx sdk.Context, epochNum uint64) (types.CheckpointStatus, error) {
	ckptWithMeta, err := k.GetRawCheckpoint(ctx, epochNum)
	if err != nil {
		return -1, err
	}
	return ckptWithMeta.Status, nil
}

// AddRawCheckpoint adds a raw checkpoint into the storage
func (k Keeper) AddRawCheckpoint(ctx sdk.Context, ckptWithMeta *types.RawCheckpointWithMeta) error {
	return k.CheckpointsState(ctx).CreateRawCkptWithMeta(ckptWithMeta)
}

func (k Keeper) BuildRawCheckpoint(ctx sdk.Context, epochNum uint64, lch types.LastCommitHash) (*types.RawCheckpointWithMeta, error) {
	ckptWithMeta := types.NewCheckpointWithMeta(types.NewCheckpoint(epochNum, lch), types.Accumulating)
	ckptWithMeta.RecordStateUpdate(ctx, types.Accumulating) // record the state update of Accumulating
	err := k.AddRawCheckpoint(ctx, ckptWithMeta)
	if err != nil {
		return nil, err
	}
	ctx.Logger().Info(fmt.Sprintf("Checkpointing: a new raw checkpoint is built for epoch %v", epochNum))

	return ckptWithMeta, nil
}

// CheckpointEpoch verifies checkpoint from BTC and returns epoch number if
// it equals to the existing raw checkpoint. Otherwise, it further verifies
// the raw checkpoint and decides whether it is an invalid checkpoint or a
// conflicting checkpoint. A conflicting checkpoint indicates the existence
// of a fork
func (k Keeper) CheckpointEpoch(ctx sdk.Context, btcCkptBytes []byte) (uint64, error) {
	ckptWithMeta, err := k.verifyCkptBytes(ctx, btcCkptBytes)
	if err != nil {
		if errors.Is(err, types.ErrConflictingCheckpoint) {
			panic(err)
		}
		return 0, err
	}
	return ckptWithMeta.Ckpt.EpochNum, nil
}

// verifyCkptBytes verifies checkpoint from BTC. A checkpoint is valid if
// it equals to the existing raw checkpoint. Otherwise, it further verifies
// the raw checkpoint and decides whether it is an invalid checkpoint or a
// conflicting checkpoint. A conflicting checkpoint indicates the existence
// of a fork
func (k Keeper) verifyCkptBytes(ctx sdk.Context, btcCkptBytes []byte) (*types.RawCheckpointWithMeta, error) {
	ckpt, err := types.FromBTCCkptBytesToRawCkpt(btcCkptBytes)
	if err != nil {
		return nil, err
	}
	// sanity check
	err = ckpt.ValidateBasic()
	if err != nil {
		return nil, err
	}
	ckptWithMeta, err := k.GetRawCheckpoint(ctx, ckpt.EpochNum)
	if err != nil {
		return nil, err
	}

	// can skip the checks if it is identical with the local checkpoint that is not accumulating
	if ckptWithMeta.Ckpt.Equal(ckpt) && ckptWithMeta.Status != types.Accumulating {
		return ckptWithMeta, nil
	}

	// next verify if the multi signature is valid
	// check whether sufficient voting power is accumulated
	totalPower := k.GetTotalVotingPower(ctx, ckpt.EpochNum)
	signerSet, err := k.GetValidatorSet(ctx, ckpt.EpochNum).FindSubset(ckpt.Bitmap)
	if err != nil {
		return nil, err
	}
	var sum int64
	signersPubKeys := make([]bls12381.PublicKey, len(signerSet))
	for i, v := range signerSet {
		signersPubKeys[i], err = k.GetBlsPubKey(ctx, v.Addr)
		if err != nil {
			return nil, err
		}
		sum += v.Power
	}
	if sum <= totalPower*1/3 {
		return nil, types.ErrInvalidRawCheckpoint.Wrap("insufficient voting power")
	}
	msgBytes := append(sdk.Uint64ToBigEndian(ckpt.GetEpochNum()), *ckpt.LastCommitHash...)
	ok, err := bls12381.VerifyMultiSig(*ckpt.BlsMultiSig, signersPubKeys, msgBytes)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, types.ErrInvalidRawCheckpoint.Wrap("invalid BLS multi-sig")
	}

	// now the checkpoint's multi-sig is valid, if the lastcommithash is the
	// same with that of the local checkpoint, it means it is valid except that
	// it is signed by a different signer set
	if ckptWithMeta.Ckpt.LastCommitHash.Equal(*ckpt.LastCommitHash) {
		return ckptWithMeta, nil
	}

	// multi-sig is valid but the quorum is on a different branch, meaning conflicting is observed
	ctx.Logger().Error(types.ErrConflictingCheckpoint.Wrapf("epoch %v", ckpt.EpochNum).Error())
	// report conflicting checkpoint event
	err = ctx.EventManager().EmitTypedEvent(
		&types.EventConflictingCheckpoint{
			ConflictingCheckpoint: ckpt,
			LocalCheckpoint:       ckptWithMeta,
		},
	)
	if err != nil {
		panic(err)
	}

	return nil, types.ErrConflictingCheckpoint
}

// SetCheckpointSubmitted sets the status of a checkpoint to SUBMITTED,
// and records the associated state update in lifecycle
func (k Keeper) SetCheckpointSubmitted(ctx sdk.Context, epoch uint64) {
	ckpt := k.setCheckpointStatus(ctx, epoch, types.Sealed, types.Submitted)
	err := ctx.EventManager().EmitTypedEvent(
		&types.EventCheckpointSubmitted{Checkpoint: ckpt},
	)
	if err != nil {
		ctx.Logger().Error("failed to emit checkpoint submitted event for epoch %v", ckpt.Ckpt.EpochNum)
	}
}

// SetCheckpointConfirmed sets the status of a checkpoint to CONFIRMED,
// and records the associated state update in lifecycle
func (k Keeper) SetCheckpointConfirmed(ctx sdk.Context, epoch uint64) {
	ckpt := k.setCheckpointStatus(ctx, epoch, types.Submitted, types.Confirmed)
	err := ctx.EventManager().EmitTypedEvent(
		&types.EventCheckpointConfirmed{Checkpoint: ckpt},
	)
	if err != nil {
		ctx.Logger().Error("failed to emit checkpoint confirmed event for epoch %v: %v", ckpt.Ckpt.EpochNum, err)
	}
	// invoke hook
	if err := k.AfterRawCheckpointConfirmed(ctx, epoch); err != nil {
		ctx.Logger().Error("failed to trigger checkpoint confirmed hook for epoch %v: %v", ckpt.Ckpt.EpochNum, err)
	}
}

// SetCheckpointFinalized sets the status of a checkpoint to FINALIZED,
// and records the associated state update in lifecycle
func (k Keeper) SetCheckpointFinalized(ctx sdk.Context, epoch uint64) {
	ckpt := k.setCheckpointStatus(ctx, epoch, types.Confirmed, types.Finalized)
	err := ctx.EventManager().EmitTypedEvent(
		&types.EventCheckpointFinalized{Checkpoint: ckpt},
	)
	if err != nil {
		ctx.Logger().Error("failed to emit checkpoint finalized event for epoch %v: %v", ckpt.Ckpt.EpochNum, err)
	}
	// invoke hook, which is currently subscribed by ZoneConcierge
	if err := k.AfterRawCheckpointFinalized(ctx, epoch); err != nil {
		ctx.Logger().Error("failed to trigger checkpoint finalized hook for epoch %v: %v", ckpt.Ckpt.EpochNum, err)
	}
}

// SetCheckpointForgotten rolls back the status of a checkpoint to Sealed,
// and records the associated state update in lifecycle
func (k Keeper) SetCheckpointForgotten(ctx sdk.Context, epoch uint64) {
	ckpt := k.setCheckpointStatus(ctx, epoch, types.Submitted, types.Sealed)
	err := ctx.EventManager().EmitTypedEvent(
		&types.EventCheckpointForgotten{Checkpoint: ckpt},
	)
	if err != nil {
		ctx.Logger().Error("failed to emit checkpoint forgotten event for epoch %v", ckpt.Ckpt.EpochNum)
	}
}

// setCheckpointStatus sets a ckptWithMeta to the given state,
// and records the state update in its lifecycle
func (k Keeper) setCheckpointStatus(ctx sdk.Context, epoch uint64, from types.CheckpointStatus, to types.CheckpointStatus) *types.RawCheckpointWithMeta {
	ckptWithMeta, err := k.GetRawCheckpoint(ctx, epoch)
	if err != nil {
		// TODO: ignore err for now
		return nil
	}
	if ckptWithMeta.Status != from {
		err = types.ErrInvalidCkptStatus.Wrapf("the status of the checkpoint should be %s", from.String())
		if err != nil {
			// TODO: ignore err for now
			return nil
		}
	}
	ckptWithMeta.Status = to                    // set status
	ckptWithMeta.RecordStateUpdate(ctx, to)     // record state update to the lifecycle
	err = k.UpdateCheckpoint(ctx, ckptWithMeta) // write back to KVStore
	if err != nil {
		panic("failed to update checkpoint status")
	}
	statusChangeMsg := fmt.Sprintf("Checkpointing: checkpoint status for epoch %v successfully changed from %v to %v", epoch, from.String(), to.String())
	ctx.Logger().Info(statusChangeMsg)
	return ckptWithMeta
}

func (k Keeper) UpdateCheckpoint(ctx sdk.Context, ckptWithMeta *types.RawCheckpointWithMeta) error {
	return k.CheckpointsState(ctx).UpdateCheckpoint(ckptWithMeta)
}

func (k Keeper) CreateRegistration(ctx sdk.Context, blsPubKey bls12381.PublicKey, valAddr sdk.ValAddress) error {
	return k.RegistrationState(ctx).CreateRegistration(blsPubKey, valAddr)
}

// GetBLSPubKeySet returns the set of BLS public keys in the same order of the validator set for a given epoch
func (k Keeper) GetBLSPubKeySet(ctx sdk.Context, epochNumber uint64) ([]*types.ValidatorWithBlsKey, error) {
	valset := k.GetValidatorSet(ctx, epochNumber)
	valWithblsKeys := make([]*types.ValidatorWithBlsKey, len(valset))
	for i, val := range valset {
		pubkey, err := k.GetBlsPubKey(ctx, val.Addr)
		if err != nil {
			return nil, err
		}
		valWithblsKeys[i] = &types.ValidatorWithBlsKey{
			ValidatorAddress: val.GetValAddressStr(),
			BlsPubKey:        pubkey,
			VotingPower:      uint64(val.Power),
		}
	}

	return valWithblsKeys, nil
}

func (k Keeper) GetBlsPubKey(ctx sdk.Context, address sdk.ValAddress) (bls12381.PublicKey, error) {
	return k.RegistrationState(ctx).GetBlsPubKey(address)
}

func (k Keeper) GetEpoch(ctx sdk.Context) *epochingtypes.Epoch {
	return k.epochingKeeper.GetEpoch(ctx)
}

func (k Keeper) GetValidatorSet(ctx sdk.Context, epochNumber uint64) epochingtypes.ValidatorSet {
	return k.epochingKeeper.GetValidatorSet(ctx, epochNumber)
}

func (k Keeper) GetTotalVotingPower(ctx sdk.Context, epochNumber uint64) int64 {
	return k.epochingKeeper.GetTotalVotingPower(ctx, epochNumber)
}
