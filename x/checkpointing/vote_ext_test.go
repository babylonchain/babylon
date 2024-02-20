package checkpointing_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/babylonchain/babylon/testutil/datagen"
	testhelper "github.com/babylonchain/babylon/testutil/helper"
	"github.com/babylonchain/babylon/x/checkpointing/types"
)

func TestVoteExtensionHandler_ExtendVote(t *testing.T) {
	tests := []struct {
		name        string
		expectError bool
	}{
		{
			name:        "Not in validator set",
			expectError: true,
		},
		{
			name:        "Nil BLS private key",
			expectError: true,
		},
		{
			name:        "Invalid block hash",
			expectError: true,
		},

		// TODO: Without the the fix this panics, even though it should not
		// {
		// 	name:        "List with only empty vote extensions",
		// 	votes:       GenerateNExtendedVotes(t, 4),
		// 	expectError: true,
		// },
		// TODO: Add more failing scenarios and at least one successful one
		// TODO: Add scenarios testing compatibility of prepareProposal, processProposal and preBlocker
	}
}

// FuzzAddBLSSigVoteExtension_MultipleVals tests adding BLS signatures via VoteExtension
// with multiple validators
func FuzzAddBLSSigVoteExtension_MultipleVals(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		// generate the validator set with 10 validators as genesis
		genesisValSet, privSigner, err := datagen.GenesisValidatorSetWithPrivSigner(10)
		require.NoError(t, err)
		helper := testhelper.NewHelperWithValSet(t, genesisValSet, privSigner)
		ek := helper.App.EpochingKeeper
		ck := helper.App.CheckpointingKeeper

		epoch := ek.GetEpoch(helper.Ctx)
		require.Equal(t, uint64(1), epoch.EpochNumber)

		// go to block 11, ensure the checkpoint is finalized
		interval := ek.GetParams(helper.Ctx).EpochInterval
		for i := uint64(0); i < interval; i++ {
			_, err := helper.ApplyEmptyBlockWithVoteExtension(r)
			require.NoError(t, err)
		}

		epoch = ek.GetEpoch(helper.Ctx)
		require.Equal(t, uint64(2), epoch.EpochNumber)

		ckpt, err := ck.GetRawCheckpoint(helper.Ctx, epoch.EpochNumber-1)
		require.NoError(t, err)
		require.Equal(t, types.Sealed, ckpt.Status)
	})
}

// FuzzAddBLSSigVoteExtension_InsufficientVotingPower tests adding BLS signatures
// with insufficient voting power
func FuzzAddBLSSigVoteExtension_InsufficientVotingPower(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		// generate the validator set with 10 validators as genesis
		genesisValSet, privSigner, err := datagen.GenesisValidatorSetWithPrivSigner(10)
		require.NoError(t, err)
		helper := testhelper.NewHelperWithValSet(t, genesisValSet, privSigner)
		ek := helper.App.EpochingKeeper

		epoch := ek.GetEpoch(helper.Ctx)
		require.Equal(t, uint64(1), epoch.EpochNumber)

		// the number of validators is less than 2/3 if the total set
		numOfValidators := datagen.RandomInt(r, 5) + 1
		genesisValSet.Keys = genesisValSet.Keys[:numOfValidators]
		interval := ek.GetParams(helper.Ctx).EpochInterval
		for i := uint64(0); i < interval-1; i++ {
			_, err := helper.ApplyEmptyBlockWithValSet(r, genesisValSet)
			if i < interval-2 {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		}
	})
}

// FuzzAddBLSSigVoteExtension_InvalidBLSSig tests adding BLS signatures
// with invalid BLS signature
func FuzzAddBLSSigVoteExtension_InvalidBLSSig(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		helper := testhelper.NewHelper(t)
		ek := helper.App.EpochingKeeper

		epoch := ek.GetEpoch(helper.Ctx)
		require.Equal(t, uint64(1), epoch.EpochNumber)

		interval := ek.GetParams(helper.Ctx).EpochInterval
		for i := uint64(0); i < interval-1; i++ {
			_, err := helper.ApplyEmptyBlockWithInvalidBLSSig(r)
			if i < interval-2 {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		}
	})
}

// FuzzAddBLSSigVoteExtension_EmptyVoteExtensions tests resilience against
// empty vote extensions
func FuzzAddBLSSigVoteExtension_EmptyVoteExtensions(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		// generate the validator set with 10 validators as genesis
		genesisValSet, privSigner, err := datagen.GenesisValidatorSetWithPrivSigner(10)
		require.NoError(t, err)
		helper := testhelper.NewHelperWithValSet(t, genesisValSet, privSigner)
		ek := helper.App.EpochingKeeper
		ck := helper.App.CheckpointingKeeper

		epoch := ek.GetEpoch(helper.Ctx)
		require.Equal(t, uint64(1), epoch.EpochNumber)

		// go to block 10, ensure the checkpoint is finalized
		interval := ek.GetParams(helper.Ctx).EpochInterval
		for i := uint64(0); i < interval-2; i++ {
			_, err := helper.ApplyEmptyBlockWithSomeEmptyVoteExtensions(r)
			require.NoError(t, err)
		}
		// height 11, i.e., 1st block of next epoch
		_, err = helper.ApplyEmptyBlockWithSomeEmptyVoteExtensions(r)
		require.NoError(t, err)

		epoch = ek.GetEpoch(helper.Ctx)
		require.Equal(t, uint64(2), epoch.EpochNumber)

		ckpt, err := ck.GetRawCheckpoint(helper.Ctx, epoch.EpochNumber-1)
		require.NoError(t, err)
		require.Equal(t, types.Sealed, ckpt.Status)
	})
}
