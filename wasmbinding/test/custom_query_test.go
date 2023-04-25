package wasmbinding

import (
	"encoding/json"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	"github.com/babylonchain/babylon/app"
	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/babylonchain/babylon/wasmbinding/bindings"
	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/ed25519"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/require"
)

// TODO consider doing it by enviromental variables as currently it may fail on some
// weird architectures
func getArtifactPath() string {
	if runtime.GOARCH == "amd64" {
		return "../testdata/artifacts/testdata.wasm"
	} else if runtime.GOARCH == "arm64" {
		return "../testdata/artifacts/testdata-aarch64.wasm"
	} else {
		panic("Unsupported architecture")
	}
}

var pathToContract = getArtifactPath()

func TestQueryEpoch(t *testing.T) {
	acc := randomAccountAddress()
	babylonApp, ctx := setupAppWithContext(t)
	fundAccount(t, ctx, babylonApp, acc)

	contractAddress := deployTestContract(t, ctx, babylonApp, acc, pathToContract)

	query := bindings.BabylonQuery{
		Epoch: &struct{}{},
	}
	resp := bindings.CurrentEpochResponse{}
	queryCustom(t, ctx, babylonApp, contractAddress, query, &resp)
	require.Equal(t, resp.Epoch, uint64(0))

	newEpoch := babylonApp.EpochingKeeper.IncEpoch(ctx)

	resp = bindings.CurrentEpochResponse{}
	queryCustom(t, ctx, babylonApp, contractAddress, query, &resp)
	require.Equal(t, resp.Epoch, newEpoch.EpochNumber)
}

func TestFinalizedEpoch(t *testing.T) {
	acc := randomAccountAddress()
	babylonApp, ctx := setupAppWithContext(t)
	fundAccount(t, ctx, babylonApp, acc)

	// babylonApp.ZoneConciergeKeeper
	contractAddress := deployTestContract(t, ctx, babylonApp, acc, pathToContract)

	query := bindings.BabylonQuery{
		LatestFinalizedEpochInfo: &struct{}{},
	}

	// There is no finalized epoch yet so we require an error
	queryCustomErr(t, ctx, babylonApp, contractAddress, query)

	epoch := babylonApp.EpochingKeeper.IncEpoch(ctx)

	_ = babylonApp.ZoneConciergeKeeper.Hooks().AfterRawCheckpointFinalized(ctx, epoch.EpochNumber)

	resp := bindings.LatestFinalizedEpochInfoResponse{}
	queryCustom(t, ctx, babylonApp, contractAddress, query, &resp)
	require.Equal(t, resp.EpochInfo.EpochNumber, epoch.EpochNumber)
	require.Equal(t, resp.EpochInfo.LastBlockHeight, epoch.GetLastBlockHeight())
}

func TestQueryBtcTip(t *testing.T) {
	acc := randomAccountAddress()
	babylonApp, ctx := setupAppWithContext(t)
	fundAccount(t, ctx, babylonApp, acc)

	contractAddress := deployTestContract(t, ctx, babylonApp, acc, pathToContract)

	query := bindings.BabylonQuery{
		BtcTip: &struct{}{},
	}

	resp := bindings.BtcTipResponse{}
	queryCustom(t, ctx, babylonApp, contractAddress, query, &resp)

	tip := babylonApp.BTCLightClientKeeper.GetTipInfo(ctx)
	tipAsInfo := bindings.AsBtcBlockHeaderInfo(tip)

	require.Equal(t, resp.HeaderInfo.Height, tip.Height)
	require.Equal(t, tipAsInfo, resp.HeaderInfo)
}

func TestQueryBtcBase(t *testing.T) {
	acc := randomAccountAddress()
	babylonApp, ctx := setupAppWithContext(t)
	fundAccount(t, ctx, babylonApp, acc)

	contractAddress := deployTestContract(t, ctx, babylonApp, acc, pathToContract)

	query := bindings.BabylonQuery{
		BtcBaseHeader: &struct{}{},
	}

	resp := bindings.BtcBaseHeaderResponse{}
	queryCustom(t, ctx, babylonApp, contractAddress, query, &resp)

	base := babylonApp.BTCLightClientKeeper.GetBaseBTCHeader(ctx)
	baseAsInfo := bindings.AsBtcBlockHeaderInfo(base)

	require.Equal(t, baseAsInfo, resp.HeaderInfo)
}

