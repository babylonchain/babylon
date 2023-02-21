package simapp

import (
	"fmt"
	"os"
	"testing"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdksimapp "github.com/cosmos/cosmos-sdk/simapp"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/babylonchain/babylon/app"
)

// Profile with:
// /usr/local/go/bin/go test -benchmem -run=^$ github.com/babylonchain/babylon/simapp -bench ^BenchmarkFullAppSimulation$ -Commit=true -cpuprofile cpu.out
func BenchmarkFullAppSimulation(b *testing.B) {
	b.ReportAllocs()
	config, db, dir, logger, skip, err := SetupSimulation("goleveldb-app-sim", "Simulation")
	if err != nil {
		b.Fatalf("simulation setup failed: %s", err.Error())
	}

	if skip {
		b.Skip("skipping benchmark application simulation")
	}

	defer func() {
		db.Close()
		err = os.RemoveAll(dir)
		if err != nil {
			b.Fatal(err)
		}
	}()

	privSigner, err := app.SetupPrivSigner()
	if err != nil {
		b.Fatal(err)
	}
	babylon := app.NewBabylonApp(logger, db, nil, true, map[int64]bool{}, app.DefaultNodeHome, FlagPeriodValue, app.GetEncodingConfig(), privSigner, sdksimapp.EmptyAppOptions{}, interBlockCacheOpt())

	// run randomized simulation
	_, simParams, simErr := simulation.SimulateFromSeed(
		b,
		os.Stdout,
		babylon.BaseApp,
		AppStateFn(babylon.AppCodec(), babylon.SimulationManager()),
		simtypes.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
		SimulationOperations(babylon, babylon.AppCodec(), config),
		babylon.ModuleAccountAddrs(),
		config,
		babylon.AppCodec(),
	)

	// export state and simParams before the simulation error is checked
	if err = sdksimapp.CheckExportSimulation(babylon, config, simParams); err != nil {
		b.Fatal(err)
	}

	if simErr != nil {
		b.Fatal(simErr)
	}

	if config.Commit {
		sdksimapp.PrintStats(db)
	}
}

func BenchmarkInvariants(b *testing.B) {
	b.ReportAllocs()
	config, db, dir, logger, skip, err := SetupSimulation("leveldb-app-invariant-bench", "Simulation")
	if err != nil {
		b.Fatalf("simulation setup failed: %s", err.Error())
	}

	if skip {
		b.Skip("skipping benchmark application simulation")
	}

	config.AllInvariants = false

	defer func() {
		db.Close()
		err = os.RemoveAll(dir)
		if err != nil {
			b.Fatal(err)
		}
	}()

	privSigner, err := app.SetupPrivSigner()
	if err != nil {
		b.Fatal(err)
	}
	babylon := app.NewBabylonApp(logger, db, nil, true, map[int64]bool{}, app.DefaultNodeHome, FlagPeriodValue, app.GetEncodingConfig(), privSigner, sdksimapp.EmptyAppOptions{}, interBlockCacheOpt())

	// run randomized simulation
	_, simParams, simErr := simulation.SimulateFromSeed(
		b,
		os.Stdout,
		babylon.BaseApp,
		AppStateFn(babylon.AppCodec(), babylon.SimulationManager()),
		simtypes.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
		SimulationOperations(babylon, babylon.AppCodec(), config),
		babylon.ModuleAccountAddrs(),
		config,
		babylon.AppCodec(),
	)

	// export state and simParams before the simulation error is checked
	if err = sdksimapp.CheckExportSimulation(babylon, config, simParams); err != nil {
		b.Fatal(err)
	}

	if simErr != nil {
		b.Fatal(simErr)
	}

	if config.Commit {
		sdksimapp.PrintStats(db)
	}

	ctx := babylon.NewContext(true, tmproto.Header{Height: babylon.LastBlockHeight() + 1})

	// 3. Benchmark each invariant separately
	//
	// NOTE: We use the crisis keeper as it has all the invariants registered with
	// their respective metadata which makes it useful for testing/benchmarking.
	for _, cr := range babylon.CrisisKeeper.Routes() {
		cr := cr
		b.Run(fmt.Sprintf("%s/%s", cr.ModuleName, cr.Route), func(b *testing.B) {
			if res, stop := cr.Invar(ctx); stop {
				b.Fatalf(
					"broken invariant at block %d of %d\n%s",
					ctx.BlockHeight()-1, config.NumBlocks, res,
				)
			}
		})
	}
}
