package types_test

import (
	sdkmath "cosmossdk.io/math"
	"testing"

	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/stretchr/testify/require"
)

func TestGenesisState_Validate(t *testing.T) {
	tests := []struct {
		desc     string
		genState *types.GenesisState
		valid    bool
	}{
		{
			desc:     "default is valid",
			genState: types.DefaultGenesis(),
			valid:    true,
		},
		{
			desc: "valid genesis state",
			genState: &types.GenesisState{
				Params: types.Params{
					JuryPk:              types.DefaultParams().JuryPk,
					SlashingAddress:     types.DefaultParams().SlashingAddress,
					MinSlashingTxFeeSat: 500,
					MinCommissionRate:   sdkmath.LegacyMustNewDecFromStr("0.5"),
				},
			},
			valid: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.genState.Validate()
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
