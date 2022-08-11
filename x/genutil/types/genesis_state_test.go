package types_test

import (
	"encoding/json"
	bbnapp "github.com/babylonchain/babylon/app"
	"github.com/babylonchain/babylon/crypto/bls12381"
	"github.com/babylonchain/babylon/privval"
	checkpointingtypes "github.com/babylonchain/babylon/x/checkpointing/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/babylonchain/babylon/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

var (
	accpriv1    = secp256k1.GenPrivKey()
	accpriv2    = secp256k1.GenPrivKey()
	accpk1      = accpriv1.PubKey()
	accpk2      = accpriv2.PubKey()
	valpriv1    = ed25519.GenPrivKey()
	valpriv2    = ed25519.GenPrivKey()
	valpk1      = valpriv1.PubKey()
	valpk2      = valpriv2.PubKey()
	blspriv1    = bls12381.GenPrivKey()
	blspriv2    = bls12381.GenPrivKey()
	blspk1      = blspriv1.PubKey()
	blspk2      = blspriv2.PubKey()
	valKeys1, _ = privval.NewValidatorKeys(valpriv1, blspriv1)
	valKeys2, _ = privval.NewValidatorKeys(valpriv2, blspriv2)
	addr1       = sdk.AccAddress(accpk1.Address())
	addr2       = sdk.AccAddress(accpk2.Address())
	desc        = stakingtypes.NewDescription("testname", "", "", "", "")
	comm        = stakingtypes.CommissionRates{}
)

func TestNetGenesisState(t *testing.T) {
	gen := types.NewGenesisState(nil)
	assert.NotNil(t, gen.GenTxs) // https://github.com/cosmos/cosmos-sdk/issues/5086

	gen = types.NewGenesisState(
		[]json.RawMessage{
			[]byte(`{"foo":"bar"}`),
		},
	)
	assert.Equal(t, string(gen.GenTxs[0]), `{"foo":"bar"}`)
}

func TestValidateGenesisMultipleMessages(t *testing.T) {
	var err error
	amount := sdk.NewInt64Coin(sdk.DefaultBondDenom, 50)
	one := sdk.OneInt()
	cosmosValpubkey1, err := cryptocodec.FromTmPubKeyInterface(valpk1)
	require.NoError(t, err)
	msgcreateval1, err := stakingtypes.NewMsgCreateValidator(
		sdk.ValAddress(addr1), cosmosValpubkey1, amount, desc, comm, one)
	require.NoError(t, err)
	cosmosValpubkey2, err := cryptocodec.FromTmPubKeyInterface(valpk1)
	require.NoError(t, err)
	msgcreateval2, err := stakingtypes.NewMsgCreateValidator(
		sdk.ValAddress(addr2), cosmosValpubkey2, amount, desc, comm, one)
	msg1 := checkpointingtypes.NewMsgWrappedCreateValidator(valKeys1.BlsPubkey, valKeys1.PoP, msgcreateval1)
	msg2 := checkpointingtypes.NewMsgWrappedCreateValidator(valKeys2.BlsPubkey, valKeys2.PoP, msgcreateval2)
	require.NoError(t, err)

	txGen := bbnapp.MakeTestEncodingConfig().TxConfig
	txBuilder := txGen.NewTxBuilder()
	require.NoError(t, txBuilder.SetMsgs(msg1, msg2))

	tx := txBuilder.GetTx()
	genesisState := types.NewGenesisStateFromTx(txGen.TxJSONEncoder(), []sdk.Tx{tx})

	err = types.ValidateGenesis(genesisState, bbnapp.MakeTestEncodingConfig().TxConfig.TxJSONDecoder())
	require.Error(t, err)
}

func TestValidateGenesisBadMessage(t *testing.T) {
	desc := stakingtypes.NewDescription("testname", "", "", "", "")

	msg1 := stakingtypes.NewMsgEditValidator(sdk.ValAddress(addr1), desc, nil, nil)

	txGen := bbnapp.MakeTestEncodingConfig().TxConfig
	txBuilder := txGen.NewTxBuilder()
	err := txBuilder.SetMsgs(msg1)
	require.NoError(t, err)

	tx := txBuilder.GetTx()
	genesisState := types.NewGenesisStateFromTx(txGen.TxJSONEncoder(), []sdk.Tx{tx})

	err = types.ValidateGenesis(genesisState, bbnapp.MakeTestEncodingConfig().TxConfig.TxJSONDecoder())
	require.Error(t, err)
}

func TestGenesisStateFromGenFile(t *testing.T) {
	cdc := codec.NewLegacyAmino()

	genFile := "../../../tests/fixtures/adr-024-coin-metadata_genesis.json"
	genesisState, _, err := types.GenesisStateFromGenFile(genFile)
	require.NoError(t, err)

	var bankGenesis banktypes.GenesisState
	cdc.MustUnmarshalJSON(genesisState[banktypes.ModuleName], &bankGenesis)

	require.True(t, bankGenesis.Params.DefaultSendEnabled)
	require.Equal(t, "1000nametoken,100000000stake", bankGenesis.Balances[0].GetCoins().String())
	//require.Equal(t, "bbl106vrzv5xkheqhjm023pxcxlqmcjvuhtfyachz4", bankGenesis.Balances[0].GetAddress().String())
	require.Equal(t, "The native staking token of the Cosmos Hub.", bankGenesis.DenomMetadata[0].GetDescription())
	require.Equal(t, "uatom", bankGenesis.DenomMetadata[0].GetBase())
	require.Equal(t, "matom", bankGenesis.DenomMetadata[0].GetDenomUnits()[1].GetDenom())
	require.Equal(t, []string{"milliatom"}, bankGenesis.DenomMetadata[0].GetDenomUnits()[1].GetAliases())
	require.Equal(t, uint32(3), bankGenesis.DenomMetadata[0].GetDenomUnits()[1].GetExponent())

}
