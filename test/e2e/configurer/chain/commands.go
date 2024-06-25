package chain

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	txformat "github.com/babylonchain/babylon/btctxformatter"
	"github.com/babylonchain/babylon/test/e2e/containers"
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

const (
	flagKeyringTest = "--keyring-backend=test"
)

func (n *NodeConfig) GetWallet(walletName string) string {
	n.LogActionF("retrieving wallet %s", walletName)
	cmd := []string{"babylond", "keys", "show", walletName, flagKeyringTest, containers.FlagHome}
	outBuf, _, err := n.containerManager.ExecCmd(n.t, n.Name, cmd, "")
	require.NoError(n.t, err)
	re := regexp.MustCompile("bbn(.{39})")
	walletAddr := fmt.Sprintf("%s\n", re.FindString(outBuf.String()))
	walletAddr = strings.TrimSuffix(walletAddr, "\n")
	n.LogActionF("wallet %s found, wallet address - %s", walletName, walletAddr)
	return walletAddr
}

// KeysAdd creates a new key in the keyring
func (n *NodeConfig) KeysAdd(walletName string, overallFlags ...string) string {
	n.LogActionF("adding new wallet %s", walletName)
	cmd := []string{"babylond", "keys", "add", walletName, flagKeyringTest, containers.FlagHome}
	outBuf, _, err := n.containerManager.ExecCmd(n.t, n.Name, append(cmd, overallFlags...), "")
	require.NoError(n.t, err)
	re := regexp.MustCompile("bbn(.{39})")
	walletAddr := fmt.Sprintf("%s\n", re.FindString(outBuf.String()))
	walletAddr = strings.TrimSuffix(walletAddr, "\n")
	n.LogActionF("wallet %s created, address - %s", walletName, walletAddr)
	return walletAddr
}

// QueryParams extracts the params for a given subspace and key. This is done generically via json to avoid having to
// specify the QueryParamResponse type (which may not exist for all params).
// TODO for now all commands are not used and left here as an example
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

func (n *NodeConfig) BankSendFromNode(receiveAddress, amount string) {
	n.BankSend(n.WalletName, receiveAddress, amount)
}

func (n *NodeConfig) BankSend(fromWallet, to, amount string, overallFlags ...string) {
	fromAddr := n.GetWallet(fromWallet)
	n.LogActionF("bank sending %s from wallet %s to %s", amount, fromWallet, to)
	cmd := []string{"babylond", "tx", "bank", "send", fromAddr, to, amount, fmt.Sprintf("--from=%s", fromWallet)}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, append(cmd, overallFlags...))
	require.NoError(n.t, err)
	n.LogActionF("successfully sent bank sent %s from address %s to %s", amount, fromWallet, to)
}

func (n *NodeConfig) BankSendOutput(fromWallet, to, amount string, overallFlags ...string) (out bytes.Buffer, errBuff bytes.Buffer, err error) {
	fromAddr := n.GetWallet(fromWallet)
	n.LogActionF("bank sending %s from wallet %s to %s", amount, fromWallet, to)
	cmd := []string{
		"babylond", "tx", "bank", "send", fromAddr, to, amount, fmt.Sprintf("--from=%s", fromWallet),
		n.FlagChainID(), "-b=sync", "--yes", "--keyring-backend=test", "--log_format=json", "--home=/home/babylon/babylondata",
	}
	return n.containerManager.ExecCmd(n.t, n.Name, append(cmd, overallFlags...), "")
}

func (n *NodeConfig) SendHeaderHex(headerHex string) {
	n.LogActionF("btclightclient sending header %s", headerHex)
	cmd := []string{"babylond", "tx", "btclightclient", "insert-headers", headerHex, "--from=val", "--gas=500000"}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully inserted header %s", headerHex)
}