func TestQueryBtcByHash(t *testing.T) {
	acc := randomAccountAddress()
	babylonApp, ctx := setupAppWithContext(t)
	fundAccount(t, ctx, babylonApp, acc)

	contractAddress := deployTestContract(t, ctx, babylonApp, acc, pathToContract)
	tip := babylonApp.BTCLightClientKeeper.GetTipInfo(ctx)

	query := bindings.BabylonQuery{
		BtcHeaderByHash: &bindings.BtcHeaderByHash{
			Hash: tip.Hash.String(),
		},
	}

	headerAsInfo := bindings.AsBtcBlockHeaderInfo(tip)
	resp := bindings.BtcHeaderQueryResponse{}
	queryCustom(t, ctx, babylonApp, contractAddress, query, &resp)

	require.Equal(t, resp.HeaderInfo, headerAsInfo)
}

func TestQueryBtcByNumber(t *testing.T) {
	acc := randomAccountAddress()
	babylonApp, ctx := setupAppWithContext(t)
	fundAccount(t, ctx, babylonApp, acc)

	contractAddress := deployTestContract(t, ctx, babylonApp, acc, pathToContract)
	tip := babylonApp.BTCLightClientKeeper.GetTipInfo(ctx)

	query := bindings.BabylonQuery{
		BtcHeaderByHeight: &bindings.BtcHeaderByHeight{
			Height: tip.Height,
		},
	}

	headerAsInfo := bindings.AsBtcBlockHeaderInfo(tip)
	resp := bindings.BtcHeaderQueryResponse{}
	queryCustom(t, ctx, babylonApp, contractAddress, query, &resp)

	require.Equal(t, resp.HeaderInfo, headerAsInfo)
}

func TestQueryNonExistingHeader(t *testing.T) {
	acc := randomAccountAddress()
	babylonApp, ctx := setupAppWithContext(t)
	fundAccount(t, ctx, babylonApp, acc)

	contractAddress := deployTestContract(t, ctx, babylonApp, acc, pathToContract)

	queryNonExisitingHeight := bindings.BabylonQuery{
		BtcHeaderByHeight: &bindings.BtcHeaderByHeight{
			Height: 1,
		},
	}
	resp := bindings.BtcHeaderQueryResponse{}
	queryCustom(t, ctx, babylonApp, contractAddress, queryNonExisitingHeight, &resp)
	require.Nil(t, resp.HeaderInfo)

	queryNonExisitingHash := bindings.BabylonQuery{
		BtcHeaderByHash: &bindings.BtcHeaderByHash{
			Hash: datagen.GenRandomBtcdHash().String(),
		},
	}
	resp1 := bindings.BtcHeaderQueryResponse{}
	queryCustom(t, ctx, babylonApp, contractAddress, queryNonExisitingHash, &resp1)
	require.Nil(t, resp1.HeaderInfo)
}

func setupAppWithContext(t *testing.T) (*app.BabylonApp, sdk.Context) {
	return setupAppWithContextAndCustomHeight(t, 1)
}

func setupAppWithContextAndCustomHeight(t *testing.T, height int64) (*app.BabylonApp, sdk.Context) {
	babylonApp := app.Setup(t, false)
	ctx := babylonApp.BaseApp.NewContext(false, tmproto.Header{Height: height, Time: time.Now().UTC()})
	return babylonApp, ctx
}

func keyPubAddr() (crypto.PrivKey, crypto.PubKey, sdk.AccAddress) {
	key := ed25519.GenPrivKey()
	pub := key.PubKey()
	addr := sdk.AccAddress(pub.Address())
	return key, pub, addr
}

func randomAccountAddress() sdk.AccAddress {
	_, _, addr := keyPubAddr()
	return addr
}

