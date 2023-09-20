package cmd

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"time"

	babylonApp "github.com/babylonchain/babylon/app"
	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	btcstypes "github.com/babylonchain/babylon/x/btcstaking/types"
	tmrand "github.com/cometbft/cometbft/libs/rand"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
)

const (
	flagMaxActiveValidators    = "max-active-validators"
	flagBtcConfirmationDepth   = "btc-confirmation-depth"
	flagEpochInterval          = "epoch-interval"
	flagBtcFinalizationTimeout = "btc-finalization-timeout"
	flagCheckpointTag          = "checkpoint-tag"
	flagBaseBtcHeaderHex       = "btc-base-header"
	flagBaseBtcHeaderHeight    = "btc-base-header-height"
	flagInflationRateChange    = "inflation-rate-change"
	flagInflationMax           = "inflation-max"
	flagInflationMin           = "inflation-min"
	flagGoalBonded             = "goal-bonded"
	flagBlocksPerYear          = "blocks-per-year"
	flagGenesisTime            = "genesis-time"
	flagBlockGasLimit          = "block-gas-limit"
	flagJuryPk                 = "jury-pk"
	flagSlashingAddress        = "slashing-address"
	flagMinSlashingFee         = "min-slashing-fee-sat"
	flagMinPubRand             = "min-pub-rand"
	flagMinCommissionRate      = "min-commission-rate"
)

type GenesisCLIArgs struct {
	ChainID                      string
	MaxActiveValidators          uint32
	BtcConfirmationDepth         uint64
	BtcFinalizationTimeout       uint64
	CheckpointTag                string
	EpochInterval                uint64
	BaseBtcHeaderHex             string
	BaseBtcHeaderHeight          uint64
	InflationRateChange          float64
	InflationMax                 float64
	InflationMin                 float64
	GoalBonded                   float64
	BlocksPerYear                uint64
	GenesisTime                  time.Time
	BlockGasLimit                int64
	JuryPK                       string
	SlashingAddress              string
	MinSlashingTransactionFeeSat int64
	MinPubRand                   uint64
	MinCommissionRate            sdk.Dec
}

func addGenesisFlags(cmd *cobra.Command) {
	cmd.Flags().String(flags.FlagChainID, "", "genesis file chain-id, if left blank will be randomly created")
	// staking flags
	cmd.Flags().Uint32(flagMaxActiveValidators, 10, "Maximum number of validators.")
	// btccheckpoint flags
	cmd.Flags().Uint64(flagBtcConfirmationDepth, 6, "Confirmation depth for Bitcoin headers.")
	cmd.Flags().Uint64(flagBtcFinalizationTimeout, 20, "Finalization timeout for Bitcoin headers.")
	cmd.Flags().String(flagCheckpointTag, btcctypes.DefaultCheckpointTag, "Hex encoded tag for babylon checkpoint on btc")
	// epoch args
	cmd.Flags().Uint64(flagEpochInterval, 400, "Number of blocks between epochs. Must be more than 0.")
	// btclightclient args
	// Genesis header for the simnet
	cmd.Flags().String(flagBaseBtcHeaderHex, "0100000000000000000000000000000000000000000000000000000000000000000000003ba3edfd7a7b12b27ac72c3e67768f617fc81bc3888a51323a9fb8aa4b1e5e4a45068653ffff7f2002000000", "Hex of the base Bitcoin header.")
	cmd.Flags().Uint64(flagBaseBtcHeaderHeight, 0, "Height of the base Bitcoin header.")
	// btcstaking args
	cmd.Flags().String(flagJuryPk, btcstypes.DefaultParams().JuryPk.MarshalHex(), "Bitcoin staking jury public key")
	cmd.Flags().String(flagSlashingAddress, btcstypes.DefaultParams().SlashingAddress, "Bitcoin staking slashing address")
	cmd.Flags().Int64(flagMinSlashingFee, 1000, "Bitcoin staking minimum slashing fee")
	cmd.Flags().String(flagMinCommissionRate, "0", "Bitcoin staking validator minimum commission rate")
	// finality args
	cmd.Flags().Uint64(flagMinPubRand, 100, "Bitcoin staking minimum public randomness commit")
	// inflation args
	cmd.Flags().Float64(flagInflationRateChange, 0.13, "Inflation rate change")
	cmd.Flags().Float64(flagInflationMax, 0.2, "Maximum inflation")
	cmd.Flags().Float64(flagInflationMin, 0.07, "Minimum inflation")
	cmd.Flags().Float64(flagGoalBonded, 0.67, "Bonded tokens goal")
	cmd.Flags().Uint64(flagBlocksPerYear, 6311520, "Blocks per year")
	// genesis args
	cmd.Flags().Int64(flagGenesisTime, time.Now().Unix(), "Genesis time")
	// blocks args
	cmd.Flags().Int64(flagBlockGasLimit, babylonApp.DefaultGasLimit, "Block gas limit")
}

