package datagen

import (
	bbl "github.com/babylonchain/babylon/types"
	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
)

func SpvProofFromHeaderAndTransactions(headerBytes []byte, transactions [][]byte, transactionIdx uint) (*btcctypes.BTCSpvProof, error) {
	proof, e := bbl.CreateProofForIdx(transactions, transactionIdx)

	if e != nil {
		return nil, e
	}

	var flatProof []byte

	for _, h := range proof {
		flatProof = append(flatProof, h.CloneBytes()...)
	}

	spvProof := btcctypes.BTCSpvProof{
		BtcTransaction:      transactions[transactionIdx],
		BtcTransactionIndex: uint32(transactionIdx),
		MerkleNodes:         flatProof,
		ConfirmingBtcHeader: headerBytes,
	}

	return &spvProof, nil
}
