package cmd

import (
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	"time"
)

const (
	flagMaxActiveValidators    = "max-active-validators"
	flagBtcConfirmationDepth   = "btc-confirmation-depth"
	flagEpochInterval          = "epoch-interval"
	flagBtcFinalizationTimeout = "btc-finalization-timeout"
	flagBaseBtcHeaderHex       = "btc-base-header"
	flagBaseBtcHeaderHeight    = "btc-base-header-height"
	flagGenesisTime            = "genesis-time"
)

type GenesisCLIArgs struct {
	ChainID                string
	MaxActiveValidators    uint32
	BtcConfirmationDepth   uint64
	BtcFinalizationTimeout uint64
	EpochInterval          uint64
	BaseBtcHeaderHex       string
	BaseBtcHeaderHeight    uint64
	GenesisTime            time.Time
}

func addGenesisFlags(cmd *cobra.Command) {
	cmd.Flags().String(flags.FlagChainID, "", "genesis file chain-id, if left blank will be randomly created")
	// staking flags
	cmd.Flags().Uint32(flagMaxActiveValidators, 10, "Maximum number of validators.")
	// btccheckpoint flags
	cmd.Flags().Uint64(flagBtcConfirmationDepth, 6, "Confirmation depth for Bitcoin headers.")
	cmd.Flags().Uint64(flagBtcFinalizationTimeout, 20, "Finalization timeout for Bitcoin headers.")
	// epoch args
	cmd.Flags().Uint64(flagEpochInterval, 400, "Number of blocks between epochs. Must be more than 0.")
	// btclightclient args
	// Genesis header for the simnet
	cmd.Flags().String(flagBaseBtcHeaderHex, "0100000000000000000000000000000000000000000000000000000000000000000000003ba3edfd7a7b12b27ac72c3e67768f617fc81bc3888a51323a9fb8aa4b1e5e4a45068653ffff7f2002000000", "Hex of the base Bitcoin header.")
	cmd.Flags().Uint64(flagBaseBtcHeaderHeight, 0, "Height of the base Bitcoin header.")
	cmd.Flags().Int64(flagGenesisTime, time.Now().Unix(), "Genesis time")
}

func parseGenesisFlags(cmd *cobra.Command) *GenesisCLIArgs {
	chainID, _ := cmd.Flags().GetString(flags.FlagChainID)
	maxActiveValidators, _ := cmd.Flags().GetUint32(flagMaxActiveValidators)
	btcConfirmationDepth, _ := cmd.Flags().GetUint64(flagBtcConfirmationDepth)
	btcFinalizationTimeout, _ := cmd.Flags().GetUint64(flagBtcFinalizationTimeout)
	epochInterval, _ := cmd.Flags().GetUint64(flagEpochInterval)
	baseBtcHeaderHex, _ := cmd.Flags().GetString(flagBaseBtcHeaderHex)
	baseBtcHeaderHeight, _ := cmd.Flags().GetUint64(flagBaseBtcHeaderHeight)
	genesisTimeUnix, _ := cmd.Flags().GetInt64(flagGenesisTime)

	if chainID == "" {
		chainID = "chain-" + tmrand.NewRand().Str(6)
	}

	genesisTime := time.Unix(genesisTimeUnix, 0)

	return &GenesisCLIArgs{
		ChainID:                chainID,
		MaxActiveValidators:    maxActiveValidators,
		BtcConfirmationDepth:   btcConfirmationDepth,
		BtcFinalizationTimeout: btcFinalizationTimeout,
		EpochInterval:          epochInterval,
		BaseBtcHeaderHeight:    baseBtcHeaderHeight,
		BaseBtcHeaderHex:       baseBtcHeaderHex,
		GenesisTime:            genesisTime,
	}
}
