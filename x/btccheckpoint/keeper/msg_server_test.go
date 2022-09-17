package keeper_test

import (
	"bytes"
	"encoding/hex"
	"math/rand"
	"testing"
	"time"

	txformat "github.com/babylonchain/babylon/btctxformatter"
	dg "github.com/babylonchain/babylon/testutil/datagen"
	keepertest "github.com/babylonchain/babylon/testutil/keeper"
	bkeeper "github.com/babylonchain/babylon/x/btccheckpoint/keeper"
	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	"github.com/btcsuite/btcd/chaincfg"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func BlockCreationResultToProofs(inputs []*dg.BlockCreationResult) []*btcctypes.BTCSpvProof {
	var spvs []*btcctypes.BTCSpvProof

	for _, input := range inputs {
		headerBytes := input.HeaderBytes

		var txBytes [][]byte

		for _, t := range input.Transactions {
			tbytes, err := hex.DecodeString(t)

			if err != nil {
				panic("Inputs should contain valid hex encoded transactions")
			}

			txBytes = append(txBytes, tbytes)
		}

		spv, err := btcctypes.SpvProofFromHeaderAndTransactions(headerBytes, txBytes, uint(input.BbnTxIndex))

		if err != nil {
			panic("Inputs should contain valid spv hex encoded data")
		}

		spvs = append(spvs, spv)
	}

	return spvs
}

type testCheckpointData struct {
	epoch            uint64
	lastCommitHash   []byte
	bitmap           []byte
	blsSig           []byte
	submitterAddress []byte
}

func getRandomCheckpointDataForEpoch(e uint64) testCheckpointData {
	return testCheckpointData{
		epoch:            e,
		lastCommitHash:   dg.GenRandomByteArray(txformat.LastCommitHashLength),
		bitmap:           dg.GenRandomByteArray(txformat.BitMapLength),
		blsSig:           dg.GenRandomByteArray(txformat.BlsSigLength),
		submitterAddress: dg.GenRandomByteArray(txformat.AddressLength),
	}
}

// both f and s must be parts retrived from txformat.Encode
func getExpectedOpReturn(f []byte, s []byte) []byte {
	firstPartNoHeader, err := txformat.GetCheckpointData(
		txformat.MainTag(),
		txformat.CurrentVersion,
		0,
		f,
	)

	if err != nil {
		panic("ExpectedOpReturn provided first part should be valid checkpoint data")
	}

	secondPartNoHeader, err := txformat.GetCheckpointData(
		txformat.MainTag(),
		txformat.CurrentVersion,
		1,
		s,
	)

	if err != nil {
		panic("ExpectedOpReturn provided second part should be valid checkpoint data")
	}

	connected, err := txformat.ComposeRawCheckpointData(txformat.CurrentVersion, firstPartNoHeader, secondPartNoHeader)

	if err != nil {
		panic("ExpectedOpReturn parts should be connected")
	}

	return connected
}

func TestSubmitValidNewCheckpoint(t *testing.T) {
	rand.Seed(time.Now().Unix())
	epoch := uint64(1)
	defaultParams := btcctypes.DefaultParams()
	kDeep := defaultParams.BtcConfirmationDepth
	checkpointData := getRandomCheckpointDataForEpoch(epoch)

	data1, data2 := txformat.MustEncodeCheckpointData(
		txformat.MainTag(),
		txformat.CurrentVersion,
		checkpointData.epoch,
		checkpointData.lastCommitHash,
		checkpointData.bitmap,
		checkpointData.blsSig,
		checkpointData.submitterAddress,
	)

	blck1 := dg.CreateBlock(1, 7, 7, data1)
	blck2 := dg.CreateBlock(2, 14, 3, data2)

	expectedOpReturn := getExpectedOpReturn(data1, data2)

	// here we will only have valid unconfirmed submissions
	lc := btcctypes.NewMockBTCLightClientKeeper(int64(kDeep) - 1)

	cc := btcctypes.NewMockCheckpointingKeeper(epoch)

	k, ctx := keepertest.NewBTCCheckpointKeeper(t, lc, cc, chaincfg.SimNetParams.PowLimit)

	proofs := BlockCreationResultToProofs([]*dg.BlockCreationResult{blck1, blck2})

	pk, _ := dg.NewPV().GetPubKey()

	address := sdk.AccAddress(pk.Address().Bytes())

	msg := btcctypes.MsgInsertBTCSpvProof{
		Proofs:    proofs,
		Submitter: address.String(),
	}

	srv := bkeeper.NewMsgServerImpl(*k)

	sdkCtx := sdk.WrapSDKContext(ctx)

	_, err := srv.InsertBTCSpvProof(sdkCtx, &msg)

	if err != nil {
		// fatal as other tests will panic if this fails
		t.Fatalf("Unexpected message processing error: %v", err)
	}

	ed := k.GetEpochData(ctx, epoch)

	if len(ed.Key) == 0 {
		t.Errorf("There should be at least one key in epoch %d", epoch)
	}

	if ed.Status != btcctypes.Submitted {
		t.Errorf("Epoch should be in submitted state after processing message")
	}

	if !bytes.Equal(expectedOpReturn, ed.RawCheckpoint) {
		t.Errorf("Epoch does not contain expected op return data")
	}

	submissionKey := ed.Key[0]

	submissionData := k.GetSubmissionData(ctx, *submissionKey)

	if submissionData == nil {
		t.Fatalf("Unexpected missing submission")
	}

	if submissionData.Epoch != epoch {
		t.Errorf("Submission data with invalid epoch")
	}

	allUnconfirmedSubmissions := k.GetAllUnconfirmedSubmissions(ctx)

	// TODO Add custom equal fo submission key and transaction key to check
	// it is expected key
	if len(allUnconfirmedSubmissions) == 0 {
		t.Errorf("Unexpected missing unconfirmed submissions")
	}
}

func TestStateTransitionOfValidSubmission(t *testing.T) {
	rand.Seed(time.Now().Unix())
	epoch := uint64(1)
	defaultParams := btcctypes.DefaultParams()
	kDeep := defaultParams.BtcConfirmationDepth
	wDeep := defaultParams.CheckpointFinalizationTimeout
	checkpointData := getRandomCheckpointDataForEpoch(epoch)

	data1, data2 := txformat.MustEncodeCheckpointData(
		txformat.MainTag(),
		txformat.CurrentVersion,
		checkpointData.epoch,
		checkpointData.lastCommitHash,
		checkpointData.bitmap,
		checkpointData.blsSig,
		checkpointData.submitterAddress,
	)

	blck1 := dg.CreateBlock(1, 7, 7, data1)
	blck2 := dg.CreateBlock(2, 14, 3, data2)

	// here we will only have valid unconfirmed submissions
	lc := btcctypes.NewMockBTCLightClientKeeper(int64(kDeep) - 1)
	cc := btcctypes.NewMockCheckpointingKeeper(epoch)

	k, ctx := keepertest.NewBTCCheckpointKeeper(t, lc, cc, chaincfg.SimNetParams.PowLimit)

	proofs := BlockCreationResultToProofs([]*dg.BlockCreationResult{blck1, blck2})

	pk, _ := dg.NewPV().GetPubKey()

	address := sdk.AccAddress(pk.Address().Bytes())

	msg := btcctypes.MsgInsertBTCSpvProof{
		Proofs:    proofs,
		Submitter: address.String(),
	}

	srv := bkeeper.NewMsgServerImpl(*k)

	sdkCtx := sdk.WrapSDKContext(ctx)

	_, err := srv.InsertBTCSpvProof(sdkCtx, &msg)

	if err != nil {
		t.Errorf("Unexpected message processing error: %v", err)
	}

	// TODO customs Equality for submission keys
	unc := k.GetAllUnconfirmedSubmissions(ctx)

	if len(unc) != 1 {
		t.Errorf("Unexpected missing unconfirmed submissions")
	}

	// Now we will return depth enough for moving submission to confirmed
	lc.SetDepth(int64(kDeep))

	// fire tip change callback
	k.OnTipChange(ctx)
	// TODO customs Equality for submission keys to check this are really keys
	// we are looking for
	unc = k.GetAllUnconfirmedSubmissions(ctx)
	conf := k.GetAllConfirmedSubmissions(ctx)

	if len(unc) != 0 {
		t.Errorf("Unexpected not promoted submission")
	}

	if len(conf) != 1 {
		t.Errorf("Unexpected missing confirmed submission")
	}

	ed := k.GetEpochData(ctx, epoch)

	if ed == nil || ed.Status != btcctypes.Confirmed {
		t.Errorf("Epoch Data missing of in unexpected state")
	}

	lc.SetDepth(int64(wDeep))
	k.OnTipChange(ctx)

	unc = k.GetAllUnconfirmedSubmissions(ctx)
	conf = k.GetAllConfirmedSubmissions(ctx)
	fin := k.GetAllFinalizedSubmissions(ctx)

	if len(unc) != 0 {
		t.Errorf("Unexpected not promoted unconfirmed submission")
	}

	if len(conf) != 0 {
		t.Errorf("Unexpected not promoted confirmed submission")
	}

	if len(fin) != 1 {
		t.Errorf("Unexpected missing finalized submission")
	}

	ed = k.GetEpochData(ctx, epoch)

	if ed == nil || ed.Status != btcctypes.Finalized {
		t.Errorf("Epoch Data missing of in unexpected state")
	}
}
