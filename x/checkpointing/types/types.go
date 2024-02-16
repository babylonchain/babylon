package types

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/boljen/go-bitmap"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	txformat "github.com/babylonchain/babylon/btctxformatter"
	"github.com/babylonchain/babylon/crypto/bls12381"
	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"
)

const (
	// HashSize is the size in bytes of a hash
	HashSize   = sha256.Size
	BitmapBits = txformat.BitMapLength * 8 // 104 bits for 104 validators at top
)

type BlockHash []byte

type BlsSigHash []byte

type RawCkptHash []byte

func NewCheckpoint(epochNum uint64, blockHash BlockHash) *RawCheckpoint {
	return &RawCheckpoint{
		EpochNum:    epochNum,
		BlockHash:   &blockHash,
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
	if int64(cm.PowerSum)*3 > totalPower*2 {
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
	height, time := sdkCtx.HeaderInfo().Height, sdkCtx.HeaderInfo().Time
	stateUpdate := &CheckpointStateUpdate{
		State:       status,
		BlockHeight: uint64(height),
		BlockTime:   &time,
	}
	cm.Lifecycle = append(cm.Lifecycle, stateUpdate)
}

func (bh *BlockHash) Unmarshal(bz []byte) error {
	if len(bz) != HashSize {
		return fmt.Errorf(
			"invalid block hash length, expected: %d, got: %d",
			HashSize, len(bz))
	}
	*bh = bz
	return nil
}

func (bh *BlockHash) Size() (n int) {
	if bh == nil {
		return 0
	}
	return len(*bh)
}

func (bh *BlockHash) Equal(l BlockHash) bool {
	return bh.String() == l.String()
}

func (bh *BlockHash) String() string {
	return hex.EncodeToString(*bh)
}

func (bh *BlockHash) MustMarshal() []byte {
	bz, err := bh.Marshal()
	if err != nil {
		panic(err)
	}
	return bz
}

func (bh *BlockHash) Marshal() ([]byte, error) {
	return *bh, nil
}

func (bh BlockHash) MarshalTo(data []byte) (int, error) {
	copy(data, bh)
	return len(data), nil
}

func (bh *BlockHash) ValidateBasic() error {
	if bh == nil {
		return errors.New("invalid block hash")
	}
	if len(*bh) != HashSize {
		return errors.New("invalid block hash")
	}
	return nil
}

// ValidateBasic does sanity checks on a raw checkpoint
func (ckpt RawCheckpoint) ValidateBasic() error {
	if ckpt.Bitmap == nil {
		return ErrInvalidRawCheckpoint.Wrapf("bitmap cannot be empty")
	}
	err := ckpt.BlockHash.ValidateBasic()
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

func (ve *VoteExtension) ToBLSSig() *BlsSig {
	return &BlsSig{
		EpochNum:         ve.EpochNum,
		BlockHash:        ve.BlockHash,
		BlsSig:           ve.BlsSig,
		SignerAddress:    ve.Signer,
		ValidatorAddress: ve.ValidatorAddress,
	}
}
