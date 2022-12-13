package types

import (
	"github.com/babylonchain/babylon/btctxformatter"
	"github.com/babylonchain/babylon/crypto/bls12381"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
)

func (m BlsSig) Hash() BlsSigHash {
	fields := [][]byte{
		sdk.Uint64ToBigEndian(m.EpochNum),
		m.LastCommitHash.MustMarshal(),
		m.BlsSig.MustMarshal(),
		[]byte(m.SignerAddress),
	}
	return hash(fields)
}

func (m RawCheckpoint) Hash() RawCkptHash {
	fields := [][]byte{
		sdk.Uint64ToBigEndian(m.EpochNum),
		m.LastCommitHash.MustMarshal(),
		m.BlsMultiSig.MustMarshal(),
		m.Bitmap,
	}
	return hash(fields)
}

// SignedMsg is the message corresponding to the BLS sig in this raw checkpoint
// Its value should be (epoch_number || last_commit_hash)
func (m RawCheckpoint) SignedMsg() []byte {
	return append(sdk.Uint64ToBigEndian(m.EpochNum), *m.LastCommitHash...)
}

func hash(fields [][]byte) []byte {
	var bz []byte
	for _, b := range fields {
		bz = append(bz, b...)
	}
	return tmhash.Sum(bz)
}

func (m BlsSigHash) Bytes() []byte {
	return m
}

func (m RawCkptHash) Bytes() []byte {
	return m
}

func FromBTCCkptBytesToRawCkpt(btcCkptBytes []byte) (*RawCheckpoint, error) {
	btcCkpt, err := btctxformatter.DecodeRawCheckpoint(btctxformatter.CurrentVersion, btcCkptBytes)
	if err != nil {
		return nil, err
	}
	var lch LastCommitHash
	err = lch.Unmarshal(btcCkpt.LastCommitHash)
	if err != nil {
		return nil, err
	}
	var blsSig bls12381.Signature
	err = blsSig.Unmarshal(btcCkpt.BlsSig)
	if err != nil {
		return nil, err
	}
	rawCheckpoint := &RawCheckpoint{
		EpochNum:       btcCkpt.Epoch,
		LastCommitHash: &lch,
		Bitmap:         btcCkpt.BitMap,
		BlsMultiSig:    &blsSig,
	}

	return rawCheckpoint, nil
}

func FromRawCkptToBTCCkpt(rawCkpt *RawCheckpoint, address []byte) (*btctxformatter.RawBtcCheckpoint, error) {
	lchBytes, err := rawCkpt.LastCommitHash.Marshal()
	if err != nil {
		return nil, err
	}
	blsSigBytes, err := rawCkpt.BlsMultiSig.Marshal()
	if err != nil {
		return nil, err
	}

	btcCkpt := &btctxformatter.RawBtcCheckpoint{
		Epoch:            rawCkpt.EpochNum,
		LastCommitHash:   lchBytes,
		BitMap:           rawCkpt.Bitmap,
		SubmitterAddress: address,
		BlsSig:           blsSigBytes,
	}

	return btcCkpt, nil
}