func (n *NodeConfig) InsertNewEmptyBtcHeader(r *rand.Rand) *blc.BTCHeaderInfo {
	tipResp, err := n.QueryTip()
	require.NoError(n.t, err)
	n.t.Logf("Retrieved current tip of btc headerchain. Height: %d", tipResp.Height)

	tip, err := ParseBTCHeaderInfoResponseToInfo(tipResp)
	require.NoError(n.t, err)

	child := datagen.GenRandomValidBTCHeaderInfoWithParent(r, *tip)
	n.SendHeaderHex(child.Header.MarshalHex())
	n.WaitUntilBtcHeight(tipResp.Height + 1)
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

		currentBtcTipResp, err := n.QueryTip()
		require.NoError(n.t, err)

		_, submitterAddr, err := bech32.DecodeAndConvert(n.PublicAddress)
		require.NoError(n.t, err)

		rawCheckpoint, err := checkpoint.Ckpt.ToRawCheckpoint()
		require.NoError(n.t, err)

		btcCheckpoint, err := cttypes.FromRawCkptToBTCCkpt(rawCheckpoint, submitterAddr)
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
		currentBtcTip, err := ParseBTCHeaderInfoResponseToInfo(currentBtcTipResp)
		require.NoError(n.t, err)

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

// TxMultisigSign sign a tx in a file with one wallet for a multisig address.
func (n *NodeConfig) TxMultisigSign(walletName, multisigAddr, txFileFullPath, fileName string, overallFlags ...string) (fullFilePathInContainer string) {
	return n.TxSign(walletName, txFileFullPath, fileName, fmt.Sprintf("--multisig=%s", multisigAddr))
}

// TxSign sign a tx in a file with one wallet.
func (n *NodeConfig) TxSign(walletName, txFileFullPath, fileName string, overallFlags ...string) (fullFilePathInContainer string) {
	n.LogActionF("wallet %s sign tx file %s", walletName, txFileFullPath)
	cmd := []string{
		"babylond", "tx", "sign", txFileFullPath,
		fmt.Sprintf("--from=%s", walletName),
		n.FlagChainID(), flagKeyringTest, containers.FlagHome,
	}
	outBuf, _, err := n.containerManager.ExecCmd(n.t, n.Name, append(cmd, overallFlags...), "")
	require.NoError(n.t, err)

	return n.WriteFile(fileName, outBuf.String())
}

// TxMultisign sign a tx in a file.
func (n *NodeConfig) TxMultisign(walletNameMultisig, txFileFullPath, outputFileName string, signedFiles []string, overallFlags ...string) (signedTxFilePath string) {
	n.LogActionF("%s multisig tx file %s", walletNameMultisig, txFileFullPath)
	cmd := []string{
		"babylond", "tx", "multisign", txFileFullPath, walletNameMultisig,
		n.FlagChainID(),
		flagKeyringTest, containers.FlagHome,
	}
	cmd = append(cmd, signedFiles...)
	outBuf, _, err := n.containerManager.ExecCmd(n.t, n.Name, append(cmd, overallFlags...), "")
	require.NoError(n.t, err)

	return n.WriteFile(outputFileName, outBuf.String())
}

// TxBroadcast broadcast a signed transaction to the chain.
func (n *NodeConfig) TxBroadcast(txSignedFileFullPath string, overallFlags ...string) {
	n.LogActionF("broadcast tx file %s", txSignedFileFullPath)
	cmd := []string{
		"babylond", "tx", "broadcast", txSignedFileFullPath,
		n.FlagChainID(),
	}
	_, _, err := n.containerManager.ExecCmd(n.t, n.Name, append(cmd, overallFlags...), "")
	require.NoError(n.t, err)
}

// TxFeeGrant creates a fee grant tx. Which the granter is the one that will
// pay the fees for the grantee to submit txs for free.
func (n *NodeConfig) TxFeeGrant(granter, grantee string, overallFlags ...string) {
	n.LogActionF("tx fee grant, granter: %s - grantee: %s", granter, grantee)
	cmd := []string{
		"babylond", "tx", "feegrant", "grant", granter, grantee,
		n.FlagChainID(),
	}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, append(cmd, overallFlags...))
	require.NoError(n.t, err)
}

// TxSignBroadcast signs the tx from the wallet and broadcast to chain.
func (n *NodeConfig) TxSignBroadcast(walletName, txFileFullPath string) {
	fileName := fmt.Sprintf("tx-signed-%s.json", walletName)
	signedTxToBroadcast := n.TxSign(walletName, txFileFullPath, fileName)
	n.TxBroadcast(signedTxToBroadcast)
}

// TxMultisignBroadcast signs the tx from each wallet and the multisig and broadcast to chain.
func (n *NodeConfig) TxMultisignBroadcast(walletNameMultisig, txFileFullPath string, walleNameSigners []string) {
	multisigAddr := n.GetWallet(walletNameMultisig)

	signedFiles := make([]string, len(walleNameSigners))
	for i, wName := range walleNameSigners {
		fileName := fmt.Sprintf("tx-signed-%s.json", wName)
		signedFiles[i] = n.TxMultisigSign(wName, multisigAddr, txFileFullPath, fileName)
	}

	signedTxToBroadcast := n.TxMultisign(walletNameMultisig, txFileFullPath, "tx-multisigned.json", signedFiles)
	n.TxBroadcast(signedTxToBroadcast)
}

// WriteFile writes a new file in the config dir of the node where it is volume mounted to the
// babylon home inside the container and returns the full file path inside the container.
func (n *NodeConfig) WriteFile(fileName, content string) (fullFilePathInContainer string) {
	b := bytes.NewBufferString(content)
	fileFullPath := filepath.Join(n.ConfigDir, fileName)

	err := os.WriteFile(fileFullPath, b.Bytes(), 0644)
	require.NoError(n.t, err)

	return filepath.Join(containers.BabylonHomePath, fileName)
}

// FlagChainID returns the flag of the chainID.
func (n *NodeConfig) FlagChainID() string {
	return fmt.Sprintf("--chain-id=%s", n.chainId)
}

// ParseBTCHeaderInfoResponseToInfo turns an BTCHeaderInfoResponse back to BTCHeaderInfo.
func ParseBTCHeaderInfoResponseToInfo(r *blc.BTCHeaderInfoResponse) (*blc.BTCHeaderInfo, error) {
	header, err := bbn.NewBTCHeaderBytesFromHex(r.HeaderHex)
	if err != nil {
		return nil, err
	}

	hash, err := bbn.NewBTCHeaderHashBytesFromHex(r.HashHex)
	if err != nil {
		return nil, err
	}

	return &blc.BTCHeaderInfo{
		Header: &header,
		Hash:   &hash,
		Height: r.Height,
		Work:   &r.Work,
	}, nil
}
