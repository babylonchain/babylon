package genhelpers

import (
	"fmt"

	btcstktypes "github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	"github.com/spf13/cobra"
)

// CmdSetBtcDels CLI sets bitcoin delegations into the genesis state.
func CmdSetBtcDels() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-btc-delegations [path/to/btc_delegations.json]",
		Short: "Set the BTC delegations from the given json file into the genesis.json",
		Long: `Reads BTC delegation structures from the given json file and update the genesis.json file
in place to include the delegations in the btcstaking module's genesis state.
Duplicated BTC delegations are not allowed and it will prompt an error.
`,
		Example: `babylond gen-helpers set-btc-delegations path/to/btc_delegations.json
Possible content of 'btc_delegations.json' is
{
	"btc_delegations": [
		{
			"babylon_pk": {
				"key": "AkCmkLYWjr0/49MCK3lzTqKC1MiPSldXgoqP0EdtV8m1"
			},
			"btc_pk": "5ea8d7dfe920f5aecb1f4411b1b3e6684031ecb7a282df65107a615adc562a48",
			"pop": {
				"btc_sig_type": "BIP340",
				"babylon_sig": "gjqXTQdgzC6AX/lnj3T3QV5vV+jtKDKwJOP9jMpZdb90MoIKRGV7j/A6lwbgsnUzbsvd4L19U7/3cOZqYBfQAg==",
				"btc_sig": "Gx0VxuXgjiLvdCIIx1eN9JE9mqo+qvIo48cUhTCbM6uTxt59HTJiVGloHE518OXk+vQrNcHB+bZFqX+T5wMS9w=="
			},
			"fp_btc_pk_list": [
				"3d6a1db7bffaebed71831fae74d32bf78e97884c41634cc1a80b0ebbffe1c7dd"
			],
			"start_height": "127",
			"end_height": "10127",
			"total_sat": "1000000",
			"staking_tx": "AQAAAAF/cxe2A5/IsfYkwiVTA795y/X16NDBpqETC3kiv27AKgAAAABrSDBFAiEA8inc8+grbO6iJbl1VwJjR4Y89dK7ATcEz3XnJPNd37MCIGd0wpfu3Hz4+hjVXatnKCTViUG2SPiD83k/+yhy9bIIASEDXqjX3+kg9a7LH0QRsbPmaEAx7Leigt9lEHphWtxWKkj/////AkBCDwAAAAAAIlEg8eeUnPklLO9pkkhYMkoQFW7gxAmd4W781NwChbc9WSe0mPYpAQAAABl2qRS9MnviWBrF4wvpgMpBhIFGQHnOWYisAAAAAA==",
			"staking_output_idx": 0,
			"slashing_tx": "AQAAAAES8FBlJfWXZSLzImp6AjfD1bosph2q02QhTIjDo9T8bQAAAAAA/////wKghgEAAAAAABl2qRQBAQEBAQEBAQEBAQEBAQEBAQEBAYisuLcNAAAAAAAiUSD1hd0Q7BOqlMAZY6fqBidPwKe7jLRXtdGcpgQjVIhr8wAAAAA=",
			"delegator_sig": "2i6kTpmzn0pJAbY+v5Ktf0yo6YNhucVZC4clg4YVVZKU7dVh9DKEzrichq5C1eFgaRr+YYtEyxtTlVwhyOg2Ag==",
			"covenant_sigs": [
				{
					"cov_pk": "e76137fda8dff755a0879264c4eb3c3574ab77b6c5589f47b090203e4938d065",
					"adaptor_sigs": [
						"Ann3xe0L1gIHWS2+hhHwRDTPz5RoCFzBBte0xkhJ0VRUWDB2a46Pl1miZ9y+jyseZNYcps+DeE1cB/S54oeVQYEA"
					]
				}
			],
			"unbonding_time": 101,
			"btc_undelegation": {
				"unbonding_tx": "AgAAAAES8FBlJfWXZSLzImp6AjfD1bosph2q02QhTIjDo9T8bQAAAAAA/////wGsMA8AAAAAACJRIEObJoQoxADSVZ3K8dV0Req4coALcrefxibIBmWItvVDAAAAAA==",
				"slashing_tx": "AQAAAAGGZtqTIrJN8TDOMw+bbENzXLECv1Z1E4idjZ2UiHnsbAAAAAAA/////wLehAEAAAAAABl2qRQBAQEBAQEBAQEBAQEBAQEBAQEBAYis5qcNAAAAAAAiUSD1hd0Q7BOqlMAZY6fqBidPwKe7jLRXtdGcpgQjVIhr8wAAAAA=",
				"delegator_unbonding_sig": null,
				"delegator_slashing_sig": "P6GSaJdbIhEED+k74T1sRYzTYHUhzg8/NDGvwR+sBHXhKsT8O82XRaHyvjFXAVCLK0sa5oxSOTndMmY3SKHNsw==",
				"covenant_slashing_sigs": [
					{
						"cov_pk": "e76137fda8dff755a0879264c4eb3c3574ab77b6c5589f47b090203e4938d065",
						"adaptor_sigs": [
							"AjZrivh6++7Wu7OWFRxLAv2SHPvf0VhZM7pZK7oWP3KoWKdPpBcNotFqjdCQE2TTSshS4qpEuwYMLtMMzYqIOKYA"
						]
					}
				],
				"covenant_unbonding_sig_list": [
					{
						"pk": "e76137fda8dff755a0879264c4eb3c3574ab77b6c5589f47b090203e4938d065",
						"sig": "t+zUA4qtNCKREviFUtwWh46QNEov033whhF2i89oF2mEoa55k7Hq68TAY+jiGeeTJU1BM7Q4Qfu8XFuv3c1WAw=="
					}
				]
			},
			"params_version": 0
		}
	],
}
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			config := server.GetServerContextFromCmd(cmd).Config
			config.SetRoot(clientCtx.HomeDir)

			inputBtcDels, err := getBtcStakingGenStateFromFile(clientCtx.Codec, args[0])
			if err != nil {
				return err
			}

			genFile := config.GenesisFile()
			appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}
			btcstkGenState := btcstktypes.GenesisStateFromAppState(clientCtx.Codec, appState)

			genStateBtcDelsByTxHash := make(map[string]struct{}, 0)
			for _, del := range btcstkGenState.BtcDelegations {
				hash, err := del.GetStakingTxHash()
				if err != nil {
					return err
				}

				key := hash.String()
				if _, ok := genStateBtcDelsByTxHash[key]; ok {
					return fmt.Errorf("bad genesis state, there is more than one btc delegator with the same staking tx hash %s", key)
				}
				genStateBtcDelsByTxHash[key] = struct{}{}
			}

			genStateFpsByBtcPk, err := mapFinalityProvidersByBtcPk(btcstkGenState.FinalityProviders)
			if err != nil {
				return fmt.Errorf("bad gen state: %w", err)
			}

			newDels := make([]*btcstktypes.BTCDelegation, 0, len(inputBtcDels.BtcDelegations))
			for _, del := range inputBtcDels.BtcDelegations {
				if err := del.ValidateBasic(); err != nil {
					return fmt.Errorf("failed to validate basic btc delegation: %w", err)
				}

				hash, err := del.GetStakingTxHash()
				if err != nil {
					return err
				}

				key := hash.String()
				if _, ok := genStateBtcDelsByTxHash[key]; ok {
					return fmt.Errorf("error: btc delegation: %+v\nwas already set on genesis, or contains the same staking tx hash %s than another btc delegation", del, key)
				}

				for _, fpBTCpk := range del.FpBtcPkList {
					fpKey := fpBTCpk.MarshalHex()
					if _, ok := genStateFpsByBtcPk[fpKey]; !ok {
						return fmt.Errorf("error: btc delegation: %+v\nhas an associated finality provider that is not set on genesis %s", del, fpKey)
					}
				}

				genStateBtcDelsByTxHash[key] = struct{}{}
				newDels = append(newDels, del)
			}
			btcstkGenState.BtcDelegations = append(btcstkGenState.BtcDelegations, newDels...)

			err = replaceModOnGenesis(clientCtx.Codec, genDoc, appState, btcstktypes.ModuleName, &btcstkGenState)
			if err != nil {
				return err
			}

			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	return cmd
}
