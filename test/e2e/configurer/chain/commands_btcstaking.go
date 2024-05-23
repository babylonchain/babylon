package chain

import (
	"encoding/hex"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"

	sdkmath "cosmossdk.io/math"
	cmtcrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/stretchr/testify/require"

	asig "github.com/babylonchain/babylon/crypto/schnorr-adaptor-signature"
	bbn "github.com/babylonchain/babylon/types"
	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	bstypes "github.com/babylonchain/babylon/x/btcstaking/types"
)

func (n *NodeConfig) CreateFinalityProvider(babylonPK *secp256k1.PubKey, btcPK *bbn.BIP340PubKey, pop *bstypes.ProofOfPossession, moniker, identity, website, securityContract, details string, commission *sdkmath.LegacyDec) {
	n.LogActionF("creating finality provider")

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
		"babylond", "tx", "btcstaking", "create-finality-provider", babylonPKHex, btcPKHex, popHex, "--from=val", "--moniker", moniker, "--identity", identity, "--website", website, "--security-contact", securityContract, "--details", details, "--commission-rate", commission.String(),
	}
	_, _, err = n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully created finality provider")
}

func (n *NodeConfig) CreateBTCDelegation(
	babylonPK *secp256k1.PubKey,
	btcPk *bbn.BIP340PubKey,
	pop *bstypes.ProofOfPossession,
	stakingTxInfo *btcctypes.TransactionInfo,
	fpPK *bbn.BIP340PubKey,
	stakingTimeBlocks uint16,
	stakingValue btcutil.Amount,
	slashingTx *bstypes.BTCSlashingTx,
	delegatorSig *bbn.BIP340Signature,
	unbondingTx *wire.MsgTx,
	unbondingSlashingTx *bstypes.BTCSlashingTx,
	unbondingTime uint16,
	unbondingValue btcutil.Amount,
	delUnbondingSlashingSig *bbn.BIP340Signature,
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

	fpPKHex := fpPK.MarshalHex()

	stakingTimeString := sdkmath.NewUint(uint64(stakingTimeBlocks)).String()
	stakingValueString := sdkmath.NewInt(int64(stakingValue)).String()

	// get slashing tx hex
	slashingTxHex := slashingTx.ToHexStr()
	// get delegator sig hex
	delegatorSigHex := delegatorSig.ToHexStr()

	// on-demand unbonding related
	unbondingTxBytes, err := bbn.SerializeBTCTx(unbondingTx)
	require.NoError(n.t, err)
	unbondingTxHex := hex.EncodeToString(unbondingTxBytes)
	unbondingSlashingTxHex := unbondingSlashingTx.ToHexStr()
	unbondingTimeStr := sdkmath.NewUint(uint64(unbondingTime)).String()
	unbondingValueStr := sdkmath.NewInt(int64(unbondingValue)).String()
	delUnbondingSlashingSigHex := delUnbondingSlashingSig.ToHexStr()

	cmd := []string{"babylond", "tx", "btcstaking", "create-btc-delegation", babylonPKHex, btcPkHex, popHex, stakingTxInfoHex, fpPKHex, stakingTimeString, stakingValueString, slashingTxHex, delegatorSigHex, unbondingTxHex, unbondingSlashingTxHex, unbondingTimeStr, unbondingValueStr, delUnbondingSlashingSigHex, "--from=val"}
	_, _, err = n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully created BTC delegation")
}

func (n *NodeConfig) AddCovenantSigs(covPK *bbn.BIP340PubKey, stakingTxHash string, slashingSigs [][]byte, unbondingSig *bbn.BIP340Signature, unbondingSlashingSigs [][]byte) {
	n.LogActionF("adding covenant signature")

	covPKHex := covPK.MarshalHex()

	cmd := []string{"babylond", "tx", "btcstaking", "add-covenant-sigs", covPKHex, stakingTxHash}

	// slashing signatures
	var slashingSigStrList []string
	for _, sig := range slashingSigs {
		slashingSigStrList = append(slashingSigStrList, hex.EncodeToString(sig))
	}
	slashingSigStr := strings.Join(slashingSigStrList, ",")
	cmd = append(cmd, slashingSigStr)

	// on-demand unbonding stuff
	cmd = append(cmd, unbondingSig.ToHexStr())
	var unbondingSlashingSigStrList []string
	for _, sig := range unbondingSlashingSigs {
		unbondingSlashingSigStrList = append(unbondingSlashingSigStrList, hex.EncodeToString(sig))
	}
	unbondingSlashingSigStr := strings.Join(unbondingSlashingSigStrList, ",")
	cmd = append(cmd, unbondingSlashingSigStr)

	// used key
	cmd = append(cmd, "--from=val")
	// gas
	cmd = append(cmd, "--gas=auto", "--gas-prices=1ubbn", "--gas-adjustment=1.3")

	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully added covenant signatures")
}

