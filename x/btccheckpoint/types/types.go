package types

import (
	"github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btccheckpoint/btcutils"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Semantically valid checkpoint submission with:
// - valid submitter address
// - at least 2 parsed proof
// Modelling proofs as separate Proof1 and Proof2, as this is more explicit than
// []*ParsedProof.
type RawCheckpointSubmission struct {
	Submitter sdk.AccAddress
	Proof1    btcutils.ParsedProof
	Proof2    btcutils.ParsedProof
}

func NewRawCheckpointSubmission(a sdk.AccAddress, p1 btcutils.ParsedProof, p2 btcutils.ParsedProof) RawCheckpointSubmission {
	r := RawCheckpointSubmission{
		Submitter: a,
		Proof1:    p1,
		Proof2:    p2,
	}

	return r
}

func (s *RawCheckpointSubmission) GetProofs() []*btcutils.ParsedProof {
	return []*btcutils.ParsedProof{&s.Proof1, &s.Proof2}
}

func (s *RawCheckpointSubmission) GetRawCheckPointBytes() []byte {
	var rawCheckpointData []byte
	rawCheckpointData = append(rawCheckpointData, s.Proof1.OpReturnData...)
	rawCheckpointData = append(rawCheckpointData, s.Proof2.OpReturnData...)
	return rawCheckpointData
}

func (s *RawCheckpointSubmission) GetFirstBlockHash() types.BTCHeaderHashBytes {
	return s.Proof1.BlockHash
}

func (s *RawCheckpointSubmission) GetSecondBlockHash() types.BTCHeaderHashBytes {
	return s.Proof2.BlockHash
}

func toTransactionKey(p *btcutils.ParsedProof) TransactionKey {
	hashBytes := p.BlockHash
	return TransactionKey{
		Index: p.TransactionIdx,
		Hash:  &hashBytes,
	}
}

func (rsc *RawCheckpointSubmission) GetSubmissionKey() SubmissionKey {
	var keys []*TransactionKey
	k1 := toTransactionKey(&rsc.Proof1)
	keys = append(keys, &k1)
	k2 := toTransactionKey(&rsc.Proof2)
	keys = append(keys, &k2)
	return SubmissionKey{
		Key: keys,
	}
}

func (rsc *RawCheckpointSubmission) GetSubmissionData(epochNum uint64) SubmissionData {

	tBytes := [][]byte{rsc.Proof1.TransactionBytes, rsc.Proof2.TransactionBytes}
	return SubmissionData{
		Submitter:      rsc.Submitter.Bytes(),
		Btctransaction: tBytes,
		Epoch:          epochNum,
	}
}

func (sk *SubmissionKey) GetKeyBlockHashes() []*types.BTCHeaderHashBytes {
	var hashes []*types.BTCHeaderHashBytes

	for _, k := range sk.Key {
		h := k.Hash
		hashes = append(hashes, h)
	}

	return hashes
}

func GetEpochIndexKey(e uint64) []byte {
	return sdk.Uint64ToBigEndian(e)
}

func NewEmptyEpochData(rawCheckpointBytes []byte) EpochData {
	return EpochData{
		Key:           []*SubmissionKey{},
		Status:        Signed,
		RawCheckpoint: rawCheckpointBytes,
	}
}

func (s *EpochData) AppendKey(k SubmissionKey) {
	key := &k
	s.Key = append(s.Key, key)
}
