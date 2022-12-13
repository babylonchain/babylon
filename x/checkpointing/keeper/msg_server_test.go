package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/app"
	appparams "github.com/babylonchain/babylon/app/params"
	"github.com/babylonchain/babylon/crypto/bls12381"
	"github.com/babylonchain/babylon/privval"
	"github.com/babylonchain/babylon/testutil/datagen"
	checkpointingkeeper "github.com/babylonchain/babylon/x/checkpointing/keeper"
	"github.com/babylonchain/babylon/x/checkpointing/types"
	"github.com/babylonchain/babylon/x/epoching/testepoching"
	"github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

// FuzzWrappedCreateValidator tests adding new validators via
// MsgWrappedCreateValidator, which first registers BLS pubkey
// and then unwrapped into MsgCreateValidator and enqueued into
// the epoching module, and delivered to the staking module
// at epoch ends for execution
func FuzzWrappedCreateValidator(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 4)

	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)

		// a genesis validator is generate for setup
		helper := testepoching.NewHelper(t)
		ek := helper.EpochingKeeper
		ck := helper.App.CheckpointingKeeper
		msgServer := checkpointingkeeper.NewMsgServerImpl(ck)

		// BeginBlock of block 1, and thus entering epoch 1
		ctx := helper.BeginBlock()
		epoch := ek.GetEpoch(ctx)
		require.Equal(t, uint64(1), epoch.EpochNumber)

		// add n new validators via MsgWrappedCreateValidator
		n := rand.Intn(3)
		addrs := app.AddTestAddrs(helper.App, helper.Ctx, n, sdk.NewInt(100000000))

		wcvMsgs := make([]*types.MsgWrappedCreateValidator, n)
		for i := 0; i < n; i++ {
			msg, err := buildMsgWrappedCreateValidator(addrs[i])
			require.NoError(t, err)
			wcvMsgs[i] = msg
			_, err = msgServer.WrappedCreateValidator(ctx, msg)
			require.NoError(t, err)
			blsPK, err := ck.GetBlsPubKey(ctx, sdk.ValAddress(addrs[i]))
			require.NoError(t, err)
			require.True(t, msg.Key.Pubkey.Equal(blsPK))
		}
		require.Len(t, ek.GetCurrentEpochMsgs(ctx), n)

		// EndBlock of block 1
		ctx = helper.EndBlock()

		// go to BeginBlock of block 11, and thus entering epoch 2
		for i := uint64(0); i < ek.GetParams(ctx).EpochInterval; i++ {
			ctx = helper.GenAndApplyEmptyBlock()
		}
		epoch = ek.GetEpoch(ctx)
		require.Equal(t, uint64(2), epoch.EpochNumber)
		// ensure epoch 2 has initialised an empty msg queue
		require.Empty(t, ek.GetCurrentEpochMsgs(ctx))

		// check whether the length of current validator set equals to 1 + n
		// since one genesis validator was added when setup
		valSet = ck.GetValidatorSet(ctx, 2)
		require.Equal(t, len(wcvMsgs)+1, len(valSet))
		for _, msg := range wcvMsgs {
			found := false
			for _, val := range valSet {
				if msg.MsgCreateValidator.ValidatorAddress == val.GetValAddressStr() {
					found = true
				}
			}
			require.True(t, found)
		}
	})
}

func buildMsgWrappedCreateValidator(addr sdk.AccAddress) (*types.MsgWrappedCreateValidator, error) {
	tmValPrivkey := ed25519.GenPrivKey()
	bondTokens := sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction)
	bondCoin := sdk.NewCoin(appparams.DefaultBondDenom, bondTokens)
	description := stakingtypes.NewDescription("foo_moniker", "", "", "", "")
	commission := stakingtypes.NewCommissionRates(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec())

	pk, err := codec.FromTmPubKeyInterface(tmValPrivkey.PubKey())
	if err != nil {
		return nil, err
	}

	createValidatorMsg, err := stakingtypes.NewMsgCreateValidator(
		sdk.ValAddress(addr), pk, bondCoin, description, commission, sdk.OneInt(),
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
