package btclightclient_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/babylonchain/babylon/testutil/datagen"
	thelper "github.com/babylonchain/babylon/testutil/helper"
	keepertest "github.com/babylonchain/babylon/testutil/keeper"
	"github.com/babylonchain/babylon/testutil/nullify"
	"github.com/babylonchain/babylon/x/btclightclient"
	"github.com/babylonchain/babylon/x/btclightclient/keeper"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestGenesis(t *testing.T) {
	baseHeaderInfo := types.SimnetGenesisBlock()
	genesisState := types.GenesisState{
		BtcHeaders: []types.BTCHeaderInfo{baseHeaderInfo},
	}

	k, ctx := keepertest.BTCLightClientKeeper(t)
	btclightclient.InitGenesis(ctx, *k, genesisState)
	got := btclightclient.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)
}

func TestImportExport(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	sender1 := secp256k1.GenPrivKey()
	address1, err := sdk.AccAddressFromHexUnsafe(sender1.PubKey().Address().String())
	require.NoError(t, err)
	sender2 := secp256k1.GenPrivKey()
	address2, err := sdk.AccAddressFromHexUnsafe(sender2.PubKey().Address().String())
	require.NoError(t, err)

	params := types.NewParams(
		// only sender1 and sender2 are allowed to update
		[]string{address1.String(), address2.String()},
	)

	k, ctx, stServ := keepertest.BTCLightClientKeeperWithCustomParams(t, params)
	srv := keeper.NewMsgServerImpl(*k)

	_, chain := keepertest.BTCLightGenRandomChain(t, r, k, ctx, 0, 10)
	initTip := chain.GetTipInfo()

	chainExtension := datagen.GenRandomValidChainStartingFrom(
		r,
		initTip.Height,
		initTip.Header.ToBlockHeader(),
		nil,
		10,
	)

	// sender 1 is allowed to update, it should succeed
	msg := &types.MsgInsertHeaders{Signer: address1.String(), Headers: keepertest.ChainToChainBytes(chainExtension)}
	_, err = srv.InsertHeaders(ctx, msg)
	require.NoError(t, err)

	newTip := k.GetTipInfo(ctx)
	require.NotNil(t, newTip)

	newChainExt := datagen.GenRandomValidChainStartingFrom(
		r,
		newTip.Height,
		newTip.Header.ToBlockHeader(),
		nil,
		10,
	)

	msg1 := &types.MsgInsertHeaders{Signer: address2.String(), Headers: keepertest.ChainToChainBytes(newChainExt)}
	_, err = srv.InsertHeaders(ctx, msg1)
	require.NoError(t, err)

	genState := btclightclient.ExportGenesis(ctx, *k)
	KvA := stServ.OpenKVStore(ctx)

	kB, ctxb, stServB := keepertest.BTCLightClientKeeperWithCustomParams(t, params)
	btclightclient.InitGenesis(ctxb, *kB, *genState)

	infos := kB.GetAllHeaderInfos(ctxb)
	require.Equal(t, len(infos), len(genState.BtcHeaders), "it should have the same amount of headers from before")

	KvB := stServB.OpenKVStore(ctxb)

	failedKVAs, failedKVBs := thelper.DiffKVStores(KvA, KvB, [][]byte{})
	require.Equal(t, len(failedKVAs), len(failedKVBs), "unequal sets of key-values to compare btcligthclient")
	require.Equal(t, len(failedKVAs), 0, "should not exist any difference froms states.")
}
