package cli_test

import (
	"context"
	sdkmath "cosmossdk.io/math"
	"fmt"
	"github.com/babylonchain/babylon/app"
	"github.com/cosmos/cosmos-sdk/testutil"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	"io"
	"path/filepath"
	"testing"

	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/stretchr/testify/suite"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtconfig "github.com/cometbft/cometbft/config"
	tmbytes "github.com/cometbft/cometbft/libs/bytes"
	tmos "github.com/cometbft/cometbft/libs/os"
	rpcclient "github.com/cometbft/cometbft/rpc/client"
	rpcclientmock "github.com/cometbft/cometbft/rpc/client/mock"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/babylonchain/babylon/app/params"
	"github.com/babylonchain/babylon/privval"
	testutilcli "github.com/babylonchain/babylon/testutil/cli"
	checkpointcli "github.com/babylonchain/babylon/x/checkpointing/client/cli"
)

type mockCometRPC struct {
	rpcclientmock.Client

	responseQuery abci.ResponseQuery
}

func newMockCometRPC(respQuery abci.ResponseQuery) mockCometRPC {
	return mockCometRPC{responseQuery: respQuery}
}

func (mockCometRPC) BroadcastTxSync(_ context.Context, _ tmtypes.Tx) (*coretypes.ResultBroadcastTx, error) {
	return &coretypes.ResultBroadcastTx{}, nil
}

func (m mockCometRPC) ABCIQueryWithOptions(
	_ context.Context,
	_ string, _ tmbytes.HexBytes,
	_ rpcclient.ABCIQueryOptions,
) (*coretypes.ResultABCIQuery, error) {
	return &coretypes.ResultABCIQuery{Response: m.responseQuery}, nil
}

type CLITestSuite struct {
	suite.Suite

	kr        keyring.Keyring
	encCfg    *params.EncodingConfig
	baseCtx   client.Context
	clientCtx client.Context
	addrs     []sdk.AccAddress
}

