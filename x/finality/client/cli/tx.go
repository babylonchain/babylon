package cli

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/finality/types"
	cmtcrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		NewCommitPubRandListCmd(),
		NewAddFinalitySigCmd(),
	)

	return cmd
}

func NewCommitPubRandListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commit-pubrand-list [fp_btc_pk] [start_height] [num_pub_rand] [commitment] [sig]",
		Args:  cobra.MinimumNArgs(5),
		Short: "Commit a list of public randomness",
		Long: strings.TrimSpace(
			`Commit a list of public randomness.`, // TODO: example
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// get finality provider BTC PK
			fpBTCPK, err := bbn.NewBIP340PubKeyFromHex(args[0])
			if err != nil {
				return err
			}

			// get start height
			startHeight, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			numPubRand, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return err
			}

			commitment, err := hex.DecodeString(args[3])
			if err != nil {
				return err
			}

			// get signature
			sig, err := bbn.NewBIP340SignatureFromHex(args[4])
			if err != nil {
				return err
			}

			msg := types.MsgCommitPubRandList{
				Signer:      clientCtx.FromAddress.String(),
				FpBtcPk:     fpBTCPK,
				StartHeight: startHeight,
				NumPubRand:  numPubRand,
				Commitment:  commitment,
				Sig:         sig,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewAddFinalitySigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-finality-sig [fp_btc_pk] [block_height] [pub_rand] [proof] [block_app_hash] [finality_sig]",
		Args:  cobra.ExactArgs(6),
		Short: "Add a finality signature",
		Long: strings.TrimSpace(
			`Add a finality signature.`, // TODO: example
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// get finality provider BTC PK
			fpBTCPK, err := bbn.NewBIP340PubKeyFromHex(args[0])
			if err != nil {
				return err
			}

			// get block height
			blockHeight, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			// get public randomness
			pubRand, err := bbn.NewSchnorrPubRandFromHex(args[2])
			if err != nil {
				return err
			}

			// get proof
			proofBytes, err := hex.DecodeString(args[3])
			if err != nil {
				return err
			}
			var proof cmtcrypto.Proof
			if err := clientCtx.Codec.Unmarshal(proofBytes, &proof); err != nil {
				return err
			}

			// get block app hash
			appHash, err := hex.DecodeString(args[4])
			if err != nil {
				return err
			}

			// get finality signature
			finalitySig, err := bbn.NewSchnorrEOTSSigFromHex(args[5])
			if err != nil {
				return err
			}

			msg := types.MsgAddFinalitySig{
				Signer:       clientCtx.FromAddress.String(),
				FpBtcPk:      fpBTCPK,
				BlockHeight:  blockHeight,
				PubRand:      pubRand,
				Proof:        &proof,
				BlockAppHash: appHash,
				FinalitySig:  finalitySig,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
