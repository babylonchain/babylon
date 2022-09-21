package cmd

// DONTCOVER

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/babylonchain/babylon/app"
	txformat "github.com/babylonchain/babylon/btctxformatter"
	btclightclienttypes "github.com/babylonchain/babylon/x/btclightclient/types"
	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"

	btccheckpointtypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"

	"github.com/babylonchain/babylon/privval"
	"github.com/babylonchain/babylon/testutil/datagen"
	bbn "github.com/babylonchain/babylon/types"
	checkpointingtypes "github.com/babylonchain/babylon/x/checkpointing/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"

	"github.com/spf13/cobra"
	tmconfig "github.com/tendermint/tendermint/config"
	tmos "github.com/tendermint/tendermint/libs/os"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	"github.com/tendermint/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/server"
	srvconfig "github.com/cosmos/cosmos-sdk/server/config"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var (
	flagNodeDirPrefix          = "node-dir-prefix"
	flagNumValidators          = "v"
	flagOutputDir              = "output-dir"
	flagNodeDaemonHome         = "node-daemon-home"
	flagStartingIPAddress      = "starting-ip-address"
	flagBtcNetwork             = "btc-network"
	flagBtcCheckpointTag       = "btc-checkpoint-tag"
	flagBtcConfirmationDepth   = "btc-confirmation-depth"
	flagBtcFinalizationTimeout = "btc-finalization-timeout"
	flagEpochInterval          = "epoch-interval"
	flagRetrySleepTime         = "retry-sleep-time"
	flagMaxRetrySleepTime      = "max-retry-sleep-time"
	flagBaseBtcHeaderHex       = "btc-base-header"
	flagBaseBtcHeaderHeight    = "btc-base-header-height"
	flagMaxActiveValidators    = "max-active-validators"
)