func parseGenesisFlags(cmd *cobra.Command) *GenesisCLIArgs {
	chainID, _ := cmd.Flags().GetString(flags.FlagChainID)
	maxActiveValidators, _ := cmd.Flags().GetUint32(flagMaxActiveValidators)
	btcConfirmationDepth, _ := cmd.Flags().GetUint64(flagBtcConfirmationDepth)
	btcFinalizationTimeout, _ := cmd.Flags().GetUint64(flagBtcFinalizationTimeout)
	checkpointTag, _ := cmd.Flags().GetString(flagCheckpointTag)
	epochInterval, _ := cmd.Flags().GetUint64(flagEpochInterval)
	baseBtcHeaderHex, _ := cmd.Flags().GetString(flagBaseBtcHeaderHex)
	baseBtcHeaderHeight, _ := cmd.Flags().GetUint64(flagBaseBtcHeaderHeight)
	juryPk, _ := cmd.Flags().GetString(flagJuryPk)
	slashingAddress, _ := cmd.Flags().GetString(flagSlashingAddress)
	minSlashingFee, _ := cmd.Flags().GetInt64(flagMinSlashingFee)
	minCommissionRate, _ := cmd.Flags().GetString(flagMinCommissionRate)
	minPubRand, _ := cmd.Flags().GetUint64(flagMinPubRand)
	genesisTimeUnix, _ := cmd.Flags().GetInt64(flagGenesisTime)
	inflationRateChange, _ := cmd.Flags().GetFloat64(flagInflationRateChange)
	inflationMax, _ := cmd.Flags().GetFloat64(flagInflationMax)
	inflationMin, _ := cmd.Flags().GetFloat64(flagInflationMin)
	goalBonded, _ := cmd.Flags().GetFloat64(flagGoalBonded)
	blocksPerYear, _ := cmd.Flags().GetUint64(flagBlocksPerYear)
	blockGasLimit, _ := cmd.Flags().GetInt64(flagBlockGasLimit)

	if chainID == "" {
		chainID = "chain-" + tmrand.NewRand().Str(6)
	}

	genesisTime := time.Unix(genesisTimeUnix, 0)

	return &GenesisCLIArgs{
		ChainID:                      chainID,
		MaxActiveValidators:          maxActiveValidators,
		BtcConfirmationDepth:         btcConfirmationDepth,
		BtcFinalizationTimeout:       btcFinalizationTimeout,
		CheckpointTag:                checkpointTag,
		EpochInterval:                epochInterval,
		BaseBtcHeaderHeight:          baseBtcHeaderHeight,
		BaseBtcHeaderHex:             baseBtcHeaderHex,
		JuryPK:                       juryPk,
		SlashingAddress:              slashingAddress,
		MinSlashingTransactionFeeSat: minSlashingFee,
		MinCommissionRate:            sdk.MustNewDecFromStr(minCommissionRate),
		MinPubRand:                   minPubRand,
		GenesisTime:                  genesisTime,
		InflationRateChange:          inflationRateChange,
		InflationMax:                 inflationMax,
		InflationMin:                 inflationMin,
		GoalBonded:                   goalBonded,
		BlocksPerYear:                blocksPerYear,
		BlockGasLimit:                blockGasLimit,
	}
}
