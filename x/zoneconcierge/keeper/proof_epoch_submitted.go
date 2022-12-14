package keeper

import (
	"fmt"
	"math/big"

	txformat "github.com/babylonchain/babylon/btctxformatter"
	"github.com/babylonchain/babylon/types"
	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	checkpointingtypes "github.com/babylonchain/babylon/x/checkpointing/types"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ProveEpochSubmitted generates proof that the epoch's checkpoint is submitted to BTC
// i.e., the two `TransactionInfo`s for the checkpoint
func (k Keeper) ProveEpochSubmitted(ctx sdk.Context, sk btcctypes.SubmissionKey) []*btcctypes.TransactionInfo {
	bestSubmissionData := k.btccKeeper.GetSubmissionData(ctx, sk)
	if bestSubmissionData == nil {
		err := fmt.Errorf("the best submission key for epoch has no submission data")
		panic(err) // this can only be a programming error
	}
	return bestSubmissionData.TxsInfo
}

// VerifyEpochSubmitted verifies whether an epoch's checkpoint has been included in BTC or not
// verifications include:
// - basic sanity checks
// - Merkle proofs in txsInfo are valid
// - the raw ckpt decoded from txsInfo is same as the expected rawCkpt
func VerifyEpochSubmitted(rawCkpt *checkpointingtypes.RawCheckpoint, txsInfo []*btcctypes.TransactionInfo, btcHeaders []*wire.BlockHeader, btcNetParams wire.BitcoinNet, babylonTag txformat.BabylonTag) error {
	// basic sanity check
	if rawCkpt == nil {
		return fmt.Errorf("rawCkpt is nil")
	} else if txsInfo == nil {
		return fmt.Errorf("txsInfo is nil")
	} else if len(txsInfo) != txformat.NumberOfParts {
		return fmt.Errorf("txsInfo contains %d parts rather than %d", len(txsInfo), txformat.NumberOfParts)
	} else if btcHeaders == nil {
		return fmt.Errorf("btcHeaders is nil")
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
	powLimit := getPoWLimit(btcNetParams)
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
	var checkpointData [][]byte
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

func getPoWLimit(btcNetParams wire.BitcoinNet) *big.Int {
	switch btcNetParams {
	case wire.MainNet:
		return chaincfg.MainNetParams.PowLimit
	case wire.TestNet:
		return chaincfg.RegressionNetParams.PowLimit
	case wire.TestNet3:
		return chaincfg.TestNet3Params.PowLimit
	case wire.SimNet:
		return chaincfg.SimNetParams.PowLimit
	default:
		panic(fmt.Errorf("invalid btcNetParams: %v", btcNetParams))
	}
}
