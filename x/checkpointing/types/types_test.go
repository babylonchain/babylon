package types_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/stretchr/testify/require"

	"github.com/babylonchain/babylon/testutil/datagen"
	testkeeper "github.com/babylonchain/babylon/testutil/keeper"
	"github.com/babylonchain/babylon/x/checkpointing/types"
)

// a single validator
func TestRawCheckpointWithMeta_Accumulate1(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	epochNum := uint64(2)
	n := 1
	totalPower := int64(10)
	ckptkeeper, ctx, _ := testkeeper.CheckpointingKeeper(t, nil, nil, client.Context{})
	lch := datagen.GenRandomAppHash(r)
	msg := types.GetSignBytes(epochNum, lch)
	blsPubkeys, blsSigs := datagen.GenRandomPubkeysAndSigs(n, msg)
	ckpt, err := ckptkeeper.BuildRawCheckpoint(ctx, epochNum, lch)
	require.NoError(t, err)
	valSet := datagen.GenRandomValSet(n)
	err = ckpt.Accumulate(valSet, valSet[0].Addr, blsPubkeys[0], blsSigs[0], totalPower)
	require.NoError(t, err)
	require.Equal(t, types.Sealed, ckpt.Status)

	// accumulate the same BLS sig
	err = ckpt.Accumulate(valSet, valSet[0].Addr, blsPubkeys[0], blsSigs[0], totalPower)
	require.ErrorIs(t, err, types.ErrCkptNotAccumulating)
	require.Equal(t, types.Sealed, ckpt.Status)
}

// 4 validators
func TestRawCheckpointWithMeta_Accumulate4(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	epochNum := uint64(2)
	n := 4
	totalPower := int64(10) * int64(n)
	ckptkeeper, ctx, _ := testkeeper.CheckpointingKeeper(t, nil, nil, client.Context{})
	lch := datagen.GenRandomAppHash(r)
	msg := types.GetSignBytes(epochNum, lch)
	blsPubkeys, blsSigs := datagen.GenRandomPubkeysAndSigs(n, msg)
	ckpt, err := ckptkeeper.BuildRawCheckpoint(ctx, epochNum, lch)
	require.NoError(t, err)
	valSet := datagen.GenRandomValSet(n)
	for i := 0; i < n; i++ {
		err = ckpt.Accumulate(valSet, valSet[i].Addr, blsPubkeys[i], blsSigs[i], totalPower)
		if i == 0 {
			require.NoError(t, err)
			require.Equal(t, types.Accumulating, ckpt.Status)
		}
		if i == 1 {
			require.NoError(t, err)
			require.Equal(t, types.Sealed, ckpt.Status)
		}
		if i >= 2 {
			require.ErrorIs(t, err, types.ErrCkptNotAccumulating)
			require.Equal(t, types.Sealed, ckpt.Status)
		}
	}
}
