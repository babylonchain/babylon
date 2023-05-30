package keeper_test

import (
	"math/rand"
	"testing"

	"cosmossdk.io/math"
	"github.com/babylonchain/babylon/app"
	appparams "github.com/babylonchain/babylon/app/params"
	"github.com/babylonchain/babylon/crypto/bls12381"
	"github.com/babylonchain/babylon/privval"
	"github.com/babylonchain/babylon/testutil/datagen"
	checkpointingkeeper "github.com/babylonchain/babylon/x/checkpointing/keeper"
	"github.com/babylonchain/babylon/x/checkpointing/types"
	"github.com/babylonchain/babylon/x/epoching/testepoching"
	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"
	"github.com/cometbft/cometbft/crypto/ed25519"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
)

// FuzzWrappedCreateValidator_InsufficientTokens tests adding new validators with zero voting power
// It ensures that validators with zero voting power (i.e., with tokens fewer than sdk.DefaultPowerReduction)
// are unbonded, thus are not included in the validator set
func FuzzWrappedCreateValidator_InsufficientTokens(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 4)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		// a genesis validator is generate for setup
		helper := testepoching.NewHelper(t)
		ek := helper.EpochingKeeper
		ck := helper.App.CheckpointingKeeper
		msgServer := checkpointingkeeper.NewMsgServerImpl(ck)

		// BeginBlock of block 1, and thus entering epoch 1
		ctx := helper.BeginBlock(r)
		epoch := ek.GetEpoch(ctx)
		require.Equal(t, uint64(1), epoch.EpochNumber)

		n := r.Intn(3) + 1
		addrs := app.AddTestAddrs(helper.App, helper.Ctx, n, sdk.NewInt(100000000))

		// add n new validators with zero voting power via MsgWrappedCreateValidator
		wcvMsgs := make([]*types.MsgWrappedCreateValidator, n)
		for i := 0; i < n; i++ {
			msg, err := buildMsgWrappedCreateValidatorWithAmount(addrs[i], sdk.DefaultPowerReduction.SubRaw(1))
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
			ctx = helper.GenAndApplyEmptyBlock(r)
		}
		epoch = ek.GetEpoch(ctx)
		require.Equal(t, uint64(2), epoch.EpochNumber)
		// ensure epoch 2 has initialised an empty msg queue
		require.Empty(t, ek.GetCurrentEpochMsgs(ctx))

		// ensure the length of current validator set equals to 1
		// since one genesis validator was added when setup
		// the rest n validators have zero voting power and thus are ruled out
		valSet = ck.GetValidatorSet(ctx, 2)
		require.Equal(t, 1, len(valSet))

		// ensure all validators (not just validators in the val set) have correct bond status
		// - the 1st validator is bonded
		// - all the rest are unbonded since they have zero voting power
		iterator := helper.StakingKeeper.ValidatorsPowerStoreIterator(ctx)
		defer iterator.Close()
		count := 0
		for ; iterator.Valid(); iterator.Next() {
			valAddr := sdk.ValAddress(iterator.Value())
			val, found := helper.StakingKeeper.GetValidator(ctx, valAddr)
			require.True(t, found)
			count++
			if count == 1 {
				require.Equal(t, stakingtypes.Bonded, val.Status)
			} else {
				require.Equal(t, stakingtypes.Unbonded, val.Status)
			}
		}
		require.Equal(t, len(wcvMsgs)+1, count)
	})
}

