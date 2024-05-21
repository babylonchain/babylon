package keeper_test

import (
	"context"
	"errors"
	"math/rand"
	"testing"
	"time"

	"cosmossdk.io/core/header"
	"github.com/babylonchain/babylon/testutil/datagen"
	keepertest "github.com/babylonchain/babylon/testutil/keeper"
	bbn "github.com/babylonchain/babylon/types"
	bstypes "github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/babylonchain/babylon/x/finality/keeper"
	"github.com/babylonchain/babylon/x/finality/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func setupMsgServer(t testing.TB) (*keeper.Keeper, types.MsgServer, context.Context) {
	fKeeper, ctx := keepertest.FinalityKeeper(t, nil, nil)
	return fKeeper, keeper.NewMsgServerImpl(*fKeeper), ctx
}

func TestMsgServer(t *testing.T) {
	_, ms, ctx := setupMsgServer(t)
	require.NotNil(t, ms)
	require.NotNil(t, ctx)
}

func FuzzAddFinalitySig(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		bsKeeper := types.NewMockBTCStakingKeeper(ctrl)
		fKeeper, ctx := keepertest.FinalityKeeper(t, bsKeeper, nil)
		ms := keeper.NewMsgServerImpl(*fKeeper)

		// create and register a random finality provider
		btcSK, btcPK, err := datagen.GenRandomBTCKeyPair(r)
		require.NoError(t, err)
		fp, err := datagen.GenRandomFinalityProviderWithBTCSK(r, btcSK)
		require.NoError(t, err)
		fpBTCPK := bbn.NewBIP340PubKeyFromBTCPK(btcPK)
		fpBTCPKBytes := fpBTCPK.MustMarshal()
		require.NoError(t, err)
		bsKeeper.EXPECT().HasFinalityProvider(gomock.Any(), gomock.Eq(fpBTCPKBytes)).Return(true).AnyTimes()
		startHeight := uint64(0)

		// generate a vote
		blockHeight := uint64(1)
		sr, _, err := msr.DeriveRandPair(uint32(startHeight + blockHeight))
		require.NoError(t, err)
		blockHash := datagen.GenRandomByteArray(r, 32)
		signer := datagen.GenRandomAccount().Address
		msg, err := types.NewMsgAddFinalitySig(signer, btcSK, sr, blockHeight, blockHash)
		require.NoError(t, err)

		// Case 1: slashed finality proivder cannot vote
		fp.SlashedBabylonHeight = blockHeight
		bsKeeper.EXPECT().GetFinalityProvider(gomock.Any(), gomock.Eq(fpBTCPKBytes)).Return(fp, nil).Times(1)
		_, err = ms.AddFinalitySig(ctx, msg)
		require.True(t, errors.Is(err, bstypes.ErrFpAlreadySlashed))

		// reset slashed height
		fp.SlashedBabylonHeight = 0

		// Case 2: fail if the finality provider's registered epoch is not finalised
		// by BTC timestamping yet
		bsKeeper.EXPECT().GetFinalityProvider(gomock.Any(), gomock.Eq(fpBTCPKBytes)).Return(fp, nil).Times(1)
		finalizedEpoch := registeredEpoch - 1
		bsKeeper.EXPECT().GetLastFinalizedEpoch(gomock.Any()).Return(finalizedEpoch).Times(1)
		_, err = ms.AddFinalitySig(ctx, msg)
		require.Error(t, err)
		require.True(t, errors.Is(err, bstypes.ErrFpNotBTCTimestamped))

		// make the registered finality provider BTC-timestamped
		finalizedEpoch = registeredEpoch
		bsKeeper.EXPECT().GetLastFinalizedEpoch(gomock.Any()).Return(finalizedEpoch).AnyTimes()

		// Case 3: fail if the finality provider does not have voting power
		bsKeeper.EXPECT().GetVotingPower(gomock.Any(), gomock.Eq(fpBTCPKBytes), gomock.Eq(blockHeight)).Return(uint64(0)).Times(1)
		bsKeeper.EXPECT().GetFinalityProvider(gomock.Any(), gomock.Eq(fpBTCPKBytes)).Return(fp, nil).Times(1)
		_, err = ms.AddFinalitySig(ctx, msg)
		require.Error(t, err)

		// mock voting power
		bsKeeper.EXPECT().GetVotingPower(gomock.Any(), gomock.Eq(fpBTCPKBytes), gomock.Eq(blockHeight)).Return(uint64(1)).AnyTimes()

		// Case 4: successful if the finality provider has voting power and has not casted this vote yet
		// index this block first
		ctx = ctx.WithHeaderInfo(header.Info{Height: int64(blockHeight), AppHash: blockHash})
		fKeeper.IndexBlock(ctx)
		bsKeeper.EXPECT().GetFinalityProvider(gomock.Any(), gomock.Eq(fpBTCPKBytes)).Return(fp, nil).Times(1)
		// add vote and it should work
		_, err = ms.AddFinalitySig(ctx, msg)
		require.NoError(t, err)
		// query this vote and assert
		sig, err := fKeeper.GetSig(ctx, blockHeight, fpBTCPK)
		require.NoError(t, err)
		require.Equal(t, msg.FinalitySig.MustMarshal(), sig.MustMarshal())

		// Case 5: In case of duplicate vote return success
		bsKeeper.EXPECT().GetFinalityProvider(gomock.Any(), gomock.Eq(fpBTCPKBytes)).Return(fp, nil).Times(1)
		resp, err := ms.AddFinalitySig(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Case 6: the finality provider is slashed if it votes for a fork
		blockHash2 := datagen.GenRandomByteArray(r, 32)
		msg2, err := types.NewMsgAddFinalitySig(signer, btcSK, sr, blockHeight, blockHash2)
		require.NoError(t, err)
		bsKeeper.EXPECT().GetFinalityProvider(gomock.Any(), gomock.Eq(fpBTCPKBytes)).Return(fp, nil).Times(1)
		// mock slashing interface
		bsKeeper.EXPECT().SlashFinalityProvider(gomock.Any(), gomock.Eq(fpBTCPKBytes)).Return(nil).Times(1)
		// NOTE: even though this finality provider is slashed, the msg should be successful
		// Otherwise the saved evidence will be rolled back
		_, err = ms.AddFinalitySig(ctx, msg2)
		require.NoError(t, err)
		// ensure the evidence has been stored
		evidence, err := fKeeper.GetEvidence(ctx, fpBTCPK, blockHeight)
		require.NoError(t, err)
		require.Equal(t, msg2.BlockHeight, evidence.BlockHeight)
		require.Equal(t, msg2.FpBtcPk.MustMarshal(), evidence.FpBtcPk.MustMarshal())
		require.Equal(t, msg2.BlockAppHash, evidence.ForkAppHash)
		require.Equal(t, msg2.FinalitySig.MustMarshal(), evidence.ForkFinalitySig.MustMarshal())
		// extract the SK and assert the extracted SK is correct
		btcSK2, err := evidence.ExtractBTCSK()
		require.NoError(t, err)
		// ensure btcSK and btcSK2 are same or inverse, AND correspond to the same PK
		// NOTE: it's possible that different SKs derive to the same PK
		// In this scenario, signature of any of these SKs can be verified with this PK
		// exclude the first byte here since it denotes the y axis of PubKey, which does
		// not affect verification
		require.True(t, btcSK.Key.Equals(&btcSK2.Key) || btcSK.Key.Negate().Equals(&btcSK2.Key))
		require.Equal(t, btcSK.PubKey().SerializeCompressed()[1:], btcSK2.PubKey().SerializeCompressed()[1:])
	})
}

