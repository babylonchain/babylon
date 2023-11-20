package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	btcstakingtypes "github.com/babylonchain/babylon/x/btcstaking/types"
	finalitytypes "github.com/babylonchain/babylon/x/finality/types"

	appparams "github.com/babylonchain/babylon/app/params"
	bbn "github.com/babylonchain/babylon/types"
	btccheckpointtypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	btclightclienttypes "github.com/babylonchain/babylon/x/btclightclient/types"
	checkpointingtypes "github.com/babylonchain/babylon/x/checkpointing/types"
	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"
	comettypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/spf13/cobra"
)

func PrepareGenesisCmd(defaultNodeHome string, mbm module.BasicManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prepare-genesis <testnet|mainnet> <chain-id>",
		Args:  cobra.ExactArgs(2),
		Short: "Prepare a genesis file",
		Long: `Prepare a genesis file.
Example:
	babylond prepare-genesis testnet babylon-test-1
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			genesisCliArgs := parseGenesisFlags(cmd)

			genFile := config.GenesisFile()

			appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %s", err)
			}

			network := args[0]
			chainID := args[1]

			var genesisParams GenesisParams
			if network == "testnet" {
				genesisParams = TestnetGenesisParams(genesisCliArgs.MaxActiveValidators,
					genesisCliArgs.BtcConfirmationDepth, genesisCliArgs.BtcFinalizationTimeout, genesisCliArgs.CheckpointTag,
					genesisCliArgs.EpochInterval, genesisCliArgs.BaseBtcHeaderHex,
					genesisCliArgs.BaseBtcHeaderHeight, genesisCliArgs.CovenantPK,
					genesisCliArgs.SlashingAddress, genesisCliArgs.MinSlashingTransactionFeeSat,
					genesisCliArgs.MinCommissionRate, genesisCliArgs.SlashingRate,
					genesisCliArgs.MinPubRand, genesisCliArgs.InflationRateChange,
					genesisCliArgs.InflationMin, genesisCliArgs.InflationMax, genesisCliArgs.GoalBonded,
					genesisCliArgs.BlocksPerYear, genesisCliArgs.GenesisTime, genesisCliArgs.BlockGasLimit)
			} else if network == "mainnet" {
				// TODO: mainnet genesis params
				panic("Mainnet params not implemented.")
			} else {
				return fmt.Errorf("please choose testnet or mainnet")
			}

			err = PrepareGenesis(clientCtx, appState, genDoc, genesisParams, chainID)
			if err != nil {
				return fmt.Errorf("failed to prepare genesis: %w", err)
			}

			if err = mbm.ValidateGenesis(clientCtx.Codec, clientCtx.TxConfig, appState); err != nil {
				return fmt.Errorf("error validating genesis file: %s", err)
			}

			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	addGenesisFlags(cmd)

	return cmd
}

func PrepareGenesis(
	clientCtx client.Context,
	appState map[string]json.RawMessage,
	genDoc *comettypes.GenesisDoc,
	genesisParams GenesisParams,
	chainID string,
) error {
	if genDoc == nil {
		return fmt.Errorf("provided genesis must not be nil")
	}

	depCdc := clientCtx.Codec
	cdc := depCdc

	// Add ChainID
	genDoc.ChainID = chainID
	genDoc.GenesisTime = genesisParams.GenesisTime

	if genDoc.ConsensusParams == nil {
		genDoc.ConsensusParams = comettypes.DefaultConsensusParams()
	}

	// Set gas limit
	genDoc.ConsensusParams.Block.MaxGas = genesisParams.BlockGasLimit

	// Set the confirmation and finalization parameters
	btccheckpointGenState := btccheckpointtypes.DefaultGenesis()
	btccheckpointGenState.Params = genesisParams.BtccheckpointParams
	appState[btccheckpointtypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(btccheckpointGenState)

	// btclightclient genesis
	btclightclientGenState := btclightclienttypes.DefaultGenesis()
	btclightclientGenState.BaseBtcHeader = genesisParams.BtclightclientBaseBtcHeader
	appState[btclightclienttypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(btclightclientGenState)

	// epoching module genesis
	epochingGenState := epochingtypes.DefaultGenesis()
	epochingGenState.Params = genesisParams.EpochingParams
	appState[epochingtypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(epochingGenState)

	// checkpointing module genesis
	checkpointingGenState := checkpointingtypes.DefaultGenesis()
	checkpointingGenState.GenesisKeys = genesisParams.CheckpointingGenKeys
	appState[checkpointingtypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(checkpointingGenState)

	// btcstaking module genesis
	btcstakingGenState := btcstakingtypes.DefaultGenesis()
	btcstakingGenState.Params = genesisParams.BtcstakingParams
	appState[btcstakingtypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(btcstakingGenState)

	// finality module genesis
	finalityGenState := finalitytypes.DefaultGenesis()
	finalityGenState.Params = genesisParams.FinalityParams
	appState[finalitytypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(finalityGenState)

	// staking module genesis
	stakingGenState := stakingtypes.GetGenesisStateFromAppState(depCdc, appState)
	clientCtx.Codec.MustUnmarshalJSON(appState[stakingtypes.ModuleName], stakingGenState)
	stakingGenState.Params = genesisParams.StakingParams
	appState[stakingtypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(stakingGenState)

	// mint module genesis
	mintGenState := minttypes.DefaultGenesisState()
	mintGenState.Params = genesisParams.MintParams
	appState[minttypes.ModuleName] = cdc.MustMarshalJSON(mintGenState)

	// distribution module genesis
	distributionGenState := distributiontypes.DefaultGenesisState()
	distributionGenState.Params = genesisParams.DistributionParams
	appState[distributiontypes.ModuleName] = cdc.MustMarshalJSON(distributionGenState)

	// gov module genesis
	govGenState := govv1.DefaultGenesisState()
	govGenState.Params = &genesisParams.GovParams
	appState[govtypes.ModuleName] = cdc.MustMarshalJSON(govGenState)

	// crisis module genesis
	crisisGenState := crisistypes.DefaultGenesisState()
	crisisGenState.ConstantFee = genesisParams.CrisisConstantFee
	appState[crisistypes.ModuleName] = cdc.MustMarshalJSON(crisisGenState)

	// auth module genesis
	authGenState := authtypes.DefaultGenesisState()
	authGenState.Accounts = genesisParams.AuthAccounts
	appState[authtypes.ModuleName] = cdc.MustMarshalJSON(authGenState)

	// bank module genesis
	bankGenState := banktypes.DefaultGenesisState()
	bankGenState.Balances = banktypes.SanitizeGenesisBalances(genesisParams.BankGenBalances)
	for _, bal := range bankGenState.Balances {
		bankGenState.Supply = bankGenState.Supply.Add(bal.Coins...)
	}
	appState[banktypes.ModuleName] = cdc.MustMarshalJSON(bankGenState)

	appGenStateJSON, err := json.MarshalIndent(appState, "", "  ")

	if err != nil {
		return err
	}

	genDoc.AppState = appGenStateJSON

	return nil
}

type GenesisParams struct {
	GenesisTime time.Time

	NativeCoinMetadatas []banktypes.Metadata

	StakingParams      stakingtypes.Params
	MintParams         minttypes.Params
	DistributionParams distributiontypes.Params
	GovParams          govv1.Params

	CrisisConstantFee    sdk.Coin
	AuthAccounts         []*cdctypes.Any
	BankGenBalances      []banktypes.Balance
	CheckpointingGenKeys []*checkpointingtypes.GenesisKey

	BtccheckpointParams         btccheckpointtypes.Params
	EpochingParams              epochingtypes.Params
	BtcstakingParams            btcstakingtypes.Params
	FinalityParams              finalitytypes.Params
	BtclightclientBaseBtcHeader btclightclienttypes.BTCHeaderInfo
	BlockGasLimit               int64
}

func TestnetGenesisParams(maxActiveValidators uint32, btcConfirmationDepth uint64,
	btcFinalizationTimeout uint64, checkpointTag string, epochInterval uint64, baseBtcHeaderHex string,
	baseBtcHeaderHeight uint64, covenantPk string, slashingAddress string, minSlashingFee int64,
	minCommissionRate sdk.Dec, slashingRate sdk.Dec, minPubRand uint64, inflationRateChange float64,
	inflationMin float64, inflationMax float64, goalBonded float64,
	blocksPerYear uint64, genesisTime time.Time, blockGasLimit int64) GenesisParams {

	genParams := GenesisParams{}

	genParams.GenesisTime = genesisTime

	genParams.NativeCoinMetadatas = []banktypes.Metadata{
		{
			Description: "The native token of Babylon",
			DenomUnits: []*banktypes.DenomUnit{
				{
					Denom:    appparams.BaseCoinUnit,
					Exponent: 0,
					Aliases:  nil,
				},
				{
					Denom:    appparams.HumanCoinUnit,
					Exponent: appparams.BbnExponent,
					Aliases:  nil,
				},
			},
			Base:    appparams.BaseCoinUnit,
			Display: appparams.HumanCoinUnit,
		},
	}

	genParams.StakingParams = stakingtypes.DefaultParams()
	if maxActiveValidators == 0 {
		panic(fmt.Sprintf("Invalid max active validators value %d", maxActiveValidators))
	}
	genParams.StakingParams.MaxValidators = maxActiveValidators
	genParams.StakingParams.BondDenom = genParams.NativeCoinMetadatas[0].Base

	genParams.MintParams = minttypes.DefaultParams()
	genParams.MintParams.MintDenom = genParams.NativeCoinMetadatas[0].Base
	genParams.MintParams.BlocksPerYear = blocksPerYear
	// This should always work as inflation rate is already a float64
	genParams.MintParams.InflationRateChange = sdk.MustNewDecFromStr(fmt.Sprintf("%f", inflationRateChange))
	genParams.MintParams.InflationMin = sdk.MustNewDecFromStr(fmt.Sprintf("%f", inflationMin))
	genParams.MintParams.InflationMax = sdk.MustNewDecFromStr(fmt.Sprintf("%f", inflationMax))
	genParams.MintParams.GoalBonded = sdk.MustNewDecFromStr(fmt.Sprintf("%f", goalBonded))

	genParams.GovParams = govv1.DefaultParams()
	genParams.GovParams.MinDeposit = sdk.NewCoins(sdk.NewCoin(
		genParams.NativeCoinMetadatas[0].Base,
		sdk.NewInt(2_500_000_000),
	))

	genParams.CrisisConstantFee = sdk.NewCoin(
		genParams.NativeCoinMetadatas[0].Base,
		sdk.NewInt(500_000_000_000),
	)

	genParams.BtccheckpointParams = btccheckpointtypes.DefaultParams()
	genParams.BtccheckpointParams.BtcConfirmationDepth = btcConfirmationDepth
	genParams.BtccheckpointParams.CheckpointFinalizationTimeout = btcFinalizationTimeout
	genParams.BtccheckpointParams.CheckpointTag = checkpointTag

	if err := genParams.BtccheckpointParams.Validate(); err != nil {
		panic(err)
	}

	// set the base BTC header in the genesis state
	baseBtcHeader, err := bbn.NewBTCHeaderBytesFromHex(baseBtcHeaderHex)
	if err != nil {
		panic(err)
	}
	work := btclightclienttypes.CalcWork(&baseBtcHeader)
	baseBtcHeaderInfo := btclightclienttypes.NewBTCHeaderInfo(&baseBtcHeader, baseBtcHeader.Hash(), baseBtcHeaderHeight, &work)
	genParams.BtclightclientBaseBtcHeader = *baseBtcHeaderInfo

	genParams.BtcstakingParams = btcstakingtypes.DefaultParams()
	covenantPK, err := bbn.NewBIP340PubKeyFromHex(covenantPk)
	if err != nil {
		panic(err)
	}
	genParams.BtcstakingParams.CovenantPks = []bbn.BIP340PubKey{*covenantPK}
	genParams.BtcstakingParams.CovenantQuorum = 1 // TODO: multisig covenant
	genParams.BtcstakingParams.SlashingAddress = slashingAddress
	genParams.BtcstakingParams.MinSlashingTxFeeSat = minSlashingFee
	genParams.BtcstakingParams.MinCommissionRate = minCommissionRate
	genParams.BtcstakingParams.SlashingRate = slashingRate
	if err := genParams.BtcstakingParams.Validate(); err != nil {
		panic(err)
	}

	genParams.FinalityParams = finalitytypes.DefaultParams()
	genParams.FinalityParams.MinPubRand = minPubRand
	if err := genParams.FinalityParams.Validate(); err != nil {
		panic(err)
	}

	if epochInterval == 0 {
		panic(fmt.Sprintf("Invalid epoch interval %d", epochInterval))
	}
	genParams.EpochingParams = epochingtypes.DefaultParams()
	genParams.EpochingParams.EpochInterval = epochInterval
	if err := genParams.EpochingParams.Validate(); err != nil {
		panic(err)
	}

	genParams.BlockGasLimit = blockGasLimit
	return genParams
}
