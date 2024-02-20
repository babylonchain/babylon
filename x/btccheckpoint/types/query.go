package types

import (
	"encoding/hex"

	"github.com/cosmos/cosmos-sdk/types"
)

// ToResponse parses a TransactionInfo into a query response tx info struct.
func (ti *TransactionInfo) ToResponse() *TransactionInfoResponse {
	return &TransactionInfoResponse{
		Key:         ti.Key.ToResponse(),
		Transaction: hex.EncodeToString(ti.Transaction),
		Proof:       hex.EncodeToString(ti.Proof),
	}
}

// ToResponse parses a TransactionKeyResponse into a query response tx key struct.
func (tk *TransactionKey) ToResponse() *TransactionKeyResponse {
	return &TransactionKeyResponse{
		Index: tk.Index,
		Hash:  hex.EncodeToString(*tk.Hash),
	}
}

// ToResponse parses a CheckpointAddresses into a query response checkpoint addresses struct.
func (ca *CheckpointAddresses) ToResponse() *CheckpointAddressesResponse {
	return &CheckpointAddressesResponse{
		Submitter: types.AccAddress(ca.Submitter).String(),
		Reporter:  types.AccAddress(ca.Reporter).String(),
	}
}

// ToResponse parses a BTCCheckpointInfo into a query response for btc checkpoint info struct.
func (b BTCCheckpointInfo) ToResponse() *BTCCheckpointInfoResponse {
	bestSubTxs := make([]*TransactionInfoResponse, len(b.BestSubmissionTransactions))
	for i, tx := range b.BestSubmissionTransactions {
		bestSubTxs[i] = tx.ToResponse()
	}
	bestSubVigAddrs := make([]*CheckpointAddressesResponse, len(b.BestSubmissionVigilanteAddressList))
	for i, addrs := range b.BestSubmissionVigilanteAddressList {
		bestSubVigAddrs[i] = addrs.ToResponse()
	}

	return &BTCCheckpointInfoResponse{
		EpochNumber:                        b.EpochNumber,
		BestSubmissionBtcBlockHeight:       b.BestSubmissionBtcBlockHeight,
		BestSubmissionBtcBlockHash:         hex.EncodeToString(*b.BestSubmissionBtcBlockHash),
		BestSubmissionTransactions:         bestSubTxs,
		BestSubmissionVigilanteAddressList: bestSubVigAddrs,
	}
}
