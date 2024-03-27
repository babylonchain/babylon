package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"

	btcstakingtypes "github.com/babylonchain/babylon/x/btcstaking/types"
	finalitytypes "github.com/babylonchain/babylon/x/finality/types"

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

	appparams "github.com/babylonchain/babylon/app/params"
	bbn "github.com/babylonchain/babylon/types"
	btccheckpointtypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	btclightclienttypes "github.com/babylonchain/babylon/x/btclightclient/types"
	checkpointingtypes "github.com/babylonchain/babylon/x/checkpointing/types"
	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"
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

			genesisState, genesis, err := genutiltypes.GenesisStateFromGenFile(genFile)
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
					genesisCliArgs.BaseBtcHeaderHeight, genesisCliArgs.AllowedReporterAddresses,
					genesisCliArgs.CovenantPKs, genesisCliArgs.CovenantQuorum,
					genesisCliArgs.SlashingAddress, genesisCliArgs.MinSlashingTransactionFeeSat,
					genesisCliArgs.MinCommissionRate, genesisCliArgs.SlashingRate, genesisCliArgs.MaxActiveFinalityProviders,
					genesisCliArgs.MinUnbondingTime, genesisCliArgs.MinUnbondingRate, genesisCliArgs.MinPubRand, genesisCliArgs.InflationRateChange,
					genesisCliArgs.InflationMin, genesisCliArgs.InflationMax, genesisCliArgs.GoalBonded,
					genesisCliArgs.BlocksPerYear, genesisCliArgs.GenesisTime, genesisCliArgs.BlockGasLimit, genesisCliArgs.VoteExtensionEnableHeight)
			} else if network == "mainnet" {
				// TODO: mainnet genesis params
				panic("Mainnet params not implemented.")
			} else {
				return fmt.Errorf("please choose testnet or mainnet")
			}

			err = PrepareGenesis(clientCtx, genesisState, genesis, genesisParams, chainID)
			if err != nil {
				return fmt.Errorf("failed to prepare genesis: %w", err)
			}

			if err = mbm.ValidateGenesis(clientCtx.Codec, clientCtx.TxConfig, genesisState); err != nil {
				return fmt.Errorf("error validating genesis file: %s", err)
			}

			return genutil.ExportGenesisFile(genesis, genFile)
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	addGenesisFlags(cmd)

	return cmd
}

