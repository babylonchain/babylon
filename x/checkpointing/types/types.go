package types

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"

	"github.com/babylonchain/babylon/crypto/bls12381"
	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"
	"github.com/boljen/go-bitmap"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// HashSize is the size in bytes of a hash
	HashSize   = sha256.Size
	BitmapBits = 104 // 104 bits for 104 validators at top
)

type AppHash []byte

type BlsSigHash []byte

type RawCkptHash []byte

func NewCheckpoint(epochNum uint64, appHash AppHash) *RawCheckpoint {
	return &RawCheckpoint{
		EpochNum:    epochNum,
		AppHash:     &appHash,
		Bitmap:      bitmap.New(BitmapBits), // 13 bytes, holding 100 validators
		BlsMultiSig: nil,
	}
}

func NewCheckpointWithMeta(ckpt *RawCheckpoint, status CheckpointStatus) *RawCheckpointWithMeta {
	return &RawCheckpointWithMeta{
		Ckpt:      ckpt,
		Status:    status,
		Lifecycle: []*CheckpointStateUpdate{},
	}
}

// Accumulate does the following things
// 1. aggregates the BLS signature
// 2. aggregates the BLS public key
// 3. updates Bitmap
// 4. accumulates voting power
// it returns nil if the checkpoint is updated, otherwise it returns an error
func (cm *RawCheckpointWithMeta) Accumulate(
	vals epochingtypes.ValidatorSet,
	signerAddr sdk.ValAddress,
	signerBlsKey bls12381.PublicKey,
	sig bls12381.Signature,
	totalPower int64) error {

	// the checkpoint should be accumulating
	if cm.Status != Accumulating {
		return ErrCkptNotAccumulating
	}

	// get validator and its index
	val, index, err := vals.FindValidatorWithIndex(signerAddr)
	if err != nil {
		return err
	}

	// return an error if the validator has already voted
	if bitmap.Get(cm.Ckpt.Bitmap, index) {
		return ErrCkptAlreadyVoted
	}

	// aggregate BLS sig
	if cm.Ckpt.BlsMultiSig != nil {
		aggSig, err := bls12381.AggrSig(*cm.Ckpt.BlsMultiSig, sig)
		if err != nil {
			return err
		}
		cm.Ckpt.BlsMultiSig = &aggSig
	} else {
		cm.Ckpt.BlsMultiSig = &sig
	}

	// aggregate BLS public key
	if cm.BlsAggrPk != nil {
		aggPK, err := bls12381.AggrPK(*cm.BlsAggrPk, signerBlsKey)
		if err != nil {
			return err
		}
		cm.BlsAggrPk = &aggPK
	} else {
		cm.BlsAggrPk = &signerBlsKey
	}

	// update bitmap
	bitmap.Set(cm.Ckpt.Bitmap, index, true)

	// accumulate voting power and update status when the threshold is reached
	cm.PowerSum += uint64(val.Power)
	if int64(cm.PowerSum) > totalPower/3 {
		cm.Status = Sealed
	}

	return nil
}

func (cm *RawCheckpointWithMeta) IsMoreMatureThanStatus(status CheckpointStatus) bool {
	return cm.Status > status
}

// RecordStateUpdate appends a new state update to the raw ckpt with meta
// where the time/height are captured by the current ctx
func (cm *RawCheckpointWithMeta) RecordStateUpdate(ctx context.Context, status CheckpointStatus) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	height, time := sdkCtx.BlockHeight(), sdkCtx.BlockTime()
	stateUpdate := &CheckpointStateUpdate{
		State:       status,
		BlockHeight: uint64(height),
		BlockTime:   &time,
	}
	cm.Lifecycle = append(cm.Lifecycle, stateUpdate)
}

func NewAppHashFromHex(s string) (AppHash, error) {
	bz, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}

	var appHash AppHash

	err = appHash.Unmarshal(bz)
	if err != nil {
		return nil, err
	}

	return appHash, nil
}

func (appHash *AppHash) Unmarshal(bz []byte) error {
	if len(bz) != HashSize {
		return errors.New("invalid appHash length")
	}
	*appHash = bz
	return nil
}

func (appHash *AppHash) Size() (n int) {
	if appHash == nil {
		return 0
	}
	return len(*appHash)
}

func (appHash *AppHash) Equal(l AppHash) bool {
	return appHash.String() == l.String()
}

func (appHash *AppHash) String() string {
	return hex.EncodeToString(*appHash)
}

func (appHash *AppHash) MustMarshal() []byte {
	bz, err := appHash.Marshal()
	if err != nil {
		panic(err)
	}
	return bz
}

func (appHash *AppHash) Marshal() ([]byte, error) {
	return *appHash, nil
}

func (appHash AppHash) MarshalTo(data []byte) (int, error) {
	copy(data, appHash)
	return len(data), nil
}

func (appHash *AppHash) ValidateBasic() error {
	if appHash == nil {
		return errors.New("invalid appHash")
	}
	if len(*appHash) != HashSize {
		return errors.New("invalid appHash")
	}
	return nil
}

// ValidateBasic does sanity checks on a raw checkpoint
func (ckpt RawCheckpoint) ValidateBasic() error {
	if ckpt.Bitmap == nil {
		return ErrInvalidRawCheckpoint.Wrapf("bitmap cannot be empty")
	}
	err := ckpt.AppHash.ValidateBasic()
	if err != nil {
		return ErrInvalidRawCheckpoint.Wrapf(err.Error())
	}
	err = ckpt.BlsMultiSig.ValidateBasic()
	if err != nil {
		return ErrInvalidRawCheckpoint.Wrapf(err.Error())
	}

	return nil
}

func CkptWithMetaToBytes(cdc codec.BinaryCodec, ckptWithMeta *RawCheckpointWithMeta) []byte {
	return cdc.MustMarshal(ckptWithMeta)
}

func BytesToCkptWithMeta(cdc codec.BinaryCodec, bz []byte) (*RawCheckpointWithMeta, error) {
	ckptWithMeta := new(RawCheckpointWithMeta)
	err := cdc.Unmarshal(bz, ckptWithMeta)
	return ckptWithMeta, err
}
