package types

import (
	"encoding/hex"
	"fmt"

	"github.com/babylonchain/babylon/btctxformatter"
	"github.com/babylonchain/babylon/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Semantically valid checkpoint submission with:
// - valid submitter address
// - at least 2 parsed proof
// Modelling proofs as separate Proof1 and Proof2, as this is more explicit than
// []*ParsedProof.
type RawCheckpointSubmission struct {
	Reporter       sdk.AccAddress
	Proof1         ParsedProof
	Proof2         ParsedProof
	CheckpointData btctxformatter.RawBtcCheckpoint
}

// SubmissionBtcInfo encapsualte important information about submission posistion
// on btc ledger
type SubmissionBtcInfo struct {
	SubmissionKey SubmissionKey
	// Depth of the oldest btc header of the submission
	OldestBlockDepth uint64

	// Depth of the youngest btc header of the submission
	YoungestBlockDepth uint64

	// Hash of the youngest btc header of the submission
	YoungestBlockHash types.BTCHeaderHashBytes

	// Index of the latest transaction in youngest submission block
	YoungestBlockInitialTxIdx uint32
}

func NewRawCheckpointSubmission(
	a sdk.AccAddress,
	p1 ParsedProof,
	p2 ParsedProof,
	checkpointData btctxformatter.RawBtcCheckpoint,
) RawCheckpointSubmission {
	r := RawCheckpointSubmission{
		Reporter:       a,
		Proof1:         p1,
		Proof2:         p2,
		CheckpointData: checkpointData,
	}

	return r
}

func (s *RawCheckpointSubmission) GetProofs() []*ParsedProof {
	return []*ParsedProof{&s.Proof1, &s.Proof2}
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

func (rsc *RawCheckpointSubmission) GetSubmissionData(epochNum uint64, txsInfo []*TransactionInfo) SubmissionData {
	return SubmissionData{
		VigilanteAddresses: &CheckpointAddresses{
			Reporter:  rsc.Reporter.Bytes(),
			Submitter: rsc.CheckpointData.SubmitterAddress,
		},
		TxsInfo: txsInfo,
		Epoch:   epochNum,
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

func NewEmptyEpochData() EpochData {
	return EpochData{
		Key:    []*SubmissionKey{},
		Status: Submitted,
	}
}

func (s *EpochData) AppendKey(k SubmissionKey) {
	key := &k
	s.Key = append(s.Key, key)
}

// HappenedAfter returns true if `this` submission happened after `that` submission
func (submission *SubmissionBtcInfo) HappenedAfter(parentEpochSubmission *SubmissionBtcInfo) bool {
	return submission.OldestBlockDepth < parentEpochSubmission.YoungestBlockDepth
}

// SubmissionDepth return depth of the submission. Due to the fact that submissions
// are splitted between several btc blocks, in Babylon subbmission depth is the depth
// of the youngest btc block
func (submission *SubmissionBtcInfo) SubmissionDepth() uint64 {
	return submission.YoungestBlockDepth
}

func (newSubmission *SubmissionBtcInfo) IsBetterThan(currentBestSubmission *SubmissionBtcInfo) bool {
	if newSubmission.SubmissionDepth() > currentBestSubmission.SubmissionDepth() {
		return true
	}

	if newSubmission.SubmissionDepth() < currentBestSubmission.SubmissionDepth() {
		return false
	}

	// at this point we know that both submissions youngest part happens to be in
	// the same block. To resolve the tie we need to take into account index of
	// latest transaction of the submissions
	return newSubmission.YoungestBlockInitialTxIdx < currentBestSubmission.YoungestBlockInitialTxIdx
}

func NewTransactionInfo(txKey *TransactionKey, txBytes []byte, proof []byte) *TransactionInfo {
	return &TransactionInfo{
		Key:         txKey,
		Transaction: txBytes,
		Proof:       proof,
	}
}

func (ti *TransactionInfo) ValidateBasic() error {
	if ti.Key == nil {
		return fmt.Errorf("key in TransactionInfo is nil")
	}
	if ti.Transaction == nil {
		return fmt.Errorf("transaction in TransactionInfo is nil")
	}
	if ti.Proof == nil {
		return fmt.Errorf("proof in TransactionInfo is nil")
	}
	return nil
}

func NewSpvProofFromHexBytes(c codec.Codec, proof string) (*BTCSpvProof, error) {
	bytes, err := hex.DecodeString(proof)

	if err != nil {
		return nil, err
	}

	var p BTCSpvProof
	err = c.Unmarshal(bytes, &p)

	if err != nil {
		return nil, err
	}

	return &p, nil
}