func mintCoinsTo(
	bankKeeper bankkeeper.Keeper,
	ctx sdk.Context,
	addr sdk.AccAddress,
	amounts sdk.Coins) error {
	if err := bankKeeper.MintCoins(ctx, minttypes.ModuleName, amounts); err != nil {
		return err
	}

	return bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addr, amounts)
}

func fundAccount(
	t *testing.T,
	ctx sdk.Context,
	bbn *app.BabylonApp,
	acc sdk.AccAddress) {

	err := mintCoinsTo(bbn.BankKeeper, ctx, acc, sdk.NewCoins(
		sdk.NewCoin("ubbn", sdk.NewInt(10000000000)),
	))
	require.NoError(t, err)
}

func storeTestCodeCode(
	t *testing.T,
	ctx sdk.Context,
	babylonApp *app.BabylonApp,
	addr sdk.AccAddress,
	codePath string,
) (uint64, []byte) {
	wasmCode, err := os.ReadFile(codePath)

	require.NoError(t, err)
	permKeeper := keeper.NewPermissionedKeeper(babylonApp.WasmKeeper, keeper.DefaultAuthorizationPolicy{})
	id, checksum, err := permKeeper.Create(ctx, addr, wasmCode, nil)
	require.NoError(t, err)
	return id, checksum
}

func instantiateExampleContract(
	t *testing.T,
	ctx sdk.Context,
	bbn *app.BabylonApp,
	funder sdk.AccAddress,
	codeId uint64,
) sdk.AccAddress {
	initMsgBz := []byte("{}")
	contractKeeper := keeper.NewDefaultPermissionKeeper(bbn.WasmKeeper)
	addr, _, err := contractKeeper.Instantiate(ctx, codeId, funder, funder, initMsgBz, "demo contract", nil)
	require.NoError(t, err)
	return addr
}

func deployTestContract(
	t *testing.T,
	ctx sdk.Context,
	bbn *app.BabylonApp,
	deployer sdk.AccAddress,
	codePath string,
) sdk.AccAddress {

	codeId, _ := storeTestCodeCode(t, ctx, bbn, deployer, codePath)

	contractAddr := instantiateExampleContract(t, ctx, bbn, deployer, codeId)

	return contractAddr
}

type ExampleQuery struct {
	Chain *ChainRequest `json:"chain,omitempty"`
}

type ChainRequest struct {
	Request wasmvmtypes.QueryRequest `json:"request"`
}

type ChainResponse struct {
	Data []byte `json:"data"`
}

func queryCustom(
	t *testing.T,
	ctx sdk.Context,
	bbn *app.BabylonApp,
	contract sdk.AccAddress,
	request bindings.BabylonQuery,
	response interface{},
) {
	msgBz, err := json.Marshal(request)
	require.NoError(t, err)

	query := ExampleQuery{
		Chain: &ChainRequest{
			Request: wasmvmtypes.QueryRequest{Custom: msgBz},
		},
	}
	queryBz, err := json.Marshal(query)
	require.NoError(t, err)

	resBz, err := bbn.WasmKeeper.QuerySmart(ctx, contract, queryBz)
	require.NoError(t, err)
	var resp ChainResponse
	err = json.Unmarshal(resBz, &resp)
	require.NoError(t, err)
	err = json.Unmarshal(resp.Data, response)
	require.NoError(t, err)
}

func queryCustomErr(
	t *testing.T,
	ctx sdk.Context,
	bbn *app.BabylonApp,
	contract sdk.AccAddress,
	request bindings.BabylonQuery,
) {
	msgBz, err := json.Marshal(request)
	require.NoError(t, err)

	query := ExampleQuery{
		Chain: &ChainRequest{
			Request: wasmvmtypes.QueryRequest{Custom: msgBz},
		},
	}
	queryBz, err := json.Marshal(query)
	require.NoError(t, err)

	_, err = bbn.WasmKeeper.QuerySmart(ctx, contract, queryBz)
	require.Error(t, err)
}
