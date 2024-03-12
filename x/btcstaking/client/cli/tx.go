package cli

import (
	"encoding/hex"
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/spf13/cobra"

	asig "github.com/babylonchain/babylon/crypto/schnorr-adaptor-signature"
	bbn "github.com/babylonchain/babylon/types"
	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	"github.com/babylonchain/babylon/x/btcstaking/types"
)

const (
	FlagMoniker         = "moniker"
	FlagIdentity        = "identity"
	FlagWebsite         = "website"
	FlagSecurityContact = "security-contact"
	FlagDetails         = "details"
	FlagCommissionRate  = "commission-rate"
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
		NewCreateFinalityProvicerCmd(),
		NewEditFinalityProvicerCmd(),
		NewCreateBTCDelegationCmd(),
		NewAddCovenantSigsCmd(),
		NewBTCUndelegateCmd(),
		NewSelectiveSlashingEvidenceCmd(),
	)

	return cmd
}

func NewCreateFinalityProvicerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-finality-provider [babylon_pk] [btc_pk] [pop]",
		Args:  cobra.ExactArgs(3),
		Short: "Create a finality provider",
		Long: strings.TrimSpace(
			`Create a finality provider.`, // TODO: example
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			fs := cmd.Flags()

			// get description
			moniker, _ := fs.GetString(FlagMoniker)
			identity, _ := fs.GetString(FlagIdentity)
			website, _ := fs.GetString(FlagWebsite)
			security, _ := fs.GetString(FlagSecurityContact)
			details, _ := fs.GetString(FlagDetails)
			description := stakingtypes.NewDescription(
				moniker,
				identity,
				website,
				security,
				details,
			)
			// get commission
			rateStr, _ := fs.GetString(FlagCommissionRate)
			rate, err := sdkmath.LegacyNewDecFromStr(rateStr)
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

			msg := types.MsgCreateFinalityProvider{
				Signer:      clientCtx.FromAddress.String(),
				Description: &description,
				Commission:  &rate,
				BabylonPk:   &babylonPK,
				BtcPk:       btcPK,
				Pop:         pop,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	fs := cmd.Flags()
	fs.String(FlagMoniker, "", "The finality provider's (optional) moniker")
	fs.String(FlagWebsite, "", "The finality provider's (optional) website")
	fs.String(FlagSecurityContact, "", "The finality provider's (optional) security contact email")
	fs.String(FlagDetails, "", "The finality provider's (optional) details")
	fs.String(FlagIdentity, "", "The (optional) identity signature (ex. UPort or Keybase)")
	fs.String(FlagCommissionRate, "0", "The initial commission rate percentage")

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewEditFinalityProvicerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit-finality-provider [btc_pk]",
		Args:  cobra.ExactArgs(1),
		Short: "Edit an existing finality provider",
		Long: strings.TrimSpace(
			`Edit an existing finality provider.`, // TODO: example
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			fs := cmd.Flags()

			// get description
			moniker, _ := fs.GetString(FlagMoniker)
			identity, _ := fs.GetString(FlagIdentity)
			website, _ := fs.GetString(FlagWebsite)
			security, _ := fs.GetString(FlagSecurityContact)
			details, _ := fs.GetString(FlagDetails)
			description := stakingtypes.NewDescription(
				moniker,
				identity,
				website,
				security,
				details,
			)
			// get commission
			rateStr, _ := fs.GetString(FlagCommissionRate)
			rate, err := sdkmath.LegacyNewDecFromStr(rateStr)
			if err != nil {
				return err
			}

			// get BTC PK
			btcPK, err := hex.DecodeString(args[1])
			if err != nil {
				return err
			}

			msg := types.MsgEditFinalityProvider{
				Signer:      clientCtx.FromAddress.String(),
				BtcPk:       btcPK,
				Description: &description,
				Commission:  &rate,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	fs := cmd.Flags()
	fs.String(FlagMoniker, "", "The finality provider's (optional) moniker")
	fs.String(FlagWebsite, "", "The finality provider's (optional) website")
	fs.String(FlagSecurityContact, "", "The finality provider's (optional) security contact email")
	fs.String(FlagDetails, "", "The finality provider's (optional) details")
	fs.String(FlagIdentity, "", "The (optional) identity signature (ex. UPort or Keybase)")
	fs.String(FlagCommissionRate, "0", "The initial commission rate percentage")

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewCreateBTCDelegationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-btc-delegation [babylon_pk] [btc_pk] [pop] [staking_tx_info] [fp_pk] [staking_time] [staking_value] [slashing_tx] [delegator_slashing_sig] [unbonding_tx] [unbonding_slashing_tx] [unbonding_time] [unbonding_value] [delegator_unbonding_slashing_sig]",
		Args:  cobra.ExactArgs(14),
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

			// staker pk
			btcPK, err := bbn.NewBIP340PubKeyFromHex(args[1])

			if err != nil {
				return err
			}

			// get PoP
			pop, err := types.NewPoPFromHex(args[2])
			if err != nil {
				return err
			}

			// get staking tx info
			stakingTxInfo, err := btcctypes.NewTransactionInfoFromHex(args[3])
			if err != nil {
				return err
			}

			// TODO: Support multiple finality providers
			// get finality provider PK
			fpPK, err := bbn.NewBIP340PubKeyFromHex(args[4])
			if err != nil {
				return err
			}

			// get staking time
			stakingTime, err := parseLockTime(args[5])
			if err != nil {
				return err
			}

			stakingValue, err := parseBtcAmount(args[6])
			if err != nil {
				return err
			}

			// get slashing tx
			slashingTx, err := types.NewBTCSlashingTxFromHex(args[7])
			if err != nil {
				return err
			}

			// get delegator sig on slashing tx
			delegatorSlashingSig, err := bbn.NewBIP340SignatureFromHex(args[8])
			if err != nil {
				return err
			}

			// get unbonding tx
			_, unbondingTxBytes, err := bbn.NewBTCTxFromHex(args[9])
			if err != nil {
				return err
			}

			// get unbonding slashing tx
			unbondingSlashingTx, err := types.NewBTCSlashingTxFromHex(args[10])
			if err != nil {
				return err
			}

			// get staking time
			unbondingTime, err := parseLockTime(args[11])
			if err != nil {
				return err
			}

			unbondingValue, err := parseBtcAmount(args[12])
			if err != nil {
				return err
			}

			// get delegator sig on unbonding slashing tx
			delegatorUnbondingSlashingSig, err := bbn.NewBIP340SignatureFromHex(args[13])
			if err != nil {
				return err
			}

			msg := types.MsgCreateBTCDelegation{
				Signer:                        clientCtx.FromAddress.String(),
				BabylonPk:                     &babylonPK,
				BtcPk:                         btcPK,
				FpBtcPkList:                   []bbn.BIP340PubKey{*fpPK},
				Pop:                           pop,
				StakingTime:                   uint32(stakingTime),
				StakingValue:                  int64(stakingValue),
				StakingTx:                     stakingTxInfo,
				SlashingTx:                    slashingTx,
				DelegatorSlashingSig:          delegatorSlashingSig,
				UnbondingTx:                   unbondingTxBytes,
				UnbondingTime:                 uint32(unbondingTime),
				UnbondingValue:                int64(unbondingValue),
				UnbondingSlashingTx:           unbondingSlashingTx,
				DelegatorUnbondingSlashingSig: delegatorUnbondingSlashingSig,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewAddCovenantSigsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-covenant-sigs [covenant_pk] [staking_tx_hash] [slashing_tx_sig1],[slashing_tx_sig2],... [unbonding_tx_sig] [slashing_unbonding_tx_sig1],[slashing_unbonding_tx_sig2],...",
		Args:  cobra.ExactArgs(5),
		Short: "Add a covenant signature",
		Long: strings.TrimSpace(
			`Add a covenant signature.`, // TODO: example
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			covPK, err := bbn.NewBIP340PubKeyFromHex(args[0])
			if err != nil {
				return fmt.Errorf("invalid public key: %w", err)
			}

			// get staking tx hash
			stakingTxHash := args[1]

			// parse slashing tx sigs
			slashingTxSigs := [][]byte{}
			for _, sigHex := range strings.Split(args[2], ",") {
				sig, err := asig.NewAdaptorSignatureFromHex(sigHex)
				if err != nil {
					return fmt.Errorf("invalid covenant signature: %w", err)
				}
				slashingTxSigs = append(slashingTxSigs, sig.MustMarshal())
			}

			// get covenant signature for unbonding tx
			unbondingTxSig, err := bbn.NewBIP340SignatureFromHex(args[3])
			if err != nil {
				return err
			}

			// parse unbonding slashing tx sigs
			unbondingSlashingSigs := [][]byte{}
			for _, sigHex := range strings.Split(args[4], ",") {
				slashingSig, err := asig.NewAdaptorSignatureFromHex(sigHex)
				if err != nil {
					return fmt.Errorf("invalid covenant signature: %w", err)
				}
				unbondingSlashingSigs = append(unbondingSlashingSigs, slashingSig.MustMarshal())
			}

			msg := types.MsgAddCovenantSigs{
				Signer:                  clientCtx.FromAddress.String(),
				Pk:                      covPK,
				StakingTxHash:           stakingTxHash,
				SlashingTxSigs:          slashingTxSigs,
				UnbondingTxSig:          unbondingTxSig,
				SlashingUnbondingTxSigs: unbondingSlashingSigs,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewBTCUndelegateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "btc-undelegate [staking_tx_hash] [unbonding_tx_sig]",
		Args:  cobra.ExactArgs(2),
		Short: "Add a signature on the unbonding tx of a BTC delegation identified by a given staking tx hash. ",
		Long: strings.TrimSpace(
			`Add a signature on the unbonding tx of a BTC delegation identified by a given staking tx hash signed by the delegator. The signature proves that delegator wants to unbond, and Babylon will consider the BTC delegation unbonded.`, // TODO: example
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// get staking tx hash
			stakingTxHash := args[0]

			// get delegator signature for unbonding tx
			unbondingTxSig, err := bbn.NewBIP340SignatureFromHex(args[1])
			if err != nil {
				return err
			}

			msg := types.MsgBTCUndelegate{
				Signer:         clientCtx.FromAddress.String(),
				StakingTxHash:  stakingTxHash,
				UnbondingTxSig: unbondingTxSig,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewSelectiveSlashingEvidenceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "selective-slashing-evidence [staking_tx_hash] [recovered_fp_btc_sk]",
		Args:  cobra.ExactArgs(2),
		Short: "Add the recovered BTC SK of a finality provider that launched selective slashing offence.",
		Long: strings.TrimSpace(
			`Add the recovered BTC SK of a finality provider that launched selective slashing offence. The SK is recovered from a pair of Schnorr/adaptor signatures`, // TODO: example
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// get staking tx hash
			stakingTxHash := args[0]

			// get delegator signature for unbonding tx
			fpSKBytes, err := hex.DecodeString(args[1])
			if err != nil {
				return err
			}

			msg := types.MsgSelectiveSlashingEvidence{
				Signer:           clientCtx.FromAddress.String(),
				StakingTxHash:    stakingTxHash,
				RecoveredFpBtcSk: fpSKBytes,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