func (s *CLITestSuite) SetupSuite() {
	s.encCfg = app.GetEncodingConfig()
	s.kr = keyring.NewInMemory(s.encCfg.Codec)
	s.baseCtx = client.Context{}.
		WithKeyring(s.kr).
		WithTxConfig(s.encCfg.TxConfig).
		WithCodec(s.encCfg.Codec).
		WithClient(mockCometRPC{Client: rpcclientmock.Client{}}).
		WithAccountRetriever(client.MockAccountRetriever{}).
		WithOutput(io.Discard).
		WithChainID("test-chain")

	ctxGen := func() client.Context {
		bz, _ := s.encCfg.Codec.Marshal(&sdk.TxResponse{})
		c := newMockCometRPC(abci.ResponseQuery{
			Value: bz,
		})
		return s.baseCtx.WithClient(c)
	}
	s.clientCtx = ctxGen()

	s.addrs = make([]sdk.AccAddress, 0)
	for i := 0; i < 3; i++ {
		k, _, err := s.clientCtx.Keyring.NewMnemonic("NewValidator", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
		s.Require().NoError(err)

		pub, err := k.GetPubKey()
		s.Require().NoError(err)

		newAddr := sdk.AccAddress(pub.Address())
		s.addrs = append(s.addrs, newAddr)
	}
}

// test cases copied from https://github.com/cosmos/cosmos-sdk/blob/v0.50.1/x/staking/client/cli/tx_test.go#L163
func (s *CLITestSuite) TestCmdWrappedCreateValidator() {
	require := s.Require()
	homeDir := s.T().TempDir()
	nodeCfg := cmtconfig.DefaultConfig()
	pvKeyFile := filepath.Join(homeDir, nodeCfg.PrivValidatorKeyFile())
	err := tmos.EnsureDir(filepath.Dir(pvKeyFile), 0777)
	require.NoError(err)
	pvStateFile := filepath.Join(homeDir, nodeCfg.PrivValidatorStateFile())
	err = tmos.EnsureDir(filepath.Dir(pvStateFile), 0777)
	require.NoError(err)
	wrappedPV := privval.LoadOrGenWrappedFilePV(pvKeyFile, pvStateFile)
	cmd := checkpointcli.CmdWrappedCreateValidator(authcodec.NewBech32Codec("cosmosvaloper"))

	consPrivKey := wrappedPV.GetValPrivKey()
	consPubKey, err := cryptocodec.FromCmtPubKeyInterface(consPrivKey.PubKey())
	require.NoError(err)
	consPubKeyBz, err := s.clientCtx.Codec.MarshalInterfaceJSON(consPubKey)
	require.NoError(err)
	require.NotNil(consPubKeyBz)

	validJSON := fmt.Sprintf(`
	{
  		"pubkey": %s,
  		"amount": "%dstake",
  		"moniker": "NewValidator",
		"identity": "AFAF00C4",
		"website": "https://newvalidator.io",
		"security": "contact@newvalidator.io",
		"details": "'Hey, I am a new validator. Please delegate!'",
  		"commission-rate": "0.5",
  		"commission-max-rate": "1.0",
  		"commission-max-change-rate": "0.1",
  		"min-self-delegation": "1"
	}`, consPubKeyBz, 100)
	validJSONFile := testutil.WriteToNewTempFile(s.T(), validJSON)
	defer validJSONFile.Close()

	validJSONWithoutOptionalFields := fmt.Sprintf(`
	{
  		"pubkey": %s,
  		"amount": "%dstake",
  		"moniker": "NewValidator",
  		"commission-rate": "0.5",
  		"commission-max-rate": "1.0",
  		"commission-max-change-rate": "0.1",
  		"min-self-delegation": "1"
	}`, consPubKeyBz, 100)
	validJSONWOOptionalFile := testutil.WriteToNewTempFile(s.T(), validJSONWithoutOptionalFields)
	defer validJSONWOOptionalFile.Close()

	noAmountJSON := fmt.Sprintf(`
	{
  		"pubkey": %s,
  		"moniker": "NewValidator",
  		"commission-rate": "0.5",
  		"commission-max-rate": "1.0",
  		"commission-max-change-rate": "0.1",
  		"min-self-delegation": "1"
	}`, consPubKeyBz)
	noAmountJSONFile := testutil.WriteToNewTempFile(s.T(), noAmountJSON)
	defer noAmountJSONFile.Close()

	noPubKeyJSON := fmt.Sprintf(`
	{
  		"amount": "%dstake",
  		"moniker": "NewValidator",
  		"commission-rate": "0.5",
  		"commission-max-rate": "1.0",
  		"commission-max-change-rate": "0.1",
  		"min-self-delegation": "1"
	}`, 100)
	noPubKeyJSONFile := testutil.WriteToNewTempFile(s.T(), noPubKeyJSON)
	defer noPubKeyJSONFile.Close()

	noMonikerJSON := fmt.Sprintf(`
	{
  		"pubkey": {"@type":"/cosmos.crypto.ed25519.PubKey","key":"oWg2ISpLF405Jcm2vXV+2v4fnjodh6aafuIdeoW+rUw="},
  		"amount": "%dstake",
  		"commission-rate": "0.5",
  		"commission-max-rate": "1.0",
  		"commission-max-change-rate": "0.1",
  		"min-self-delegation": "1"
	}`, 100)
	noMonikerJSONFile := testutil.WriteToNewTempFile(s.T(), noMonikerJSON)
	defer noMonikerJSONFile.Close()

	testCases := []struct {
		name         string
		args         []string
		expectErrMsg string
	}{
		{
			"invalid transaction (missing amount)",
			[]string{
				noAmountJSONFile.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addrs[0]),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 10)).String()),
				fmt.Sprintf("--%s=%s", flags.FlagHome, homeDir),
			},
			"must specify amount of coins to bond",
		},
		{
			"invalid transaction (missing pubkey)",
			[]string{
				noPubKeyJSONFile.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addrs[0]),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 10)).String()),
				fmt.Sprintf("--%s=%s", flags.FlagHome, homeDir),
			},
			"must specify the JSON encoded pubkey",
		},
		{
			"invalid transaction (missing moniker)",
			[]string{
				noMonikerJSONFile.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addrs[0]),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 10)).String()),
				fmt.Sprintf("--%s=%s", flags.FlagHome, homeDir),
			},
			"must specify the moniker name",
		},
		{
			"valid transaction with all fields",
			[]string{
				validJSONFile.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addrs[0]),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10))).String()),
				fmt.Sprintf("--%s=%s", flags.FlagHome, homeDir),
			},
			"",
		},
		{
			"valid transaction without optional fields",
			[]string{
				validJSONWOOptionalFile.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addrs[0]),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10))).String()),
				fmt.Sprintf("--%s=%s", flags.FlagHome, homeDir),
			},
			"",
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			out, err := testutilcli.ExecTestCLICmd(s.clientCtx, cmd, tc.args)
			if tc.expectErrMsg != "" {
				require.Error(err)
				require.Contains(err.Error(), tc.expectErrMsg)
			} else {
				require.NoError(err, "test: %s\noutput: %s", tc.name, out.String())
				resp := &sdk.TxResponse{}
				err = s.clientCtx.Codec.UnmarshalJSON(out.Bytes(), resp)
				require.NoError(err, out.String(), "test: %s, output\n:", tc.name, out.String())
			}
		})
	}
}

func TestCLITestSuite(t *testing.T) {
	// t.Skip()
	suite.Run(t, new(CLITestSuite))
}
