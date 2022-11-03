package cmd_test

import (
	"github.com/babylonchain/babylon/app"
	"github.com/babylonchain/babylon/cmd/babylond/cmd"
	"github.com/cosmos/cosmos-sdk/client"
	types2 "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/types"
	"testing"
)

var misMatchGenesis = `
{
  "chain_id": "chain-test",
  "app_state": {
    "checkpointing": {
      "params": {},
      "genesis_keys": [
        {
          "validator_address": "bbnvaloper18unlvcpj9kaa5y27ghgjtmmkcsm4gk075cz2sv",
          "bls_key": {
            "pubkey": "swnOf6PuVF1YDXeShKCx1M3RpNsN+rTyzUoNm9O7UJVseYbmIbqZ3WlAhHA+1bCFBImr3+bjKu0S1RZY8bOhfbRpNBNOIiSKoyGPDqKj5+BSwmIFU4IgKOd10KvYfb/J",
            "pop": {
              "ed25519_sig": "VAjE+l9R2ilrZxTZrmpYvEO0IIi7Y8VQacltHeAtau8MoXnxeBUbLJPIuENUBPRVzObGPpU0QMmmzJkpexdWBw==",
              "bls_sig": "rF6wt1ZOVYM/xhWvh5RrIT3Lpwwtx6qRxuQh84fEInl2x5dNDSyrrA/60MIEMmm8"
            }
          },
          "val_pubkey": {
            "key": "PUoM/ErXICyPaiByrt7X/7/AgbP0URmtC7foTECOmoc="
          }
        }
      ]
    },
    "genutil": {
      "gen_txs": [
        {
          "body": {
            "messages": [
              {
                "@type": "/cosmos.staking.v1beta1.MsgCreateValidator",
                "description": {
                  "moniker": "node0",
                  "identity": "",
                  "website": "",
                  "security_contact": "",
                  "details": ""
                },
                "commission": {
                  "rate": "1.000000000000000000",
                  "max_rate": "1.000000000000000000",
                  "max_change_rate": "1.000000000000000000"
                },
                "min_self_delegation": "1",
                "delegator_address": "bbn1qck5qppfs7wkj20u94q60s8j5lsqy772rfktkm",
                "validator_address": "bbnvaloper1qck5qppfs7wkj20u94q60s8j5lsqy7727tuk66",
                "pubkey": {
                  "@type": "/cosmos.crypto.ed25519.PubKey",
                  "key": "ICl5MC/3coQYKSLGQqhIgU2Qr09fBv4tkYJ/j0d41As="
                },
                "value": {
                  "denom": "ubbn",
                  "amount": "100000000"
                }
              }
            ],
            "memo": "f9f9f5613f2010edbb6c6ed01633efadad8af269@192.168.10.2:26656",
            "timeout_height": "0",
            "extension_options": [],
            "non_critical_extension_options": []
          },
          "auth_info": {
            "signer_infos": [
              {
                "public_key": {
                  "@type": "/cosmos.crypto.secp256k1.PubKey",
                  "key": "AsOXpLQmKg88CzhYa4+c9LHKX0tlUlwyW1Lr0rf52KOp"
                },
                "mode_info": {
                  "single": {
                    "mode": "SIGN_MODE_DIRECT"
                  }
                },
                "sequence": "0"
              }
            ],
            "fee": {
              "amount": [],
              "gas_limit": "0",
              "payer": "",
              "granter": ""
            },
            "tip": null
          },
          "signatures": [
            "CvhDhApWLoK/Hl7PmAfXh8sG8ZOzZI4KGKvwWF/65yxTwJYFnfb43u8sa3hKkpKEIZWJpiem662yTdR6mKZWmQ=="
          ]
        }
      ]
    }
  }
}
`

