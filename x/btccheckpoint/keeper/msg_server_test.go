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
	initialLighClientDepth int64,
	epoch uint64,
) *TestKeepers {
	lc := btcctypes.NewMockBTCLightClientKeeper(initialLighClientDepth)

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

func (k *TestKeepers) getUnconfirmedSubmissions() []btcctypes.SubmissionKey {
	return k.BTCCheckpoint.GetAllUnconfirmedSubmissions(k.SdkCtx)
}

func (k *TestKeepers) getConfirmedSubmissions() []btcctypes.SubmissionKey {
	return k.BTCCheckpoint.GetAllConfirmedSubmissions(k.SdkCtx)
}

func (k *TestKeepers) getFinalizedSubmissions() []btcctypes.SubmissionKey {
	return k.BTCCheckpoint.GetAllFinalizedSubmissions(k.SdkCtx)
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
	defaultParams := btcctypes.DefaultParams()
	kDeep := defaultParams.BtcConfirmationDepth
	raw := RandomRawCheckpointDataForEpoch(epoch)

	blck1 := dg.CreateBlock(1, 7, 7, raw.FirstPart)

	blck2 := dg.CreateBlock(2, 14, 3, raw.SecondPart)

	tk := InitTestKeepers(t, int64(kDeep)-1, epoch)

	msg := GenerateMessageWithRandomSubmitter([]*dg.BlockCreationResult{blck1, blck2})

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

func TestSubmitValidNewCheckpoint(t *testing.T) {
	rand.Seed(time.Now().Unix())
	epoch := uint64(1)
	defaultParams := btcctypes.DefaultParams()
	kDeep := defaultParams.BtcConfirmationDepth
	raw := RandomRawCheckpointDataForEpoch(epoch)
	blck1 := dg.CreateBlock(1, 7, 7, raw.FirstPart)
	blck2 := dg.CreateBlock(2, 14, 3, raw.SecondPart)

	// here we will only have valid unconfirmed submissions
	tk := InitTestKeepers(t, int64(kDeep)-1, epoch)

	msg := GenerateMessageWithRandomSubmitter([]*dg.BlockCreationResult{blck1, blck2})

	_, err := tk.insertProofMsg(msg)

	if err != nil {
		// fatal as other tests will panic if this fails
		t.Fatalf("Unexpected message processing error: %v", err)
	}

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

	allUnconfirmedSubmissions := tk.getUnconfirmedSubmissions()

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
	raw := RandomRawCheckpointDataForEpoch(epoch)

	blck1 := dg.CreateBlock(1, 7, 7, raw.FirstPart)
	blck2 := dg.CreateBlock(2, 14, 3, raw.SecondPart)

	// here we will only have valid unconfirmed submissions
	// here we will only have valid unconfirmed submissions
	tk := InitTestKeepers(t, int64(kDeep)-1, epoch)

	msg := GenerateMessageWithRandomSubmitter([]*dg.BlockCreationResult{blck1, blck2})

	_, err := tk.insertProofMsg(msg)

	if err != nil {
		t.Errorf("Unexpected message processing error: %v", err)
	}

	// TODO customs Equality for submission keys
	unc := tk.getUnconfirmedSubmissions()

	if len(unc) != 1 {
		t.Errorf("Unexpected missing unconfirmed submissions")
	}

	// Now we will return depth enough for moving submission to confirmed
	tk.BTCLightClient.SetDepth(int64(kDeep))

	// fire tip change callback
	tk.onTipChange()
	// TODO customs Equality for submission keys to check this are really keys
	// we are looking for
	unc = tk.getUnconfirmedSubmissions()
	conf := tk.getConfirmedSubmissions()

	if len(unc) != 0 {
		t.Errorf("Unexpected not promoted submission")
	}

	if len(conf) != 1 {
		t.Errorf("Unexpected missing confirmed submission")
	}

	ed := tk.getEpochData(epoch)

	if ed == nil || ed.Status != btcctypes.Confirmed {
		t.Errorf("Epoch Data missing of in unexpected state")
	}

	tk.BTCLightClient.SetDepth(int64(wDeep))
	tk.onTipChange()

	unc = tk.getUnconfirmedSubmissions()
	conf = tk.getConfirmedSubmissions()
	fin := tk.getFinalizedSubmissions()

	if len(unc) != 0 {
		t.Errorf("Unexpected not promoted unconfirmed submission")
	}

	if len(conf) != 0 {
		t.Errorf("Unexpected not promoted confirmed submission")
	}

	if len(fin) != 1 {
		t.Errorf("Unexpected missing finalized submission")
	}

	ed = tk.getEpochData(epoch)

	if ed == nil || ed.Status != btcctypes.Finalized {
		t.Errorf("Epoch Data missing of in unexpected state")
	}
}
