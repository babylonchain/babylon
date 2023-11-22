package testepoching

import (
	"math/rand"
	"testing"

	"cosmossdk.io/core/header"
	"github.com/babylonchain/babylon/crypto/bls12381"
	"github.com/babylonchain/babylon/privval"
	checkpointingtypes "github.com/babylonchain/babylon/x/checkpointing/types"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/ed25519"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cosmosed "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"

	"cosmossdk.io/math"
	proto "github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	appparams "github.com/babylonchain/babylon/app/params"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/babylonchain/babylon/app"
	"github.com/babylonchain/babylon/x/epoching/keeper"
	"github.com/babylonchain/babylon/x/epoching/types"
)

type GenesisValidators struct {
	GenesisKeys []*checkpointingtypes.GenesisKey
	BlsPrivKeys []bls12381.PrivateKey
}

// Helper is a structure which wraps the entire app and exposes functionalities for testing the epoching module
type Helper struct {
	t *testing.T

	Ctx            sdk.Context
	App            *app.BabylonApp
	EpochingKeeper *keeper.Keeper
	MsgSrvr        types.MsgServer
	QueryClient    types.QueryClient
	StakingKeeper  *stakingkeeper.Keeper

	GenAccs       []authtypes.GenesisAccount
	GenValidators *GenesisValidators
}

// NewHelper creates the helper for testing the epoching module
func NewHelper(t *testing.T) *Helper {
	valSet, err := GenesisValidatorSet(1)
	require.NoError(t, err)
	// generate genesis account
	acc := authtypes.NewBaseAccount(valSet.GenesisKeys[0].ValPubkey.Address().Bytes(), valSet.GenesisKeys[0].ValPubkey, 0, 0)
	balance := banktypes.Balance{
		Address: acc.GetAddress().String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(appparams.DefaultBondDenom, math.NewInt(100000000000000))),
	}

	app := app.SetupWithGenesisValSet(t, valSet.GenesisKeys, []authtypes.GenesisAccount{acc}, balance)
	ctx := app.BaseApp.NewContext(false).WithBlockHeight(1).WithHeaderInfo(header.Info{Height: 1}) // NOTE: height is 1

	epochingKeeper := app.EpochingKeeper

	querier := keeper.Querier{Keeper: epochingKeeper}
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, querier)
	queryClient := types.NewQueryClient(queryHelper)
	msgSrvr := keeper.NewMsgServerImpl(epochingKeeper)

	return &Helper{
		t,
		ctx,
		app,
		&epochingKeeper,
		msgSrvr,
		queryClient,
		app.StakingKeeper,
		nil,
		valSet,
	}
}

// NewHelperWithValSet is same as NewHelper, except that it creates a set of validators
func NewHelperWithValSet(t *testing.T) *Helper {
	// generate the validator set with 10 validators
	valSet, err := GenesisValidatorSet(10)
	require.NoError(t, err)

	// generate the genesis account
	senderPrivKey := secp256k1.GenPrivKey()
	acc := authtypes.NewBaseAccount(senderPrivKey.PubKey().Address().Bytes(), senderPrivKey.PubKey(), 0, 0)
	// ensure the genesis account has a sufficient amount of tokens
	balance := banktypes.Balance{
		Address: acc.GetAddress().String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(appparams.DefaultBondDenom, sdk.DefaultPowerReduction.MulRaw(10000000))),
	}
	GenAccs := []authtypes.GenesisAccount{acc}

	// setup the app and ctx
	app := app.SetupWithGenesisValSet(t, valSet.GenesisKeys, GenAccs, balance)
	ctx := app.BaseApp.NewContext(false).WithBlockHeight(1).WithHeaderInfo(header.Info{Height: 1}) // NOTE: height is 1

	// get necessary subsets of the app/keeper
	epochingKeeper := app.EpochingKeeper
	querier := keeper.Querier{Keeper: epochingKeeper}
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, querier)
	queryClient := types.NewQueryClient(queryHelper)
	msgSrvr := keeper.NewMsgServerImpl(epochingKeeper)

	return &Helper{
		t,
		ctx,
		app,
		&epochingKeeper,
		msgSrvr,
		queryClient,
		app.StakingKeeper,
		GenAccs,
		valSet,
	}
}

// GenAndApplyEmptyBlock generates a new empty block and appends it to the current blockchain
func (h *Helper) GenAndApplyEmptyBlock(r *rand.Rand) (sdk.Context, error) {
	newHeight := h.App.LastBlockHeight() + 1
	valSet, err := h.StakingKeeper.GetLastValidators(h.Ctx)
	if err != nil {
		return sdk.Context{}, err
	}
	valhash := CalculateValHash(valSet)
	newHeader := tmproto.Header{
		Height:             newHeight,
		ValidatorsHash:     valhash,
		NextValidatorsHash: valhash,
	}

	resp, err := h.App.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height:             newHeader.Height,
		NextValidatorsHash: newHeader.NextValidatorsHash,
	})
	if err != nil {
		return sdk.Context{}, err
	}

	newHeader.AppHash = resp.AppHash
	ctxDuringHeight := h.Ctx.WithHeaderInfo(header.Info{
		Height:  newHeader.Height,
		AppHash: resp.AppHash,
	}).WithBlockHeader(newHeader)

	_, err = h.App.Commit()
	if err != nil {
		return sdk.Context{}, err
	}

	if newHeight == 1 {
		// do it again
		// TODO: Figure out why when ctx height is 1, GenAndApplyEmptyBlock
		// will still give ctx height 1 once, then start to increment
		return h.GenAndApplyEmptyBlock(r)
	}

	return ctxDuringHeight, nil
}

