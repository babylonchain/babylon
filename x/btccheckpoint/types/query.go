package types

import (
	"encoding/hex"

	"github.com/cosmos/cosmos-sdk/types"
)

// ToResponse parses a TransactionInfo into a query response tx info struct.
func (ti *TransactionInfo) ToResponse() *TransactionInfoResponse {
	return &TransactionInfoResponse{
		Index:       ti.Key.Index,
		Hash:        ti.Key.Hash.MarshalHex(),
		Transaction: hex.EncodeToString(ti.Transaction),
		Proof:       hex.EncodeToString(ti.Proof),
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
		BestSubmissionBtcBlockHash:         b.BestSubmissionBtcBlockHash.MarshalHex(),
		BestSubmissionTransactions:         bestSubTxs,
		BestSubmissionVigilanteAddressList: bestSubVigAddrs,
	}
}