var validGenesis = `
{
  "chain_id": "chain-test",
  "app_state": {
    "checkpointing": {
      "params": {},
      "genesis_keys": [
        {
          "validator_address": "bbnvaloper1qck5qppfs7wkj20u94q60s8j5lsqy7727tuk66",
          "bls_key": {
            "pubkey": "qOS3pHu3OQWJAqjlFG18T+9OaQx/uY1cQ9OClmGUknL2CrO+VPpveRne7SKZojYFFuifNmpjN4bUGiRYmea7hdixpeIwFkArjxKcg264MqEcKM/UthduM+1o+lNjoxN5",
            "pop": {
              "ed25519_sig": "rgfes5KUA3B4lF5JjG6HRHIMb3kL+VJnMyIx4v08nBSjy+sqKvPqpxvNv6Wn+UfTXuWZ3yqRzKQyMWGsA6kPCQ==",
              "bls_sig": "l/BmZn7fvctenvPqq1MB0emwKtcUfgpjvQuy+gI/AvUR27TyZNhlKcWAq+GRz/n3"
            }
          },
          "val_pubkey": {
            "key": "ICl5MC/3coQYKSLGQqhIgU2Qr09fBv4tkYJ/j0d41As="
          }
        },
        {
          "validator_address": "bbnvaloper18unlvcpj9kaa5y27ghgjtmmkcsm4gk075cz2sv",
          "bls_key": {
            "pubkey": "swnOf6PuVF1YDXeShKCx1M3RpNsN+rTyzUoNm9O7UJVseYbmIbqZ3WlAhHA+1bCFBImr3+bjKu0S1RZY8bOhfbRpNBNOIiSKoyGPDqKj5+BSwmIFU4IgKOd10KvYfb/J",
            "pop": {
              "ed25519_sig": "VAjE+l9R2ilrZxTZrmpYvEO0IIi7Y8VQacltHeAtau8MoXnxeBUbLJPIuENUBPRVzObGPpU0QMmmzJkpexdWBw==",
              "bls_sig": "rF6wt1ZOVYM/xhWvh5RrIT3Lpwwtx6qRxuQh84fEInl2x5dNDSyrrA/60MIEMmm8"
            }
          },
          "val_pubkey": {
            "key": "PUoM/ErXICyPaiByrt7X/7/AgbP0URmtC7foTECOmoc="
          }
        }
      ]
    },
    "genutil": {
      "gen_txs": [
        {
          "body": {
            "messages": [
              {
                "@type": "/cosmos.staking.v1beta1.MsgCreateValidator",
                "description": {
                  "moniker": "node0",
                  "identity": "",
                  "website": "",
                  "security_contact": "",
                  "details": ""
                },
                "commission": {
                  "rate": "1.000000000000000000",
                  "max_rate": "1.000000000000000000",
                  "max_change_rate": "1.000000000000000000"
                },
                "min_self_delegation": "1",
                "delegator_address": "bbn1qck5qppfs7wkj20u94q60s8j5lsqy772rfktkm",
                "validator_address": "bbnvaloper1qck5qppfs7wkj20u94q60s8j5lsqy7727tuk66",
                "pubkey": {
                  "@type": "/cosmos.crypto.ed25519.PubKey",
                  "key": "ICl5MC/3coQYKSLGQqhIgU2Qr09fBv4tkYJ/j0d41As="
                },
                "value": {
                  "denom": "ubbn",
                  "amount": "100000000"
                }
              }
            ],
            "memo": "f9f9f5613f2010edbb6c6ed01633efadad8af269@192.168.10.2:26656",
            "timeout_height": "0",
            "extension_options": [],
            "non_critical_extension_options": []
          },
          "auth_info": {
            "signer_infos": [
              {
                "public_key": {
                  "@type": "/cosmos.crypto.secp256k1.PubKey",
                  "key": "AsOXpLQmKg88CzhYa4+c9LHKX0tlUlwyW1Lr0rf52KOp"
                },
                "mode_info": {
                  "single": {
                    "mode": "SIGN_MODE_DIRECT"
                  }
                },
                "sequence": "0"
              }
            ],
            "fee": {
              "amount": [],
              "gas_limit": "0",
              "payer": "",
              "granter": ""
            },
            "tip": null
          },
          "signatures": [
            "CvhDhApWLoK/Hl7PmAfXh8sG8ZOzZI4KGKvwWF/65yxTwJYFnfb43u8sa3hKkpKEIZWJpiem662yTdR6mKZWmQ=="
          ]
        },
        {
          "body": {
            "messages": [
              {
                "@type": "/cosmos.staking.v1beta1.MsgCreateValidator",
                "description": {
                  "moniker": "node1",
                  "identity": "",
                  "website": "",
                  "security_contact": "",
                  "details": ""
                },
                "commission": {
                  "rate": "1.000000000000000000",
                  "max_rate": "1.000000000000000000",
                  "max_change_rate": "1.000000000000000000"
                },
                "min_self_delegation": "1",
                "delegator_address": "bbn18unlvcpj9kaa5y27ghgjtmmkcsm4gk07f6ghud",
                "validator_address": "bbnvaloper18unlvcpj9kaa5y27ghgjtmmkcsm4gk075cz2sv",
                "pubkey": {
                  "@type": "/cosmos.crypto.ed25519.PubKey",
                  "key": "PUoM/ErXICyPaiByrt7X/7/AgbP0URmtC7foTECOmoc="
                },
                "value": {
                  "denom": "ubbn",
                  "amount": "100000000"
                }
              }
            ],
            "memo": "b06768d9d68a3d5a0631b0540fb901559ab89964@192.168.10.3:26656",
            "timeout_height": "0",
            "extension_options": [],
            "non_critical_extension_options": []
          },
          "auth_info": {
            "signer_infos": [
              {
                "public_key": {
                  "@type": "/cosmos.crypto.secp256k1.PubKey",
                  "key": "AsgcC0fVVoNoQ70AJgw/7N3exxGUHVAJ+97EMPpkL+nP"
                },
                "mode_info": {
                  "single": {
                    "mode": "SIGN_MODE_DIRECT"
                  }
                },
                "sequence": "0"
              }
            ],
            "fee": {
              "amount": [],
              "gas_limit": "0",
              "payer": "",
              "granter": ""
            },
            "tip": null
          },
          "signatures": [
            "SAgaaPAXNofY0wbbc9CQDcAW4HPU827i/ufD8MJup0x/aYUSMLhOr600EPtTtvoLk5sUp9o3kDDMvscb/nYdFQ=="
          ]
        }
      ]
    }
  }
}
`

func TestCheckCorrespondence(t *testing.T) {

	encodingCft := app.MakeTestEncodingConfig()
	clientCtx := client.Context{}.WithCodec(encodingCft.Marshaler).WithTxConfig(encodingCft.TxConfig)

	testCases := []struct {
		name    string
		genesis string
		expErr  bool
	}{
		{
			"valid genesis gentx and BLS key pair",
			validGenesis,
			false,
		},
		{
			"mismatched genesis state",
			misMatchGenesis,
			true,
		},
	}

	for _, tc := range testCases {
		genDoc, err := types.GenesisDocFromJSON([]byte(tc.genesis))
		require.NoError(t, err)
		require.NotEmpty(t, genDoc)
		genesisState, err := types2.GenesisStateFromGenDoc(*genDoc)
		require.NoError(t, err)
		require.NotEmpty(t, genesisState)
		err = cmd.CheckCorrespondence(clientCtx, genesisState)
		if tc.expErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}
