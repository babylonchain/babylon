package app

import (
	"encoding/json"
	"math/rand"
	"os"
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	pruningtypes "cosmossdk.io/store/pruning/types"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/ed25519"
	tmjson "github.com/cometbft/cometbft/libs/json"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cosmosed "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/types"
	simsutils "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/mempool"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/docker/docker/pkg/ioutils"
	"github.com/stretchr/testify/require"

	appparams "github.com/babylonchain/babylon/app/params"
	"github.com/babylonchain/babylon/crypto/bls12381"
	"github.com/babylonchain/babylon/privval"
	bbn "github.com/babylonchain/babylon/types"
	checkpointingtypes "github.com/babylonchain/babylon/x/checkpointing/types"
)

// SetupOptions defines arguments that are passed into `Simapp` constructor.
type SetupOptions struct {
	Logger             log.Logger
	DB                 *dbm.MemDB
	InvCheckPeriod     uint
	SkipUpgradeHeights map[int64]bool
	AppOpts            types.AppOptions
}

func setup(t *testing.T, ps *PrivSigner, withGenesis bool, invCheckPeriod uint) (*BabylonApp, GenesisState) {
	db := dbm.NewMemDB()
	nodeHome := t.TempDir()

	appOptions := make(simsutils.AppOptionsMap, 0)
	appOptions[flags.FlagHome] = nodeHome // ensure unique folder
	appOptions[server.FlagInvCheckPeriod] = invCheckPeriod
	appOptions["btc-config.network"] = string(bbn.BtcSimnet)
	appOptions[server.FlagPruning] = pruningtypes.PruningOptionDefault
	appOptions[server.FlagMempoolMaxTxs] = mempool.DefaultMaxTx
	appOptions[flags.FlagChainID] = "chain-test"
	baseAppOpts := server.DefaultBaseappOptions(appOptions)
	app := NewBabylonApp(
		log.NewNopLogger(),
		db,
		nil,
		true,
		map[int64]bool{},
		invCheckPeriod,
		ps,
		appOptions,
		EmptyWasmOpts,
		baseAppOpts...,
	)
	if withGenesis {
		return app, app.DefaultGenesis()
	}
	return app, GenesisState{}
}

// NewBabylonAppWithCustomOptions initializes a new BabylonApp with custom options.
// Created Babylon application will have one validator with hardcoed amount of tokens.
// This is necessary as from cosmos-sdk 0.46 it is required that there is at least
// one validator in validator set during InitGenesis abci call - https://github.com/cosmos/cosmos-sdk/pull/9697
func NewBabylonAppWithCustomOptions(t *testing.T, isCheckTx bool, privSigner *PrivSigner, options SetupOptions) *BabylonApp {
	t.Helper()
	// create validator set with single validator
	valKeys, err := privval.NewValidatorKeys(ed25519.GenPrivKey(), bls12381.GenPrivKey())
	require.NoError(t, err)
	valPubkey, err := cryptocodec.FromCmtPubKeyInterface(valKeys.ValPubkey)
	require.NoError(t, err)
	genesisKey, err := checkpointingtypes.NewGenesisKey(
		sdk.ValAddress(valKeys.ValPubkey.Address()),
		&valKeys.BlsPubkey,
		valKeys.PoP,
		&cosmosed.PubKey{Key: valPubkey.Bytes()},
	)
	require.NoError(t, err)
	genesisValSet := []*checkpointingtypes.GenesisKey{genesisKey}

	acc := authtypes.NewBaseAccount(valPubkey.Address().Bytes(), valPubkey, 0, 0)
	balance := banktypes.Balance{
		Address: acc.GetAddress().String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(appparams.DefaultBondDenom, math.NewInt(100000000000000))),
	}

	app := NewBabylonApp(
		options.Logger,
		options.DB,
		nil,
		true,
		options.SkipUpgradeHeights,
		options.InvCheckPeriod,
		privSigner,
		options.AppOpts,
		EmptyWasmOpts,
	)
	genesisState := app.DefaultGenesis()
	genesisState = genesisStateWithValSet(t, app, genesisState, genesisValSet, []authtypes.GenesisAccount{acc}, balance)

	if !isCheckTx {
		// init chain must be called to stop deliverState from being nil
		stateBytes, err := tmjson.MarshalIndent(genesisState, "", " ")
		require.NoError(t, err)

		// Initialize the chain
		consensusParams := simsutils.DefaultConsensusParams
		initialHeight := app.LastBlockHeight() + 1
		consensusParams.Abci = &cmtproto.ABCIParams{VoteExtensionsEnableHeight: initialHeight}
		_, err = app.InitChain(
			&abci.RequestInitChain{
				Validators:      []abci.ValidatorUpdate{},
				ConsensusParams: consensusParams,
				AppStateBytes:   stateBytes,
				InitialHeight:   initialHeight,
			},
		)
		require.NoError(t, err)
	}

	return app
}

