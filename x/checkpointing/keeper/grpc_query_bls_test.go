package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/app"
	"github.com/babylonchain/babylon/testutil/datagen"
	checkpointingkeeper "github.com/babylonchain/babylon/x/checkpointing/keeper"
	"github.com/babylonchain/babylon/x/checkpointing/types"
	"github.com/babylonchain/babylon/x/epoching/testepoching"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
)

// FuzzQueryBLSKeySet does the following checks
// 1. check the query when there's only a genesis validator
// 2. check the query when there are n+1 validators without pagination
// 3. check the query when there are n+1 validators with pagination
func FuzzQueryBLSKeySet(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		// a genesis validator is generated for setup
		helper := testepoching.NewHelper(t)
		ek := helper.EpochingKeeper
		ck := helper.App.CheckpointingKeeper
		queryHelper := baseapp.NewQueryServerTestHelper(helper.Ctx, helper.App.InterfaceRegistry())
		types.RegisterQueryServer(queryHelper, ck)
		queryClient := types.NewQueryClient(queryHelper)
		msgServer := checkpointingkeeper.NewMsgServerImpl(ck)
		genesisVal := ek.GetValidatorSet(helper.Ctx, 0)[0]
		genesisBLSPubkey, err := ck.GetBlsPubKey(helper.Ctx, genesisVal.Addr)
		require.NoError(t, err)

		// BeginBlock of block 1, and thus entering epoch 1
		ctx := helper.BeginBlock(r)
		epoch := ek.GetEpoch(ctx)
		require.Equal(t, uint64(1), epoch.EpochNumber)

		// 1. query public key list when there's only a genesis validator
		queryRequest := &types.QueryBlsPublicKeyListRequest{
			EpochNum: 1,
		}
		res, err := queryClient.BlsPublicKeyList(ctx, queryRequest)
		require.NoError(t, err)
		require.Len(t, res.ValidatorWithBlsKeys, 1)
		require.Equal(t, res.ValidatorWithBlsKeys[0].BlsPubKey, genesisBLSPubkey.Bytes())
		require.Equal(t, res.ValidatorWithBlsKeys[0].VotingPower, uint64(1000))
		require.Equal(t, res.ValidatorWithBlsKeys[0].ValidatorAddress, genesisVal.GetValAddressStr())

		// add n new validators via MsgWrappedCreateValidator
		n := r.Intn(3) + 1
		addrs := app.AddTestAddrs(helper.App, helper.Ctx, n, sdk.NewInt(100000000))

		wcvMsgs := make([]*types.MsgWrappedCreateValidator, n)
		for i := 0; i < n; i++ {
			msg, err := buildMsgWrappedCreateValidator(addrs[i])
			require.NoError(t, err)
			wcvMsgs[i] = msg
			_, err = msgServer.WrappedCreateValidator(ctx, msg)
			require.NoError(t, err)
		}

		// EndBlock of block 1
		ctx = helper.EndBlock()

		// go to BeginBlock of block 11, and thus entering epoch 2
		for i := uint64(0); i < ek.GetParams(ctx).EpochInterval; i++ {
			ctx = helper.GenAndApplyEmptyBlock(r)
		}
		epoch = ek.GetEpoch(ctx)
		require.Equal(t, uint64(2), epoch.EpochNumber)

		// 2. query BLS public keys when there are n+1 validators without pagination
		req := types.QueryBlsPublicKeyListRequest{
			EpochNum: 2,
		}
		resp, err := queryClient.BlsPublicKeyList(ctx, &req)
		require.NoError(t, err)
		require.Len(t, resp.ValidatorWithBlsKeys, n+1)
		expectedValSet := ek.GetValidatorSet(ctx, 2)
		require.Len(t, expectedValSet, n+1)
		for i, expectedVal := range expectedValSet {
			require.Equal(t, uint64(expectedVal.Power), resp.ValidatorWithBlsKeys[i].VotingPower)
			require.Equal(t, expectedVal.GetValAddressStr(), resp.ValidatorWithBlsKeys[i].ValidatorAddress)
		}

		// 3.1 query BLS public keys when there are n+1 validators with limit pagination
		req = types.QueryBlsPublicKeyListRequest{
			EpochNum: 2,
			Pagination: &query.PageRequest{
				Limit: 1,
			},
		}
		resp, err = queryClient.BlsPublicKeyList(ctx, &req)
		require.NoError(t, err)
		require.Len(t, resp.ValidatorWithBlsKeys, 1)

		// 3.2 query BLS public keys when there are n+1 validators with offset pagination
		req = types.QueryBlsPublicKeyListRequest{
			EpochNum: 2,
			Pagination: &query.PageRequest{
				Offset: 1,
			},
		}
		resp, err = queryClient.BlsPublicKeyList(ctx, &req)
		require.NoError(t, err)
		require.Len(t, resp.ValidatorWithBlsKeys, n)
	})
}
