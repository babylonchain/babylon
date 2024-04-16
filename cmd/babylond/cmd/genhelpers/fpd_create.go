package genhelpers

import (
	"encoding/json"
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/babylonchain/babylon/app"
	btcstktypes "github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	stktypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

const (
	FlagMoniker   = "moniker"
	FlagWebsite   = "website"
	FlagIdentity  = "identity"
	FlagContact   = "contact"
	FlagDetails   = "details"
	FlagComission = "commision"
)

// CmdCreateFinalityProvider CLI command to create finality provider.
func CmdCreateFinalityProvider() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-fpd [key-name] [commission]",
		Short: "Create genesis Finality provider",
		Args:  cobra.ExactArgs(2),
		Long: strings.TrimSpace(`create-fpd creates the finality provider structure needed to start
the genesis with finality provider already active and available.

The pre-conditions of running the create-fpd are the existence of the keyring,
and the existence of btc pv key.


Example:
$ babylond create-fpd name-of-your-finality-provider 0.5 --home ./
`),

		RunE: func(cmd *cobra.Command, args []string) error {
			f := cmd.Flags()
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			keyName, strCommission := args[0], args[1]

			k, err := clientCtx.Keyring.Key(keyName)
			if err != nil {
				return err
			}

			bbnPK, err := k.GetPubKey()
			if err != nil {
				return err
			}

			secp256k1PK, ok := bbnPK.(*secp256k1.PubKey)
			if !ok {
				return fmt.Errorf("failed to assert bbnPK to *secp256k1.PubKey")
			}

			desc, err := loadDescription(f)
			if err != nil {
				return err
			}

			commission, err := sdkmath.LegacyNewDecFromStr(strCommission)
			if err != nil {
				return err
			}

			fp := btcstktypes.FinalityProvider{
				// BtcPk:           req.BtcPk,
				// Pop:             req.Pop,
				// MasterPubRand:   req.MasterPubRand,
				BabylonPk:            secp256k1PK,
				Description:          desc,
				Commission:           &commission,
				RegisteredEpoch:      0,
				SlashedBabylonHeight: 0,
				SlashedBtcHeight:     0,
			}

			out, err := json.Marshal(fp)
			if err != nil {
				return err
			}

			if _, err := fmt.Fprintln(cmd.OutOrStdout(), string(out)); err != nil {
				return err
			}
			return nil
		},
	}

	cmd.Flags().String(flags.FlagHome, app.DefaultNodeHome, "The node home directory")
	cmd.Flags().String(FlagMoniker, "", "The human-readable name of your finality provider")
	cmd.Flags().String(FlagIdentity, "", "Optional identity signature (ex. UPort or Keybase)")
	cmd.Flags().String(FlagWebsite, "", "Website to recognize finality provider")
	cmd.Flags().String(FlagContact, "", "Contact information (email/cellphone/twitter)")
	cmd.Flags().String(FlagDetails, "", "Any other detail information")

	return cmd
}

func loadDescription(f *pflag.FlagSet) (*stktypes.Description, error) {
	moniker, err := f.GetString(FlagMoniker)
	if err != nil {
		return nil, err
	}
	identity, err := f.GetString(FlagIdentity)
	if err != nil {
		return nil, err
	}
	website, err := f.GetString(FlagWebsite)
	if err != nil {
		return nil, err
	}
	contact, err := f.GetString(FlagContact)
	if err != nil {
		return nil, err
	}
	details, err := f.GetString(FlagDetails)
	if err != nil {
		return nil, err
	}

	d := stktypes.NewDescription(moniker, identity, website, contact, details)
	return &d, nil
}
