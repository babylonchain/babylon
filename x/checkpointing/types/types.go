package types

import (
	"bytes"
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
	HashSize = sha256.Size
)

type LastCommitHash []byte

type BlsSigHash []byte

type RawCkptHash []byte

func NewCheckpoint(epochNum uint64, lch LastCommitHash) *RawCheckpoint {
	return &RawCheckpoint{
		EpochNum:       epochNum,
		LastCommitHash: &lch,
		Bitmap:         nil,
		BlsMultiSig:    nil,
	}
}

func NewCheckpointWithMeta(ckpt *RawCheckpoint, status CheckpointStatus) *RawCheckpointWithMeta {
	return &RawCheckpointWithMeta{
		Ckpt:   ckpt,
		Status: status,
	}
}

// Accumulate does the following things
// 1. aggregates the BLS signature
// 2. aggregates the BLS public key
// 3. updates Bitmap
// 4. accumulates voting power
// it returns True if the checkpoint is updated
func (cm *RawCheckpointWithMeta) Accumulate(
	vals epochingtypes.ValidatorSet,
	signerAddr sdk.ValAddress,
	signerBlsKey bls12381.PublicKey,
	sig bls12381.Signature,
	totalPower int64) (bool, error) {

	// the checkpoint should be accumulating
	if cm.Status != Accumulating {
		return false, ErrCkptNotAccumulating
	}

	// get validator and its index
	val, index, err := vals.FindValidatorWithIndex(signerAddr)
	if err != nil {
		return false, err
	}

	// return an error if the validator has already voted
	if bitmap.Get(cm.Ckpt.Bitmap, index) {
		return false, ErrCkptAlreadyVoted
	}

	// aggregate BLS sig
	aggSig, err := bls12381.AggrSig(*cm.Ckpt.BlsMultiSig, sig)
	if err != nil {
		return false, err
	}
	cm.Ckpt.BlsMultiSig = &aggSig

	// aggregate BLS public key
	aggPK, err := bls12381.AggrPK(*cm.BlsAggrPk, signerBlsKey)
	if err != nil {
		return false, err
	}
	cm.BlsAggrPk = &aggPK

	// update bitmap
	bitmap.Set(cm.Ckpt.Bitmap, index, true)

	// accumulate voting power and update status when the threshold is reached
	cm.PowerSum += uint64(val.Power)
	if int64(cm.PowerSum) > totalPower/3 {
		cm.Status = Sealed
	}

	return true, nil
}

func NewLastCommitHashFromHex(s string) (LastCommitHash, error) {
	bz, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}

	var lch LastCommitHash

	err = lch.Unmarshal(bz)
	if err != nil {
		return nil, err
	}

	return lch, nil
}

func (lch *LastCommitHash) Unmarshal(bz []byte) error {
	if len(bz) != HashSize {
		return errors.New("invalid lastCommitHash length")
	}
	*lch = bz
	return nil
}

func (lch *LastCommitHash) Size() (n int) {
	if lch == nil {
		return 0
	}
	return len(*lch)
}

func (lch *LastCommitHash) Equal(l LastCommitHash) bool {
	return lch.String() == l.String()
}

func (lch *LastCommitHash) String() string {
	return hex.EncodeToString(*lch)
}

func (lch *LastCommitHash) Marshal() ([]byte, error) {
	return *lch, nil
}

func (lch *LastCommitHash) MarshalTo(dAtA []byte) (int, error) {
	size := lch.Size()
	return lch.MarshalToSizedBuffer(dAtA[:size])
}

func (lch *LastCommitHash) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (lch *LastCommitHash) ValidateBasic() error {
	if lch == nil {
		return errors.New("invalid lastCommitHash")
	}
	if len(*lch) != HashSize {
		return errors.New("invalid lastCommitHash")
	}
	return nil
}

func BytesToRawCkpt(cdc codec.BinaryCodec, bz []byte) (*RawCheckpoint, error) {
	ckpt := new(RawCheckpoint)
	err := cdc.Unmarshal(bz, ckpt)
	return ckpt, err
}

// ValidateBasic does sanity checks on a raw checkpoint
func (ckpt RawCheckpoint) ValidateBasic() error {
	if ckpt.EpochNum == 0 {
		return ErrInvalidRawCheckpoint.Wrapf("epoch number cannot be zero")
	}
	if ckpt.Bitmap == nil {
		return ErrInvalidRawCheckpoint.Wrapf("bitmap cannot be empty")
	}
	err := ckpt.LastCommitHash.ValidateBasic()
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

func (m RawCkptHash) Equals(h RawCkptHash) bool {
	if bytes.Compare(m.Bytes(), h.Bytes()) == 0 {
		return true
	}
	return false
}
