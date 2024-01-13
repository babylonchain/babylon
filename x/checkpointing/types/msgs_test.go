package types_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	appparams "github.com/babylonchain/babylon/app/params"
	"github.com/babylonchain/babylon/crypto/bls12381"
	"github.com/babylonchain/babylon/privval"
	"github.com/babylonchain/babylon/x/checkpointing/types"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
)

var (
	pk1      = ed25519.GenPrivKey().PubKey()
	valAddr1 = sdk.ValAddress(pk1.Address())
)

func TestMsgDecode(t *testing.T) {
	registry := codectypes.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(registry)
	types.RegisterInterfaces(registry)
	stakingtypes.RegisterInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)

	// build MsgWrappedCreateValidator
	msg, err := buildMsgWrappedCreateValidatorWithAmount(
		sdk.AccAddress(valAddr1),
		sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction),
	)
	require.NoError(t, err)

	// marshal
	msgBytes, err := cdc.MarshalInterface(msg)
	require.NoError(t, err)

	// unmarshal to sdk.Msg interface
	var msg2 sdk.Msg
	err = cdc.UnmarshalInterface(msgBytes, &msg2)
	require.NoError(t, err)

	// type assertion
	msgWithType, ok := msg2.(*types.MsgWrappedCreateValidator)
	require.True(t, ok)

	// ensure msgWithType.MsgCreateValidator.Pubkey with type Any is unmarshaled successfully
	require.NotNil(t, msgWithType.MsgCreateValidator.Pubkey.GetCachedValue())
}

func buildMsgWrappedCreateValidatorWithAmount(addr sdk.AccAddress, bondTokens sdkmath.Int) (*types.MsgWrappedCreateValidator, error) {
	tmValPrivkey := ed25519.GenPrivKey()
	bondCoin := sdk.NewCoin(appparams.DefaultBondDenom, bondTokens)
	description := stakingtypes.NewDescription("foo_moniker", "", "", "", "")
	commission := stakingtypes.NewCommissionRates(sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec())

	pk, err := cryptocodec.FromCmtPubKeyInterface(tmValPrivkey.PubKey())
	if err != nil {
		return nil, err
	}

	createValidatorMsg, err := stakingtypes.NewMsgCreateValidator(
		sdk.ValAddress(addr).String(), pk, bondCoin, description, commission, sdkmath.OneInt(),
	)
	if err != nil {
		return nil, err
	}
	blsPrivKey := bls12381.GenPrivKey()
	pop, err := privval.BuildPoP(tmValPrivkey, blsPrivKey)
	if err != nil {
		return nil, err
	}
	blsPubKey := blsPrivKey.PubKey()

	return types.NewMsgWrappedCreateValidator(createValidatorMsg, &blsPubKey, pop)
}
