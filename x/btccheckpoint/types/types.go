package types

import (
	"github.com/babylonchain/babylon/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Semantically valid checkpoint submission with:
// - valid submitter address
// - at least 2 parsed proof
// Modelling proofs as separate Proof1 and Proof2, as this is more explicit than
// []*ParsedProof.
type RawCheckpointSubmission struct {
	Submitter      sdk.AccAddress
	Proof1         ParsedProof
	Proof2         ParsedProof
	checkpointData []byte
}

func NewRawCheckpointSubmission(
	a sdk.AccAddress,
	p1 ParsedProof,
	p2 ParsedProof,
	checkpointData []byte,
) RawCheckpointSubmission {
	r := RawCheckpointSubmission{
		Submitter:      a,
		Proof1:         p1,
		Proof2:         p2,
		checkpointData: checkpointData,
	}

	return r
}

func (s *RawCheckpointSubmission) GetProofs() []*ParsedProof {
	return []*ParsedProof{&s.Proof1, &s.Proof2}
}

func (s *RawCheckpointSubmission) GetRawCheckPointBytes() []byte {
	checkpointDataCopy := make([]byte, len(s.checkpointData))
	// return copy, to avoid someone modifing original
	copy(checkpointDataCopy, s.checkpointData)
	return checkpointDataCopy
}

func (s *RawCheckpointSubmission) GetFirstBlockHash() types.BTCHeaderHashBytes {
	return s.Proof1.BlockHash
}

func (s *RawCheckpointSubmission) GetSecondBlockHash() types.BTCHeaderHashBytes {
	return s.Proof2.BlockHash
}

func (s *RawCheckpointSubmission) InOneBlock() bool {
	fh := s.GetFirstBlockHash()
	sh := s.GetSecondBlockHash()
	return fh.Eq(&sh)
}

func toTransactionKey(p *ParsedProof) TransactionKey {
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

func NewEmptyEpochData(rawCheckpointBytes []byte) EpochData {
	return EpochData{
		Key:           []*SubmissionKey{},
		Status:        Submitted,
		RawCheckpoint: rawCheckpointBytes,
	}
}

func (s *EpochData) AppendKey(k SubmissionKey) {
	key := &k
	s.Key = append(s.Key, key)
}