func (n *NodeConfig) CommitPubRandList(fpBTCPK *bbn.BIP340PubKey, startHeight uint64, numPubrand uint64, commitment []byte, sig *bbn.BIP340Signature) {
	n.LogActionF("committing public randomness list")

	cmd := []string{"babylond", "tx", "finality", "commit-pubrand-list"}

	// add finality provider BTC PK to cmd
	fpBTCPKHex := fpBTCPK.MarshalHex()
	cmd = append(cmd, fpBTCPKHex)

	// add start height to cmd
	startHeightStr := strconv.FormatUint(startHeight, 10)
	cmd = append(cmd, startHeightStr)

	// add num_pub_rand to cmd
	numPubRandStr := strconv.FormatUint(numPubrand, 10)
	cmd = append(cmd, numPubRandStr)

	// add commitment to cmd
	commitmentHex := hex.EncodeToString(commitment)
	cmd = append(cmd, commitmentHex)

	// add sig to cmd
	sigHex := sig.ToHexStr()
	cmd = append(cmd, sigHex)

	// specify used key
	cmd = append(cmd, "--from=val")

	// gas
	cmd = append(cmd, "--gas=500000")

	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully committed public randomness list")
}

func (n *NodeConfig) AddFinalitySig(fpBTCPK *bbn.BIP340PubKey, blockHeight uint64, pubRand *bbn.SchnorrPubRand, proof cmtcrypto.Proof, appHash []byte, finalitySig *bbn.SchnorrEOTSSig) {
	n.LogActionF("add finality signature")

	fpBTCPKHex := fpBTCPK.MarshalHex()
	blockHeightStr := strconv.FormatUint(blockHeight, 10)
	pubRandHex := pubRand.MarshalHex()
	proofBytes, err := proof.Marshal()
	require.NoError(n.t, err)
	proofHex := hex.EncodeToString(proofBytes)
	appHashHex := hex.EncodeToString(appHash)
	finalitySigHex := finalitySig.ToHexStr()

	cmd := []string{"babylond", "tx", "finality", "add-finality-sig", fpBTCPKHex, blockHeightStr, pubRandHex, proofHex, appHashHex, finalitySigHex, "--from=val", "--gas=500000"}
	_, _, err = n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully added finality signature")
}

func (n *NodeConfig) AddCovenantUnbondingSigs(
	covPK *bbn.BIP340PubKey,
	stakingTxHash string,
	unbondingTxSig *bbn.BIP340Signature,
	slashUnbondingTxSigs []*asig.AdaptorSignature) {
	n.LogActionF("adding finality provider signature")

	covPKHex := covPK.MarshalHex()
	unbondingTxSigHex := unbondingTxSig.ToHexStr()

	cmd := []string{"babylond", "tx", "btcstaking", "add-covenant-unbonding-sigs", covPKHex, stakingTxHash, unbondingTxSigHex}
	for _, sig := range slashUnbondingTxSigs {
		cmd = append(cmd, sig.MarshalHex())
	}
	cmd = append(cmd, "--from=val")
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully added covenant unbonding sigs")
}

func (n *NodeConfig) BTCUndelegate(stakingTxHash *chainhash.Hash, delUnbondingSig *schnorr.Signature) {
	n.LogActionF("undelegate by using signature on unbonding tx from delegator")

	sigHex := bbn.NewBIP340SignatureFromBTCSig(delUnbondingSig).ToHexStr()
	cmd := []string{"babylond", "tx", "btcstaking", "btc-undelegate", stakingTxHash.String(), sigHex, "--from=val"}

	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully added signature on unbonding tx from delegator")
}