func TestVoteForConflictingHashShouldRetrieveEvidenceAndSlash(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	bsKeeper := types.NewMockBTCStakingKeeper(ctrl)
	fKeeper, ctx := keepertest.FinalityKeeper(t, bsKeeper, nil)
	ms := keeper.NewMsgServerImpl(*fKeeper)
	// create and register a random finality provider
	btcSK, btcPK, err := datagen.GenRandomBTCKeyPair(r)
	require.NoError(t, err)
	fp, err := datagen.GenRandomFinalityProviderWithBTCSK(r, btcSK)
	require.NoError(t, err)
	fpBTCPK := bbn.NewBIP340PubKeyFromBTCPK(btcPK)
	fpBTCPKBytes := fpBTCPK.MustMarshal()
	require.NoError(t, err)
	bsKeeper.EXPECT().HasFinalityProvider(gomock.Any(),
		gomock.Eq(fpBTCPKBytes)).Return(true).AnyTimes()
	// commit some public randomness
	startHeight := uint64(0)

	// set a block height of 1 and some random list
	blockHeight := uint64(1)
	sr, _, err := msr.DeriveRandPair(uint32(startHeight + blockHeight))
	require.NoError(t, err)

	// generate two random hashes, one for the canonical block and
	// one for a fork block
	canonicalHash := datagen.GenRandomByteArray(r, 32)
	forkHash := datagen.GenRandomByteArray(r, 32)
	signer := datagen.GenRandomAccount().Address
	require.NoError(t, err)
	// (1) Set a canonical hash at height 1
	ctx = ctx.WithHeaderInfo(header.Info{Height: int64(blockHeight), AppHash: canonicalHash})
	fKeeper.IndexBlock(ctx)
	// (2) Vote for a different block at height 1, this will make us have
	// some "evidence"
	ctx = ctx.WithHeaderInfo(header.Info{Height: int64(blockHeight), AppHash: forkHash})
	msg1, err := types.NewMsgAddFinalitySig(signer, btcSK, sr,
		blockHeight, forkHash)

	require.NoError(t, err)
	bsKeeper.EXPECT().GetVotingPower(gomock.Any(),
		gomock.Eq(fpBTCPKBytes),
		gomock.Eq(blockHeight)).Return(uint64(1)).AnyTimes()
	bsKeeper.EXPECT().GetFinalityProvider(gomock.Any(),
		gomock.Eq(fpBTCPKBytes)).Return(fp, nil).Times(1)
	_, err = ms.AddFinalitySig(ctx, msg1)
	require.NoError(t, err)
	// (3) Now vote for the canonical block at height 1. This should slash Finality provider
	msg, err := types.NewMsgAddFinalitySig(signer, btcSK, sr,
		blockHeight, canonicalHash)
	ctx = ctx.WithHeaderInfo(header.Info{Height: int64(blockHeight), AppHash: canonicalHash})
	require.NoError(t, err)
	bsKeeper.EXPECT().GetVotingPower(gomock.Any(),
		gomock.Eq(fpBTCPKBytes),
		gomock.Eq(blockHeight)).Return(uint64(1)).AnyTimes()
	bsKeeper.EXPECT().GetFinalityProvider(gomock.Any(),
		gomock.Eq(fpBTCPKBytes)).Return(fp, nil).Times(1)
	bsKeeper.EXPECT().SlashFinalityProvider(gomock.Any(),
		gomock.Eq(fpBTCPKBytes)).Return(nil).Times(1)
	_, err = ms.AddFinalitySig(ctx, msg)
	require.NoError(t, err)
	sig, err := fKeeper.GetSig(ctx, blockHeight, fpBTCPK)
	require.NoError(t, err)
	require.Equal(t, msg.FinalitySig.MustMarshal(),
		sig.MustMarshal())
}
