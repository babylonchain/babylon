package chain

import (
	"encoding/hex"
	"strconv"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/wire"

	"cosmossdk.io/math"
	sdkmath "cosmossdk.io/math"
	bbn "github.com/babylonchain/babylon/types"
	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	bstypes "github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/stretchr/testify/require"
)

func (n *NodeConfig) CreateBTCValidator(babylonPK *secp256k1.PubKey, btcPK *bbn.BIP340PubKey, pop *bstypes.ProofOfPossession, moniker, identity, website, securityContract, details string, commission *math.LegacyDec) {
	n.LogActionF("creating BTC validator")

	// get babylon PK hex
	babylonPKBytes, err := babylonPK.Marshal()
	require.NoError(n.t, err)
	babylonPKHex := hex.EncodeToString(babylonPKBytes)
	// get BTC PK hex
	btcPKHex := btcPK.MarshalHex()
	// get pop hex
	popHex, err := pop.ToHexStr()
	require.NoError(n.t, err)

	cmd := []string{
		"babylond", "tx", "btcstaking", "create-btc-validator", babylonPKHex, btcPKHex, popHex, "--from=val", "--moniker", moniker, "--identity", identity, "--website", website, "--security-contact", securityContract, "--details", details, "--commission-rate", commission.String(),
	}
	_, _, err = n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully created BTC validator")
}

func (n *NodeConfig) CreateBTCDelegation(
	babylonPK *secp256k1.PubKey,
	btcPk *bbn.BIP340PubKey,
	pop *bstypes.ProofOfPossession,
	stakingTxInfo *btcctypes.TransactionInfo,
	valPK *bbn.BIP340PubKey,
	stakingTimeBlocks uint16,
	stakingValue btcutil.Amount,
	slashingTx *bstypes.BTCSlashingTx,
	delegatorSig *bbn.BIP340Signature,
) {
	n.LogActionF("creating BTC delegation")

	// get babylon PK hex
	babylonPKBytes, err := babylonPK.Marshal()
	require.NoError(n.t, err)
	babylonPKHex := hex.EncodeToString(babylonPKBytes)

	btcPkHex := btcPk.MarshalHex()

	// get pop hex
	popHex, err := pop.ToHexStr()
	require.NoError(n.t, err)

	// get staking tx info hex
	stakingTxInfoHex, err := stakingTxInfo.ToHexStr()
	require.NoError(n.t, err)

	valPKHex := valPK.MarshalHex()

	stakingTimeString := sdkmath.NewUint(uint64(stakingTimeBlocks)).String()
	stakingValueString := sdkmath.NewInt(int64(stakingValue)).String()

	// get slashing tx hex
	slashingTxHex := slashingTx.ToHexStr()
	// get delegator sig hex
	delegatorSigHex := delegatorSig.ToHexStr()

	cmd := []string{"babylond", "tx", "btcstaking", "create-btc-delegation", babylonPKHex, btcPkHex, popHex, stakingTxInfoHex, valPKHex, stakingTimeString, stakingValueString, slashingTxHex, delegatorSigHex, "--from=val"}
	_, _, err = n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully created BTC delegation")
}

func (n *NodeConfig) AddCovenantSig(covPK *bbn.BIP340PubKey, stakingTxHash string, sig *bbn.BIP340Signature) {
	n.LogActionF("adding covenant signature")

	covPKHex := covPK.MarshalHex()
	sigHex := sig.ToHexStr()

	cmd := []string{"babylond", "tx", "btcstaking", "add-covenant-sig", covPKHex, stakingTxHash, sigHex, "--from=val"}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully added covenant sig")
}

