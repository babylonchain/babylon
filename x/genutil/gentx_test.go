package genutil_test

import (
	"encoding/json"
	"fmt"
	bbnapp "github.com/babylonchain/babylon/app"
	appparams "github.com/babylonchain/babylon/app/params"
	"github.com/babylonchain/babylon/crypto/bls12381"
	"github.com/babylonchain/babylon/privval"
	checkpointingtypes "github.com/babylonchain/babylon/x/checkpointing/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"testing"

	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/babylonchain/babylon/x/genutil"
	"github.com/babylonchain/babylon/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
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

// GenTxTestSuite is a test suite to be used with gentx tests.
type GenTxTestSuite struct {
	suite.Suite

	ctx            sdk.Context
	app            *bbnapp.BabylonApp
	encodingConfig appparams.EncodingConfig

	msg1, msg2 *checkpointingtypes.MsgWrappedCreateValidator
}

func (suite *GenTxTestSuite) SetupTest() {
	checkTx := false
	app := bbnapp.Setup(checkTx)
	suite.ctx = app.BaseApp.NewContext(checkTx, tmproto.Header{})
	suite.app = app
	suite.encodingConfig = bbnapp.MakeTestEncodingConfig()

	var err error
	amount := sdk.NewInt64Coin(sdk.DefaultBondDenom, 50)
	one := sdk.OneInt()
	cosmosValpubkey1, err := cryptocodec.FromTmPubKeyInterface(valpk1)
	suite.NoError(err)
	msgcreateval1, err := stakingtypes.NewMsgCreateValidator(
		sdk.ValAddress(addr1), cosmosValpubkey1, amount, desc, comm, one)
	suite.NoError(err)
	cosmosValpubkey2, err := cryptocodec.FromTmPubKeyInterface(valpk1)
	suite.NoError(err)
	msgcreateval2, err := stakingtypes.NewMsgCreateValidator(
		sdk.ValAddress(addr2), cosmosValpubkey2, amount, desc, comm, one)
	suite.msg1 = checkpointingtypes.NewMsgWrappedCreateValidator(valKeys1.BlsPubkey, valKeys1.PoP, msgcreateval1)
	suite.msg2 = checkpointingtypes.NewMsgWrappedCreateValidator(valKeys2.BlsPubkey, valKeys2.PoP, msgcreateval2)
	suite.NoError(err)
}

func (suite *GenTxTestSuite) setAccountBalance(addr sdk.AccAddress, amount int64) json.RawMessage {
	acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr)
	suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

	err := bbnapp.FundAccount(suite.app.BankKeeper, suite.ctx, addr, sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, amount)})
	suite.Require().NoError(err)

	bankGenesisState := suite.app.BankKeeper.ExportGenesis(suite.ctx)
	bankGenesis, err := suite.encodingConfig.Amino.MarshalJSON(bankGenesisState) // TODO switch this to use Marshaler
	suite.Require().NoError(err)

	return bankGenesis
}

