package keeper_test

import (
	"bytes"
	"context"
	"math/rand"
	"testing"
	"time"

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

func b2TxIdx(m *btcctypes.MsgInsertBTCSpvProof) uint32 {
	return m.Proofs[1].BtcTransactionIndex
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

func (k *TestKeepers) setEpoch(epoch uint64) {
	k.Checkpointing.SetEpoch(epoch)
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

func TestRejectDuplicatedSubmission(t *testing.T) {
	rand.Seed(time.Now().Unix())
	epoch := uint64(1)
	raw := dg.RandomRawCheckpointDataForEpoch(epoch)

	blck1 := dg.CreateBlock(1, 7, 7, raw.FirstPart)

	blck2 := dg.CreateBlock(2, 14, 3, raw.SecondPart)

	tk := InitTestKeepers(t, epoch)

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
	rand.Seed(time.Now().Unix())
	epoch := uint64(1)
	raw := dg.RandomRawCheckpointDataForEpoch(epoch)

	blck1 := dg.CreateBlock(1, 7, 7, raw.FirstPart)
	blck2 := dg.CreateBlock(2, 14, 3, raw.SecondPart)

	tk := InitTestKeepers(t, epoch)

	msg := dg.GenerateMessageWithRandomSubmitter([]*dg.BlockCreationResult{blck1, blck2})

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
	raw := dg.RandomRawCheckpointDataForEpoch(epoch)

	blck1 := dg.CreateBlock(1, 7, 7, raw.FirstPart)
	blck2 := dg.CreateBlock(2, 14, 3, raw.SecondPart)

	tk := InitTestKeepers(t, epoch)

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
	rand.Seed(time.Now().Unix())
	epoch := uint64(1)
	raw := dg.RandomRawCheckpointDataForEpoch(epoch)
	blck1 := dg.CreateBlock(1, 7, 7, raw.FirstPart)
	blck2 := dg.CreateBlock(2, 14, 3, raw.SecondPart)

	// here we will only have valid unconfirmed submissions
	tk := InitTestKeepers(t, epoch)

	msg := dg.GenerateMessageWithRandomSubmitter([]*dg.BlockCreationResult{blck1, blck2})

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
	raw := dg.RandomRawCheckpointDataForEpoch(epoch)
	blck1 := dg.CreateBlock(1, 7, 7, raw.FirstPart)
	blck2 := dg.CreateBlock(2, 14, 3, raw.SecondPart)

	// here we will only have valid unconfirmed submissions
	tk := InitTestKeepers(t, epoch)

	msg := dg.GenerateMessageWithRandomSubmitter([]*dg.BlockCreationResult{blck1, blck2})

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
	raw := dg.RandomRawCheckpointDataForEpoch(epoch)
	epoch1Block1 := dg.CreateBlock(1, 7, 7, raw.FirstPart)
	epoch1Block2 := dg.CreateBlock(2, 14, 3, raw.SecondPart)

	// here we will only have valid unconfirmed submissions
	tk := InitTestKeepers(t, epoch)
	msg := dg.GenerateMessageWithRandomSubmitter([]*dg.BlockCreationResult{epoch1Block1, epoch1Block2})

	// Now we will return depth enough for moving submission to be submitted
	tk.BTCLightClient.SetDepth(epoch1Block1.HeaderBytes.Hash(), int64(5))
	tk.BTCLightClient.SetDepth(epoch1Block2.HeaderBytes.Hash(), int64(4))

	_, err := tk.insertProofMsg(msg)

	require.NoErrorf(t, err, "Unexpected message processing error: %v", err)

	epoch2 := uint64(2)
	raw2 := dg.RandomRawCheckpointDataForEpoch(epoch2)
	epoch2Block1 := dg.CreateBlock(1, 19, 2, raw2.FirstPart)
	epoch2Block2 := dg.CreateBlock(2, 14, 7, raw2.SecondPart)
	// Submitting checkpoints for epoch 2, there should be at least one submission
	// for epoch 1, with headers deeper in chain that in this new submission
	tk.Checkpointing.SetEpoch(epoch2)
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
	rand.Seed(time.Now().Unix())
	tk := InitTestKeepers(t, uint64(1))

	msg1 := dg.GenerateMessageWithRandomSubmitterForEpoch(1)
	tk.BTCLightClient.SetDepth(b1Hash(msg1), int64(5))
	tk.BTCLightClient.SetDepth(b2Hash(msg1), int64(4))
	_, err := tk.insertProofMsg(msg1)
	require.NoError(t, err, "failed to insert submission for epoch 1")

	msg1a := dg.GenerateMessageWithRandomSubmitterForEpoch(1)
	tk.BTCLightClient.SetDepth(b1Hash(msg1a), int64(4))
	tk.BTCLightClient.SetDepth(b2Hash(msg1a), int64(5))
	_, err = tk.insertProofMsg(msg1a)
	require.NoError(t, err, "failed to insert submission for epoch 1")

	tk.setEpoch(2)
	msg2 := dg.GenerateMessageWithRandomSubmitterForEpoch(1)
	tk.BTCLightClient.SetDepth(b1Hash(msg2), int64(3))
	tk.BTCLightClient.SetDepth(b2Hash(msg2), int64(2))
	_, err = tk.insertProofMsg(msg2)
	require.NoError(t, err, "failed to insert submission for epoch 2")

	msg2a := dg.GenerateMessageWithRandomSubmitterForEpoch(1)
	tk.BTCLightClient.SetDepth(b1Hash(msg2a), int64(3))
	tk.BTCLightClient.SetDepth(b2Hash(msg2a), int64(2))
	_, err = tk.insertProofMsg(msg2a)
	require.NoError(t, err, "failed to insert submission for epoch 2")

	tk.setEpoch(3)
	msg3 := dg.GenerateMessageWithRandomSubmitterForEpoch(1)
	tk.BTCLightClient.SetDepth(b1Hash(msg3), int64(1))
	tk.BTCLightClient.SetDepth(b2Hash(msg3), int64(0))
	_, err = tk.insertProofMsg(msg3)
	require.NoError(t, err, "failed to insert submission for epoch 3")

	msg3a := dg.GenerateMessageWithRandomSubmitterForEpoch(1)
	tk.BTCLightClient.SetDepth(b1Hash(msg3a), int64(1))
	tk.BTCLightClient.SetDepth(b2Hash(msg3a), int64(0))
	_, err = tk.insertProofMsg(msg3a)
	require.NoError(t, err, "failed to insert submission for epoch 3")

	for i := 1; i <= 3; i++ {
		// all 3 epoch must have two  submissions
		ed := tk.getEpochData(uint64(i))
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
		ed := tk.getEpochData(uint64(i))
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
		ed := tk.getEpochData(uint64(i))
		require.NotNil(t, ed)
		require.Len(t, ed.Key, 0)
		require.EqualValues(t, ed.Status, btcctypes.Submitted)
	}
}

func TestLeaveOnlyBestSubmissionWhenEpochFinalized(t *testing.T) {
	rand.Seed(time.Now().Unix())
	tk := InitTestKeepers(t, uint64(1))
	defaultParams := btcctypes.DefaultParams()
	wDeep := defaultParams.CheckpointFinalizationTimeout

	msg1 := dg.GenerateMessageWithRandomSubmitterForEpoch(1)
	tk.BTCLightClient.SetDepth(b1Hash(msg1), int64(1))
	tk.BTCLightClient.SetDepth(b2Hash(msg1), int64(0))
	_, err := tk.insertProofMsg(msg1)
	require.NoError(t, err, "failed to insert submission")

	msg2 := dg.GenerateMessageWithRandomSubmitterForEpoch(1)
	tk.BTCLightClient.SetDepth(b1Hash(msg2), int64(1))
	tk.BTCLightClient.SetDepth(b2Hash(msg2), int64(0))
	_, err = tk.insertProofMsg(msg2)
	require.NoError(t, err, "failed to insert submission")

	msg3 := dg.GenerateMessageWithRandomSubmitterForEpoch(1)
	tk.BTCLightClient.SetDepth(b1Hash(msg3), int64(1))
	tk.BTCLightClient.SetDepth(b2Hash(msg3), int64(0))
	_, err = tk.insertProofMsg(msg3)
	require.NoError(t, err, "failed to insert submission")

	ed := tk.getEpochData(uint64(1))
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

	ed = tk.getEpochData(uint64(1))
	require.NotNil(t, ed)
	require.Len(t, ed.Key, 1)
	require.Equal(t, ed.Status, btcctypes.Finalized)

	finalSubKey := ed.Key[0]

	require.Equal(t, finalSubKey.Key[0].Hash, b1Hash(msg3))
	require.Equal(t, finalSubKey.Key[1].Hash, b2Hash(msg3))
}

func TestTxIdxShouldBreakTies(t *testing.T) {
	rand.Seed(time.Now().Unix())
	tk := InitTestKeepers(t, uint64(1))
	defaultParams := btcctypes.DefaultParams()
	wDeep := defaultParams.CheckpointFinalizationTimeout

	msg1 := dg.GenerateMessageWithRandomSubmitterForEpoch(1)
	tk.BTCLightClient.SetDepth(b1Hash(msg1), int64(1))
	tk.BTCLightClient.SetDepth(b2Hash(msg1), int64(0))
	_, err := tk.insertProofMsg(msg1)
	require.NoError(t, err, "failed to insert submission")

	msg2 := dg.GenerateMessageWithRandomSubmitterForEpoch(1)
	tk.BTCLightClient.SetDepth(b1Hash(msg2), int64(1))
	tk.BTCLightClient.SetDepth(b2Hash(msg2), int64(0))
	_, err = tk.insertProofMsg(msg2)
	require.NoError(t, err, "failed to insert submission")

	ed := tk.getEpochData(uint64(1))
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

	ed = tk.getEpochData(uint64(1))
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
	rand.Seed(time.Now().Unix())
	epoch := uint64(1)
	defaultParams := btcctypes.DefaultParams()
	kDeep := defaultParams.BtcConfirmationDepth
	wDeep := defaultParams.CheckpointFinalizationTimeout
	raw := dg.RandomRawCheckpointDataForEpoch(epoch)

	blck1 := dg.CreateBlock(1, 7, 7, raw.FirstPart)
	blck2 := dg.CreateBlock(2, 14, 3, raw.SecondPart)

	tk := InitTestKeepers(t, epoch)

	msg := dg.GenerateMessageWithRandomSubmitter([]*dg.BlockCreationResult{blck1, blck2})

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
