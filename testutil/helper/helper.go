package helper

import (
	"bytes"
	"testing"

	"cosmossdk.io/core/header"
	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	cosmosed "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	protoio "github.com/cosmos/gogoproto/io"

	"github.com/babylonchain/babylon/crypto/bls12381"
	"github.com/babylonchain/babylon/testutil/datagen"
	checkpointingtypes "github.com/babylonchain/babylon/x/checkpointing/types"

	"cosmossdk.io/math"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/babylonchain/babylon/app"
	appparams "github.com/babylonchain/babylon/app/params"
	btcstakingtypes "github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/babylonchain/babylon/x/epoching/keeper"
	"github.com/babylonchain/babylon/x/epoching/types"
)

// Helper is a structure which wraps the entire app and exposes functionalities for testing the epoching module
type Helper struct {
	t *testing.T

	Ctx         sdk.Context
	App         *app.BabylonApp
	MsgSrvr     types.MsgServer
	QueryClient types.QueryClient

	GenAccs       []authtypes.GenesisAccount
	GenValidators *datagen.GenesisValidators
}

// NewHelper creates the helper for testing the epoching module
func NewHelper(t *testing.T) *Helper {
	valSet, privSigner, err := datagen.GenesisValidatorSetWithPrivSigner(1)
	require.NoError(t, err)

	return NewHelperWithValSet(t, valSet, privSigner)
}

// NewHelperWithValSet is same as NewHelper, except that it creates a set of validators
// the privSigner is the 0th validator in valSet
func NewHelperWithValSet(t *testing.T, valSet *datagen.GenesisValidators, privSigner *app.PrivSigner) *Helper {
	// generate the genesis account
	signerPubKey := privSigner.WrappedPV.Key.PubKey
	acc := authtypes.NewBaseAccount(signerPubKey.Address().Bytes(), &cosmosed.PubKey{Key: signerPubKey.Bytes()}, 0, 0)
	privSigner.WrappedPV.Key.DelegatorAddress = acc.Address
	valSet.Keys[0].ValidatorAddress = privSigner.WrappedPV.GetAddress().String()
	// ensure the genesis account has a sufficient amount of tokens
	balance := banktypes.Balance{
		Address: acc.GetAddress().String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(appparams.DefaultBondDenom, sdk.DefaultPowerReduction.MulRaw(10000000))),
	}
	GenAccs := []authtypes.GenesisAccount{acc}

	// setup the app and ctx
	app := app.SetupWithGenesisValSet(t, valSet.GetGenesisKeys(), privSigner, GenAccs, balance)
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
		msgSrvr,
		queryClient,
		GenAccs,
		valSet,
	}
}

// NewHelperWithValSetNoSigner is same as NewHelperWithValSet, except that the privSigner is not
// included in the validator set
func NewHelperWithValSetNoSigner(t *testing.T, valSet *datagen.GenesisValidators, privSigner *app.PrivSigner) *Helper {
	// generate the genesis account
	signerPubKey := privSigner.WrappedPV.Key.PubKey
	acc := authtypes.NewBaseAccount(signerPubKey.Address().Bytes(), &cosmosed.PubKey{Key: signerPubKey.Bytes()}, 0, 0)
	privSigner.WrappedPV.Key.DelegatorAddress = acc.Address
	// set a random validator address instead of the privSigner's
	valSet.Keys[0].ValidatorAddress = datagen.GenRandomValidatorAddress().String()
	// ensure the genesis account has a sufficient amount of tokens
	balance := banktypes.Balance{
		Address: acc.GetAddress().String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(appparams.DefaultBondDenom, sdk.DefaultPowerReduction.MulRaw(10000000))),
	}
	GenAccs := []authtypes.GenesisAccount{acc}

	// setup the app and ctx
	app := app.SetupWithGenesisValSet(t, valSet.GetGenesisKeys(), privSigner, GenAccs, balance)
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
		msgSrvr,
		queryClient,
		GenAccs,
		valSet,
	}
}

func (h *Helper) NoError(err error) {
	require.NoError(h.t, err)
}

func (h *Helper) Error(err error) {
	require.Error(h.t, err)
}

