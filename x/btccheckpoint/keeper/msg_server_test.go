package keeper_test

import (
	"bytes"
	"context"
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/babylonchain/babylon/testutil/datagen"

	dg "github.com/babylonchain/babylon/testutil/datagen"
	keepertest "github.com/babylonchain/babylon/testutil/keeper"
	bbn "github.com/babylonchain/babylon/types"
	bkeeper "github.com/babylonchain/babylon/x/btccheckpoint/keeper"
	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	"github.com/btcsuite/btcd/chaincfg"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

type TestKeepers struct {
	SdkCtx         sdk.Context
	Ctx            context.Context
	BTCLightClient *btcctypes.MockBTCLightClientKeeper
	Checkpointing  *btcctypes.MockCheckpointingKeeper
	BTCCheckpoint  *bkeeper.Keeper
	MsgSrv         btcctypes.MsgServer
}

func b1Hash(m *btcctypes.MsgInsertBTCSpvProof) *bbn.BTCHeaderHashBytes {
	return m.Proofs[0].ConfirmingBtcHeader.Hash()
}

func b1TxIdx(m *btcctypes.MsgInsertBTCSpvProof) uint32 {
	return m.Proofs[0].BtcTransactionIndex
}

func b2Hash(m *btcctypes.MsgInsertBTCSpvProof) *bbn.BTCHeaderHashBytes {
	return m.Proofs[1].ConfirmingBtcHeader.Hash()
}

//nolint:unused
func b2TxIdx(m *btcctypes.MsgInsertBTCSpvProof) uint32 {
	return m.Proofs[1].BtcTransactionIndex
}

func InitTestKeepers(
	t *testing.T,
) *TestKeepers {
	lc := btcctypes.NewMockBTCLightClientKeeper()

	cc := btcctypes.NewMockCheckpointingKeeper()

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

func (k *TestKeepers) GetEpochData(e uint64) *btcctypes.EpochData {
	return k.BTCCheckpoint.GetEpochData(k.SdkCtx, e)
}

func (k *TestKeepers) getSubmissionData(key btcctypes.SubmissionKey) *btcctypes.SubmissionData {
	return k.BTCCheckpoint.GetSubmissionData(k.SdkCtx, key)
}

func (k *TestKeepers) onTipChange() {
	k.BTCCheckpoint.OnTipChange(k.SdkCtx)
}

func TestRejectDuplicatedSubmission(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	epoch := uint64(1)
	raw, _ := dg.RandomRawCheckpointDataForEpoch(r, epoch)

	blck1 := dg.CreateBlock(r, 1, 7, 7, raw.FirstPart)

	blck2 := dg.CreateBlock(r, 2, 14, 3, raw.SecondPart)

	tk := InitTestKeepers(t)

	msg := dg.GenerateMessageWithRandomSubmitter([]*dg.BlockCreationResult{blck1, blck2})

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
	r := rand.New(rand.NewSource(time.Now().Unix()))
	epoch := uint64(1)
	raw, _ := dg.RandomRawCheckpointDataForEpoch(r, epoch)

	blck1 := dg.CreateBlock(r, 1, 7, 7, raw.FirstPart)
	blck2 := dg.CreateBlock(r, 2, 14, 3, raw.SecondPart)

	tk := InitTestKeepers(t)

	msg := dg.GenerateMessageWithRandomSubmitter([]*dg.BlockCreationResult{blck1, blck2})

	_, err := tk.insertProofMsg(msg)

	require.ErrorContainsf(t, err, btcctypes.ErrInvalidHeader.Error(), "Processing should return invalid header error")

	// even if one header is known, submission should still be considered invalid
	tk.BTCLightClient.SetDepth(blck1.HeaderBytes.Hash(), int64(1))

	_, err = tk.insertProofMsg(msg)

	require.ErrorContainsf(t, err, btcctypes.ErrInvalidHeader.Error(), "Processing should return invalid header error")
}

func TestRejectSubmissionsNotOnMainchain(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	epoch := uint64(1)
	raw, _ := dg.RandomRawCheckpointDataForEpoch(r, epoch)

	blck1 := dg.CreateBlock(r, 1, 7, 7, raw.FirstPart)
	blck2 := dg.CreateBlock(r, 2, 14, 3, raw.SecondPart)

	tk := InitTestKeepers(t)

	msg := dg.GenerateMessageWithRandomSubmitter([]*dg.BlockCreationResult{blck1, blck2})

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
	r := rand.New(rand.NewSource(time.Now().Unix()))
	epoch := uint64(1)
	raw, rawBtcCheckpoint := dg.RandomRawCheckpointDataForEpoch(r, epoch)
	blck1 := dg.CreateBlock(r, 1, 7, 7, raw.FirstPart)
	blck2 := dg.CreateBlock(r, 2, 14, 3, raw.SecondPart)

	// here we will only have valid unconfirmed submissions
	tk := InitTestKeepers(t)

	msg := dg.GenerateMessageWithRandomSubmitter([]*dg.BlockCreationResult{blck1, blck2})

	// Now we will return depth enough for moving submission to be submitted
	tk.BTCLightClient.SetDepth(blck1.HeaderBytes.Hash(), int64(1))
	tk.BTCLightClient.SetDepth(blck2.HeaderBytes.Hash(), int64(1))

	_, err := tk.insertProofMsg(msg)

	require.NoErrorf(t, err, "Unexpected message processing error: %v", err)

	ed := tk.GetEpochData(epoch)

	if len(ed.Key) == 0 {
		t.Errorf("There should be at least one key in epoch %d", epoch)
	}

	if ed.Status != btcctypes.Submitted {
		t.Errorf("Epoch should be in submitted state after processing message")
	}

	submissionKey := ed.Key[0]

	submissionData := tk.getSubmissionData(*submissionKey)

	if submissionData == nil {
		t.Fatalf("Unexpected missing submission")
	}

	if submissionData.Epoch != epoch {
		t.Errorf("Submission data with invalid epoch")
	}

	if len(submissionData.TxsInfo) != 2 {
		t.Errorf("Submission data with invalid TransactionInfo")
	}

	if !bytes.Equal(rawBtcCheckpoint.SubmitterAddress, submissionData.VigilanteAddresses.Submitter) {
		t.Errorf("Submission data does not contain expected submitter address")
	}

	for i, txInfo := range submissionData.TxsInfo {
		require.Equal(t, submissionKey.Key[i].Index, txInfo.Key.Index)
		require.True(t, submissionKey.Key[i].Hash.Eq(txInfo.Key.Hash))
		require.Equal(t, msg.Proofs[i].BtcTransaction, txInfo.Transaction)
		require.Equal(t, msg.Proofs[i].MerkleNodes, txInfo.Proof)
	}

	ed1 := tk.GetEpochData(epoch)

	// TODO Add custom equal fo submission key and transaction key to check
	// it is expected key
	if len(ed1.Key) == 0 {
		t.Errorf("Unexpected missing unconfirmed submissions")
	}
}

func TestRejectSubmissionWithoutSubmissionsForPreviousEpoch(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	epoch := uint64(2)
	raw, _ := dg.RandomRawCheckpointDataForEpoch(r, epoch)
	blck1 := dg.CreateBlock(r, 1, 7, 7, raw.FirstPart)
	blck2 := dg.CreateBlock(r, 2, 14, 3, raw.SecondPart)

	// here we will only have valid unconfirmed submissions
	tk := InitTestKeepers(t)

	msg := dg.GenerateMessageWithRandomSubmitter([]*dg.BlockCreationResult{blck1, blck2})

	// Now we will return depth enough for moving submission to be submitted
	tk.BTCLightClient.SetDepth(blck1.HeaderBytes.Hash(), int64(0))
	tk.BTCLightClient.SetDepth(blck2.HeaderBytes.Hash(), int64(1))

	_, err := tk.insertProofMsg(msg)

	require.ErrorContainsf(
		t,
		err,
		btcctypes.ErrNoCheckpointsForPreviousEpoch.Error(),
		"Processing msg should return ErrNoCheckpointsForPreviousEpoch error",
	)
}

func TestRejectSubmissionWithoutAncestorsOnMainchainInPreviousEpoch(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	epoch := uint64(1)
	raw, _ := dg.RandomRawCheckpointDataForEpoch(r, epoch)
	epoch1Block1 := dg.CreateBlock(r, 1, 7, 7, raw.FirstPart)
	epoch1Block2 := dg.CreateBlock(r, 2, 14, 3, raw.SecondPart)

	// here we will only have valid unconfirmed submissions
	tk := InitTestKeepers(t)
	msg := dg.GenerateMessageWithRandomSubmitter([]*dg.BlockCreationResult{epoch1Block1, epoch1Block2})

	// Now we will return depth enough for moving submission to be submitted
	tk.BTCLightClient.SetDepth(epoch1Block1.HeaderBytes.Hash(), int64(5))
	tk.BTCLightClient.SetDepth(epoch1Block2.HeaderBytes.Hash(), int64(4))

	_, err := tk.insertProofMsg(msg)

	require.NoErrorf(t, err, "Unexpected message processing error: %v", err)

	epoch2 := uint64(2)
	raw2, _ := dg.RandomRawCheckpointDataForEpoch(r, epoch2)
	epoch2Block1 := dg.CreateBlock(r, 1, 19, 2, raw2.FirstPart)
	epoch2Block2 := dg.CreateBlock(r, 2, 14, 7, raw2.SecondPart)
	msg2 := dg.GenerateMessageWithRandomSubmitter([]*dg.BlockCreationResult{epoch2Block1, epoch2Block2})

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

func TestClearChildEpochsWhenNoParenNotOnMainChain(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	tk := InitTestKeepers(t)

	msg1 := dg.GenerateMessageWithRandomSubmitterForEpoch(r, 1)
	tk.BTCLightClient.SetDepth(b1Hash(msg1), int64(5))
	tk.BTCLightClient.SetDepth(b2Hash(msg1), int64(4))
	_, err := tk.insertProofMsg(msg1)
	require.NoError(t, err, "failed to insert submission for epoch 1")

	msg1a := dg.GenerateMessageWithRandomSubmitterForEpoch(r, 1)
	tk.BTCLightClient.SetDepth(b1Hash(msg1a), int64(4))
	tk.BTCLightClient.SetDepth(b2Hash(msg1a), int64(5))
	_, err = tk.insertProofMsg(msg1a)
	require.NoError(t, err, "failed to insert submission for epoch 1")

	msg2 := dg.GenerateMessageWithRandomSubmitterForEpoch(r, 2)
	tk.BTCLightClient.SetDepth(b1Hash(msg2), int64(3))
	tk.BTCLightClient.SetDepth(b2Hash(msg2), int64(2))
	_, err = tk.insertProofMsg(msg2)
	require.NoError(t, err, "failed to insert submission for epoch 2")

	msg2a := dg.GenerateMessageWithRandomSubmitterForEpoch(r, 2)
	tk.BTCLightClient.SetDepth(b1Hash(msg2a), int64(3))
	tk.BTCLightClient.SetDepth(b2Hash(msg2a), int64(2))
	_, err = tk.insertProofMsg(msg2a)
	require.NoError(t, err, "failed to insert submission for epoch 2")

	msg3 := dg.GenerateMessageWithRandomSubmitterForEpoch(r, 3)
	tk.BTCLightClient.SetDepth(b1Hash(msg3), int64(1))
	tk.BTCLightClient.SetDepth(b2Hash(msg3), int64(0))
	_, err = tk.insertProofMsg(msg3)
	require.NoError(t, err, "failed to insert submission for epoch 3")

	msg3a := dg.GenerateMessageWithRandomSubmitterForEpoch(r, 3)
	tk.BTCLightClient.SetDepth(b1Hash(msg3a), int64(1))
	tk.BTCLightClient.SetDepth(b2Hash(msg3a), int64(0))
	_, err = tk.insertProofMsg(msg3a)
	require.NoError(t, err, "failed to insert submission for epoch 3")

	for i := 1; i <= 3; i++ {
		// all 3 epoch must have two  submissions
		ed := tk.GetEpochData(uint64(i))
		require.NotNil(t, ed)
		require.Len(t, ed.Key, 2)
		require.EqualValues(t, ed.Status, btcctypes.Submitted)
	}

	// Due to reorg one submission from epoch 1 lands on fork, which means it is no
	// longer vaiable. It should be pruned. Other subbmissions should be left
	// intact
	tk.BTCLightClient.SetDepth(b2Hash(msg1), -1)

	tk.onTipChange()

	for i := 1; i <= 3; i++ {
		ed := tk.GetEpochData(uint64(i))
		require.NotNil(t, ed)

		if i == 1 {
			// forked submission got pruned
			require.Len(t, ed.Key, 1)
		} else {
			// other submissions still have parent so they are left intact
			require.Len(t, ed.Key, 2)
		}
		require.EqualValues(t, ed.Status, btcctypes.Submitted)
	}

	// second submission from epoch 1 got orphaned. Clear it, and submissions from
	// child epochs
	tk.BTCLightClient.SetDepth(b2Hash(msg1a), -1)

	tk.onTipChange()

	for i := 1; i <= 3; i++ {
		// all 3 epoch must have two  submissions
		ed := tk.GetEpochData(uint64(i))
		require.NotNil(t, ed)
		require.Len(t, ed.Key, 0)
		require.EqualValues(t, ed.Status, btcctypes.Submitted)
	}
}

func TestLeaveOnlyBestSubmissionWhenEpochFinalized(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	tk := InitTestKeepers(t)
	defaultParams := btcctypes.DefaultParams()
	wDeep := defaultParams.CheckpointFinalizationTimeout

	msg1 := dg.GenerateMessageWithRandomSubmitterForEpoch(r, 1)
	tk.BTCLightClient.SetDepth(b1Hash(msg1), int64(1))
	tk.BTCLightClient.SetDepth(b2Hash(msg1), int64(0))
	_, err := tk.insertProofMsg(msg1)
	require.NoError(t, err, "failed to insert submission")

	msg2 := dg.GenerateMessageWithRandomSubmitterForEpoch(r, 1)
	tk.BTCLightClient.SetDepth(b1Hash(msg2), int64(1))
	tk.BTCLightClient.SetDepth(b2Hash(msg2), int64(0))
	_, err = tk.insertProofMsg(msg2)
	require.NoError(t, err, "failed to insert submission")

	msg3 := dg.GenerateMessageWithRandomSubmitterForEpoch(r, 1)
	tk.BTCLightClient.SetDepth(b1Hash(msg3), int64(1))
	tk.BTCLightClient.SetDepth(b2Hash(msg3), int64(0))
	_, err = tk.insertProofMsg(msg3)
	require.NoError(t, err, "failed to insert submission")

	ed := tk.GetEpochData(uint64(1))
	require.NotNil(t, ed)
	require.Len(t, ed.Key, 3)

	// deepest submission is submission in msg3
	tk.BTCLightClient.SetDepth(b1Hash(msg1), int64(wDeep))
	tk.BTCLightClient.SetDepth(b2Hash(msg1), int64(wDeep+1))
	tk.BTCLightClient.SetDepth(b1Hash(msg2), int64(wDeep+2))
	tk.BTCLightClient.SetDepth(b2Hash(msg2), int64(wDeep+3))
	tk.BTCLightClient.SetDepth(b1Hash(msg3), int64(wDeep+4))
	tk.BTCLightClient.SetDepth(b2Hash(msg3), int64(wDeep+5))

	tk.onTipChange()

	ed = tk.GetEpochData(uint64(1))
	require.NotNil(t, ed)
	require.Len(t, ed.Key, 1)
	require.Equal(t, ed.Status, btcctypes.Finalized)

	finalSubKey := ed.Key[0]

	require.Equal(t, finalSubKey.Key[0].Hash, b1Hash(msg3))
	require.Equal(t, finalSubKey.Key[1].Hash, b2Hash(msg3))
}

func TestTxIdxShouldBreakTies(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	tk := InitTestKeepers(t)
	defaultParams := btcctypes.DefaultParams()
	wDeep := defaultParams.CheckpointFinalizationTimeout

	msg1 := dg.GenerateMessageWithRandomSubmitterForEpoch(r, 1)
	tk.BTCLightClient.SetDepth(b1Hash(msg1), int64(1))
	tk.BTCLightClient.SetDepth(b2Hash(msg1), int64(0))
	_, err := tk.insertProofMsg(msg1)
	require.NoError(t, err, "failed to insert submission")

	msg2 := dg.GenerateMessageWithRandomSubmitterForEpoch(r, 1)
	tk.BTCLightClient.SetDepth(b1Hash(msg2), int64(1))
	tk.BTCLightClient.SetDepth(b2Hash(msg2), int64(0))
	_, err = tk.insertProofMsg(msg2)
	require.NoError(t, err, "failed to insert submission")

	ed := tk.GetEpochData(uint64(1))
	require.NotNil(t, ed)
	require.Len(t, ed.Key, 2)

	// Both submissions have the same depth the most fresh block i.e
	// it is the same block
	// When finalizing the one with lower TxIx should be treated as better
	tk.BTCLightClient.SetDepth(b1Hash(msg1), int64(wDeep))
	tk.BTCLightClient.SetDepth(b2Hash(msg1), int64(wDeep+1))
	tk.BTCLightClient.SetDepth(b1Hash(msg2), int64(wDeep))
	tk.BTCLightClient.SetDepth(b2Hash(msg2), int64(wDeep+3))

	tk.onTipChange()

	ed = tk.GetEpochData(uint64(1))
	require.NotNil(t, ed)
	require.Len(t, ed.Key, 1)
	require.Equal(t, ed.Status, btcctypes.Finalized)
	finalSubKey := ed.Key[0]

	// There is small chance that we can draw the same transactions indexes, which
	// cannot happend in real life i.e in real life if block has the same depth
	// then this is one block and transaction indexes.
	// In that case this test is noop to avoid spoutius failures
	if b1TxIdx(msg1) < b1TxIdx(msg2) {
		require.Equal(t, finalSubKey.Key[0].Hash, b1Hash(msg1))
	} else if b1TxIdx(msg2) < b1TxIdx(msg1) {
		require.Equal(t, finalSubKey.Key[0].Hash, b1Hash(msg2))
	}
}

func TestStateTransitionOfValidSubmission(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	epoch := uint64(1)
	defaultParams := btcctypes.DefaultParams()
	kDeep := defaultParams.BtcConfirmationDepth
	wDeep := defaultParams.CheckpointFinalizationTimeout
	raw, _ := dg.RandomRawCheckpointDataForEpoch(r, epoch)

	blck1 := dg.CreateBlock(r, 1, 7, 7, raw.FirstPart)
	blck2 := dg.CreateBlock(r, 2, 14, 3, raw.SecondPart)

	tk := InitTestKeepers(t)

	msg := dg.GenerateMessageWithRandomSubmitter([]*dg.BlockCreationResult{blck1, blck2})

	// Now we will return depth enough for moving submission to confirmed
	tk.BTCLightClient.SetDepth(blck1.HeaderBytes.Hash(), int64(1))
	tk.BTCLightClient.SetDepth(blck2.HeaderBytes.Hash(), int64(1))

	_, err := tk.insertProofMsg(msg)

	if err != nil {
		t.Errorf("Unexpected message processing error: %v", err)
	}

	// TODO customs Equality for submission keys
	ed := tk.GetEpochData(epoch)

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
	ed = tk.GetEpochData(epoch)

	if len(ed.Key) != 1 {
		t.Errorf("Unexpected missing submission")
	}

	if ed.Status != btcctypes.Confirmed {
		t.Errorf("Epoch should be in submitted stated")
	}

	tk.BTCLightClient.SetDepth(blck1.HeaderBytes.Hash(), int64(wDeep))
	tk.BTCLightClient.SetDepth(blck2.HeaderBytes.Hash(), int64(wDeep))

	tk.onTipChange()

	ed = tk.GetEpochData(epoch)

	if ed == nil || ed.Status != btcctypes.Finalized {
		t.Errorf("Epoch Data missing of in unexpected state")
	}
}

func FuzzConfirmAndDinalizeManyEpochs(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 20)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		tk := InitTestKeepers(t)
		defaultParams := btcctypes.DefaultParams()
		kDeep := defaultParams.BtcConfirmationDepth
		wDeep := defaultParams.CheckpointFinalizationTimeout

		numFinalizedEpochs := r.Intn(10) + 1
		numConfirmedEpochs := r.Intn(5) + 1
		numSubmittedEpochs := 1

		finalizationDepth := math.MaxUint32
		confirmationDepth := wDeep - 1
		sumbissionDepth := kDeep - 1

		numOfEpochs := numFinalizedEpochs + numConfirmedEpochs + numSubmittedEpochs

		bestSumbissionInfos := make(map[uint64]uint64)

		for i := 1; i <= numOfEpochs; i++ {
			epoch := uint64(i)
			raw, _ := dg.RandomRawCheckpointDataForEpoch(r, epoch)
			numSubmissionsPerEpoch := r.Intn(3) + 1

			for j := 1; j <= numSubmissionsPerEpoch; j++ {
				numTx1 := uint32(r.Intn(30) + 10)
				numTx2 := uint32(r.Intn(30) + 10)
				blck1 := dg.CreateBlock(r, 0, numTx1, 1, raw.FirstPart)
				blck2 := dg.CreateBlock(r, 0, numTx2, 2, raw.SecondPart)

				msg := dg.GenerateMessageWithRandomSubmitter([]*dg.BlockCreationResult{blck1, blck2})

				if epoch <= uint64(numFinalizedEpochs) {
					tk.BTCLightClient.SetDepth(blck1.HeaderBytes.Hash(), int64(finalizationDepth))
					finalizationDepth = finalizationDepth - 1
					tk.BTCLightClient.SetDepth(blck2.HeaderBytes.Hash(), int64(finalizationDepth))

					// first submission is always deepest one, and second block is the most recent one
					if j == 1 {
						bestSumbissionInfos[epoch] = uint64(finalizationDepth)
					}
					finalizationDepth = finalizationDepth - 1
				} else if epoch <= uint64(numFinalizedEpochs+numConfirmedEpochs) {
					tk.BTCLightClient.SetDepth(blck1.HeaderBytes.Hash(), int64(confirmationDepth))
					confirmationDepth = confirmationDepth - 1
					tk.BTCLightClient.SetDepth(blck2.HeaderBytes.Hash(), int64(confirmationDepth))
					// first submission is always deepest one, and second block is the most recent one
					if j == 1 {
						bestSumbissionInfos[epoch] = uint64(confirmationDepth)
					}
					confirmationDepth = confirmationDepth - 1
				} else {
					tk.BTCLightClient.SetDepth(blck1.HeaderBytes.Hash(), int64(sumbissionDepth))
					sumbissionDepth = sumbissionDepth - 1
					tk.BTCLightClient.SetDepth(blck2.HeaderBytes.Hash(), int64(sumbissionDepth))
					// first submission is always deepest one, and second block is the most recent one
					if j == 1 {
						bestSumbissionInfos[epoch] = uint64(sumbissionDepth)
					}
					sumbissionDepth = sumbissionDepth - 1
				}

				_, err := tk.insertProofMsg(msg)
				require.NoError(t, err, "failed to insert submission for epoch %d", epoch)
			}
		}

		// Check that all epochs are in submitted state
		for i := 1; i <= numOfEpochs; i++ {
			epoch := uint64(i)
			ed := tk.GetEpochData(epoch)
			require.NotNil(t, ed)
			require.Equal(t, ed.Status, btcctypes.Submitted)
		}

		// Fire up tip change callback. All epochs should reach their correct state
		tk.onTipChange()

		for i := 1; i <= numOfEpochs; i++ {
			epoch := uint64(i)
			ed := tk.GetEpochData(epoch)
			require.NotNil(t, ed)

			if epoch <= uint64(numFinalizedEpochs) {
				require.Equal(t, ed.Status, btcctypes.Finalized)
				// finalized epochs should have only best submission
				require.Equal(t, len(ed.Key), 1)
			} else if epoch <= uint64(numFinalizedEpochs+numConfirmedEpochs) {
				require.Equal(t, ed.Status, btcctypes.Confirmed)
			} else {
				require.Equal(t, ed.Status, btcctypes.Submitted)
			}

			bestSubInfo := tk.BTCCheckpoint.GetEpochBestSubmissionBtcInfo(tk.SdkCtx, ed)
			require.NotNil(t, bestSubInfo)
			expectedBestSubmissionDepth := bestSumbissionInfos[epoch]
			require.Equal(t, bestSubInfo.SubmissionDepth(), expectedBestSubmissionDepth)
		}
	})
}