// FuzzWrappedCreateValidator_InsufficientBalance tests adding a new validator
// but the delegator has insufficient balance to perform delegating
func FuzzWrappedCreateValidator_InsufficientBalance(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 4)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		// a genesis validator is generate for setup
		helper := testepoching.NewHelper(t)
		ek := helper.EpochingKeeper
		ck := helper.App.CheckpointingKeeper
		msgServer := checkpointingkeeper.NewMsgServerImpl(ck)

		// BeginBlock of block 1, and thus entering epoch 1
		ctx := helper.BeginBlock(r)
		epoch := ek.GetEpoch(ctx)
		require.Equal(t, uint64(1), epoch.EpochNumber)

		n := r.Intn(3) + 1
		balance := r.Int63n(100)
		addrs := app.AddTestAddrs(helper.App, helper.Ctx, n, sdk.NewInt(balance))

		// add n new validators with value more than the delegator balance via MsgWrappedCreateValidator
		wcvMsgs := make([]*types.MsgWrappedCreateValidator, n)
		for i := 0; i < n; i++ {
			// make sure the value is more than the balance
			value := sdk.NewInt(balance).Add(sdk.NewInt(r.Int63n(100)))
			msg, err := buildMsgWrappedCreateValidatorWithAmount(addrs[i], value)
			require.NoError(t, err)
			wcvMsgs[i] = msg
			_, err = msgServer.WrappedCreateValidator(ctx, msg)
			require.ErrorIs(t, err, epochingtypes.ErrInsufficientBalance)
		}
	})
}

// FuzzWrappedCreateValidator tests adding new validators via
// MsgWrappedCreateValidator, which first registers BLS pubkey
// and then unwrapped into MsgCreateValidator and enqueued into
// the epoching module, and delivered to the staking module
// at epoch ends for execution
func FuzzWrappedCreateValidator(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 4)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		// a genesis validator is generate for setup
		helper := testepoching.NewHelper(t)
		ek := helper.EpochingKeeper
		ck := helper.App.CheckpointingKeeper
		msgServer := checkpointingkeeper.NewMsgServerImpl(ck)

		// BeginBlock of block 1, and thus entering epoch 1
		ctx := helper.BeginBlock(r)
		epoch := ek.GetEpoch(ctx)
		require.Equal(t, uint64(1), epoch.EpochNumber)

		// add n new validators via MsgWrappedCreateValidator
		n := r.Intn(3)
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
			ctx = helper.GenAndApplyEmptyBlock(r)
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

// FuzzAddBlsSig_NoError tests adding BLS signatures via MsgAddBlsSig
// it covers the following scenarios that would not cause errors:
// 1. a BLS signature is successfully accumulated and the checkpoint remains ACCUMULATING
// 2. a BLS signature is successfully accumulated and the checkpoint is changed to SEALED
// 3. a BLS signature is rejected if the checkpoint is not ACCUMULATING
func FuzzAddBlsSig_NoError(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 4)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		helper := testepoching.NewHelperWithValSet(t)
		ek := helper.EpochingKeeper
		ck := helper.App.CheckpointingKeeper
		msgServer := checkpointingkeeper.NewMsgServerImpl(ck)

		// BeginBlock of block 1, and thus entering epoch 1
		ctx := helper.BeginBlock(r)
		epoch := ek.GetEpoch(ctx)
		require.Equal(t, uint64(1), epoch.EpochNumber)

		// apply 2 blocks to ensure that a raw checkpoint for the previous epoch is built
		for i := uint64(0); i < 2; i++ {
			ctx = helper.GenAndApplyEmptyBlock(r)
		}
		endingEpoch := ek.GetEpoch(ctx).EpochNumber - 1
		_, err := ck.GetRawCheckpoint(ctx, endingEpoch)
		require.NoError(t, err)

		// add BLS signatures
		n := len(helper.ValBlsPrivKeys)
		totalPower := uint64(ck.GetTotalVotingPower(ctx, endingEpoch))
		for i := 0; i < n; i++ {
			lch := ctx.BlockHeader().LastCommitHash
			blsPrivKey := helper.ValBlsPrivKeys[i].BlsKey
			addr := helper.ValBlsPrivKeys[i].Address
			signBytes := types.GetSignBytes(endingEpoch, lch)
			blsSig := bls12381.Sign(blsPrivKey, signBytes)

			// create MsgAddBlsSig message
			msg := types.NewMsgAddBlsSig(sdk.AccAddress(addr), endingEpoch, lch, blsSig, addr)
			_, err = msgServer.AddBlsSig(ctx, msg)
			require.NoError(t, err)
			afterCkpt, err := ck.GetRawCheckpoint(ctx, endingEpoch)
			require.NoError(t, err)
			if afterCkpt.PowerSum <= totalPower/3 {
				require.True(t, afterCkpt.Status == types.Accumulating)
			} else {
				require.True(t, afterCkpt.Status == types.Sealed)
			}
		}
	})
}