// get cmd to initialize all files for tendermint testnet and application
func testnetCmd(mbm module.BasicManager, genBalIterator banktypes.GenesisBalancesIterator) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "testnet",
		Short: "Initialize files for a babylon testnet",
		Long: `testnet will create "v" number of directories and populate each with
necessary files (private validator, genesis, config, etc.).

Note, strict routability for addresses is turned off in the config file.

Example:
	babylond testnet --v 4 --output-dir ./output --starting-ip-address 192.168.10.2
	`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			outputDir, _ := cmd.Flags().GetString(flagOutputDir)
			keyringBackend, _ := cmd.Flags().GetString(flags.FlagKeyringBackend)
			chainID, _ := cmd.Flags().GetString(flags.FlagChainID)
			minGasPrices, _ := cmd.Flags().GetString(server.FlagMinGasPrices)
			nodeDirPrefix, _ := cmd.Flags().GetString(flagNodeDirPrefix)
			nodeDaemonHome, _ := cmd.Flags().GetString(flagNodeDaemonHome)
			startingIPAddress, _ := cmd.Flags().GetString(flagStartingIPAddress)
			numValidators, _ := cmd.Flags().GetInt(flagNumValidators)
			algo, _ := cmd.Flags().GetString(flags.FlagKeyAlgorithm)
			// staking args
			maxActiveValidators, _ := cmd.Flags().GetUint32(flagMaxActiveValidators)
			// btccheckpoint args
			btcNetwork, _ := cmd.Flags().GetString(flagBtcNetwork)
			btcCheckpointTag, _ := cmd.Flags().GetString(flagBtcCheckpointTag)
			btcConfirmationDepth, _ := cmd.Flags().GetUint64(flagBtcConfirmationDepth)
			btcFinalizationTimeout, _ := cmd.Flags().GetUint64(flagBtcFinalizationTimeout)
			// checkpointing args
			retrySleepTime, _ := cmd.Flags().GetString(flagRetrySleepTime)
			maxRetrySleepTime, _ := cmd.Flags().GetString(flagMaxRetrySleepTime)
			// epoching args
			epochInterval, _ := cmd.Flags().GetUint64(flagEpochInterval)
			// btclightclient args
			baseBtcHeaderHex, _ := cmd.Flags().GetString(flagBaseBtcHeaderHex)
			baseBtcHeaderHeight, err := cmd.Flags().GetUint64(flagBaseBtcHeaderHeight)
			if err != nil {
				return errors.New("base Bitcoin header height should be a uint64")
			}

			return InitTestnet(
				clientCtx, cmd, config, mbm, genBalIterator, outputDir, chainID, minGasPrices,
				nodeDirPrefix, nodeDaemonHome, startingIPAddress, keyringBackend, algo, numValidators,
				maxActiveValidators, btcNetwork, btcCheckpointTag, btcConfirmationDepth, btcFinalizationTimeout,
				retrySleepTime, maxRetrySleepTime,
				epochInterval, baseBtcHeaderHex, baseBtcHeaderHeight,
			)
		},
	}

	cmd.Flags().Int(flagNumValidators, 4, "Number of validators to initialize the testnet with")
	cmd.Flags().StringP(flagOutputDir, "o", "./mytestnet", "Directory to store initialization data for the testnet")
	cmd.Flags().String(flagNodeDirPrefix, "node", "Prefix the directory name for each node with (node results in node0, node1, ...)")
	cmd.Flags().String(flagNodeDaemonHome, "babylond", "Home directory of the node's daemon configuration")
	cmd.Flags().String(flagStartingIPAddress, "192.168.0.1", "Starting IP address (192.168.0.1 results in persistent peers list ID0@192.168.0.1:46656, ID1@192.168.0.2:46656, ...)")
	cmd.Flags().String(flags.FlagChainID, "", "genesis file chain-id, if left blank will be randomly created")
	cmd.Flags().String(server.FlagMinGasPrices, fmt.Sprintf("0.000006%s", sdk.DefaultBondDenom), "Minimum gas prices to accept for transactions; All fees in a tx must meet this minimum (e.g. 0.01photino,0.001stake)")
	cmd.Flags().String(flags.FlagKeyringBackend, flags.DefaultKeyringBackend, "Select keyring's backend (os|file|test)")
	cmd.Flags().String(flags.FlagKeyAlgorithm, string(hd.Secp256k1Type), "Key signing algorithm to generate keys for")
	// btccheckpoint args
	cmd.Flags().String(flagBtcNetwork, string(bbn.BtcSimnet), "Bitcoin network to use. Available networks: simnet, testnet, mainnet")
	cmd.Flags().String(flagBtcCheckpointTag, string(txformat.DefautTestTagStr), "Tag to use for Bitcoin checkpoints.")
	cmd.Flags().Uint64(flagBtcConfirmationDepth, 6, "Confirmation depth for Bitcoin headers.")
	cmd.Flags().Uint64(flagBtcFinalizationTimeout, 20, "Finalization timeout for Bitcoin headers.")
	// checkpointing args
	cmd.Flags().String(flagRetrySleepTime, "1s", "Time waited for retry")
	cmd.Flags().String(flagMaxRetrySleepTime, "1m", "Max time waited for retry")
	// epoch args
	cmd.Flags().Uint64(flagEpochInterval, 10, "Number of blocks between epochs. Must be more than 0.")
	// btclightclient args
	// Simnet genesis header
	cmd.Flags().String(flagBaseBtcHeaderHex, "0100000000000000000000000000000000000000000000000000000000000000000000003ba3edfd7a7b12b27ac72c3e67768f617fc81bc3888a51323a9fb8aa4b1e5e4a45068653ffff7f2002000000", "Hex of the base Bitcoin header.")
	cmd.Flags().Uint64(flagBaseBtcHeaderHeight, 0, "Height of the base Bitcoin header.")
	cmd.Flags().Uint32(flagMaxActiveValidators, 10, "Maximum number of validators.")

	return cmd
}

const nodeDirPerm = 0755