func (h *Helper) getExtendedVotesFromValSet(epochNum, height uint64, blockHash checkpointingtypes.BlockHash, valSet *datagen.GenesisValidators) ([]abci.ExtendedVoteInfo, error) {
	valPrivKey := valSet.GetValPrivKeys()
	blsPrivKeys := valSet.GetBLSPrivKeys()
	genesisKeys := valSet.GetGenesisKeys()
	signBytes := checkpointingtypes.GetSignBytes(epochNum, blockHash)
	extendedVotes := make([]abci.ExtendedVoteInfo, 0, len(valSet.Keys))
	for i, sk := range blsPrivKeys {
		// 1. set build vote extension
		sig := bls12381.Sign(sk, signBytes)
		ve := checkpointingtypes.VoteExtension{
			Signer:    genesisKeys[i].ValidatorAddress,
			BlockHash: &blockHash,
			EpochNum:  epochNum,
			Height:    height,
			BlsSig:    &sig,
		}
		veBytes, err := ve.Marshal()
		if err != nil {
			return nil, err
		}

		// 2. sign validator signature over the vote extension
		cve := cmtproto.CanonicalVoteExtension{
			Extension: veBytes,
			Height:    int64(height),
			Round:     int64(0),
			ChainId:   h.App.ChainID(),
		}
		var cveBuffer bytes.Buffer
		err = protoio.NewDelimitedWriter(&cveBuffer).WriteMsg(&cve)
		if err != nil {
			return nil, err
		}
		cveBytes := cveBuffer.Bytes()
		extensionSig, err := valPrivKey[i].Sign(cveBytes)
		if err != nil {
			return nil, err
		}

		// 3. set up the validator of the vote extension
		valAddress, err := sdk.ValAddressFromBech32(genesisKeys[i].ValidatorAddress)
		if err != nil {
			return nil, err
		}
		val := abci.Validator{
			Address: valAddress.Bytes(),
			Power:   1000,
		}

		// 4. construct the extended vote info
		veInfo := abci.ExtendedVoteInfo{
			Validator:          val,
			VoteExtension:      veBytes,
			BlockIdFlag:        cmtproto.BlockIDFlagCommit,
			ExtensionSignature: extensionSig,
		}
		extendedVotes = append(extendedVotes, veInfo)
	}

	return extendedVotes, nil
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
	coin := sdk.NewCoin(appparams.DefaultBondDenom, h.App.StakingKeeper.TokensFromConsensusPower(h.Ctx, power))
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

// WrappedCancelUnbondingDelegation calls handler to cancel unbonding a delegation
func (h *Helper) WrappedCancelUnbondingDelegation(delegator sdk.AccAddress, val sdk.ValAddress, amount math.Int, creationHeight int64) *sdk.Result {
	unbondAmt := sdk.NewCoin(appparams.DefaultBondDenom, amount)
	msg := stakingtypes.NewMsgCancelUnbondingDelegation(delegator.String(), val.String(), creationHeight, unbondAmt)
	wmsg := types.NewMsgWrappedCancelUnbondingDelegation(msg)
	return h.Handle(func(ctx sdk.Context) (proto.Message, error) {
		return h.MsgSrvr.WrappedCancelUnbondingDelegation(ctx, wmsg)
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
	v, err := h.App.StakingKeeper.GetValidator(h.Ctx, addr)
	require.NoError(h.t, err)
	require.Equal(h.t, jailed, v.Jailed, "wrong Jalied status")
	if status >= 0 {
		require.Equal(h.t, status, v.Status)
	}
	return v
}

// CheckDelegator asserts that a delegator exists
func (h *Helper) CheckDelegator(delegator sdk.AccAddress, val sdk.ValAddress, found bool) {
	_, ok := h.App.StakingKeeper.GetDelegation(h.Ctx, delegator, val)
	require.Equal(h.t, ok, found)
}

func (h *Helper) AddDelegation(del *btcstakingtypes.BTCDelegation) {
	err := h.App.BTCStakingKeeper.AddBTCDelegation(h.Ctx, del)
	h.NoError(err)
}

func (h *Helper) AddFinalityProvider(fp *btcstakingtypes.FinalityProvider) {
	h.App.BTCStakingKeeper.SetFinalityProvider(h.Ctx, fp)
}