func (n *NodeConfig) CommitPubRandList(valBTCPK *bbn.BIP340PubKey, startHeight uint64, pubRandList []bbn.SchnorrPubRand, sig *bbn.BIP340Signature) {
	n.LogActionF("committing public randomness list")

	cmd := []string{"babylond", "tx", "finality", "commit-pubrand-list"}

	// add val BTC PK to cmd
	valBTCPKHex := valBTCPK.MarshalHex()
	cmd = append(cmd, valBTCPKHex)

	// add start height to cmd
	startHeightStr := strconv.FormatUint(startHeight, 10)
	cmd = append(cmd, startHeightStr)

	// add each pubrand to cmd
	for _, pr := range pubRandList {
		prHex := pr.ToHexStr()
		cmd = append(cmd, prHex)
	}

	// add sig to cmd
	sigHex := sig.ToHexStr()
	cmd = append(cmd, sigHex)

	// specify used key
	cmd = append(cmd, "--from=val")

	// gas
	cmd = append(cmd, "--gas=auto", "--gas-prices=1ubbn", "--gas-adjustment=1.3")

	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully committed public randomness list")
}

func (n *NodeConfig) AddFinalitySig(valBTCPK *bbn.BIP340PubKey, blockHeight uint64, blockLch []byte, finalitySig *bbn.SchnorrEOTSSig) {
	n.LogActionF("add finality signature")

	valBTCPKHex := valBTCPK.MarshalHex()
	blockHeightStr := strconv.FormatUint(blockHeight, 10)
	blockLchHex := hex.EncodeToString(blockLch)
	finalitySigHex := finalitySig.ToHexStr()

	cmd := []string{"babylond", "tx", "finality", "add-finality-sig", valBTCPKHex, blockHeightStr, blockLchHex, finalitySigHex, "--from=val", "--gas=500000"}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully added finality signature")
}

func (n *NodeConfig) CreateBTCUndelegation(
	unbondingTx *wire.MsgTx,
	slashingTx *bstypes.BTCSlashingTx,
	unbondingTimeBlocks uint16,
	unbondingValue btcutil.Amount,
	delegatorSig *bbn.BIP340Signature) {
	n.LogActionF("creating BTC undelegation")

	txBytes, err := bstypes.SerializeBtcTx(unbondingTx)
	require.NoError(n.t, err)
	// get staking tx hex
	unbondingTxHex := hex.EncodeToString(txBytes)
	// get slashing tx hex
	slashingTxHex := slashingTx.ToHexStr()
	// get delegator sig hex
	delegatorSigHex := delegatorSig.ToHexStr()

	unbondingTimeStr := sdkmath.NewUint(uint64(unbondingTimeBlocks)).String()
	unbondingValueStr := sdkmath.NewInt(int64(unbondingValue)).String()

	cmd := []string{"babylond", "tx", "btcstaking", "create-btc-undelegation", unbondingTxHex, slashingTxHex, unbondingTimeStr, unbondingValueStr, delegatorSigHex, "--from=val"}
	_, _, err = n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully created BTC delegation")
}

func (n *NodeConfig) AddValidatorUnbondingSig(valPK *bbn.BIP340PubKey, delPK *bbn.BIP340PubKey, stakingTxHash string, sig *bbn.BIP340Signature) {
	n.LogActionF("adding validator signature")

	valPKHex := valPK.MarshalHex()
	delPKHex := delPK.MarshalHex()
	sigHex := sig.ToHexStr()

	cmd := []string{"babylond", "tx", "btcstaking", "add-validator-unbonding-sig", valPKHex, delPKHex, stakingTxHash, sigHex, "--from=val"}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully added validator unbonding sig")
}

func (n *NodeConfig) AddCovenantUnbondingSigs(
	covPK *bbn.BIP340PubKey,
	stakingTxHash string,
	unbondingTxSig *bbn.BIP340Signature,
	slashUnbondingTxSig *bbn.BIP340Signature) {
	n.LogActionF("adding validator signature")

	covPKHex := covPK.MarshalHex()
	unbondingTxSigHex := unbondingTxSig.ToHexStr()
	slashUnbondingTxSigHex := slashUnbondingTxSig.ToHexStr()

	cmd := []string{"babylond", "tx", "btcstaking", "add-covenant-unbonding-sigs", covPKHex, stakingTxHash, unbondingTxSigHex, slashUnbondingTxSigHex, "--from=val"}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully added covenant unbonding sigs")
}