// Initialize the testnet
func InitTestnet(
	clientCtx client.Context,
	cmd *cobra.Command,
	nodeConfig *tmconfig.Config,
	mbm module.BasicManager,
	genBalIterator banktypes.GenesisBalancesIterator,
	outputDir,
	chainID,
	minGasPrices,
	nodeDirPrefix,
	nodeDaemonHome,
	startingIPAddress,
	keyringBackend,
	algoStr string,
	numValidators int,
	maxActiveValidators uint32,
	btcNetwork string,
	btcCheckpointTag string,
	btcConfirmationDepth uint64,
	btcFinalizationTimeout uint64,
	retrySleepTime string,
	maxRetrySleepTime string,
	epochInterval uint64,
	baseBtcHeaderHex string,
	baseBtcHeaderHeight uint64,
) error {

	if chainID == "" {
		chainID = "chain-" + tmrand.NewRand().Str(6)
	}

	nodeIDs := make([]string, numValidators)
	valKeys := make([]*privval.ValidatorKeys, numValidators)

	babylonConfig := DefaultBabylonConfig()
	babylonConfig.MinGasPrices = minGasPrices
	babylonConfig.API.Enable = true
	babylonConfig.Telemetry.Enabled = true
	babylonConfig.Telemetry.PrometheusRetentionTime = 60
	babylonConfig.Telemetry.EnableHostnameLabel = false
	babylonConfig.Telemetry.GlobalLabels = [][]string{{"chain_id", chainID}}
	// BTC related config. Default values "simnet" and "BBT1"
	babylonConfig.BtcConfig.Network = btcNetwork
	babylonConfig.BtcConfig.CheckpointTag = btcCheckpointTag
	// Explorer related config. Allow CORS connections.
	babylonConfig.API.EnableUnsafeCORS = true

	var (
		genAccounts []authtypes.GenesisAccount
		genBalances []banktypes.Balance
		genKeys     []*checkpointingtypes.GenesisKey
		genFiles    []string
	)

	inBuf := bufio.NewReader(cmd.InOrStdin())
	// generate private keys, node IDs, and initial transactions
	for i := 0; i < numValidators; i++ {
		nodeDirName := fmt.Sprintf("%s%d", nodeDirPrefix, i)
		nodeDir := filepath.Join(outputDir, nodeDirName, nodeDaemonHome)
		gentxsDir := filepath.Join(outputDir, "gentxs")

		nodeConfig.SetRoot(nodeDir)
		nodeConfig.RPC.ListenAddress = "tcp://0.0.0.0:26657"

		// Explorer related config
		// Allow all CORS requests
		nodeConfig.RPC.CORSAllowedOrigins = []string{"*"}
		// Enable Prometheus
		nodeConfig.Instrumentation.Prometheus = true
		// Set the number of simultaneous connections to unlimited
		nodeConfig.Instrumentation.MaxOpenConnections = 0

		if err := os.MkdirAll(filepath.Join(nodeDir, "config"), nodeDirPerm); err != nil {
			_ = os.RemoveAll(outputDir)
			return err
		}

		nodeConfig.Moniker = nodeDirName

		ip, err := getIP(i, startingIPAddress)
		if err != nil {
			_ = os.RemoveAll(outputDir)
			return err
		}

		// generate account key
		kb, err := keyring.New(sdk.KeyringServiceName(), keyringBackend, nodeDir, inBuf)
		if err != nil {
			return err
		}
		keyringAlgos, _ := kb.SupportedAlgorithms()
		algo, err := keyring.NewSigningAlgoFromString(algoStr, keyringAlgos)
		if err != nil {
			return err
		}
		addr, secret, err := testutil.GenerateSaveCoinKey(kb, nodeDirName, "", true, algo)
		if err != nil {
			_ = os.RemoveAll(outputDir)
			return err
		}

		// generate validator keys
		nodeIDs[i], valKeys[i], err = datagen.InitializeNodeValidatorFiles(nodeConfig, addr)
		if err != nil {
			_ = os.RemoveAll(outputDir)
			return err
		}

		memo := fmt.Sprintf("%s@%s:26656", nodeIDs[i], ip)
		genFiles = append(genFiles, nodeConfig.GenesisFile())

		info := map[string]string{"secret": secret}

		cliPrint, err := json.Marshal(info)
		if err != nil {
			return err
		}

		// save private key seed words
		if err = writeFile(fmt.Sprintf("%v.json", "key_seed"), nodeDir, cliPrint); err != nil {
			return err
		}

		accTokens := sdk.TokensFromConsensusPower(1000, sdk.DefaultPowerReduction)
		accStakingTokens := sdk.TokensFromConsensusPower(500, sdk.DefaultPowerReduction)
		coins := sdk.Coins{
			sdk.NewCoin("testtoken", accTokens),
			sdk.NewCoin(sdk.DefaultBondDenom, accStakingTokens),
		}

		genBalances = append(genBalances, banktypes.Balance{Address: addr.String(), Coins: coins.Sort()})
		genAccounts = append(genAccounts, authtypes.NewBaseAccount(addr, nil, 0, 0))

		valTokens := sdk.TokensFromConsensusPower(100, sdk.DefaultPowerReduction)
		valPubkey, err := cryptocodec.FromTmPubKeyInterface(valKeys[i].ValPubkey)
		if err != nil {
			return err
		}

		genKey := &checkpointingtypes.GenesisKey{
			ValidatorAddress: sdk.ValAddress(addr).String(),
			BlsKey: &checkpointingtypes.BlsKey{
				Pubkey: &valKeys[i].BlsPubkey,
				Pop:    valKeys[i].PoP,
			},
			ValPubkey: valPubkey.(*ed25519.PubKey),
		}
		genKeys = append(genKeys, genKey)
		createValMsg, err := stakingtypes.NewMsgCreateValidator(
			sdk.ValAddress(addr),
			valPubkey,
			sdk.NewCoin(sdk.DefaultBondDenom, valTokens),
			stakingtypes.NewDescription(nodeDirName, "", "", "", ""),
			stakingtypes.NewCommissionRates(sdk.OneDec(), sdk.OneDec(), sdk.OneDec()),
			sdk.OneInt(),
		)
		if err != nil {
			return err
		}

		txBuilder := clientCtx.TxConfig.NewTxBuilder()
		if err = txBuilder.SetMsgs(createValMsg); err != nil {
			return err
		}

		txBuilder.SetMemo(memo)

		txFactory := tx.Factory{}
		txFactory = txFactory.
			WithChainID(chainID).
			WithMemo(memo).
			WithKeybase(kb).
			WithTxConfig(clientCtx.TxConfig)

		if err = tx.Sign(txFactory, nodeDirName, txBuilder, true); err != nil {
			return err
		}

		txBz, err := clientCtx.TxConfig.TxJSONEncoder()(txBuilder.GetTx())
		if err != nil {
			return err
		}

		if err = writeFile(fmt.Sprintf("%v.json", nodeDirName), gentxsDir, txBz); err != nil {
			return err
		}

		customTemplate := DefaultBabylonTemplate()
		srvconfig.SetConfigTemplate(customTemplate)
		srvconfig.WriteConfigFile(filepath.Join(nodeDir, "config/app.toml"), babylonConfig)

		// create and save client config
		if _, err = app.CreateClientConfig(chainID, keyringBackend, nodeDir); err != nil {
			return err
		}
	}

	if err := initGenFiles(clientCtx, mbm, chainID, genAccounts, genBalances, genFiles,
		genKeys, numValidators, maxActiveValidators, btcConfirmationDepth, btcFinalizationTimeout,
		retrySleepTime, maxRetrySleepTime,
		epochInterval, baseBtcHeaderHex, baseBtcHeaderHeight); err != nil {
		return err
	}

	err := collectGenFiles(
		clientCtx, nodeConfig, chainID, nodeIDs, genKeys, numValidators,
		outputDir, nodeDirPrefix, nodeDaemonHome, genBalIterator,
	)
	if err != nil {
		return err
	}

	cmd.PrintErrf("Successfully initialized %d node directories\n", numValidators)
	return nil
}

