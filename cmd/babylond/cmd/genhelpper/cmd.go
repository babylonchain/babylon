package genhelpper

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	"github.com/babylonchain/babylon/privval"
	checkpointingtypes "github.com/babylonchain/babylon/x/checkpointing/types"
)

// GenesisHelpers helpers for manipulating the genesis file.
func GenesisHelpers() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "gen-helper",
		Short:                      "Usefull commands for creating the genesis state",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	cmd.AddCommand(genHelperAddCheckpointGenesisKey())

	return cmd
}

func genHelperAddCheckpointGenesisKey() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ckpt-gen-key [val-addr]",
		Example: "ckpt-gen-key bbnvaloper1y6xz2ggfc0pcsmyjlekh0j9pxh6hk87yrjr7tn",
		Short:   "add the genesis key pop to checkpoint genstate",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			valAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return fmt.Errorf("argument val-addr invalid %w", err)
			}

			config := server.GetServerContextFromCmd(cmd).Config
			pvKeyPath := config.PrivValidatorKeyFile()
			pvStatePath := config.PrivValidatorStateFile()

			filePV := privval.LoadWrappedFilePV(pvKeyPath, pvStatePath)
			valPrivkey := filePV.GetValPrivKey()
			blsPrivkey := filePV.GetBlsPrivKey()
			valKeysWithPoP, err := privval.NewValidatorKeys(valPrivkey, blsPrivkey)
			if err != nil {
				return err
			}

			valPubkey, err := cryptocodec.FromCmtPubKeyInterface(valKeysWithPoP.ValPubkey)
			if err != nil {
				return err
			}

			genKey := &checkpointingtypes.GenesisKey{
				ValidatorAddress: valAddr.String(),
				BlsKey: &checkpointingtypes.BlsKey{
					Pubkey: &valKeysWithPoP.BlsPubkey,
					Pop:    valKeysWithPoP.PoP,
				},
				ValPubkey: valPubkey.(*ed25519.PubKey),
			}

			genFile := config.GenesisFile()
			genesisState, genesis, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %s", err)
			}

			ckpGenStateRaw, ok := genesisState[checkpointingtypes.ModuleName]
			if !ok {
				return fmt.Errorf("genesis state does not have %s module", checkpointingtypes.ModuleName)
			}

			var ckpGenState checkpointingtypes.GenesisState
			if err := json.Unmarshal(ckpGenStateRaw, &ckpGenState); err != nil {
				return fmt.Errorf("failed genesis state unmarshal %s module", checkpointingtypes.ModuleName)
			}
			// Updates the genesis state by adding new key to it
			ckpGenState.GenesisKeys = append(ckpGenState.GenesisKeys, genKey)

			ckpGenStateRaw, err = json.Marshal(ckpGenState)
			if err != nil {
				return fmt.Errorf("failed to marshal %s module into genesis state", checkpointingtypes.ModuleName)
			}
			// updates it back to the genesis state
			genesisState[checkpointingtypes.ModuleName] = ckpGenStateRaw

			genBz, err := json.Marshal(genesisState)
			if err != nil {
				return fmt.Errorf("failed to Marshal genesisState")
			}

			genesis.AppState = genBz
			if err := genutil.ExportGenesisFile(genesis, genFile); err != nil {
				return err
			}
			return nil
		},
	}

	return cmd
}
