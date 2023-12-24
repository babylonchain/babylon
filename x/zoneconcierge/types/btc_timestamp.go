package types

import (
	"context"
	"fmt"
	"math/big"

	errorsmod "cosmossdk.io/errors"
	"github.com/btcsuite/btcd/wire"
	tmcrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	sdk "github.com/cosmos/cosmos-sdk/types"

	txformat "github.com/babylonchain/babylon/btctxformatter"
	"github.com/babylonchain/babylon/crypto/bls12381"
	bbn "github.com/babylonchain/babylon/types"
	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	btclckeeper "github.com/babylonchain/babylon/x/btclightclient/keeper"
	checkpointingtypes "github.com/babylonchain/babylon/x/checkpointing/types"
	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"
)

func GetCZHeaderKey(chainID string, height uint64) []byte {
	key := CanonicalChainKey
	key = append(key, []byte(chainID)...)
	key = append(key, sdk.Uint64ToBigEndian(height)...)
	return key
}

func GetEpochInfoKey(epochNumber uint64) []byte {
	epochInfoKey := epochingtypes.EpochInfoKey
	epochInfoKey = append(epochInfoKey, sdk.Uint64ToBigEndian(epochNumber)...)
	return epochInfoKey
}

func GetValSetKey(epochNumber uint64) []byte {
	valSetKey := checkpointingtypes.ValidatorBlsKeySetPrefix
	valSetKey = append(valSetKey, sdk.Uint64ToBigEndian(epochNumber)...)
	return valSetKey
}

func VerifyEpochInfo(epoch *epochingtypes.Epoch, proof *tmcrypto.ProofOps) error {
	// get the Merkle root, i.e., the BlockHash of the sealer header
	root := epoch.SealerAppHash

	// Ensure The epoch medatata is committed to the app_hash of the sealer header
	// NOTE: the proof is generated when sealer header is generated. At that time
	// sealer header hash is not given to epoch metadata. Thus we need to clear the
	// sealer header hash when verifying the proof.
	epoch.SealerAppHash = []byte{}
	epochBytes, err := epoch.Marshal()
	if err != nil {
		return err
	}
	epoch.SealerAppHash = root
	if err := VerifyStore(root, epochingtypes.StoreKey, GetEpochInfoKey(epoch.EpochNumber), epochBytes, proof); err != nil {
		return errorsmod.Wrapf(ErrInvalidMerkleProof, "invalid inclusion proof for epoch metadata: %v", err)
	}

	return nil
}

func VerifyValSet(epoch *epochingtypes.Epoch, valSet *checkpointingtypes.ValidatorWithBlsKeySet, proof *tmcrypto.ProofOps) error {
	valSetBytes, err := valSet.Marshal()
	if err != nil {
		return err
	}
	if err := VerifyStore(epoch.SealerAppHash, checkpointingtypes.StoreKey, GetValSetKey(epoch.EpochNumber), valSetBytes, proof); err != nil {
		return errorsmod.Wrapf(ErrInvalidMerkleProof, "invalid inclusion proof for validator set: %v", err)
	}

	return nil
}

// VerifyEpochSealed verifies that the given `epoch` is sealed by the `rawCkpt` by using the given `proof`
// The verification rules include:
// - basic sanity checks
// - The raw checkpoint's BlockHash is same as the sealer_block_hash of the sealed epoch
// - More than 2/3 (in voting power) validators in the validator set of this epoch have signed sealer_block_hash of the sealed epoch
// - The epoch medatata is committed to the sealer_app_hash of the sealed epoch
// - The validator set is committed to the sealer_app_hash of the sealed epoch
func VerifyEpochSealed(epoch *epochingtypes.Epoch, rawCkpt *checkpointingtypes.RawCheckpoint, proof *ProofEpochSealed) error {
	// nil check
	if epoch == nil {
		return fmt.Errorf("epoch is nil")
	} else if rawCkpt == nil {
		return fmt.Errorf("rawCkpt is nil")
	} else if proof == nil {
		return fmt.Errorf("proof is nil")
	}

	// sanity check
	if err := epoch.ValidateBasic(); err != nil {
		return err
	} else if err := rawCkpt.ValidateBasic(); err != nil {
		return err
	} else if err = proof.ValidateBasic(); err != nil {
		return err
	}

	// ensure epoch number is same in epoch and rawCkpt
	if epoch.EpochNumber != rawCkpt.EpochNum {
		return fmt.Errorf("epoch.EpochNumber (%d) is not equal to rawCkpt.EpochNum (%d)", epoch.EpochNumber, rawCkpt.EpochNum)
	}

	// ensure the raw checkpoint's block_hash is same as the sealer_block_hash of the sealed epoch
	// NOTE: since this proof is assembled by a Babylon node who has verified the checkpoint,
	// the two blockhash values should always be the same, otherwise this Babylon node is malicious.
	// This is different from the checkpoint verification rules in checkpointing,
	// where a checkpoint with valid BLS multisig but different blockhashes signals a dishonest majority equivocation.
	blockHashInCkpt := rawCkpt.BlockHash
	blockHashInSealerHeader := checkpointingtypes.BlockHash(epoch.SealerBlockHash)
	if !blockHashInCkpt.Equal(blockHashInSealerHeader) {
		return fmt.Errorf("BlockHash is not same in rawCkpt (%s) and epoch's SealerHeader (%s)", blockHashInCkpt.String(), blockHashInSealerHeader.String())
	}

	/*
		Ensure more than 2/3 (in voting power) validators of this epoch have signed (epoch_num || block_hash) in the raw checkpoint
	*/
	valSet := &checkpointingtypes.ValidatorWithBlsKeySet{ValSet: proof.ValidatorSet}
	// filter validator set that contributes to the signature
	signerSet, signerSetPower, err := valSet.FindSubsetWithPowerSum(rawCkpt.Bitmap)
	if err != nil {
		return err
	}
	// ensure the signerSet has > 2/3 voting power
	if signerSetPower*3 <= valSet.GetTotalPower()*2 {
		return checkpointingtypes.ErrInsufficientVotingPower
	}
	// verify BLS multisig
	signedMsgBytes := rawCkpt.SignedMsg()
	ok, err := bls12381.VerifyMultiSig(*rawCkpt.BlsMultiSig, signerSet.GetBLSKeySet(), signedMsgBytes)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("BLS signature does not match the public key")
	}

	// Ensure The epoch medatata is committed to the app_hash of the sealer header
	if err := VerifyEpochInfo(epoch, proof.ProofEpochInfo); err != nil {
		return err
	}

	// Ensure The validator set is committed to the app_hash of the sealer header
	if err := VerifyValSet(epoch, valSet, proof.ProofEpochValSet); err != nil {
		return err
	}

	return nil
}

