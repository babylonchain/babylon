package keeper_test

import (
	"fmt"
	"testing"

	"github.com/babylonchain/babylon/app"
	"github.com/babylonchain/babylon/crypto/bls12381"
	testkeeper "github.com/babylonchain/babylon/testutil/keeper"
	"github.com/babylonchain/babylon/testutil/mocks"
	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/tmhash"
)

var (
	pk1   = ed25519.GenPrivKey().PubKey()
	pk2   = ed25519.GenPrivKey().PubKey()
	addr1 = sdk.ValAddress(pk1.Address())
	addr2 = sdk.ValAddress(pk2.Address())
	val1  = epochingtypes.Validator{
		Addr:  addr1,
		Power: 10,
	}
	val2 = epochingtypes.Validator{
		Addr:  addr2,
		Power: 10,
	}
	valSet      = epochingtypes.ValidatorSet{val1, val2}
	blsPrivKey1 = bls12381.GenPrivKey()
)

func TestKeeper_SendBlsSig(t *testing.T) {
	cfg := network.DefaultConfig()
	encodingCfg := app.MakeTestEncodingConfig()
	cfg.InterfaceRegistry = encodingCfg.InterfaceRegistry
	cfg.TxConfig = encodingCfg.TxConfig
	cfg.NumValidators = 1

	testNetwork, err := network.New(t, t.TempDir(), cfg)
	require.NoError(t, err)
	defer testNetwork.Cleanup()

	val := testNetwork.Validators[0]
	nodeDirName := fmt.Sprintf("node%d", 0)
	clientCtx := val.ClientCtx.WithHeight(2).
		WithFromAddress(val.Address).
		WithFromName(nodeDirName).
		WithBroadcastMode(flags.BroadcastAsync)
	clientCtx.SkipConfirm = true

	epochNum := uint64(10)
	lch := tmhash.Sum([]byte("last_commit_hash"))
	signBytes := append(sdk.Uint64ToBigEndian(epochNum), lch...)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ek := mocks.NewMockEpochingKeeper(ctrl)
	signer := mocks.NewMockBlsSigner(ctrl)
	ckptkeeper, ctx, _ := testkeeper.CheckpointingKeeper(t, ek, signer, clientCtx)

	ek.EXPECT().GetValidatorSet(ctx, gomock.Eq(epochNum)).Return(valSet)
	signer.EXPECT().GetAddress().Return(addr1)
	signer.EXPECT().SignMsgWithBls(gomock.Eq(signBytes)).Return(bls12381.Sign(blsPrivKey1, signBytes), nil)
	err = ckptkeeper.SendBlsSig(ctx, epochNum, lch)
	require.NoError(t, err)
}
