package cli_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"testing"

	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	"github.com/golang/mock/gomock"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/client/cli"

	abci "github.com/tendermint/tendermint/abci/types"
	tmconfig "github.com/tendermint/tendermint/config"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	tmos "github.com/tendermint/tendermint/libs/os"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	rpcclientmock "github.com/tendermint/tendermint/rpc/client/mock"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/babylonchain/babylon/app"
	"github.com/babylonchain/babylon/app/params"
	"github.com/babylonchain/babylon/privval"
	testutilcli "github.com/babylonchain/babylon/testutil/cli"
	checkpointcli "github.com/babylonchain/babylon/x/checkpointing/client/cli"
)

type mockTendermintRPC struct {
	rpcclientmock.Client

	responseQuery abci.ResponseQuery
}

func newMockTendermintRPC(respQuery abci.ResponseQuery) mockTendermintRPC {
	return mockTendermintRPC{responseQuery: respQuery}
}

func (mockTendermintRPC) BroadcastTxSync(_ context.Context, _ tmtypes.Tx) (*coretypes.ResultBroadcastTx, error) {
	return &coretypes.ResultBroadcastTx{}, nil
}

func (m mockTendermintRPC) ABCIQueryWithOptions(
	_ context.Context,
	_ string, _ tmbytes.HexBytes,
	_ rpcclient.ABCIQueryOptions,
) (*coretypes.ResultABCIQuery, error) {
	return &coretypes.ResultABCIQuery{Response: m.responseQuery}, nil
}

type CLITestSuite struct {
	suite.Suite

	kr        keyring.Keyring
	encCfg    params.EncodingConfig
	baseCtx   client.Context
	clientCtx client.Context
	addrs     []sdk.AccAddress
}