func VerifyCZHeaderInEpoch(header *IndexedHeader, epoch *epochingtypes.Epoch, proof *tmcrypto.ProofOps) error {
	// nil check
	if header == nil {
		return fmt.Errorf("header is nil")
	} else if epoch == nil {
		return fmt.Errorf("epoch is nil")
	} else if proof == nil {
		return fmt.Errorf("proof is nil")
	}

	// sanity check
	if err := header.ValidateBasic(); err != nil {
		return err
	} else if err := epoch.ValidateBasic(); err != nil {
		return err
	}

	// ensure epoch number is same in epoch and CZ header
	if epoch.EpochNumber != header.BabylonEpoch {
		return fmt.Errorf("epoch.EpochNumber (%d) is not equal to header.BabylonEpoch (%d)", epoch.EpochNumber, header.BabylonEpoch)
	}

	// get the Merkle root, i.e., the BlockHash of the sealer header
	root := epoch.SealerAppHash

	// Ensure The header is committed to the BlockHash of the sealer header
	headerBytes, err := header.Marshal()
	if err != nil {
		return err
	}

	if err := VerifyStore(root, StoreKey, GetCZHeaderKey(header.ChainId, header.Height), headerBytes, proof); err != nil {
		return errorsmod.Wrapf(ErrInvalidMerkleProof, "invalid inclusion proof for CZ header: %v", err)
	}

	return nil
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
		btcHeaderBytes := bbn.NewBTCHeaderBytesFromBlockHeader(btcHeaders[i])
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

func (ts *BTCTimestamp) Verify(
	ctx context.Context,
	btclcKeeper *btclckeeper.Keeper,
	wValue uint64,
	ckptTag txformat.BabylonTag,
) error {
	// BTC net
	btcNet := btclcKeeper.GetBTCNet()

	// verify and insert all BTC headers
	headersBytes := []bbn.BTCHeaderBytes{}
	for _, headerInfo := range ts.BtcHeaders {
		headerBytes := bbn.NewBTCHeaderBytesFromBlockHeader(headerInfo.Header.ToBlockHeader())
		headersBytes = append(headersBytes, headerBytes)
	}
	if err := btclcKeeper.InsertHeaders(ctx, headersBytes); err != nil {
		return err
	}

	// get BTC headers that include the checkpoint, and ensure at least 1 of them is w-deep
	btcHeadersWithCkpt := []*wire.BlockHeader{}
	wDeep := false
	for _, key := range ts.BtcSubmissionKey.Key {
		header := btclcKeeper.GetHeaderByHash(ctx, key.Hash)
		if header == nil {
			return fmt.Errorf("header corresponding to the inclusion proof is not on BTC light client")
		}
		btcHeadersWithCkpt = append(btcHeadersWithCkpt, header.Header.ToBlockHeader())

		depth, err := btclcKeeper.MainChainDepth(ctx, header.Hash)
		if err != nil {
			return err
		}
		if depth >= wValue {
			wDeep = true
		}
	}
	if !wDeep {
		return fmt.Errorf("checkpoint is not w-deep")
	}

	// perform stateless checks that do not rely on BTC light client
	return ts.VerifyStateless(btcHeadersWithCkpt, btcNet.PowLimit, ckptTag)
}

func (ts *BTCTimestamp) VerifyStateless(
	btcHeadersWithCkpt []*wire.BlockHeader,
	powLimit *big.Int,
	ckptTag txformat.BabylonTag,
) error {
	// ensure raw checkpoint corresponds to the epoch
	if ts.EpochInfo.EpochNumber != ts.RawCheckpoint.EpochNum {
		return fmt.Errorf("epoch number in epoch metadata and raw checkpoint is not same")
	}

	if len(ts.BtcSubmissionKey.Key) != txformat.NumberOfParts {
		return fmt.Errorf("incorrect number of txs for a checkpoint")
	}

	// verify the checkpoint txs are committed to the two headers
	err := VerifyEpochSubmitted(ts.RawCheckpoint, ts.Proof.ProofEpochSubmitted, btcHeadersWithCkpt, powLimit, ckptTag)
	if err != nil {
		return err
	}

	// verify the epoch is sealed
	if err := VerifyEpochSealed(ts.EpochInfo, ts.RawCheckpoint, ts.Proof.ProofEpochSealed); err != nil {
		return err
	}

	// verify CZ header is committed to the epoch
	if err := VerifyCZHeaderInEpoch(ts.Header, ts.EpochInfo, ts.Proof.ProofCzHeaderInEpoch); err != nil {
		return err
	}

	return nil
}