func genesisStateWithValSet(t *testing.T,
	app *BabylonApp, genesisState GenesisState,
	valSet []*checkpointingtypes.GenesisKey, genAccs []authtypes.GenesisAccount,
	balances ...banktypes.Balance,
) GenesisState {
	// set genesis accounts
	authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), genAccs)
	genesisState[authtypes.ModuleName] = app.AppCodec().MustMarshalJSON(authGenesis)

	validators := make([]stakingtypes.Validator, 0, len(valSet))
	delegations := make([]stakingtypes.Delegation, 0, len(valSet))

	bondAmt := sdk.DefaultPowerReduction.MulRaw(1000)

	for _, valGenKey := range valSet {
		pkAny, err := codectypes.NewAnyWithValue(valGenKey.ValPubkey)
		require.NoError(t, err)
		validator := stakingtypes.Validator{
			OperatorAddress:   valGenKey.ValidatorAddress,
			ConsensusPubkey:   pkAny,
			Jailed:            false,
			Status:            stakingtypes.Bonded,
			Tokens:            bondAmt,
			DelegatorShares:   math.LegacyOneDec(),
			Description:       stakingtypes.Description{},
			UnbondingHeight:   int64(0),
			UnbondingTime:     time.Unix(0, 0).UTC(),
			Commission:        stakingtypes.NewCommission(math.LegacyZeroDec(), math.LegacyZeroDec(), math.LegacyZeroDec()),
			MinSelfDelegation: math.ZeroInt(),
		}

		validators = append(validators, validator)
		delegations = append(delegations, stakingtypes.NewDelegation(genAccs[0].GetAddress().String(), valGenKey.ValidatorAddress, math.LegacyOneDec()))
		// blsKeys = append(blsKeys, checkpointingtypes.NewGenesisKey(sdk.ValAddress(val.Address), genesisBLSPubkey))
	}
	// total bond amount = bond amount * number of validators
	require.Equal(t, len(validators), len(delegations))
	totalBondAmt := bondAmt.MulRaw(int64(len(validators)))

	// set validators and delegations
	stakingGenesis := stakingtypes.NewGenesisState(stakingtypes.DefaultParams(), validators, delegations)
	stakingGenesis.Params.BondDenom = appparams.DefaultBondDenom
	genesisState[stakingtypes.ModuleName] = app.AppCodec().MustMarshalJSON(stakingGenesis)

	checkpointingGenesis := &checkpointingtypes.GenesisState{
		GenesisKeys: valSet,
	}
	genesisState[checkpointingtypes.ModuleName] = app.AppCodec().MustMarshalJSON(checkpointingGenesis)

	totalSupply := sdk.NewCoins()
	for _, b := range balances {
		// add genesis acc tokens to total supply
		totalSupply = totalSupply.Add(b.Coins...)
	}
	for range delegations {
		// add delegated tokens to total supply
		totalSupply = totalSupply.Add(sdk.NewCoin(appparams.DefaultBondDenom, bondAmt))
	}

	// add bonded amount to bonded pool module account
	balances = append(balances, banktypes.Balance{
		Address: authtypes.NewModuleAddress(stakingtypes.BondedPoolName).String(),
		Coins:   sdk.Coins{sdk.NewCoin(appparams.DefaultBondDenom, totalBondAmt)},
	})

	// update total supply
	bankGenesis := banktypes.NewGenesisState(
		banktypes.DefaultGenesisState().Params,
		balances,
		totalSupply,
		[]banktypes.Metadata{},
		[]banktypes.SendEnabled{},
	)
	genesisState[banktypes.ModuleName] = app.AppCodec().MustMarshalJSON(bankGenesis)

	return genesisState
}

// Setup initializes a new BabylonApp. A Nop logger is set in BabylonApp.
// Created Babylon application will have one validator with hardoced amount of tokens.
// This is necessary as from cosmos-sdk 0.46 it is required that there is at least
// one validator in validator set during InitGenesis abci call - https://github.com/cosmos/cosmos-sdk/pull/9697
func Setup(t *testing.T, isCheckTx bool) *BabylonApp {
	t.Helper()

	ps, err := SetupTestPrivSigner()
	require.NoError(t, err)
	valPubKey := ps.WrappedPV.Key.PubKey
	// generate genesis account
	acc := authtypes.NewBaseAccount(valPubKey.Address().Bytes(), &cosmosed.PubKey{Key: valPubKey.Bytes()}, 0, 0)
	balance := banktypes.Balance{
		Address: acc.GetAddress().String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(appparams.DefaultBondDenom, math.NewInt(100000000000000))),
	}
	ps.WrappedPV.Key.DelegatorAddress = acc.GetAddress().String()
	// create validator set with single validator
	genesisKey, err := GenesisKeyFromPrivSigner(ps)
	require.NoError(t, err)
	genesisValSet := []*checkpointingtypes.GenesisKey{genesisKey}

	app := SetupWithGenesisValSet(t, genesisValSet, ps, []authtypes.GenesisAccount{acc}, balance)

	return app
}

// SetupTestPrivSigner sets up a PrivSigner for testing
func SetupTestPrivSigner() (*PrivSigner, error) {
	// Create a temporary node directory
	nodeDir, err := ioutils.TempDir("", "tmp-signer")
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = os.RemoveAll(nodeDir)
	}()
	privSigner, _ := InitPrivSigner(nodeDir)
	return privSigner, nil
}

