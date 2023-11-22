package keeper_test

import (
	"context"
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	keepertest "github.com/babylonchain/babylon/testutil/keeper"
	bbn "github.com/babylonchain/babylon/types"
	bstypes "github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/babylonchain/babylon/x/finality/keeper"
	"github.com/babylonchain/babylon/x/finality/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
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

func FuzzCommitPubRandList(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		bsKeeper := types.NewMockBTCStakingKeeper(ctrl)
		fKeeper, ctx := keepertest.FinalityKeeper(t, bsKeeper, nil)
		ms := keeper.NewMsgServerImpl(*fKeeper)

		// create a random BTC validator
		btcSK, btcPK, err := datagen.GenRandomBTCKeyPair(r)
		require.NoError(t, err)
		valBTCPK := bbn.NewBIP340PubKeyFromBTCPK(btcPK)
		valBTCPKBytes := valBTCPK.MustMarshal()

		// Case 1: fail if the BTC validator is not registered
		bsKeeper.EXPECT().HasBTCValidator(gomock.Any(), gomock.Eq(valBTCPKBytes)).Return(false).Times(1)
		startHeight := datagen.RandomInt(r, 10)
		numPubRand := uint64(200)
		_, msg, err := datagen.GenRandomMsgCommitPubRandList(r, btcSK, startHeight, numPubRand)
		require.NoError(t, err)
		_, err = ms.CommitPubRandList(ctx, msg)
		require.Error(t, err)
		// register the BTC validator
		bsKeeper.EXPECT().HasBTCValidator(gomock.Any(), gomock.Eq(valBTCPKBytes)).Return(true).AnyTimes()

		// Case 2: commit a list of <minPubRand pubrand and it should fail
		startHeight = datagen.RandomInt(r, 10)
		numPubRand = datagen.RandomInt(r, int(fKeeper.GetParams(ctx).MinPubRand))
		_, msg, err = datagen.GenRandomMsgCommitPubRandList(r, btcSK, startHeight, numPubRand)
		require.NoError(t, err)
		_, err = ms.CommitPubRandList(ctx, msg)
		require.Error(t, err)

		// Case 3: when validator commits pubrand list and it should succeed
		startHeight = datagen.RandomInt(r, 10)
		numPubRand = 100 + datagen.RandomInt(r, int(fKeeper.GetParams(ctx).MinPubRand))
		_, msg, err = datagen.GenRandomMsgCommitPubRandList(r, btcSK, startHeight, numPubRand)
		require.NoError(t, err)
		_, err = ms.CommitPubRandList(ctx, msg)
		require.NoError(t, err)
		// query last public randomness and assert
		actualHeight, actualPubRand, err := fKeeper.GetLastPubRand(ctx, valBTCPK)
		require.NoError(t, err)
		require.Equal(t, startHeight+numPubRand-1, actualHeight)
		require.Equal(t, msg.PubRandList[len(msg.PubRandList)-1].MustMarshal(), actualPubRand.MustMarshal())

		// Case 4: commit a pubrand list with overlap of the existing pubrand in KVStore and it should fail
		overlappedStartHeight := startHeight + numPubRand - 1 - datagen.RandomInt(r, 5)
		_, msg, err = datagen.GenRandomMsgCommitPubRandList(r, btcSK, overlappedStartHeight, numPubRand)
		require.NoError(t, err)
		_, err = ms.CommitPubRandList(ctx, msg)
		require.Error(t, err)

		// Case 5: commit a pubrand list that has no overlap with existing pubrand and it should succeed
		nonOverlappedStartHeight := startHeight + numPubRand + datagen.RandomInt(r, 5)
		_, msg, err = datagen.GenRandomMsgCommitPubRandList(r, btcSK, nonOverlappedStartHeight, numPubRand)
		require.NoError(t, err)
		_, err = ms.CommitPubRandList(ctx, msg)
		require.NoError(t, err)
		// query last public randomness and assert
		actualHeight, actualPubRand, err = fKeeper.GetLastPubRand(ctx, valBTCPK)
		require.NoError(t, err)
		require.Equal(t, nonOverlappedStartHeight+numPubRand-1, actualHeight)
		require.Equal(t, msg.PubRandList[len(msg.PubRandList)-1].MustMarshal(), actualPubRand.MustMarshal())
	})
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

		// create and register a random BTC validator
		btcSK, btcPK, err := datagen.GenRandomBTCKeyPair(r)
		require.NoError(t, err)
		btcVal, err := datagen.GenRandomBTCValidatorWithBTCSK(r, btcSK)
		require.NoError(t, err)
		valBTCPK := bbn.NewBIP340PubKeyFromBTCPK(btcPK)
		valBTCPKBytes := valBTCPK.MustMarshal()
		require.NoError(t, err)
		bsKeeper.EXPECT().HasBTCValidator(gomock.Any(), gomock.Eq(valBTCPKBytes)).Return(true).AnyTimes()
		// commit some public randomness
		startHeight := uint64(0)
		numPubRand := uint64(200)
		srList, msgCommitPubRandList, err := datagen.GenRandomMsgCommitPubRandList(r, btcSK, startHeight, numPubRand)
		require.NoError(t, err)
		_, err = ms.CommitPubRandList(ctx, msgCommitPubRandList)
		require.NoError(t, err)

		// generate a vote
		blockHeight := uint64(1)
		sr, _ := srList[startHeight+blockHeight], msgCommitPubRandList.PubRandList[startHeight+blockHeight]
		blockHash := datagen.GenRandomByteArray(r, 32)
		signer := datagen.GenRandomAccount().Address
		msg, err := types.NewMsgAddFinalitySig(signer, btcSK, sr, blockHeight, blockHash)
		require.NoError(t, err)

		// Case 1: fail if the BTC validator does not have voting power
		bsKeeper.EXPECT().GetVotingPower(gomock.Any(), gomock.Eq(valBTCPKBytes), gomock.Eq(blockHeight)).Return(uint64(0)).Times(1)
		bsKeeper.EXPECT().GetBTCValidator(gomock.Any(), gomock.Eq(valBTCPKBytes)).Return(btcVal, nil).Times(1)
		_, err = ms.AddFinalitySig(ctx, msg)
		require.Error(t, err)

		// mock voting power
		bsKeeper.EXPECT().GetVotingPower(gomock.Any(), gomock.Eq(valBTCPKBytes), gomock.Eq(blockHeight)).Return(uint64(1)).AnyTimes()

		// Case 2: fail if the BTC validator has not committed public randomness at that height
		blockHeight2 := startHeight + numPubRand + 1
		bsKeeper.EXPECT().GetVotingPower(gomock.Any(), gomock.Eq(valBTCPKBytes), gomock.Eq(blockHeight2)).Return(uint64(1)).Times(1)
		bsKeeper.EXPECT().GetBTCValidator(gomock.Any(), gomock.Eq(valBTCPKBytes)).Return(btcVal, nil).Times(1)
		msg.BlockHeight = blockHeight2
		_, err = ms.AddFinalitySig(ctx, msg)
		require.Error(t, err)
		// reset block height
		msg.BlockHeight = blockHeight

		// Case 3: successful if the BTC validator has voting power and has not casted this vote yet
		// index this block first
		ctx = ctx.WithBlockHeader(tmproto.Header{Height: int64(blockHeight), AppHash: blockHash})
		fKeeper.IndexBlock(ctx)
		bsKeeper.EXPECT().GetBTCValidator(gomock.Any(), gomock.Eq(valBTCPKBytes)).Return(btcVal, nil).Times(1)
		// add vote and it should work
		_, err = ms.AddFinalitySig(ctx, msg)
		require.NoError(t, err)
		// query this vote and assert
		sig, err := fKeeper.GetSig(ctx, blockHeight, valBTCPK)
		require.NoError(t, err)
		require.Equal(t, msg.FinalitySig.MustMarshal(), sig.MustMarshal())

		// Case 4: fail if duplicate vote
		bsKeeper.EXPECT().GetBTCValidator(gomock.Any(), gomock.Eq(valBTCPKBytes)).Return(btcVal, nil).Times(1)
		_, err = ms.AddFinalitySig(ctx, msg)
		require.Error(t, err)

		// Case 5: the BTC validator is slashed if the BTC validator votes for a fork
		blockHash2 := datagen.GenRandomByteArray(r, 32)
		msg2, err := types.NewMsgAddFinalitySig(signer, btcSK, sr, blockHeight, blockHash2)
		require.NoError(t, err)
		bsKeeper.EXPECT().GetBTCValidator(gomock.Any(), gomock.Eq(valBTCPKBytes)).Return(btcVal, nil).Times(1)
		// mock slashing interface
		bsKeeper.EXPECT().SlashBTCValidator(gomock.Any(), gomock.Eq(valBTCPKBytes)).Return(nil).Times(1)
		// NOTE: even though this BTC validator is slashed, the msg should be successful
		// Otherwise the saved evidence will be rolled back
		_, err = ms.AddFinalitySig(ctx, msg2)
		require.NoError(t, err)
		// ensure the evidence has been stored
		evidence, err := fKeeper.GetEvidence(ctx, valBTCPK, blockHeight)
		require.NoError(t, err)
		require.Equal(t, msg2.BlockHeight, evidence.BlockHeight)
		require.Equal(t, msg2.ValBtcPk.MustMarshal(), evidence.ValBtcPk.MustMarshal())
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

		// Case 6: slashed BTC validator cannot vote
		btcVal.SlashedBabylonHeight = blockHeight
		bsKeeper.EXPECT().GetBTCValidator(gomock.Any(), gomock.Eq(valBTCPKBytes)).Return(btcVal, nil).Times(1)
		_, err = ms.AddFinalitySig(ctx, msg)
		require.Equal(t, bstypes.ErrBTCValAlreadySlashed, err)
	})
}
