package keeper_test

import (
	"encoding/hex"
	"math/rand"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	dg "github.com/babylonchain/babylon/testutil/datagen"
	"github.com/babylonchain/babylon/x/btccheckpoint/types"
)

func TestBtcCheckpointInfo(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	epoch := uint64(1)
	raw, btcRaw := dg.RandomRawCheckpointDataForEpoch(r, epoch)

	blck1BabylonOpReturnIdx, blck2BabylonOpReturnIdx := uint32(7), uint32(3)
	blck1 := dg.CreateBlock(r, 1, 7, blck1BabylonOpReturnIdx, raw.FirstPart)
	blck2 := dg.CreateBlock(r, 2, 14, blck2BabylonOpReturnIdx, raw.SecondPart)

	tk := InitTestKeepers(t)

	blockResults := []*dg.BlockCreationResult{blck1, blck2}
	proofs := dg.BlockCreationResultToProofs(blockResults)
	msg := dg.GenerateMessageWithRandomSubmitter(blockResults)

	tk.BTCLightClient.SetDepth(blck1.HeaderBytes.Hash(), uint64(1))
	tk.BTCLightClient.SetDepth(blck2.HeaderBytes.Hash(), uint64(1))

	_, err := tk.insertProofMsg(msg)
	require.NoErrorf(t, err, "Unexpected message processing error: %v", err)

	// test the GRPC client call.
	resp, err := tk.BTCCheckpoint.BtcCheckpointInfo(tk.SdkCtx, &types.QueryBtcCheckpointInfoRequest{EpochNum: epoch})
	require.NoErrorf(t, err, "Unexpected message processing error: %v", err)

	// gather info for verifying with response.
	btcInfo := resp.Info
	blkHeight, err := tk.BTCLightClient.BlockHeight(tk.Ctx, nil)
	require.NoErrorf(t, err, "Unexpected message processing error: %v", err)

	rawSubmission, err := types.ParseSubmission(msg, tk.BTCCheckpoint.GetPowLimit(), tk.BTCCheckpoint.GetExpectedTag(tk.SdkCtx))
	require.NoErrorf(t, err, "Unexpected message processing error: %v", err)
	blk1 := rawSubmission.GetFirstBlockHash()
	blk2 := rawSubmission.GetSecondBlockHash()

	require.Equal(t, btcInfo.EpochNumber, epoch)
	require.Equal(t, btcInfo.BestSubmissionBtcBlockHeight, blkHeight)
	require.Equal(t, btcInfo.BestSubmissionBtcBlockHash, blk1.MarshalHex())

	require.Equal(t, len(btcInfo.BestSubmissionTransactions), 2)
	tx0 := btcInfo.BestSubmissionTransactions[0]
	require.Equal(t, tx0.Index, blck1BabylonOpReturnIdx)
	require.Equal(t, tx0.Hash, blk1.MarshalHex())
	require.Equal(t, tx0.Transaction, blck1.Transactions[blck1BabylonOpReturnIdx])
	require.Equal(t, tx0.Proof, hex.EncodeToString(proofs[0].MerkleNodes))

	tx1 := btcInfo.BestSubmissionTransactions[1]
	require.Equal(t, tx1.Index, blck2BabylonOpReturnIdx)
	require.Equal(t, tx1.Hash, blk2.MarshalHex())
	require.Equal(t, tx1.Transaction, blck2.Transactions[blck2BabylonOpReturnIdx])
	require.Equal(t, tx1.Proof, hex.EncodeToString(proofs[1].MerkleNodes))

	require.Equal(t, len(btcInfo.BestSubmissionVigilanteAddressList), 1)
	require.Equal(t, btcInfo.BestSubmissionVigilanteAddressList[0].Reporter, rawSubmission.Reporter.String())
	require.Equal(t, btcInfo.BestSubmissionVigilanteAddressList[0].Submitter, sdk.AccAddress(btcRaw.SubmitterAddress).String())
}
