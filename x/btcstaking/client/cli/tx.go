package cli

import (
	"encoding/hex"
	"fmt"
	"strings"

	bbn "github.com/babylonchain/babylon/types"
	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	secp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
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
		NewCreateBTCValidatorCmd(),
		NewCreateBTCDelegationCmd(),
		NewAddJurySigCmd(),
	)

	return cmd
}

func NewCreateBTCValidatorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-btc-validator [babylon_pk] [btc_pk] [pop]",
		Args:  cobra.ExactArgs(3),
		Short: "Create a BTC validator",
		Long: strings.TrimSpace(
			`Create a BTC validator.`, // TODO: example
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// get Babylon PK
			babylonPKBytes, err := hex.DecodeString(args[0])
			if err != nil {
				return err
			}
			var babylonPK secp256k1.PubKey
			if err := babylonPK.Unmarshal(babylonPKBytes); err != nil {
				return err
			}

			// get BTC PK
			btcPK, err := bbn.NewBIP340PubKeyFromHex(args[1])
			if err != nil {
				return err
			}

			// get PoP
			pop, err := types.NewPoPFromHex(args[2])
			if err != nil {
				return err
			}

			msg := types.MsgCreateBTCValidator{
				Signer:    clientCtx.FromAddress.String(),
				BabylonPk: &babylonPK,
				BtcPk:     btcPK,
				Pop:       pop,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewCreateBTCDelegationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-btc-delegation [babylon_pk] [pop] [staking_tx] [staking_tx_info] [slashing_tx] [delegator_sig]",
		Args:  cobra.ExactArgs(6),
		Short: "Create a BTC delegation",
		Long: strings.TrimSpace(
			`Create a BTC delegation.`, // TODO: example
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// get Babylon PK
			babylonPKBytes, err := hex.DecodeString(args[0])
			if err != nil {
				return err
			}
			var babylonPK secp256k1.PubKey
			if err := babylonPK.Unmarshal(babylonPKBytes); err != nil {
				return err
			}

			// get PoP
			pop, err := types.NewPoPFromHex(args[1])
			if err != nil {
				return err
			}

			// get staking tx
			stakingTx, err := types.NewStakingTxFromHex(args[2])
			if err != nil {
				return err
			}

			// get staking tx info
			stakingTxInfo, err := btcctypes.NewTransactionInfoFromHex(args[3])
			if err != nil {
				return err
			}

			// get slashing tx
			slashingTx, err := types.NewBTCSlashingTxFromHex(args[4])
			if err != nil {
				return err
			}

			// get delegator sig
			delegatorSig, err := bbn.NewBIP340SignatureFromHex(args[5])
			if err != nil {
				return err
			}

			msg := types.MsgCreateBTCDelegation{
				Signer:        clientCtx.FromAddress.String(),
				BabylonPk:     &babylonPK,
				Pop:           pop,
				StakingTx:     stakingTx,
				StakingTxInfo: stakingTxInfo,
				SlashingTx:    slashingTx,
				DelegatorSig:  delegatorSig,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewAddJurySigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-jury-sig [val_pk] [del_pk] [staking_tx_hash] [sig]",
		Args:  cobra.ExactArgs(4),
		Short: "Add a jury signature",
		Long: strings.TrimSpace(
			`Add a jury signature.`, // TODO: example
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// get validator PK
			valPK, err := bbn.NewBIP340PubKeyFromHex(args[0])
			if err != nil {
				return err
			}

			// get delegator PK
			delPK, err := bbn.NewBIP340PubKeyFromHex(args[1])
			if err != nil {
				return err
			}

			// get staking tx hash
			stakingTxHash := args[2]

			// get jury sigature
			sig, err := bbn.NewBIP340SignatureFromHex(args[3])
			if err != nil {
				return err
			}

			msg := types.MsgAddJurySig{
				Signer:        clientCtx.FromAddress.String(),
				ValPk:         valPK,
				DelPk:         delPK,
				StakingTxHash: stakingTxHash,
				Sig:           sig,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
