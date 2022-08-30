package types

import (
	bbn "github.com/babylonchain/babylon/types"
)

func SpvProofFromHeaderAndTransactions(headerBytes []byte, transactions [][]byte, transactionIdx uint) (*BTCSpvProof, error) {
	proof, e := bbn.CreateProofForIdx(transactions, transactionIdx)

	if e != nil {
		return nil, e
	}

	var flatProof []byte

	for _, h := range proof {
		flatProof = append(flatProof, h.CloneBytes()...)
	}

	spvProof := BTCSpvProof{
		BtcTransaction:      transactions[transactionIdx],
		BtcTransactionIndex: uint32(transactionIdx),
		MerkleNodes:         flatProof,
		ConfirmingBtcHeader: headerBytes,
	}

	return &spvProof, nil
}
