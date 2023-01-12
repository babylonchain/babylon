package chain

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	btccheckpointtypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	cttypes "github.com/babylonchain/babylon/x/checkpointing/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"

	txformat "github.com/babylonchain/babylon/btctxformatter"
	bbn "github.com/babylonchain/babylon/types"

	"github.com/babylonchain/babylon/test/e2e/initialization"
	"github.com/babylonchain/babylon/test/e2e/util"
	"github.com/babylonchain/babylon/testutil/datagen"
	blc "github.com/babylonchain/babylon/x/btclightclient/types"

	"github.com/stretchr/testify/require"
)

// TODO for now all commands are not used and left here as an example
// QueryParams extracts the params for a given subspace and key. This is done generically via json to avoid having to
// specify the QueryParamResponse type (which may not exist for all params).
func (n *NodeConfig) QueryParams(subspace, key string, result any) {
	cmd := []string{"babylond", "query", "params", "subspace", subspace, key, "--output=json"}

	out, _, err := n.containerManager.ExecCmd(n.t, n.Name, cmd, "")
	require.NoError(n.t, err)

	err = json.Unmarshal(out.Bytes(), &result)
	require.NoError(n.t, err)
}

func (n *NodeConfig) FailIBCTransfer(from, recipient, amount string) {
	n.LogActionF("IBC sending %s from %s to %s", amount, from, recipient)

	cmd := []string{"babylond", "tx", "ibc-transfer", "transfer", "transfer", "channel-0", recipient, amount, fmt.Sprintf("--from=%s", from)}

	_, _, err := n.containerManager.ExecTxCmdWithSuccessString(n.t, n.chainId, n.Name, cmd, "rate limit exceeded")
	require.NoError(n.t, err)

	n.LogActionF("Failed to send IBC transfer (as expected)")
}

func (n *NodeConfig) BankSend(amount string, sendAddress string, receiveAddress string) {
	n.LogActionF("bank sending %s from address %s to %s", amount, sendAddress, receiveAddress)
	cmd := []string{"babylond", "tx", "bank", "send", sendAddress, receiveAddress, amount, "--from=val"}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully sent bank sent %s from address %s to %s", amount, sendAddress, receiveAddress)
}

func (n *NodeConfig) SendHeaderHex(headerHex string) {
	n.LogActionF("btclightclient sending header %s", headerHex)
	cmd := []string{"./babylond", "tx", "btclightclient", "insert-header", headerHex, "--from=val", "--gas=500000"}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully inserted header %s", headerHex)
}

func (n *NodeConfig) InsertNewEmptyBtcHeader() *blc.BTCHeaderInfo {
	tip, err := n.QueryTip()
	require.NoError(n.t, err)
	n.t.Logf("Retrieved current tip of btc headerchain. Height: %d", tip.Height)
	child := datagen.GenRandomValidBTCHeaderInfoWithParent(*tip)
	n.SendHeaderHex(child.Header.MarshalHex())
	n.WaitUntilBtcHeight(tip.Height + 1)
	return child
}

func (n *NodeConfig) InsertHeader(h *bbn.BTCHeaderBytes) {
	tip, err := n.QueryTip()
	require.NoError(n.t, err)
	n.t.Logf("Retrieved current tip of btc headerchain. Height: %d", tip.Height)
	n.SendHeaderHex(h.MarshalHex())
	n.WaitUntilBtcHeight(tip.Height + 1)
}

func (n *NodeConfig) InsertProofs(p1 *btccheckpointtypes.BTCSpvProof, p2 *btccheckpointtypes.BTCSpvProof) {
	n.LogActionF("btccheckpoint sending proofs")

	p1bytes, err := util.Cdc.Marshal(p1)
	require.NoError(n.t, err)
	p2bytes, err := util.Cdc.Marshal(p2)
	require.NoError(n.t, err)

	p1HexBytes := hex.EncodeToString(p1bytes)
	p2HexBytes := hex.EncodeToString(p2bytes)

	cmd := []string{"./babylond", "tx", "btccheckpoint", "insert-proofs", p1HexBytes, p2HexBytes, "--from=val"}
	_, _, err = n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully inserted btc spv proofs")
}

func (n *NodeConfig) FinalizeSealedEpochs(startingEpoch uint64, lastEpoch uint64) {
	n.LogActionF("start finalizing epoch starting from  %d", startingEpoch)

	madeProgress := false
	currEpoch := startingEpoch
	for {
		if currEpoch > lastEpoch {
			break
		}

		checkpoint, err := n.QueryCheckpointForEpoch(currEpoch)

		require.NoError(n.t, err)

		// can only finalize sealed checkpoints
		if checkpoint.Status != cttypes.Sealed {
			return
		}

		currentBtcTip, err := n.QueryTip()

		require.NoError(n.t, err)

		_, c, err := bech32.DecodeAndConvert(n.PublicAddress)

		require.NoError(n.t, err)

		btcCheckpoint, err := cttypes.FromRawCkptToBTCCkpt(checkpoint.Ckpt, c)

		require.NoError(n.t, err)

		p1, p2, err := txformat.EncodeCheckpointData(
			txformat.BabylonTag(initialization.BabylonOpReturnTag),
			txformat.CurrentVersion,
			btcCheckpoint,
		)

		require.NoError(n.t, err)

		opReturn1 := datagen.CreateBlockWithTransaction(currentBtcTip.Header.ToBlockHeader(), p1)

		opReturn2 := datagen.CreateBlockWithTransaction(opReturn1.HeaderBytes.ToBlockHeader(), p2)

		n.InsertHeader(&opReturn1.HeaderBytes)
		n.InsertHeader(&opReturn2.HeaderBytes)
		n.InsertProofs(opReturn1.SpvProof, opReturn2.SpvProof)

		n.WaitForCondition(func() bool {
			ckpt, err := n.QueryCheckpointForEpoch(currEpoch)
			require.NoError(n.t, err)
			return ckpt.Status == cttypes.Submitted
		}, "Checkpoint should be submitted ")

		madeProgress = true
		currEpoch++
	}

	if madeProgress {
		// we made progress in above loop, which means the last header of btc chain is
		// valid op return header, by finalizing it, we will also finalize all older
		// checkpoints

		for i := 0; i < initialization.BabylonBtcFinalizationPeriod; i++ {
			n.InsertNewEmptyBtcHeader()
		}
	}
}
