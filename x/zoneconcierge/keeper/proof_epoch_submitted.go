package keeper

import (
	"fmt"
	"math/big"

	txformat "github.com/babylonchain/babylon/btctxformatter"
	"github.com/babylonchain/babylon/types"
	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	checkpointingtypes "github.com/babylonchain/babylon/x/checkpointing/types"
	"github.com/btcsuite/btcd/wire"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ProveEpochSubmitted generates proof that the epoch's checkpoint is submitted to BTC
// i.e., the two `TransactionInfo`s for the checkpoint
func (k Keeper) ProveEpochSubmitted(ctx sdk.Context, sk btcctypes.SubmissionKey) ([]*btcctypes.TransactionInfo, error) {
	bestSubmissionData := k.btccKeeper.GetSubmissionData(ctx, sk)
	if bestSubmissionData == nil {
		return nil, fmt.Errorf("the best submission key for epoch %d has no submission data", bestSubmissionData.Epoch)
	}
	return bestSubmissionData.TxsInfo, nil
}

// VerifyEpochSubmitted verifies whether an epoch's checkpoint has been included in BTC or not
// verifications include:
// - basic sanity checks
// - Merkle proofs in txsInfo are valid
// - the raw ckpt decoded from txsInfo is same as the expected rawCkpt
func VerifyEpochSubmitted(rawCkpt *checkpointingtypes.RawCheckpoint, txsInfo []*btcctypes.TransactionInfo, btcHeaders []*wire.BlockHeader, powLimit *big.Int, babylonTag txformat.BabylonTag) error {
	// basic sanity check
	if rawCkpt == nil {
		return fmt.Errorf("rawCkpt is nil")
	} else if len(txsInfo) != txformat.NumberOfParts {
		return fmt.Errorf("txsInfo contains %d parts rather than %d", len(txsInfo), txformat.NumberOfParts)
	} else if len(btcHeaders) != txformat.NumberOfParts {
		return fmt.Errorf("btcHeaders contains %d parts rather than %d", len(btcHeaders), txformat.NumberOfParts)
	}

	// sanity check of each tx info
	for _, txInfo := range txsInfo {
		if err := txInfo.ValidateBasic(); err != nil {
			return err
		}
	}

	// verify Merkle proofs for each tx info
	parsedProofs := []*btcctypes.ParsedProof{}
	for i, txInfo := range txsInfo {
		btcHeaderBytes := types.NewBTCHeaderBytesFromBlockHeader(btcHeaders[i])
		parsedProof, err := btcctypes.ParseProof(
			txInfo.Transaction,
			txInfo.Key.Index,
			txInfo.Proof,
			&btcHeaderBytes,
			powLimit,
		)
		if err != nil {
			return err
		}
		parsedProofs = append(parsedProofs, parsedProof)
	}

	// decode parsedProof to checkpoint data
	checkpointData := [][]byte{}
	for i, proof := range parsedProofs {
		data, err := txformat.GetCheckpointData(
			babylonTag,
			txformat.CurrentVersion,
			uint8(i),
			proof.OpReturnData,
		)

		if err != nil {
			return err
		}
		checkpointData = append(checkpointData, data)
	}
	rawCkptData, err := txformat.ConnectParts(txformat.CurrentVersion, checkpointData[0], checkpointData[1])
	if err != nil {
		return err
	}
	decodedRawCkpt, err := checkpointingtypes.FromBTCCkptBytesToRawCkpt(rawCkptData)
	if err != nil {
		return err
	}

	// check if decodedRawCkpt is same as the expected rawCkpt
	if !decodedRawCkpt.Equal(rawCkpt) {
		return fmt.Errorf("the decoded rawCkpt (%v) is different from the expected rawCkpt (%v)", decodedRawCkpt, rawCkpt)
	}

	return nil
}