func (suite *GenTxTestSuite) TestSetGenTxsInAppGenesisState() {
	var (
		txBuilder = suite.encodingConfig.TxConfig.NewTxBuilder()
		genTxs    []sdk.Tx
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"one genesis transaction",
			func() {
				err := txBuilder.SetMsgs(suite.msg1)
				suite.Require().NoError(err)
				tx := txBuilder.GetTx()
				genTxs = []sdk.Tx{tx}
			},
			true,
		},
		{
			"two genesis transactions",
			func() {
				err := txBuilder.SetMsgs(suite.msg1, suite.msg2)
				suite.Require().NoError(err)
				tx := txBuilder.GetTx()
				genTxs = []sdk.Tx{tx}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest()
			cdc := suite.encodingConfig.Marshaler
			txJSONEncoder := suite.encodingConfig.TxConfig.TxJSONEncoder()

			tc.malleate()
			appGenesisState, err := genutil.SetGenTxsInAppGenesisState(cdc, txJSONEncoder, make(map[string]json.RawMessage), genTxs)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(appGenesisState[types.ModuleName])

				var genesisState types.GenesisState
				err := cdc.UnmarshalJSON(appGenesisState[types.ModuleName], &genesisState)
				suite.Require().NoError(err)
				suite.Require().NotNil(genesisState.GenTxs)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *GenTxTestSuite) TestValidateAccountInGenesis() {
	var (
		appGenesisState = make(map[string]json.RawMessage)
		coins           sdk.Coins
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"no accounts",
			func() {
				coins = sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)}
			},
			false,
		},
		{
			"account without balance in the genesis state",
			func() {
				coins = sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)}
				appGenesisState[banktypes.ModuleName] = suite.setAccountBalance(addr2, 50)
			},
			false,
		},
		{
			"account without enough funds of default bond denom",
			func() {
				coins = sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 50)}
				appGenesisState[banktypes.ModuleName] = suite.setAccountBalance(addr1, 25)
			},
			false,
		},
		{
			"account with enough funds of default bond denom",
			func() {
				coins = sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 10)}
				appGenesisState[banktypes.ModuleName] = suite.setAccountBalance(addr1, 25)
			},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest()
			cdc := suite.encodingConfig.Marshaler

			suite.app.StakingKeeper.SetParams(suite.ctx, stakingtypes.DefaultParams())
			stakingGenesisState := staking.ExportGenesis(suite.ctx, suite.app.StakingKeeper)
			suite.Require().Equal(stakingGenesisState.Params, stakingtypes.DefaultParams())
			stakingGenesis, err := cdc.MarshalJSON(stakingGenesisState) // TODO switch this to use Marshaler
			suite.Require().NoError(err)
			appGenesisState[stakingtypes.ModuleName] = stakingGenesis

			tc.malleate()
			err = genutil.ValidateAccountInGenesis(
				appGenesisState, banktypes.GenesisBalancesIterator{},
				addr1, coins, cdc,
			)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}

		})
	}
}

func (suite *GenTxTestSuite) TestDeliverGenTxs() {
	var (
		genTxs    []json.RawMessage
		txBuilder = suite.encodingConfig.TxConfig.NewTxBuilder()
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"no signature supplied",
			func() {
				err := txBuilder.SetMsgs(suite.msg1)
				suite.Require().NoError(err)

				genTxs = make([]json.RawMessage, 1)
				tx, err := suite.encodingConfig.TxConfig.TxJSONEncoder()(txBuilder.GetTx())
				suite.Require().NoError(err)
				genTxs[0] = tx
			},
			false,
		},
		{
			"success",
			func() {
				_ = suite.setAccountBalance(addr1, 50)
				_ = suite.setAccountBalance(addr2, 1)

				msg := banktypes.NewMsgSend(addr1, addr2, sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 1)})
				tx, err := bbnapp.GenTx(
					suite.encodingConfig.TxConfig,
					[]sdk.Msg{msg},
					sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 10)},
					bbnapp.DefaultGenTxGas,
					suite.ctx.ChainID(),
					[]uint64{0},
					[]uint64{0},
					accpriv1,
				)
				suite.Require().NoError(err)

				genTxs = make([]json.RawMessage, 1)
				genTx, err := suite.encodingConfig.TxConfig.TxJSONEncoder()(tx)
				suite.Require().NoError(err)
				genTxs[0] = genTx
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest()

			tc.malleate()

			if tc.expPass {
				suite.Require().NotPanics(func() {
					genutil.DeliverGenTxs(
						suite.ctx, genTxs, suite.app.StakingKeeper, suite.app.BaseApp.DeliverTx,
						suite.encodingConfig.TxConfig,
					)
				})
			} else {
				suite.Require().Panics(func() {
					genutil.DeliverGenTxs(
						suite.ctx, genTxs, suite.app.StakingKeeper, suite.app.BaseApp.DeliverTx,
						suite.encodingConfig.TxConfig,
					)
				})
			}
		})
	}
}

func TestGenTxTestSuite(t *testing.T) {
	suite.Run(t, new(GenTxTestSuite))
}