// FuzzAddBlsSig_Error tests adding BLS signatures via MsgAddBlsSig
// in a scenario where the signer is not in the checkpoint's validator set
func FuzzAddBlsSig_NotInValSet(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 4)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		helper := testepoching.NewHelperWithValSet(t)
		ek := helper.EpochingKeeper
		ck := helper.App.CheckpointingKeeper
		msgServer := checkpointingkeeper.NewMsgServerImpl(ck)

		// BeginBlock of block 1, and thus entering epoch 1
		ctx := helper.BeginBlock(r)
		// apply 2 blocks to ensure that a raw checkpoint for the previous epoch is built
		for i := uint64(0); i < 2; i++ {
			ctx = helper.GenAndApplyEmptyBlock(r)
		}
		endingEpoch := ek.GetEpoch(ctx).EpochNumber - 1
		_, err := ck.GetRawCheckpoint(ctx, endingEpoch)
		require.NoError(t, err)

		// build BLS sig from a random validator (not in the validator set)
		lch := ctx.BlockHeader().LastCommitHash
		blsPrivKey := bls12381.GenPrivKey()
		valAddr := datagen.GenRandomValidatorAddress()
		signBytes := types.GetSignBytes(endingEpoch, lch)
		blsSig := bls12381.Sign(blsPrivKey, signBytes)
		msg := types.NewMsgAddBlsSig(sdk.AccAddress(valAddr), endingEpoch, lch, blsSig, valAddr)

		_, err = msgServer.AddBlsSig(ctx, msg)
		require.Error(t, err, types.ErrCkptDoesNotExist)
	})
}

// FuzzAddBlsSig_CkptNotExist tests adding BLS signatures via MsgAddBlsSig
// in a scenario where the corresponding checkpoint does not exist
func FuzzAddBlsSig_CkptNotExist(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 4)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		helper := testepoching.NewHelperWithValSet(t)
		ek := helper.EpochingKeeper
		ck := helper.App.CheckpointingKeeper
		msgServer := checkpointingkeeper.NewMsgServerImpl(ck)

		// BeginBlock of block 1, and thus entering epoch 1
		ctx := helper.BeginBlock(r)
		epoch := ek.GetEpoch(ctx)
		require.Equal(t, uint64(1), epoch.EpochNumber)

		// build BLS signature from a random validator of the validator set
		n := len(helper.ValBlsPrivKeys)
		i := r.Intn(n)
		lch := ctx.BlockHeader().LastCommitHash
		blsPrivKey := helper.ValBlsPrivKeys[i].BlsKey
		addr := helper.ValBlsPrivKeys[i].Address
		signBytes := types.GetSignBytes(epoch.EpochNumber-1, lch)
		blsSig := bls12381.Sign(blsPrivKey, signBytes)
		msg := types.NewMsgAddBlsSig(sdk.AccAddress(addr), epoch.EpochNumber-1, lch, blsSig, addr)

		// add the BLS signature
		_, err := msgServer.AddBlsSig(ctx, msg)
		require.Error(t, err, types.ErrCkptDoesNotExist)
	})
}