// WrappedDelegate calls handler to delegate stake for a validator
func (h *Helper) WrappedDelegate(delegator sdk.AccAddress, val sdk.ValAddress, amount math.Int) *sdk.Result {
	coin := sdk.NewCoin(appparams.DefaultBondDenom, amount)
	msg := stakingtypes.NewMsgDelegate(delegator.String(), val.String(), coin)
	wmsg := types.NewMsgWrappedDelegate(msg)
	return h.Handle(func(ctx sdk.Context) (proto.Message, error) {
		return h.MsgSrvr.WrappedDelegate(ctx, wmsg)
	})
}

// WrappedDelegateWithPower calls handler to delegate stake for a validator
func (h *Helper) WrappedDelegateWithPower(delegator sdk.AccAddress, val sdk.ValAddress, power int64) *sdk.Result {
	coin := sdk.NewCoin(appparams.DefaultBondDenom, h.StakingKeeper.TokensFromConsensusPower(h.Ctx, power))
	msg := stakingtypes.NewMsgDelegate(delegator.String(), val.String(), coin)
	wmsg := types.NewMsgWrappedDelegate(msg)
	return h.Handle(func(ctx sdk.Context) (proto.Message, error) {
		return h.MsgSrvr.WrappedDelegate(ctx, wmsg)
	})
}

// WrappedUndelegate calls handler to unbound some stake from a validator.
func (h *Helper) WrappedUndelegate(delegator sdk.AccAddress, val sdk.ValAddress, amount math.Int) *sdk.Result {
	unbondAmt := sdk.NewCoin(appparams.DefaultBondDenom, amount)
	msg := stakingtypes.NewMsgUndelegate(delegator.String(), val.String(), unbondAmt)
	wmsg := types.NewMsgWrappedUndelegate(msg)
	return h.Handle(func(ctx sdk.Context) (proto.Message, error) {
		return h.MsgSrvr.WrappedUndelegate(ctx, wmsg)
	})
}

// WrappedBeginRedelegate calls handler to redelegate some stake from a validator to another
func (h *Helper) WrappedBeginRedelegate(delegator sdk.AccAddress, srcVal sdk.ValAddress, dstVal sdk.ValAddress, amount math.Int) *sdk.Result {
	unbondAmt := sdk.NewCoin(appparams.DefaultBondDenom, amount)
	msg := stakingtypes.NewMsgBeginRedelegate(delegator.String(), srcVal.String(), dstVal.String(), unbondAmt)
	wmsg := types.NewMsgWrappedBeginRedelegate(msg)
	return h.Handle(func(ctx sdk.Context) (proto.Message, error) {
		return h.MsgSrvr.WrappedBeginRedelegate(ctx, wmsg)
	})
}

// Handle executes an action function with the Helper's context, wraps the result into an SDK service result, and performs two assertions before returning it
func (h *Helper) Handle(action func(sdk.Context) (proto.Message, error)) *sdk.Result {
	res, err := action(h.Ctx)
	require.NoError(h.t, err)
	r, err := sdk.WrapServiceResult(h.Ctx, res, err)
	require.NoError(h.t, err)
	require.NotNil(h.t, r)
	require.NoError(h.t, err)
	return r
}

// CheckValidator asserts that a validor exists and has a given status (if status!="")
// and if has a right jailed flag.
func (h *Helper) CheckValidator(addr sdk.ValAddress, status stakingtypes.BondStatus, jailed bool) stakingtypes.Validator {
	v, err := h.StakingKeeper.GetValidator(h.Ctx, addr)
	require.NoError(h.t, err)
	require.Equal(h.t, jailed, v.Jailed, "wrong Jalied status")
	if status >= 0 {
		require.Equal(h.t, status, v.Status)
	}
	return v
}

// CheckDelegator asserts that a delegator exists
func (h *Helper) CheckDelegator(delegator sdk.AccAddress, val sdk.ValAddress, found bool) {
	_, ok := h.StakingKeeper.GetDelegation(h.Ctx, delegator, val)
	require.Equal(h.t, ok, found)
}

// GenesisValidatorSet generates a set with `numVals` genesis validators
func GenesisValidatorSet(numVals int) (*GenesisValidators, error) {
	genesisKeys := make([]*checkpointingtypes.GenesisKey, 0, numVals)
	blsPrivKeys := make([]bls12381.PrivateKey, 0, numVals)
	for i := 0; i < numVals; i++ {
		blsPrivKey := bls12381.GenPrivKey()
		// create validator set with single validator
		valKeys, err := privval.NewValidatorKeys(ed25519.GenPrivKey(), blsPrivKey)
		if err != nil {
			return nil, err
		}
		valPubkey, err := cryptocodec.FromCmtPubKeyInterface(valKeys.ValPubkey)
		if err != nil {
			return nil, err
		}
		genesisKey, err := checkpointingtypes.NewGenesisKey(
			sdk.ValAddress(valKeys.ValPubkey.Address()),
			&valKeys.BlsPubkey,
			valKeys.PoP,
			&cosmosed.PubKey{Key: valPubkey.Bytes()},
		)
		if err != nil {
			return nil, err
		}
		genesisKeys = append(genesisKeys, genesisKey)
		blsPrivKeys = append(blsPrivKeys, blsPrivKey)
	}
	return &GenesisValidators{
		GenesisKeys: genesisKeys,
		BlsPrivKeys: blsPrivKeys,
	}, nil
}
