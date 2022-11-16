package keeper_test

import (
	"bytes"
	"context"
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
	"github.com/stretchr/testify/require"
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

type TestKeepers struct {
	SdkCtx         sdk.Context
	Ctx            context.Context
	BTCLightClient *btcctypes.MockBTCLightClientKeeper
	Checkpointing  *btcctypes.MockCheckpointingKeeper
	BTCCheckpoint  *bkeeper.Keeper
	MsgSrv         btcctypes.MsgServer
}

type TestRawCheckpointData struct {
	Epoch            uint64
	FirstPart        []byte
	SecondPart       []byte
	ExpectedOpReturn []byte
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
func getExpectedOpReturn(tag txformat.BabylonTag, f []byte, s []byte) []byte {
	firstPartNoHeader, err := txformat.GetCheckpointData(
		tag,
		txformat.CurrentVersion,
		0,
		f,
	)

	if err != nil {
		panic("ExpectedOpReturn provided first part should be valid checkpoint data")
	}

	secondPartNoHeader, err := txformat.GetCheckpointData(
		tag,
		txformat.CurrentVersion,
		1,
		s,
	)

	if err != nil {
		panic("ExpectedOpReturn provided second part should be valid checkpoint data")
	}

	connected, err := txformat.ConnectParts(txformat.CurrentVersion, firstPartNoHeader, secondPartNoHeader)

	if err != nil {
		panic("ExpectedOpReturn parts should be connected")
	}

	return connected
}

func GenerateMessageWithRandomSubmitter(blockResults []*dg.BlockCreationResult) *btcctypes.MsgInsertBTCSpvProof {
	proofs := BlockCreationResultToProofs(blockResults)

	pk, _ := dg.NewPV().GetPubKey()

	address := sdk.AccAddress(pk.Address().Bytes())

	msg := btcctypes.MsgInsertBTCSpvProof{
		Proofs:    proofs,
		Submitter: address.String(),
	}

	return &msg
}

func InitTestKeepers(
	t *testing.T,
	epoch uint64,
) *TestKeepers {
	lc := btcctypes.NewMockBTCLightClientKeeper()

	cc := btcctypes.NewMockCheckpointingKeeper(epoch)

	k, ctx := keepertest.NewBTCCheckpointKeeper(t, lc, cc, chaincfg.SimNetParams.PowLimit)

	srv := bkeeper.NewMsgServerImpl(*k)

	return &TestKeepers{
		SdkCtx:         ctx,
		Ctx:            sdk.WrapSDKContext(ctx),
		BTCLightClient: lc,
		Checkpointing:  cc,
		BTCCheckpoint:  k,
		MsgSrv:         srv,
	}
}

func (k *TestKeepers) insertProofMsg(msg *btcctypes.MsgInsertBTCSpvProof) (*btcctypes.MsgInsertBTCSpvProofResponse, error) {
	return k.MsgSrv.InsertBTCSpvProof(k.Ctx, msg)
}

func (k *TestKeepers) getEpochData(e uint64) *btcctypes.EpochData {
	return k.BTCCheckpoint.GetEpochData(k.SdkCtx, e)
}

func (k *TestKeepers) getSubmissionData(key btcctypes.SubmissionKey) *btcctypes.SubmissionData {
	return k.BTCCheckpoint.GetSubmissionData(k.SdkCtx, key)
}

func (k *TestKeepers) onTipChange() {
	k.BTCCheckpoint.OnTipChange(k.SdkCtx)
}

func RandomRawCheckpointDataForEpoch(e uint64) *TestRawCheckpointData {
	checkpointData := getRandomCheckpointDataForEpoch(e)
	tag := txformat.MainTag(0)

	rawBTCCkpt := &txformat.RawBtcCheckpoint{
		Epoch:            checkpointData.epoch,
		LastCommitHash:   checkpointData.lastCommitHash,
		BitMap:           checkpointData.bitmap,
		SubmitterAddress: checkpointData.submitterAddress,
		BlsSig:           checkpointData.blsSig,
	}
	data1, data2 := txformat.MustEncodeCheckpointData(
		tag,
		txformat.CurrentVersion,
		rawBTCCkpt,
	)

	opReturn := getExpectedOpReturn(tag, data1, data2)

	return &TestRawCheckpointData{
		Epoch:            e,
		FirstPart:        data1,
		SecondPart:       data2,
		ExpectedOpReturn: opReturn,
	}
}

func TestRejectDuplicatedSubmission(t *testing.T) {
	rand.Seed(time.Now().Unix())
	epoch := uint64(1)
	raw := RandomRawCheckpointDataForEpoch(epoch)

	blck1 := dg.CreateBlock(1, 7, 7, raw.FirstPart)

	blck2 := dg.CreateBlock(2, 14, 3, raw.SecondPart)

	tk := InitTestKeepers(t, epoch)

	msg := GenerateMessageWithRandomSubmitter([]*dg.BlockCreationResult{blck1, blck2})

	// Now we will return depth enough for moving submission to be submitted
	tk.BTCLightClient.SetDepth(blck1.HeaderBytes.Hash(), int64(1))
	tk.BTCLightClient.SetDepth(blck2.HeaderBytes.Hash(), int64(1))

	_, err := tk.insertProofMsg(msg)

	if err != nil {
		// fatal as other tests will panic if this fails
		t.Fatalf("Unexpected message processing error: %v", err)
	}

	_, err = tk.insertProofMsg(msg)

	if err == nil {
		t.Fatalf("Submission should have failed due to duplicated submission")
	}

	if err != btcctypes.ErrDuplicatedSubmission {
		t.Fatalf("Error should indicate duplicated submissions")
	}
}

func TestRejectUnknownToBtcLightClient(t *testing.T) {
	rand.Seed(time.Now().Unix())
	epoch := uint64(1)
	raw := RandomRawCheckpointDataForEpoch(epoch)

	blck1 := dg.CreateBlock(1, 7, 7, raw.FirstPart)
	blck2 := dg.CreateBlock(2, 14, 3, raw.SecondPart)

	tk := InitTestKeepers(t, epoch)

	msg := GenerateMessageWithRandomSubmitter([]*dg.BlockCreationResult{blck1, blck2})

	_, err := tk.insertProofMsg(msg)

	require.ErrorContainsf(t, err, btcctypes.ErrInvalidHeader.Error(), "Processing should return invalid header error")

	// even if one header is known, submission should still be considered invalid
	tk.BTCLightClient.SetDepth(blck1.HeaderBytes.Hash(), int64(1))

	_, err = tk.insertProofMsg(msg)

	require.ErrorContainsf(t, err, btcctypes.ErrInvalidHeader.Error(), "Processing should return invalid header error")
}

func TestRejectSubmissionsNotOnMainchain(t *testing.T) {
	rand.Seed(time.Now().Unix())
	epoch := uint64(1)
	raw := RandomRawCheckpointDataForEpoch(epoch)

	blck1 := dg.CreateBlock(1, 7, 7, raw.FirstPart)
	blck2 := dg.CreateBlock(2, 14, 3, raw.SecondPart)

	tk := InitTestKeepers(t, epoch)

	msg := GenerateMessageWithRandomSubmitter([]*dg.BlockCreationResult{blck1, blck2})

	// both headers on fork, fail
	tk.BTCLightClient.SetDepth(blck1.HeaderBytes.Hash(), int64(-1))
	tk.BTCLightClient.SetDepth(blck2.HeaderBytes.Hash(), int64(-1))

	_, err := tk.insertProofMsg(msg)

	require.ErrorContainsf(t, err, btcctypes.ErrInvalidHeader.Error(), "Processing should return invalid header error")

	// one header on fork, one on main chain, fail
	tk.BTCLightClient.SetDepth(blck1.HeaderBytes.Hash(), int64(0))
	tk.BTCLightClient.SetDepth(blck2.HeaderBytes.Hash(), int64(-1))

	_, err = tk.insertProofMsg(msg)

	require.ErrorContainsf(t, err, btcctypes.ErrInvalidHeader.Error(), "Processing should return invalid header error")

	// two headers on main chain, success
	tk.BTCLightClient.SetDepth(blck1.HeaderBytes.Hash(), int64(0))
	tk.BTCLightClient.SetDepth(blck2.HeaderBytes.Hash(), int64(0))

	_, err = tk.insertProofMsg(msg)

	require.NoError(t, err, "Processing msg should succeed")
}

func TestSubmitValidNewCheckpoint(t *testing.T) {
	rand.Seed(time.Now().Unix())
	epoch := uint64(1)
	raw := RandomRawCheckpointDataForEpoch(epoch)
	blck1 := dg.CreateBlock(1, 7, 7, raw.FirstPart)
	blck2 := dg.CreateBlock(2, 14, 3, raw.SecondPart)

	// here we will only have valid unconfirmed submissions
	tk := InitTestKeepers(t, epoch)

	msg := GenerateMessageWithRandomSubmitter([]*dg.BlockCreationResult{blck1, blck2})

	// Now we will return depth enough for moving submission to be submitted
	tk.BTCLightClient.SetDepth(blck1.HeaderBytes.Hash(), int64(1))
	tk.BTCLightClient.SetDepth(blck2.HeaderBytes.Hash(), int64(1))

	_, err := tk.insertProofMsg(msg)

	require.NoErrorf(t, err, "Unexpected message processing error: %v", err)

	ed := tk.getEpochData(epoch)

	if len(ed.Key) == 0 {
		t.Errorf("There should be at least one key in epoch %d", epoch)
	}

	if ed.Status != btcctypes.Submitted {
		t.Errorf("Epoch should be in submitted state after processing message")
	}

	if !bytes.Equal(raw.ExpectedOpReturn, ed.RawCheckpoint) {
		t.Errorf("Epoch does not contain expected op return data")
	}

	submissionKey := ed.Key[0]

	submissionData := tk.getSubmissionData(*submissionKey)

	if submissionData == nil {
		t.Fatalf("Unexpected missing submission")
	}

	if submissionData.Epoch != epoch {
		t.Errorf("Submission data with invalid epoch")
	}

	ed1 := tk.getEpochData(epoch)

	// TODO Add custom equal fo submission key and transaction key to check
	// it is expected key
	if len(ed1.Key) == 0 {
		t.Errorf("Unexpected missing unconfirmed submissions")
	}
}

func TestRejectSubmissionWithoutSubmissionsForPreviousEpoch(t *testing.T) {
	rand.Seed(time.Now().Unix())
	epoch := uint64(1)
	raw := RandomRawCheckpointDataForEpoch(epoch)
	blck1 := dg.CreateBlock(1, 7, 7, raw.FirstPart)
	blck2 := dg.CreateBlock(2, 14, 3, raw.SecondPart)

	// here we will only have valid unconfirmed submissions
	tk := InitTestKeepers(t, epoch)

	msg := GenerateMessageWithRandomSubmitter([]*dg.BlockCreationResult{blck1, blck2})

	// Now we will return depth enough for moving submission to be submitted
	tk.BTCLightClient.SetDepth(blck1.HeaderBytes.Hash(), int64(0))
	tk.BTCLightClient.SetDepth(blck2.HeaderBytes.Hash(), int64(1))
	tk.Checkpointing.SetEpoch(2)

	_, err := tk.insertProofMsg(msg)

	require.ErrorContainsf(
		t,
		err,
		btcctypes.ErrNoCheckpointsForPreviousEpoch.Error(),
		"Processing msg should return ErrNoCheckpointsForPreviousEpoch error",
	)
}

func TestRejectSubmissionWithoutAncestorsOnMainchainInPreviousEpoch(t *testing.T) {
	rand.Seed(time.Now().Unix())
	epoch := uint64(1)
	raw := RandomRawCheckpointDataForEpoch(epoch)
	epoch1Block1 := dg.CreateBlock(1, 7, 7, raw.FirstPart)
	epoch1Block2 := dg.CreateBlock(2, 14, 3, raw.SecondPart)

	// here we will only have valid unconfirmed submissions
	tk := InitTestKeepers(t, epoch)
	msg := GenerateMessageWithRandomSubmitter([]*dg.BlockCreationResult{epoch1Block1, epoch1Block2})

	// Now we will return depth enough for moving submission to be submitted
	tk.BTCLightClient.SetDepth(epoch1Block1.HeaderBytes.Hash(), int64(5))
	tk.BTCLightClient.SetDepth(epoch1Block2.HeaderBytes.Hash(), int64(4))

	_, err := tk.insertProofMsg(msg)

	require.NoErrorf(t, err, "Unexpected message processing error: %v", err)

	epoch2 := uint64(2)
	raw2 := RandomRawCheckpointDataForEpoch(epoch2)
	epoch2Block1 := dg.CreateBlock(1, 19, 2, raw2.FirstPart)
	epoch2Block2 := dg.CreateBlock(2, 14, 7, raw2.SecondPart)
	// Submitting checkpoints for epoch 2, there should be at least one submission
	// for epoch 1, with headers deeper in chain that in this new submission
	tk.Checkpointing.SetEpoch(epoch2)
	msg2 := GenerateMessageWithRandomSubmitter([]*dg.BlockCreationResult{epoch2Block1, epoch2Block2})

	// Both headers are deeper than epoch 1 submission, fail
	tk.BTCLightClient.SetDepth(epoch2Block1.HeaderBytes.Hash(), int64(7))
	tk.BTCLightClient.SetDepth(epoch2Block2.HeaderBytes.Hash(), int64(6))

	_, err = tk.insertProofMsg(msg2)

	require.ErrorContainsf(
		t,
		err,
		btcctypes.ErrProvidedHeaderDoesNotHaveAncestor.Error(),
		"Processing msg should return ErrProvidedHeaderDoesNotHaveAncestor error",
	)

	// one header deeper than headers of previous epoch, one fresher, fail
	tk.BTCLightClient.SetDepth(epoch2Block1.HeaderBytes.Hash(), int64(7))
	tk.BTCLightClient.SetDepth(epoch2Block2.HeaderBytes.Hash(), int64(3))

	_, err = tk.insertProofMsg(msg2)

	require.ErrorContainsf(
		t,
		err,
		btcctypes.ErrProvidedHeaderDoesNotHaveAncestor.Error(),
		"Processing msg should return ErrProvidedHeaderDoesNotHaveAncestor error",
	)

	// one header on the same depth as previous epoch, one fresher, fail
	tk.BTCLightClient.SetDepth(epoch2Block1.HeaderBytes.Hash(), int64(4))
	tk.BTCLightClient.SetDepth(epoch2Block2.HeaderBytes.Hash(), int64(3))

	_, err = tk.insertProofMsg(msg2)

	require.ErrorContainsf(
		t,
		err,
		btcctypes.ErrProvidedHeaderDoesNotHaveAncestor.Error(),
		"Processing msg should return ErrProvidedHeaderDoesNotHaveAncestor error",
	)

	// Both Headers fresher that previous epoch, succeed
	tk.BTCLightClient.SetDepth(epoch2Block1.HeaderBytes.Hash(), int64(3))
	tk.BTCLightClient.SetDepth(epoch2Block2.HeaderBytes.Hash(), int64(2))

	_, err = tk.insertProofMsg(msg2)

	require.NoErrorf(t, err, "Unexpected message processing error: %v", err)

}

func TestStateTransitionOfValidSubmission(t *testing.T) {
	rand.Seed(time.Now().Unix())
	epoch := uint64(1)
	defaultParams := btcctypes.DefaultParams()
	kDeep := defaultParams.BtcConfirmationDepth
	wDeep := defaultParams.CheckpointFinalizationTimeout
	raw := RandomRawCheckpointDataForEpoch(epoch)

	blck1 := dg.CreateBlock(1, 7, 7, raw.FirstPart)
	blck2 := dg.CreateBlock(2, 14, 3, raw.SecondPart)

	tk := InitTestKeepers(t, epoch)

	msg := GenerateMessageWithRandomSubmitter([]*dg.BlockCreationResult{blck1, blck2})

	// Now we will return depth enough for moving submission to confirmed
	tk.BTCLightClient.SetDepth(blck1.HeaderBytes.Hash(), int64(1))
	tk.BTCLightClient.SetDepth(blck2.HeaderBytes.Hash(), int64(1))

	_, err := tk.insertProofMsg(msg)

	if err != nil {
		t.Errorf("Unexpected message processing error: %v", err)
	}

	// TODO customs Equality for submission keys
	ed := tk.getEpochData(epoch)

	if len(ed.Key) != 1 {
		t.Errorf("Unexpected missing submissions")
	}

	if ed.Status != btcctypes.Submitted {
		t.Errorf("Epoch should be in submitted stated")
	}

	// Now we will return depth enough for moving submission to confirmed
	tk.BTCLightClient.SetDepth(blck1.HeaderBytes.Hash(), int64(kDeep))
	tk.BTCLightClient.SetDepth(blck2.HeaderBytes.Hash(), int64(kDeep))

	// fire tip change callback
	tk.onTipChange()
	// TODO customs Equality for submission keys to check this are really keys
	// we are looking for
	ed = tk.getEpochData(epoch)

	if len(ed.Key) != 1 {
		t.Errorf("Unexpected missing submission")
	}

	if ed.Status != btcctypes.Confirmed {
		t.Errorf("Epoch should be in submitted stated")
	}

	tk.BTCLightClient.SetDepth(blck1.HeaderBytes.Hash(), int64(wDeep))
	tk.BTCLightClient.SetDepth(blck2.HeaderBytes.Hash(), int64(wDeep))

	tk.onTipChange()
	ed = tk.getEpochData(epoch)

	if ed == nil || ed.Status != btcctypes.Finalized {
		t.Errorf("Epoch Data missing of in unexpected state")
	}
}

func (k *TestKeepers) insertNBlocks(n int) {
	for i := 1; i <= n; i++ {
		raw := RandomRawCheckpointDataForEpoch(uint64(i))
		blck1 := dg.CreateBlock(1, 7, 7, raw.FirstPart)
		blck2 := dg.CreateBlock(2, 14, 3, raw.SecondPart)

		msg := GenerateMessageWithRandomSubmitter([]*dg.BlockCreationResult{blck1, blck2})
		k.Checkpointing.SetEpoch(uint64(i))
		k.BTCLightClient.SetDepth(blck1.HeaderBytes.Hash(), int64(n-i))
		k.BTCLightClient.SetDepth(blck2.HeaderBytes.Hash(), int64(n-i))

		_, _ = k.insertProofMsg(msg)

	}
}
