package app

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"
)

func TestBabylonBlockedAddrs(t *testing.T) {
	encCfg := GetEncodingConfig()
	db := dbm.NewMemDB()
	signer, _ := SetupPrivSigner()
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	app := NewBabyblonAppWithCustomOptions(t, false, signer, SetupOptions{
		Logger:             logger,
		DB:                 db,
		InvCheckPeriod:     0,
		EncConfig:          encCfg,
		HomePath:           DefaultNodeHome,
		SkipUpgradeHeights: map[int64]bool{},
		AppOpts:            EmptyAppOptions{},
	})

	for acc := range maccPerms {
		require.True(
			t,
			app.BankKeeper.BlockedAddr(app.AccountKeeper.GetModuleAddress(acc)),
			"ensure that blocked addresses are properly set in bank keeper",
		)
	}

	app.Commit()

	logger2 := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	// Making a new app object with the db, so that initchain hasn't been called
	app2 := NewBabylonApp(logger2, db, nil, true, map[int64]bool{}, DefaultNodeHome, 0, encCfg, signer, EmptyAppOptions{})
	_, err := app2.ExportAppStateAndValidators(false, []string{})
	require.NoError(t, err, "ExportAppStateAndValidators should not have an error")
}

func TestGetMaccPerms(t *testing.T) {
	dup := GetMaccPerms()
	require.Equal(t, maccPerms, dup, "duplicated module account permissions differed from actual module account permissions")
}

func TestUpgradeStateOnGenesis(t *testing.T) {
	encCfg := GetEncodingConfig()
	db := dbm.NewMemDB()
	privSigner, err := SetupPrivSigner()
	require.NoError(t, err)
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))

	app := NewBabyblonAppWithCustomOptions(t, false, privSigner, SetupOptions{
		Logger:             logger,
		DB:                 db,
		InvCheckPeriod:     0,
		EncConfig:          encCfg,
		HomePath:           DefaultNodeHome,
		SkipUpgradeHeights: map[int64]bool{},
		AppOpts:            EmptyAppOptions{},
	})

	// make sure the upgrade keeper has version map in state
	ctx := app.NewContext(false, tmproto.Header{})
	vm := app.UpgradeKeeper.GetModuleVersionMap(ctx)
	for v, i := range app.mm.Modules {
		require.Equal(t, vm[v], i.ConsensusVersion())
	}
}