// SetupWithGenesisValSet initializes a new BabylonApp with a validator set and genesis accounts
// that also act as delegators. For simplicity, each validator is bonded with a delegation
// of one consensus engine unit (10^6) in the default token of the babylon app from first genesis
// account. A Nop logger is set in BabylonApp.
// Note that the privSigner should be the 0th item of valSet
func SetupWithGenesisValSet(t *testing.T, valSet []*checkpointingtypes.GenesisKey, privSigner *PrivSigner, genAccs []authtypes.GenesisAccount, balances ...banktypes.Balance) *BabylonApp {
	t.Helper()
	app, genesisState := setup(t, privSigner, true, 5)
	genesisState = genesisStateWithValSet(t, app, genesisState, valSet, genAccs, balances...)

	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
	require.NoError(t, err)

	// init chain will set the validator set and initialize the genesis accounts
	consensusParams := simsutils.DefaultConsensusParams
	consensusParams.Block.MaxGas = 100 * simsutils.DefaultGenTxGas
	// it is required that the VoteExtensionsEnableHeight > 0 to enable vote extension
	initialHeight := app.LastBlockHeight() + 1
	consensusParams.Abci = &cmtproto.ABCIParams{VoteExtensionsEnableHeight: initialHeight}
	_, err = app.InitChain(&abci.RequestInitChain{
		ChainId:         app.ChainID(),
		Time:            time.Now().UTC(),
		Validators:      []abci.ValidatorUpdate{},
		ConsensusParams: consensusParams,
		InitialHeight:   initialHeight,
		AppStateBytes:   stateBytes,
	})
	require.NoError(t, err)

	_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: initialHeight,
		Hash:   app.LastCommitID().Hash,
	})
	require.NoError(t, err)

	return app
}

func GenesisKeyFromPrivSigner(ps *PrivSigner) (*checkpointingtypes.GenesisKey, error) {
	valKeys, err := privval.NewValidatorKeys(ps.WrappedPV.GetValPrivKey(), ps.WrappedPV.GetBlsPrivKey())
	if err != nil {
		return nil, err
	}
	valPubkey, err := cryptocodec.FromCmtPubKeyInterface(valKeys.ValPubkey)
	if err != nil {
		return nil, err
	}
	return checkpointingtypes.NewGenesisKey(
		ps.WrappedPV.GetAddress(),
		&valKeys.BlsPubkey,
		valKeys.PoP,
		&cosmosed.PubKey{Key: valPubkey.Bytes()},
	)
}

// createRandomAccounts is a strategy used by addTestAddrs() in order to generated addresses in random order.
func createRandomAccounts(accNum int) []sdk.AccAddress {
	testAddrs := make([]sdk.AccAddress, accNum)
	for i := 0; i < accNum; i++ {
		pk := ed25519.GenPrivKey().PubKey()
		testAddrs[i] = sdk.AccAddress(pk.Address())
	}

	return testAddrs
}

// AddTestAddrs constructs and returns accNum amount of accounts with an
// initial balance of accAmt in random order
func AddTestAddrs(app *BabylonApp, ctx sdk.Context, accNum int, accAmt math.Int) ([]sdk.AccAddress, error) {
	testAddrs := createRandomAccounts(accNum)

	bondDenom, err := app.StakingKeeper.BondDenom(ctx)
	if err != nil {
		return nil, err
	}
	initCoins := sdk.NewCoins(sdk.NewCoin(bondDenom, accAmt))

	for _, addr := range testAddrs {
		initAccountWithCoins(app, ctx, addr, initCoins)
	}

	return testAddrs, nil
}

func initAccountWithCoins(app *BabylonApp, ctx sdk.Context, addr sdk.AccAddress, coins sdk.Coins) {
	err := app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, coins)
	if err != nil {
		panic(err)
	}

	err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addr, coins)
	if err != nil {
		panic(err)
	}
}

// SignAndDeliverWithoutCommit signs and delivers a transaction. No commit
func SignAndDeliverWithoutCommit(t *testing.T, txCfg client.TxConfig, app *bam.BaseApp, msgs []sdk.Msg, fees sdk.Coins, chainID string, accNums, accSeqs []uint64, blockTime time.Time, priv ...cryptotypes.PrivKey) (*abci.ResponseFinalizeBlock, error) {
	source := rand.NewSource(time.Now().UnixNano())
	tx, err := simsutils.GenSignedMockTx(
		rand.New(source),
		txCfg,
		msgs,
		fees,
		simsutils.DefaultGenTxGas,
		chainID,
		accNums,
		accSeqs,
		priv...,
	)
	require.NoError(t, err)

	bz, err := txCfg.TxEncoder()(tx)
	require.NoError(t, err)

	return app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: app.LastBlockHeight() + 1,
		Hash:   app.LastCommitID().Hash,
		Time:   blockTime,
		Txs:    [][]byte{bz},
	})
}
