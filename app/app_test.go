package app

import (
	"fmt"
	abci "github.com/cometbft/cometbft/abci/types"
	"testing"

	"cosmossdk.io/log"
	dbm "github.com/cosmos/cosmos-db"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/stretchr/testify/require"
)

func TestBabylonBlockedAddrs(t *testing.T) {
	db := dbm.NewMemDB()
	signer, _ := SetupPrivSigner()
	logger := log.NewTestLogger(t)

	app := NewBabylonAppWithCustomOptions(t, false, signer, SetupOptions{
		Logger:             logger,
		DB:                 db,
		InvCheckPeriod:     0,
		SkipUpgradeHeights: map[int64]bool{},
		AppOpts:            EmptyAppOptions{},
	})

	for acc := range BlockedAddresses() {
		var addr sdk.AccAddress
		if modAddr, err := sdk.AccAddressFromBech32(acc); err == nil {
			addr = modAddr
		} else {
			addr = app.AccountKeeper.GetModuleAddress(acc)
		}

		require.True(
			t,
			app.BankKeeper.BlockedAddr(addr),
			fmt.Sprintf("ensure that blocked addresses are properly set in bank keeper: %s should be blocked", acc),
		)
	}

	_, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 1,
	})
	require.NoError(t, err)
	_, err = app.Commit()
	require.NoError(t, err)

	logger2 := log.NewTestLogger(t)
	// Making a new app object with the db, so that initchain hasn't been called
	app2 := NewBabylonApp(
		logger2,
		db,
		nil,
		true,
		map[int64]bool{},
		0,
		signer,
		EmptyAppOptions{},
		EmptyWasmOpts,
	)
	_, err = app2.ExportAppStateAndValidators(false, []string{}, []string{})
	require.NoError(t, err, "ExportAppStateAndValidators should not have an error")
}

func TestGetMaccPerms(t *testing.T) {
	dup := GetMaccPerms()
	require.Equal(t, maccPerms, dup, "duplicated module account permissions differed from actual module account permissions")
}

func TestUpgradeStateOnGenesis(t *testing.T) {
	db := dbm.NewMemDB()
	privSigner, err := SetupPrivSigner()
	require.NoError(t, err)
	logger := log.NewTestLogger(t)

	app := NewBabylonAppWithCustomOptions(t, false, privSigner, SetupOptions{
		Logger:             logger,
		DB:                 db,
		InvCheckPeriod:     0,
		SkipUpgradeHeights: map[int64]bool{},
		AppOpts:            EmptyAppOptions{},
	})

	// make sure the upgrade keeper has version map in state
	ctx := app.NewContext(false)
	vm, err := app.UpgradeKeeper.GetModuleVersionMap(ctx)
	require.NoError(t, err)
	for v, i := range app.ModuleManager.Modules {
		if i, ok := i.(module.HasConsensusVersion); ok {
			require.Equal(t, vm[v], i.ConsensusVersion())
		}
	}
}
