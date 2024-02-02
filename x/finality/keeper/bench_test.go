package keeper_test

import (
	"fmt"
	"math/rand"
	"os"
	"runtime/pprof"
	"testing"
	"time"

	"github.com/babylonchain/babylon/testutil/datagen"
	keepertest "github.com/babylonchain/babylon/testutil/keeper"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/finality/keeper"
	"github.com/babylonchain/babylon/x/finality/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func benchmarkCommitPubRandList(b *testing.B, numPubRand uint64) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	bsKeeper := types.NewMockBTCStakingKeeper(ctrl)
	fKeeper, ctx := keepertest.FinalityKeeper(b, bsKeeper, nil)
	ms := keeper.NewMsgServerImpl(*fKeeper)

	// create a random finality provider
	btcSK, btcPK, err := datagen.GenRandomBTCKeyPair(r)
	require.NoError(b, err)
	fpBTCPK := bbn.NewBIP340PubKeyFromBTCPK(btcPK)
	fpBTCPKBytes := fpBTCPK.MustMarshal()

	// register the finality provider
	bsKeeper.EXPECT().HasFinalityProvider(gomock.Any(), gomock.Eq(fpBTCPKBytes)).Return(true).AnyTimes()

	// Start the CPU profiler
	cpuProfileFile := fmt.Sprintf("/tmp/finality-commit-pub-rand-%d-cpu.pprof", numPubRand)
	f, err := os.Create(cpuProfileFile)
	if err != nil {
		b.Fatal(err)
	}
	defer f.Close()
	if err := pprof.StartCPUProfile(f); err != nil {
		b.Fatal(err)
	}
	defer pprof.StopCPUProfile()

	// Reset timer before the benchmark loop starts
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer() // Stop the timer to exclude measurement on GenRandomMsgCommitPubRandList

		startHeight := 1 + numPubRand*uint64(i)
		_, msg, err := datagen.GenRandomMsgCommitPubRandList(r, btcSK, startHeight, numPubRand)
		require.NoError(b, err)

		b.StartTimer() // Start the timer again to measure CommitPubRandList

		_, err = ms.CommitPubRandList(ctx, msg)
		require.NoError(b, err)
	}
}

func BenchmarkCommitPubRandList_100(b *testing.B)   { benchmarkCommitPubRandList(b, 100) }
func BenchmarkCommitPubRandList_1000(b *testing.B)  { benchmarkCommitPubRandList(b, 1000) }
func BenchmarkCommitPubRandList_10000(b *testing.B) { benchmarkCommitPubRandList(b, 10000) }
