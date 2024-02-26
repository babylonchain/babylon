package types

import (
	"encoding/hex"

	"github.com/cosmos/cosmos-sdk/types"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
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

// NewSubmissionKeyResponse parses a SubmissionKey into a query response submission key struct.
func NewSubmissionKeyResponse(sk SubmissionKey) (skr *SubmissionKeyResponse, err error) {
	if len(sk.Key) != 2 {
		return nil, status.Errorf(codes.Internal, "bad submission key %+v, does not have 2 keys", sk)
	}

	k1, k2 := sk.Key[0], sk.Key[1]
	return &SubmissionKeyResponse{
		FirstTxBlockHash:  k1.Hash.MarshalHex(),
		FirstTxIndex:      k1.Index,
		SecondTxBlockHash: k2.Hash.MarshalHex(),
		SecondTxIndex:     k2.Index,
	}, nil
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