// FuzzAddBlsSig_WrongLastCommitHash tests adding BLS signatures via MsgAddBlsSig
// in a scenario where the signature is signed over wrong last_commit_hash
// 4. a BLS signature is rejected if the signature is invalid
func FuzzAddBlsSig_WrongLastCommitHash(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 4)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		helper := testepoching.NewHelperWithValSet(t)
		ek := helper.EpochingKeeper
		ck := helper.App.CheckpointingKeeper
		msgServer := checkpointingkeeper.NewMsgServerImpl(ck)

		// BeginBlock of block 1, and thus entering epoch 1
		ctx := helper.BeginBlock(r)
		epoch := ek.GetEpoch(ctx)
		require.Equal(t, uint64(1), epoch.EpochNumber)
		// apply 2 blocks to ensure that a raw checkpoint for the previous epoch is built
		for i := uint64(0); i < 2; i++ {
			ctx = helper.GenAndApplyEmptyBlock(r)
		}
		endingEpoch := ek.GetEpoch(ctx).EpochNumber - 1
		_, err := ck.GetRawCheckpoint(ctx, endingEpoch)
		require.NoError(t, err)

		// build BLS sig from a random validator
		n := len(helper.ValBlsPrivKeys)
		i := r.Intn(n)
		// inject random last commit hash
		lch := datagen.GenRandomLastCommitHash(r)
		blsPrivKey := helper.ValBlsPrivKeys[i].BlsKey
		addr := helper.ValBlsPrivKeys[i].Address
		signBytes := types.GetSignBytes(endingEpoch, lch)
		blsSig := bls12381.Sign(blsPrivKey, signBytes)
		msg := types.NewMsgAddBlsSig(sdk.AccAddress(addr), endingEpoch, lch, blsSig, addr)

		// add the BLS signature
		_, err = msgServer.AddBlsSig(ctx, msg)
		require.Error(t, err, types.ErrInvalidLastCommitHash)
	})
}

// FuzzAddBlsSig_InvalidSignature tests adding BLS signatures via MsgAddBlsSig
// in a scenario where the signature is invalid
func FuzzAddBlsSig_InvalidSignature(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 4)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		helper := testepoching.NewHelperWithValSet(t)
		ek := helper.EpochingKeeper
		ck := helper.App.CheckpointingKeeper
		msgServer := checkpointingkeeper.NewMsgServerImpl(ck)

		// BeginBlock of block 1, and thus entering epoch 1
		ctx := helper.BeginBlock(r)
		epoch := ek.GetEpoch(ctx)
		require.Equal(t, uint64(1), epoch.EpochNumber)
		// apply 2 blocks to ensure that a raw checkpoint for the previous epoch is built
		for i := uint64(0); i < 2; i++ {
			ctx = helper.GenAndApplyEmptyBlock(r)
		}
		endingEpoch := ek.GetEpoch(ctx).EpochNumber - 1
		_, err := ck.GetRawCheckpoint(ctx, endingEpoch)
		require.NoError(t, err)

		// build BLS sig from a random validator
		n := len(helper.ValBlsPrivKeys)
		i := r.Intn(n)
		// inject random last commit hash
		lch := ctx.BlockHeader().LastCommitHash
		addr := helper.ValBlsPrivKeys[i].Address
		blsSig := datagen.GenRandomBlsMultiSig(r)
		msg := types.NewMsgAddBlsSig(sdk.AccAddress(addr), endingEpoch, lch, blsSig, addr)

		// add the BLS signature message
		_, err = msgServer.AddBlsSig(ctx, msg)
		require.Error(t, err, types.ErrInvalidBlsSignature)
	})
}

func buildMsgWrappedCreateValidator(addr sdk.AccAddress) (*types.MsgWrappedCreateValidator, error) {
	bondTokens := sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction)
	return buildMsgWrappedCreateValidatorWithAmount(addr, bondTokens)
}

func buildMsgWrappedCreateValidatorWithAmount(addr sdk.AccAddress, bondTokens math.Int) (*types.MsgWrappedCreateValidator, error) {
	tmValPrivkey := ed25519.GenPrivKey()
	bondCoin := sdk.NewCoin(appparams.DefaultBondDenom, bondTokens)
	description := stakingtypes.NewDescription("foo_moniker", "", "", "", "")
	commission := stakingtypes.NewCommissionRates(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec())

	pk, err := cryptocodec.FromTmPubKeyInterface(tmValPrivkey.PubKey())
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
