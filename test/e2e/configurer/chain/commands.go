package chain

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"

	txformat "github.com/babylonchain/babylon/btctxformatter"
	"github.com/babylonchain/babylon/test/e2e/initialization"
	"github.com/babylonchain/babylon/test/e2e/util"
	"github.com/babylonchain/babylon/testutil/datagen"
	bbn "github.com/babylonchain/babylon/types"
	btccheckpointtypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	blc "github.com/babylonchain/babylon/x/btclightclient/types"
	cttypes "github.com/babylonchain/babylon/x/checkpointing/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	sdkquerytypes "github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
)

func (n *NodeConfig) GetWallet(walletName string) string {
	n.LogActionF("retrieving wallet %s", walletName)
	cmd := []string{"babylond", "keys", "show", walletName, "--keyring-backend=test"}
	outBuf, _, err := n.containerManager.ExecCmd(n.t, n.Name, cmd, "")
	require.NoError(n.t, err)
	re := regexp.MustCompile("bbn(.{38})")
	walletAddr := fmt.Sprintf("%s\n", re.FindString(outBuf.String()))
	walletAddr = strings.TrimSuffix(walletAddr, "\n")
	n.LogActionF("wallet %s found, waller address - %s", walletName, walletAddr)
	return walletAddr
}

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

func (n *NodeConfig) SendIBCTransfer(from, recipient, memo string, token sdk.Coin) {
	n.LogActionF("IBC sending %s from %s to %s. memo: %s", token.Amount.String(), from, recipient, memo)

	cmd := []string{"babylond", "tx", "ibc-transfer", "transfer", "transfer", "channel-0", recipient, token.String(), fmt.Sprintf("--from=%s", from), "--memo", memo}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)

	n.LogActionF("successfully submitted sent IBC transfer")
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
	cmd := []string{"babylond", "tx", "btclightclient", "insert-headers", headerHex, "--from=val", "--gas=500000"}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully inserted header %s", headerHex)
}

func (n *NodeConfig) InsertNewEmptyBtcHeader(r *rand.Rand) *blc.BTCHeaderInfo {
	tip, err := n.QueryTip()
	require.NoError(n.t, err)
	n.t.Logf("Retrieved current tip of btc headerchain. Height: %d", tip.Height)
	child := datagen.GenRandomValidBTCHeaderInfoWithParent(r, *tip)
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

	cmd := []string{"babylond", "tx", "btccheckpoint", "insert-proofs", p1HexBytes, p2HexBytes, "--from=val"}
	_, _, err = n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully inserted btc spv proofs")
}

func (n *NodeConfig) FinalizeSealedEpochs(startEpoch uint64, lastEpoch uint64) {
	n.LogActionF("start finalizing epochs from  %d to %d", startEpoch, lastEpoch)
	// Random source for the generation of BTC data
	r := rand.New(rand.NewSource(time.Now().Unix()))

	madeProgress := false

	pageLimit := lastEpoch - startEpoch + 1
	pagination := &sdkquerytypes.PageRequest{
		Key:   cttypes.CkptsObjectKey(startEpoch),
		Limit: pageLimit,
	}

	resp, err := n.QueryRawCheckpoints(pagination)
	require.NoError(n.t, err)
	require.Equal(n.t, int(pageLimit), len(resp.RawCheckpoints))

	for _, checkpoint := range resp.RawCheckpoints {
		require.Equal(n.t, checkpoint.Status, cttypes.Sealed)

		currentBtcTip, err := n.QueryTip()
		require.NoError(n.t, err)

		_, submitterAddr, err := bech32.DecodeAndConvert(n.PublicAddress)
		require.NoError(n.t, err)

		btcCheckpoint, err := cttypes.FromRawCkptToBTCCkpt(checkpoint.Ckpt, submitterAddr)
		require.NoError(n.t, err)

		babylonTagBytes, err := hex.DecodeString(initialization.BabylonOpReturnTag)
		require.NoError(n.t, err)

		p1, p2, err := txformat.EncodeCheckpointData(
			babylonTagBytes,
			txformat.CurrentVersion,
			btcCheckpoint,
		)
		require.NoError(n.t, err)

		tx1 := datagen.CreatOpReturnTransaction(r, p1)
		opReturn1 := datagen.CreateBlockWithTransaction(r, currentBtcTip.Header.ToBlockHeader(), tx1)
		tx2 := datagen.CreatOpReturnTransaction(r, p2)
		opReturn2 := datagen.CreateBlockWithTransaction(r, opReturn1.HeaderBytes.ToBlockHeader(), tx2)

		n.InsertHeader(&opReturn1.HeaderBytes)
		n.InsertHeader(&opReturn2.HeaderBytes)
		n.InsertProofs(opReturn1.SpvProof, opReturn2.SpvProof)

		n.WaitForCondition(func() bool {
			ckpt, err := n.QueryRawCheckpoint(checkpoint.Ckpt.EpochNum)
			require.NoError(n.t, err)
			return ckpt.Status == cttypes.Submitted
		}, "Checkpoint should be submitted ")

		madeProgress = true
	}

	if madeProgress {
		// we made progress in above loop, which means the last header of btc chain is
		// valid op return header, by finalizing it, we will also finalize all older
		// checkpoints

		for i := 0; i < initialization.BabylonBtcFinalizationPeriod; i++ {
			n.InsertNewEmptyBtcHeader(r)
		}
	}
}

func (n *NodeConfig) StoreWasmCode(wasmFile, from string) {
	n.LogActionF("storing wasm code from file %s", wasmFile)
	cmd := []string{"babylond", "tx", "wasm", "store", wasmFile, fmt.Sprintf("--from=%s", from), "--gas=auto", "--gas-prices=1ubbn", "--gas-adjustment=1.3"}
	n.LogActionF(strings.Join(cmd, " "))
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully stored")
}

func (n *NodeConfig) InstantiateWasmContract(codeId, initMsg, from string) {
	n.LogActionF("instantiating wasm contract %s with %s", codeId, initMsg)
	cmd := []string{"babylond", "tx", "wasm", "instantiate", codeId, initMsg, fmt.Sprintf("--from=%s", from), "--no-admin", "--label=contract", "--gas=auto", "--gas-prices=1ubbn", "--gas-adjustment=1.3"}
	n.LogActionF(strings.Join(cmd, " "))
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully initialized")
}

func (n *NodeConfig) WasmExecute(contract, execMsg, from string) {
	n.LogActionF("executing %s on wasm contract %s from %s", execMsg, contract, from)
	cmd := []string{"babylond", "tx", "wasm", "execute", contract, execMsg, fmt.Sprintf("--from=%s", from)}
	n.LogActionF(strings.Join(cmd, " "))
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully executed")
}

// WithdrawReward will withdraw the rewards of the address associated with the tx signer `from`
func (n *NodeConfig) WithdrawReward(sType, from string) {
	n.LogActionF("withdraw rewards of type %s for tx signer %s", sType, from)
	cmd := []string{"babylond", "tx", "incentive", "withdraw-reward", sType, fmt.Sprintf("--from=%s", from)}
	n.LogActionF(strings.Join(cmd, " "))
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully withdrawn")
}
