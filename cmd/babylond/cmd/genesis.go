package cmd

import (
	"encoding/json"
	"fmt"
	appparams "github.com/babylonchain/babylon/app/params"
	bbn "github.com/babylonchain/babylon/types"
	btccheckpointtypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	btclightclienttypes "github.com/babylonchain/babylon/x/btclightclient/types"
	checkpointingtypes "github.com/babylonchain/babylon/x/checkpointing/types"
	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func PrepareGenesis(clientCtx client.Context, appState map[string]json.RawMessage, genesisParams GenesisParams) (map[string]json.RawMessage, error) {
	depCdc := clientCtx.Codec
	cdc := depCdc

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
	govGenState := govtypes.DefaultGenesisState()
	govGenState.DepositParams = genesisParams.GovParams.DepositParams
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

	// return appState
	return appState, nil
}

type GenesisParams struct {
	NativeCoinMetadatas []banktypes.Metadata

	StakingParams      stakingtypes.Params
	MintParams         minttypes.Params
	DistributionParams distributiontypes.Params
	GovParams          govtypes.Params

	CrisisConstantFee    sdk.Coin
	AuthAccounts         []*types.Any
	BankGenBalances      []banktypes.Balance
	CheckpointingGenKeys []*checkpointingtypes.GenesisKey

	BtccheckpointParams         btccheckpointtypes.Params
	EpochingParams              epochingtypes.Params
	BtclightclientBaseBtcHeader btclightclienttypes.BTCHeaderInfo
}

func TestnetGenesisParams(maxActiveValidators uint32, btcConfirmationDepth uint64,
	btcFinalizationTimeout uint64, epochInterval uint64, baseBtcHeaderHex string, baseBtcHeaderHeight uint64) GenesisParams {
	genParams := GenesisParams{}

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
	// Babylon should enforce this value to be 0. However Cosmos enforces it to be positive so we use the smallest value 1
	// Instead the timing of unbonding is decided by checkpoint states
	genParams.StakingParams.UnbondingTime = 1

	genParams.MintParams = minttypes.DefaultParams()
	genParams.MintParams.MintDenom = genParams.NativeCoinMetadatas[0].Base

	genParams.GovParams = govtypes.DefaultParams()
	genParams.GovParams.DepositParams.MinDeposit = sdk.NewCoins(sdk.NewCoin(
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

	// set the base BTC header in the genesis state
	baseBtcHeader, err := bbn.NewBTCHeaderBytesFromHex(baseBtcHeaderHex)
	if err != nil {
		panic(err)
	}
	work := btclightclienttypes.CalcWork(&baseBtcHeader)
	baseBtcHeaderInfo := btclightclienttypes.NewBTCHeaderInfo(&baseBtcHeader, baseBtcHeader.Hash(), baseBtcHeaderHeight, &work)
	genParams.BtclightclientBaseBtcHeader = *baseBtcHeaderInfo

	if epochInterval == 0 {
		panic(fmt.Sprintf("Invalid epoch interval %d", epochInterval))
	}
	genParams.EpochingParams = epochingtypes.DefaultParams()
	genParams.EpochingParams.EpochInterval = epochInterval

	return genParams
}
