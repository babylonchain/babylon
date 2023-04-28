package keeper_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/babylonchain/babylon/x/btccheckpoint/types"
	"github.com/stretchr/testify/require"
)

func TestKeeper_GetSubmissionBtcInfo(t *testing.T) {
	type TxKeyDesc struct {
		TxIdx uint32
		Depth uint64
	}

	type args struct {
		Key1 TxKeyDesc
		Key2 TxKeyDesc
	}

	tests := []struct {
		name                       string
		args                       args
		expectedYoungestBlockDepth uint64
		expectedTxIndex            uint32
		expectedOldestBlockDepth   uint64
	}{
		{"First header older. TxIndex larger in older header.", args{TxKeyDesc{TxIdx: 5, Depth: 10}, TxKeyDesc{TxIdx: 1, Depth: 0}}, 0, 1, 10},
		{"First header older. TxIndex larger in younger header.", args{TxKeyDesc{TxIdx: 1, Depth: 10}, TxKeyDesc{TxIdx: 5, Depth: 0}}, 0, 5, 10},
		{"Second header older. TxIndex larger in older header.", args{TxKeyDesc{TxIdx: 1, Depth: 0}, TxKeyDesc{TxIdx: 5, Depth: 10}}, 0, 1, 10},
		{"Second header older. TxIndex larger in younger header.", args{TxKeyDesc{TxIdx: 5, Depth: 0}, TxKeyDesc{TxIdx: 1, Depth: 10}}, 0, 5, 10},
		{"Same block. TxIndex larger in first transaction key.", args{TxKeyDesc{TxIdx: 5, Depth: 10}, TxKeyDesc{TxIdx: 1, Depth: 10}}, 10, 1, 10},
		{"Same block. TxIndex larger in second transaction key.", args{TxKeyDesc{TxIdx: 1, Depth: 10}, TxKeyDesc{TxIdx: 5, Depth: 10}}, 10, 1, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := rand.New(rand.NewSource(time.Now().Unix()))

			k := InitTestKeepers(t)

			hash1 := datagen.GenRandomBTCHeaderPrevBlock(r)
			hash2 := datagen.GenRandomBTCHeaderPrevBlock(r)

			sk := types.SubmissionKey{Key: []*types.TransactionKey{
				{Index: tt.args.Key1.TxIdx, Hash: hash1},
				{Index: tt.args.Key2.TxIdx, Hash: hash2},
			}}

			k.BTCLightClient.SetDepth(hash1, int64(tt.args.Key1.Depth))
			k.BTCLightClient.SetDepth(hash2, int64(tt.args.Key2.Depth))

			info, err := k.BTCCheckpoint.GetSubmissionBtcInfo(k.SdkCtx, sk)

			require.NoError(t, err)

			require.Equal(t, info.YoungestBlockDepth, tt.expectedYoungestBlockDepth, tt.name)
			require.Equal(t, info.LatestTxIndex, tt.expectedTxIndex, tt.name)
			require.Equal(t, info.OldestBlockDepth, tt.expectedOldestBlockDepth, tt.name)
		})
	}
}

func FuzzGetSubmissionBtcInfo(f *testing.F) {
	f.Add(int64(1), uint32(0), uint32(1), uint32(1), uint32(1))

	f.Fuzz(func(t *testing.T, seed int64, depth1 uint32, txidx1 uint32, depth2 uint32, txidx2 uint32) {
		r := rand.New(rand.NewSource(seed))

		if txidx1 == txidx2 {
			// transaction indexes must be different to cover the case where transactions are
			// in the same block (then they cannot have same indexes)
			t.Skip()
		}

		k := InitTestKeepers(t)

		hash1 := datagen.GenRandomBTCHeaderPrevBlock(r)
		hash2 := datagen.GenRandomBTCHeaderPrevBlock(r)

		sk := types.SubmissionKey{Key: []*types.TransactionKey{
			{Index: txidx1, Hash: hash1},
			{Index: txidx2, Hash: hash2},
		}}

		k.BTCLightClient.SetDepth(hash1, int64(depth1))
		k.BTCLightClient.SetDepth(hash2, int64(depth2))

		info, err := k.BTCCheckpoint.GetSubmissionBtcInfo(k.SdkCtx, sk)
		require.NoError(t, err)

		var expectedOldestDepth uint64
		var expectedYoungestDepth uint64
		var expectedTxIdx uint32

		if depth1 > depth2 {
			expectedOldestDepth = uint64(depth1)
			expectedYoungestDepth = uint64(depth2)
			expectedTxIdx = txidx2
		} else if depth1 < depth2 {
			expectedOldestDepth = uint64(depth2)
			expectedYoungestDepth = uint64(depth1)
			expectedTxIdx = txidx1
		} else {
			if txidx1 > txidx2 {
				expectedTxIdx = txidx2
			} else {
				expectedTxIdx = txidx1
			}
			expectedOldestDepth = uint64(depth1)
			expectedYoungestDepth = uint64(depth1)
		}

		require.Equal(t, info.YoungestBlockDepth, expectedYoungestDepth)
		require.Equal(t, info.LatestTxIndex, expectedTxIdx)
		require.Equal(t, info.OldestBlockDepth, expectedOldestDepth)
	})
}