func PrepareGenesis(
	clientCtx client.Context,
	genesisState map[string]json.RawMessage,
	genesis *genutiltypes.AppGenesis,
	genesisParams GenesisParams,
	chainID string,
) error {
	if genesis == nil {
		return fmt.Errorf("provided genesis must not be nil")
	}

	depCdc := clientCtx.Codec
	cdc := depCdc

	// Add ChainID
	genesis.ChainID = chainID
	genesis.GenesisTime = genesisParams.GenesisTime

	if genesis.Consensus == nil {
		genesis.Consensus = genutiltypes.NewConsensusGenesis(comettypes.DefaultConsensusParams().ToProto(), nil)

	}

	// Set gas limit
	genesis.Consensus.Params.Block.MaxGas = genesisParams.BlockGasLimit
	genesis.Consensus.Params.ABCI.VoteExtensionsEnableHeight = genesisParams.VoteExtensionsEnableHeight

	// Set the confirmation and finalization parameters
	btccheckpointGenState := btccheckpointtypes.DefaultGenesis()
	btccheckpointGenState.Params = genesisParams.BtccheckpointParams
	genesisState[btccheckpointtypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(btccheckpointGenState)

	// btclightclient genesis
	btclightclientGenState := btclightclienttypes.DefaultGenesis()
	btclightclientGenState.BtcHeaders = []*btclightclienttypes.BTCHeaderInfo{&genesisParams.BtclightclientBaseBtcHeader}
	btclightclientGenState.Params = genesisParams.BtclightclientParams
	genesisState[btclightclienttypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(btclightclientGenState)

	// epoching module genesis
	epochingGenState := epochingtypes.DefaultGenesis()
	epochingGenState.Params = genesisParams.EpochingParams
	genesisState[epochingtypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(epochingGenState)

	// checkpointing module genesis
	checkpointingGenState := checkpointingtypes.DefaultGenesis()
	checkpointingGenState.GenesisKeys = genesisParams.CheckpointingGenKeys
	genesisState[checkpointingtypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(checkpointingGenState)

	// btcstaking module genesis
	btcstakingGenState := btcstakingtypes.DefaultGenesis()
	// here we can start only from single params, which will be initially labelled version 0
	btcstakingGenState.Params = []*btcstakingtypes.Params{&genesisParams.BtcstakingParams}
	genesisState[btcstakingtypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(btcstakingGenState)

	// finality module genesis
	finalityGenState := finalitytypes.DefaultGenesis()
	finalityGenState.Params = genesisParams.FinalityParams
	genesisState[finalitytypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(finalityGenState)

	// staking module genesis
	stakingGenState := stakingtypes.GetGenesisStateFromAppState(depCdc, genesisState)
	clientCtx.Codec.MustUnmarshalJSON(genesisState[stakingtypes.ModuleName], stakingGenState)
	stakingGenState.Params = genesisParams.StakingParams
	genesisState[stakingtypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(stakingGenState)

	// mint module genesis
	mintGenState := minttypes.DefaultGenesisState()
	mintGenState.Params = genesisParams.MintParams
	genesisState[minttypes.ModuleName] = cdc.MustMarshalJSON(mintGenState)

	// distribution module genesis
	distributionGenState := distributiontypes.DefaultGenesisState()
	distributionGenState.Params = genesisParams.DistributionParams
	genesisState[distributiontypes.ModuleName] = cdc.MustMarshalJSON(distributionGenState)

	// gov module genesis
	govGenState := govv1.DefaultGenesisState()
	govGenState.Params = &genesisParams.GovParams
	genesisState[govtypes.ModuleName] = cdc.MustMarshalJSON(govGenState)

	// crisis module genesis
	crisisGenState := crisistypes.DefaultGenesisState()
	crisisGenState.ConstantFee = genesisParams.CrisisConstantFee
	genesisState[crisistypes.ModuleName] = cdc.MustMarshalJSON(crisisGenState)

	// auth module genesis
	authGenState := authtypes.DefaultGenesisState()
	authGenState.Accounts = genesisParams.AuthAccounts
	genesisState[authtypes.ModuleName] = cdc.MustMarshalJSON(authGenState)

	// bank module genesis
	bankGenState := banktypes.DefaultGenesisState()
	bankGenState.Balances = banktypes.SanitizeGenesisBalances(genesisParams.BankGenBalances)
	for _, bal := range bankGenState.Balances {
		bankGenState.Supply = bankGenState.Supply.Add(bal.Coins...)
	}
	genesisState[banktypes.ModuleName] = cdc.MustMarshalJSON(bankGenState)

	appGenStateJSON, err := json.MarshalIndent(genesisState, "", "  ")

	if err != nil {
		return err
	}

	genesis.AppState = appGenStateJSON

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
	BtclightclientParams        btclightclienttypes.Params
	BlockGasLimit               int64
	VoteExtensionsEnableHeight  int64
}

func TestnetGenesisParams(maxActiveValidators uint32, btcConfirmationDepth uint64,
	btcFinalizationTimeout uint64, checkpointTag string, epochInterval uint64, baseBtcHeaderHex string,
	baseBtcHeaderHeight uint64, allowedReporters []string, covenantPKs []string, covenantQuorum uint32, slashingAddress string, minSlashingFee int64,
	minCommissionRate sdkmath.LegacyDec, slashingRate sdkmath.LegacyDec, maxActiveFinalityProviders uint32, minUnbondingTime uint16, minUnbondingRate sdkmath.LegacyDec,
	minPubRand uint64, inflationRateChange float64,
	inflationMin float64, inflationMax float64, goalBonded float64,
	blocksPerYear uint64, genesisTime time.Time, blockGasLimit int64, voteExtensionEnableHeight int64) GenesisParams {

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
	genParams.MintParams.InflationRateChange = sdkmath.LegacyMustNewDecFromStr(fmt.Sprintf("%f", inflationRateChange))
	genParams.MintParams.InflationMin = sdkmath.LegacyMustNewDecFromStr(fmt.Sprintf("%f", inflationMin))
	genParams.MintParams.InflationMax = sdkmath.LegacyMustNewDecFromStr(fmt.Sprintf("%f", inflationMax))
	genParams.MintParams.GoalBonded = sdkmath.LegacyMustNewDecFromStr(fmt.Sprintf("%f", goalBonded))

	genParams.GovParams = govv1.DefaultParams()
	// TODO investigate those numbers
	genParams.GovParams.MinDeposit = sdk.NewCoins(sdk.NewCoin(
		genParams.NativeCoinMetadatas[0].Base,
		sdkmath.NewInt(2_500_000_000),
	))
	genParams.GovParams.ExpeditedMinDeposit = sdk.NewCoins(sdk.NewCoin(
		genParams.NativeCoinMetadatas[0].Base,
		sdkmath.NewInt(10_000_000_000),
	))

	genParams.CrisisConstantFee = sdk.NewCoin(
		genParams.NativeCoinMetadatas[0].Base,
		sdkmath.NewInt(500_000_000_000),
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

	params, err := btclightclienttypes.NewParamsValidate(allowedReporters)

	if err != nil {
		panic(err)
	}

	genParams.BtclightclientBaseBtcHeader = *baseBtcHeaderInfo
	genParams.BtclightclientParams = params

	genParams.BtcstakingParams = btcstakingtypes.DefaultParams()
	covenantPKsBIP340 := make([]bbn.BIP340PubKey, 0, len(covenantPKs))
	for _, pkHex := range covenantPKs {
		pk, err := bbn.NewBIP340PubKeyFromHex(pkHex)
		if err != nil {
			panic(err)
		}
		covenantPKsBIP340 = append(covenantPKsBIP340, *pk)
	}
	genParams.BtcstakingParams.CovenantPks = covenantPKsBIP340
	genParams.BtcstakingParams.CovenantQuorum = covenantQuorum
	genParams.BtcstakingParams.SlashingAddress = slashingAddress
	genParams.BtcstakingParams.MinSlashingTxFeeSat = minSlashingFee
	genParams.BtcstakingParams.MinCommissionRate = minCommissionRate
	genParams.BtcstakingParams.SlashingRate = slashingRate
	genParams.BtcstakingParams.MaxActiveFinalityProviders = maxActiveFinalityProviders
	genParams.BtcstakingParams.MinUnbondingTime = uint32(minUnbondingTime)
	genParams.BtcstakingParams.MinUnbondingRate = minUnbondingRate
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
	genParams.VoteExtensionsEnableHeight = voteExtensionEnableHeight
	return genParams
}