func initGenFiles(
	clientCtx client.Context, mbm module.BasicManager, chainID string,
	genAccounts []authtypes.GenesisAccount, genBalances []banktypes.Balance,
	genFiles []string, genKeys []*checkpointingtypes.GenesisKey, numValidators int,
	maxActiveValidators uint32, btcConfirmationDepth uint64, btcFinalizationTimeout uint64,
	retrySleepTime string, maxRetrySleepTime string,
	epochInterval uint64, baseBtcHeaderHex string, baseBtcHeaderHeight uint64,
) error {

	appGenState := mbm.DefaultGenesis(clientCtx.Codec)

	// set the accounts in the genesis state
	var authGenState authtypes.GenesisState
	clientCtx.Codec.MustUnmarshalJSON(appGenState[authtypes.ModuleName], &authGenState)

	accounts, err := authtypes.PackAccounts(genAccounts)
	if err != nil {
		return err
	}

	authGenState.Accounts = accounts
	appGenState[authtypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(&authGenState)

	// set the balances in the genesis state
	var bankGenState banktypes.GenesisState
	clientCtx.Codec.MustUnmarshalJSON(appGenState[banktypes.ModuleName], &bankGenState)

	bankGenState.Balances = banktypes.SanitizeGenesisBalances(genBalances)
	for _, bal := range bankGenState.Balances {
		bankGenState.Supply = bankGenState.Supply.Add(bal.Coins...)
	}
	appGenState[banktypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(&bankGenState)

	// set the BLS keys in the genesis state
	var checkpointGenState checkpointingtypes.GenesisState
	clientCtx.Codec.MustUnmarshalJSON(appGenState[checkpointingtypes.ModuleName], &checkpointGenState)
	checkpointGenState.Params.RetrySleepTime = retrySleepTime
	checkpointGenState.Params.MaxRetrySleepTime = maxRetrySleepTime
	checkpointGenState.GenesisKeys = genKeys
	appGenState[checkpointingtypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(&checkpointGenState)

	// Set the confirmation and finalization parameters
	var btccheckpointGenState btccheckpointtypes.GenesisState
	clientCtx.Codec.MustUnmarshalJSON(appGenState[btccheckpointtypes.ModuleName], &btccheckpointGenState)
	btccheckpointGenState.Params.BtcConfirmationDepth = btcConfirmationDepth
	btccheckpointGenState.Params.CheckpointFinalizationTimeout = btcFinalizationTimeout
	appGenState[btccheckpointtypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(&btccheckpointGenState)

	// set the base BTC header in the genesis state
	baseBtcHeader, err := bbn.NewBTCHeaderBytesFromHex(baseBtcHeaderHex)
	if err != nil {
		return err
	}
	work := btclightclienttypes.CalcWork(&baseBtcHeader)
	baseBtcHeaderInfo := btclightclienttypes.NewBTCHeaderInfo(&baseBtcHeader, baseBtcHeader.Hash(), baseBtcHeaderHeight, &work)
	var btclightclientGenState btclightclienttypes.GenesisState
	clientCtx.Codec.MustUnmarshalJSON(appGenState[btclightclienttypes.ModuleName], &btclightclientGenState)
	btclightclientGenState.BaseBtcHeader = *baseBtcHeaderInfo
	appGenState[btclightclienttypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(&btclightclientGenState)

	// set the epoch interval in the genesis state
	if epochInterval == 0 {
		return errors.New(fmt.Sprintf("Invalid epoch interval %d", epochInterval))
	}
	var epochingGenState epochingtypes.GenesisState
	clientCtx.Codec.MustUnmarshalJSON(appGenState[epochingtypes.ModuleName], &epochingGenState)
	epochingGenState.Params.EpochInterval = epochInterval
	appGenState[epochingtypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(&epochingGenState)

	if maxActiveValidators == 0 {
		return errors.New(fmt.Sprintf("Invalid max active validators value %d", maxActiveValidators))
	}
	var stakingGenState stakingtypes.GenesisState
	clientCtx.Codec.MustUnmarshalJSON(appGenState[stakingtypes.ModuleName], &stakingGenState)
	stakingGenState.Params.MaxValidators = maxActiveValidators
	appGenState[stakingtypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(&stakingGenState)

	appGenStateJSON, err := json.MarshalIndent(appGenState, "", "  ")
	if err != nil {
		return err
	}

	genDoc := types.GenesisDoc{
		ChainID:    chainID,
		AppState:   appGenStateJSON,
		Validators: nil,
	}

	// generate empty genesis files for each validator and save
	for i := 0; i < numValidators; i++ {
		if err := genDoc.SaveAs(genFiles[i]); err != nil {
			return err
		}
	}
	return nil
}

func collectGenFiles(
	clientCtx client.Context, nodeConfig *tmconfig.Config, chainID string,
	nodeIDs []string, genKeys []*checkpointingtypes.GenesisKey, numValidators int,
	outputDir, nodeDirPrefix, nodeDaemonHome string, genBalIterator banktypes.GenesisBalancesIterator,
) error {

	var appState json.RawMessage
	genTime := tmtime.Now()

	for i := 0; i < numValidators; i++ {
		nodeDirName := fmt.Sprintf("%s%d", nodeDirPrefix, i)
		nodeDir := filepath.Join(outputDir, nodeDirName, nodeDaemonHome)
		gentxsDir := filepath.Join(outputDir, "gentxs")
		nodeConfig.Moniker = nodeDirName

		nodeConfig.SetRoot(nodeDir)

		nodeID, valPubKey := nodeIDs[i], genKeys[i].ValPubkey
		initCfg := genutiltypes.NewInitConfig(chainID, gentxsDir, nodeID, valPubKey)

		genDoc, err := types.GenesisDocFromFile(nodeConfig.GenesisFile())
		if err != nil {
			return err
		}

		nodeAppState, err := genutil.GenAppStateFromConfig(clientCtx.Codec, clientCtx.TxConfig, nodeConfig, initCfg, *genDoc, genBalIterator)
		if err != nil {
			return err
		}

		if appState == nil {
			// set the canonical application state (they should not differ)
			appState = nodeAppState
		}

		genFile := nodeConfig.GenesisFile()

		// overwrite each validator's genesis file to have a canonical genesis time
		if err := genutil.ExportGenesisFileWithTime(genFile, chainID, nil, appState, genTime); err != nil {
			return err
		}
	}

	return nil
}

func getIP(i int, startingIPAddr string) (ip string, err error) {
	if len(startingIPAddr) == 0 {
		ip, err = server.ExternalIP()
		if err != nil {
			return "", err
		}
		return ip, nil
	}
	return calculateIP(startingIPAddr, i)
}

func calculateIP(ip string, i int) (string, error) {
	ipv4 := net.ParseIP(ip).To4()
	if ipv4 == nil {
		return "", fmt.Errorf("%v: non ipv4 address", ip)
	}

	for j := 0; j < i; j++ {
		ipv4[3]++
	}

	return ipv4.String(), nil
}

func writeFile(name string, dir string, contents []byte) error {
	writePath := filepath.Join(dir)
	file := filepath.Join(writePath, name)

	err := tmos.EnsureDir(writePath, 0755)
	if err != nil {
		return err
	}

	err = tmos.WriteFile(file, contents, 0644)
	if err != nil {
		return err
	}

	return nil
}