func (s *CLITestSuite) SetupSuite() {
	s.encCfg = app.GetEncodingConfig()
	s.kr = keyring.NewInMemory(s.encCfg.Marshaler)
	ctrl := gomock.NewController(s.T())
	mockAccountRetriever := mock.NewMockAccountRetriever(ctrl)
	mockAccountRetriever.EXPECT().EnsureExists(gomock.Any(), gomock.Any()).Return(nil)
	mockAccountRetriever.EXPECT().GetAccountNumberSequence(gomock.Any(), gomock.Any()).Return(uint64(0), uint64(0), nil)
	s.baseCtx = client.Context{}.
		WithKeyring(s.kr).
		WithTxConfig(s.encCfg.TxConfig).
		WithCodec(s.encCfg.Marshaler).
		WithClient(mockTendermintRPC{Client: rpcclientmock.Client{}}).
		WithAccountRetriever(mockAccountRetriever).
		WithOutput(io.Discard).
		WithChainID("test-chain")

	var outBuf bytes.Buffer
	ctxGen := func() client.Context {
		bz, _ := s.encCfg.Marshaler.Marshal(&sdk.TxResponse{})
		c := newMockTendermintRPC(abci.ResponseQuery{
			Value: bz,
		})
		return s.baseCtx.WithClient(c)
	}
	s.clientCtx = ctxGen().WithOutput(&outBuf)
	s.addrs = make([]sdk.AccAddress, 0)
	for i := 0; i < 3; i++ {
		k, _, err := s.clientCtx.Keyring.NewMnemonic(fmt.Sprintf("NewWrappedValidator%v", i), keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
		s.Require().NoError(err)

		pub, _ := k.GetPubKey()

		newAddr := sdk.AccAddress(pub.Address())
		s.addrs = append(s.addrs, newAddr)
	}
}

// test cases copied from https://github.com/cosmos/cosmos-sdk/blob/dabcedce71b43161c8357d051715d0d3a0919883/x/staking/client/cli/tx_test.go#L191
func (s *CLITestSuite) TestCmdWrappedCreateValidator() {
	require := s.Require()
	homeDir := s.T().TempDir()
	nodeCfg := tmconfig.DefaultConfig()
	pvKeyFile := filepath.Join(homeDir, nodeCfg.PrivValidatorKeyFile())
	err := tmos.EnsureDir(filepath.Dir(pvKeyFile), 0777)
	require.NoError(err)
	pvStateFile := filepath.Join(homeDir, nodeCfg.PrivValidatorStateFile())
	err = tmos.EnsureDir(filepath.Dir(pvStateFile), 0777)
	require.NoError(err)
	wrappedPV := privval.LoadOrGenWrappedFilePV(pvKeyFile, pvStateFile)
	cmd := checkpointcli.CmdWrappedCreateValidator()

	consPrivKey := wrappedPV.GetValPrivKey()
	consPubKey, err := cryptocodec.FromTmPubKeyInterface(consPrivKey.PubKey())
	require.NoError(err)
	consPubKeyBz, err := s.clientCtx.Codec.MarshalInterfaceJSON(consPubKey)
	require.NoError(err)
	require.NotNil(consPubKeyBz)

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectedCode uint32
		respType     proto.Message
	}{
		{
			"invalid transaction (missing amount)",
			[]string{
				fmt.Sprintf("--%s=AFAF00C4", cli.FlagIdentity),
				fmt.Sprintf("--%s=https://newvalidator.io", cli.FlagWebsite),
				fmt.Sprintf("--%s=contact@newvalidator.io", cli.FlagSecurityContact),
				fmt.Sprintf("--%s='Hey, I am a new validator. Please delegate!'", cli.FlagDetails),
				fmt.Sprintf("--%s=0.5", cli.FlagCommissionRate),
				fmt.Sprintf("--%s=1.0", cli.FlagCommissionMaxRate),
				fmt.Sprintf("--%s=0.1", cli.FlagCommissionMaxChangeRate),
				fmt.Sprintf("--%s=1", cli.FlagMinSelfDelegation),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addrs[0]),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--%s=%s", flags.FlagHome, homeDir),
			},
			true, 0, nil,
		},
		{
			"invalid transaction (missing pubkey)",
			[]string{
				fmt.Sprintf("--%s=%dstake", cli.FlagAmount, 100),
				fmt.Sprintf("--%s=AFAF00C4", cli.FlagIdentity),
				fmt.Sprintf("--%s=https://newvalidator.io", cli.FlagWebsite),
				fmt.Sprintf("--%s=contact@newvalidator.io", cli.FlagSecurityContact),
				fmt.Sprintf("--%s='Hey, I am a new validator. Please delegate!'", cli.FlagDetails),
				fmt.Sprintf("--%s=0.5", cli.FlagCommissionRate),
				fmt.Sprintf("--%s=1.0", cli.FlagCommissionMaxRate),
				fmt.Sprintf("--%s=0.1", cli.FlagCommissionMaxChangeRate),
				fmt.Sprintf("--%s=1", cli.FlagMinSelfDelegation),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addrs[0]),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--%s=%s", flags.FlagHome, homeDir),
			},
			true, 0, nil,
		},
		{
			"invalid transaction (missing moniker)",
			[]string{
				fmt.Sprintf("--%s=%s", cli.FlagPubKey, consPubKeyBz),
				fmt.Sprintf("--%s=%dstake", cli.FlagAmount, 100),
				fmt.Sprintf("--%s=AFAF00C4", cli.FlagIdentity),
				fmt.Sprintf("--%s=https://newvalidator.io", cli.FlagWebsite),
				fmt.Sprintf("--%s=contact@newvalidator.io", cli.FlagSecurityContact),
				fmt.Sprintf("--%s='Hey, I am a new validator. Please delegate!'", cli.FlagDetails),
				fmt.Sprintf("--%s=0.5", cli.FlagCommissionRate),
				fmt.Sprintf("--%s=1.0", cli.FlagCommissionMaxRate),
				fmt.Sprintf("--%s=0.1", cli.FlagCommissionMaxChangeRate),
				fmt.Sprintf("--%s=1", cli.FlagMinSelfDelegation),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addrs[0]),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--%s=%s", flags.FlagHome, homeDir),
			},
			true, 0, nil,
		},
		{
			"valid transaction",
			[]string{
				fmt.Sprintf("--%s=%s", cli.FlagPubKey, consPubKeyBz),
				fmt.Sprintf("--%s=%dstake", cli.FlagAmount, 100),
				fmt.Sprintf("--%s=NewValidator", cli.FlagMoniker),
				fmt.Sprintf("--%s=AFAF00C4", cli.FlagIdentity),
				fmt.Sprintf("--%s=https://newvalidator.io", cli.FlagWebsite),
				fmt.Sprintf("--%s=contact@newvalidator.io", cli.FlagSecurityContact),
				fmt.Sprintf("--%s='Hey, I am a new validator. Please delegate!'", cli.FlagDetails),
				fmt.Sprintf("--%s=0.5", cli.FlagCommissionRate),
				fmt.Sprintf("--%s=1.0", cli.FlagCommissionMaxRate),
				fmt.Sprintf("--%s=0.1", cli.FlagCommissionMaxChangeRate),
				fmt.Sprintf("--%s=1", cli.FlagMinSelfDelegation),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addrs[0]),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--%s=%s", flags.FlagHome, homeDir),
			},
			false, 0, &sdk.TxResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			out, err := testutilcli.ExecTestCLICmd(s.clientCtx, cmd, tc.args)
			if tc.expectErr {
				require.Error(err)
			} else {
				require.NoError(err, "test: %s\noutput: %s", tc.name, out.String())
				err = s.clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType)
				require.NoError(err, out.String(), "test: %s, output\n:", tc.name, out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				require.Equal(tc.expectedCode, txResp.Code,
					"test: %s, output\n:", tc.name, out.String())
			}
		})
	}
}

func TestCLITestSuite(t *testing.T) {
	// t.Skip()
	suite.Run(t, new(CLITestSuite))
}
