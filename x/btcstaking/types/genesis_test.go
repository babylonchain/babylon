package types_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"

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
				Params: []*types.Params{&types.Params{
					CovenantPks:                types.DefaultParams().CovenantPks,
					CovenantQuorum:             types.DefaultParams().CovenantQuorum,
					SlashingAddress:            types.DefaultParams().SlashingAddress,
					MinSlashingTxFeeSat:        500,
					MinCommissionRate:          sdkmath.LegacyMustNewDecFromStr("0.5"),
					SlashingRate:               sdkmath.LegacyMustNewDecFromStr("0.1"),
					MaxActiveFinalityProviders: 100,
					MinUnbondingRate:           sdkmath.LegacyMustNewDecFromStr("0.8"),
				}},
			},
			valid: true,
		},
		{
			desc: "invalid slashing rate in genesis",
			genState: &types.GenesisState{
				Params: []*types.Params{&types.Params{
					CovenantPks:                types.DefaultParams().CovenantPks,
					CovenantQuorum:             types.DefaultParams().CovenantQuorum,
					SlashingAddress:            types.DefaultParams().SlashingAddress,
					MinSlashingTxFeeSat:        500,
					MinCommissionRate:          sdkmath.LegacyMustNewDecFromStr("0.5"),
					SlashingRate:               sdkmath.LegacyZeroDec(), // invalid slashing rate
					MaxActiveFinalityProviders: 100,
					MinUnbondingRate:           sdkmath.LegacyMustNewDecFromStr("0.8"),
				},
				}},
			valid: false,
		},
		{
			desc: "invalid unbonding value in genesis",
			genState: &types.GenesisState{
				Params: []*types.Params{&types.Params{
					CovenantPks:                types.DefaultParams().CovenantPks,
					CovenantQuorum:             types.DefaultParams().CovenantQuorum,
					SlashingAddress:            types.DefaultParams().SlashingAddress,
					MinSlashingTxFeeSat:        500,
					MinCommissionRate:          sdkmath.LegacyMustNewDecFromStr("0.5"),
					SlashingRate:               sdkmath.LegacyMustNewDecFromStr("0.1"),
					MaxActiveFinalityProviders: 100,
					MinUnbondingRate:           sdkmath.LegacyZeroDec(),
				},
				}},
			valid: false,
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
