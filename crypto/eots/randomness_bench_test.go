package eots_test

import (
	"math/rand"
	"os"
	"runtime/pprof"
	"testing"
	"time"

	"github.com/babylonchain/babylon/crypto/eots"
	"github.com/stretchr/testify/require"
)

func BenchmarkDeriveRandomness(b *testing.B) {
	r := rand.New(rand.NewSource(time.Now().Unix()))

	// master randomness pair
	msr, mpr, err := eots.NewMasterRandPair(r)
	require.NoError(b, err)
	require.NoError(b, msr.Validate())
	require.NoError(b, mpr.Validate())

	// Start the CPU profiler
	cpuProfileFile := "/tmp/bench_verify_derive_randomness.pprof"
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
		_, _, err := msr.DeriveRandPair(uint32(i))
		require.NoError(b, err)
	}
}

func BenchmarkUnmarshalMasterPubRand(b *testing.B) {
	r := rand.New(rand.NewSource(time.Now().Unix()))

	// master randomness pair
	msr, mpr, err := eots.NewMasterRandPair(r)
	require.NoError(b, err)
	require.NoError(b, msr.Validate())
	require.NoError(b, mpr.Validate())

	mprStr := mpr.MarshalBase58()

	// Start the CPU profiler
	cpuProfileFile := "/tmp/bench_unmarshal_pubrand.pprof"
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
		_, err := eots.NewMasterPublicRandFromBase58(mprStr)
		require.NoError(b, err)
	}
}
