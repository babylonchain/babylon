package simulation

import (
	"fmt"
	"math/rand"

	"github.com/babylonchain/babylon/x/epoching/keeper"
	"github.com/babylonchain/babylon/x/epoching/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Simulation operation weights constants
const (
	OpWeightMsgWrappedDelegate        = "op_weight_msg_wrapped_delegate"
	OpWeightMsgWrappedUndelegate      = "op_weight_msg_wrapped_undelegate"
	OpWeightMsgWrappedBeginRedelegate = "op_weight_msg_wrapped_begin_redelegate"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec, ak types.AccountKeeper, bk types.BankKeeper, stk types.StakingKeeper, k keeper.Keeper,
) simulation.WeightedOperations {
	var (
		weightMsgWrappedDelegate        int
		weightMsgWrappedUndelegate      int
		weightMsgWrappedBeginRedelegate int
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgWrappedDelegate, &weightMsgWrappedDelegate, nil,
		func(_ *rand.Rand) {
			weightMsgWrappedDelegate = simappparams.DefaultWeightMsgDelegate // TODO: use our own (and randomised) weight rather than those from the unwrapped msgs
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgWrappedUndelegate, &weightMsgWrappedUndelegate, nil,
		func(_ *rand.Rand) {
			weightMsgWrappedUndelegate = simappparams.DefaultWeightMsgUndelegate // TODO: use our own (and randomised) weight rather than those from the unwrapped msgs
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgWrappedBeginRedelegate, &weightMsgWrappedBeginRedelegate, nil,
		func(_ *rand.Rand) {
			weightMsgWrappedBeginRedelegate = simappparams.DefaultWeightMsgBeginRedelegate // TODO: use our own (and randomised) weight rather than those from the unwrapped msgs
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgWrappedDelegate,
			SimulateMsgWrappedDelegate(ak, bk, stk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgWrappedUndelegate,
			SimulateMsgWrappedUndelegate(ak, bk, stk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgWrappedBeginRedelegate,
			SimulateMsgWrappedBeginRedelegate(ak, bk, stk, k),
		),
	}
}

// SimulateMsgDelegate generates a MsgDelegate with random values
func SimulateMsgWrappedDelegate(ak types.AccountKeeper, bk types.BankKeeper, stk types.StakingKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		epoch := k.GetEpoch(ctx)
		valSet := k.GetValidatorSet(ctx, epoch.EpochNumber)
		if len(valSet) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgWrappedDelegate, "number of validators in this epoch equal zero"), nil, nil
		}

		// pick a random validator
		i := r.Intn(len(valSet))
		val, ok := stk.GetValidator(ctx, valSet[i].Addr)
		if !ok {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgWrappedDelegate, "unable to pick a validator"), nil, nil
		}
		if val.InvalidExRate() {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgWrappedDelegate, "validator's invalid exchange rate"), nil, nil
		}

		// pick a random bondAmt
		simAccount, _ := simtypes.RandomAcc(r, accs)
		denom := stk.GetParams(ctx).BondDenom
		amount := bk.GetBalance(ctx, simAccount.Address, denom).Amount
		if !amount.IsPositive() {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgWrappedDelegate, "balance is negative"), nil, nil
		}
		amount, err := simtypes.RandPositiveInt(r, amount)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgWrappedDelegate, "unable to generate positive amount"), nil, err
		}
		bondAmt := sdk.NewCoin(denom, amount)

		// pick a random fee rate
		var fees sdk.Coins
		account := ak.GetAccount(ctx, simAccount.Address)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())
		coins, hasNeg := spendable.SafeSub(sdk.Coins{bondAmt})
		if !hasNeg {
			fees, err = simtypes.RandomFees(r, ctx, coins)
			if err != nil {
				return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgWrappedDelegate, "unable to generate fees"), nil, err
			}
		}

		msg := stakingtypes.NewMsgDelegate(simAccount.Address, val.GetOperator(), bondAmt)
		wmsg := types.NewMsgWrappedDelegate(msg)

		txCtx := simulation.OperationInput{
			App:           app,
			TxGen:         simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:           nil,
			Msg:           wmsg,
			MsgType:       wmsg.Type(),
			Context:       ctx,
			SimAccount:    simAccount,
			AccountKeeper: ak,
			ModuleName:    types.ModuleName,
		}

		return simulation.GenAndDeliverTx(txCtx, fees)
	}
}

// SimulateMsgUndelegate generates a MsgUndelegate with random values
func SimulateMsgWrappedUndelegate(ak types.AccountKeeper, bk types.BankKeeper, stk types.StakingKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		epoch := k.GetEpoch(ctx)
		valSet := k.GetValidatorSet(ctx, epoch.EpochNumber)
		if len(valSet) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgWrappedUndelegate, "number of validators in this epoch equal zero"), nil, nil
		}

		// pick a random validator
		i := r.Intn(len(valSet))
		val, ok := stk.GetValidator(ctx, valSet[i].Addr)
		if !ok {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgWrappedUndelegate, "unable to pick a validator"), nil, nil
		}
		if val.InvalidExRate() {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgWrappedUndelegate, "validator's invalid exchange rate"), nil, nil
		}

		// pick a random delegator from validator
		valAddr := val.GetOperator()
		delegations := stk.GetValidatorDelegations(ctx, val.GetOperator())
		if delegations == nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgWrappedUndelegate, "keeper does not have any delegation entries"), nil, nil
		}
		delegation := delegations[r.Intn(len(delegations))]
		delAddr := delegation.GetDelegatorAddr()

		if stk.HasMaxUnbondingDelegationEntries(ctx, delAddr, valAddr) {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgWrappedUndelegate, "keeper reaches max unbonding delegation entries"), nil, nil
		}

		// pick a random unbondAmt
		totalBond := val.TokensFromShares(delegation.GetShares()).TruncateInt()
		if !totalBond.IsPositive() {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgWrappedUndelegate, "total bond is negative"), nil, nil
		}
		unbondAmt, err := simtypes.RandPositiveInt(r, totalBond)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgWrappedUndelegate, "invalid unbond amount"), nil, err
		}
		if unbondAmt.IsZero() {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgWrappedUndelegate, "unbond amount is zero"), nil, nil
		}

		msg := stakingtypes.NewMsgUndelegate(
			delAddr, valAddr, sdk.NewCoin(stk.BondDenom(ctx), unbondAmt),
		)
		wmsg := types.NewMsgWrappedUndelegate(msg)

		// need to retrieve the simulation account associated with delegation to retrieve PrivKey
		var simAccount simtypes.Account

		for _, simAcc := range accs {
			if simAcc.Address.Equals(delAddr) {
				simAccount = simAcc
				break
			}
		}
		// if simaccount.PrivKey == nil, delegation address does not exist in accs. Return error
		if simAccount.PrivKey == nil {
			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "account private key is nil"), nil, fmt.Errorf("delegation addr: %s does not exist in simulation accounts", delAddr)
		}

		account := ak.GetAccount(ctx, delAddr)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:             nil,
			Msg:             wmsg,
			MsgType:         wmsg.Type(),
			Context:         ctx,
			SimAccount:      simAccount,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: spendable,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

// SimulateMsgBeginRedelegate generates a MsgBeginRedelegate with random values
func SimulateMsgWrappedBeginRedelegate(ak types.AccountKeeper, bk types.BankKeeper, stk types.StakingKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		epoch := k.GetEpoch(ctx)
		valSet := k.GetValidatorSet(ctx, epoch.EpochNumber)
		if len(valSet) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgWrappedBeginRedelegate, "number of validators in this epoch equal zero"), nil, nil
		}

		// pick a random source validator
		i := r.Intn(len(valSet))
		srcVal, ok := stk.GetValidator(ctx, valSet[i].Addr)
		if !ok {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgWrappedBeginRedelegate, "unable to pick a validator"), nil, nil
		}

		srcAddr := srcVal.GetOperator()
		delegations := stk.GetValidatorDelegations(ctx, srcAddr)
		if delegations == nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgWrappedBeginRedelegate, "keeper does have any delegation entries"), nil, nil
		}

		// pick a random delegator from src validator
		delegation := delegations[r.Intn(len(delegations))]
		delAddr := delegation.GetDelegatorAddr()

		if stk.HasReceivingRedelegation(ctx, delAddr, srcAddr) {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgWrappedBeginRedelegate, "receiving redelegation is not allowed"), nil, nil // skip
		}

		// pick a random destination validator
		i = r.Intn(len(valSet))
		destVal, ok := stk.GetValidator(ctx, valSet[i].Addr)
		if !ok {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgWrappedBeginRedelegate, "unable to pick a validator"), nil, nil
		}
		destAddr := destVal.GetOperator()
		if srcAddr.Equals(destAddr) || destVal.InvalidExRate() || stk.HasMaxRedelegationEntries(ctx, delAddr, srcAddr, destAddr) {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgWrappedBeginRedelegate, "checks failed"), nil, nil
		}

		// pick a random redAmt
		totalBond := srcVal.TokensFromShares(delegation.GetShares()).TruncateInt()
		if !totalBond.IsPositive() {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgWrappedBeginRedelegate, "total bond is negative"), nil, nil
		}
		redAmt, err := simtypes.RandPositiveInt(r, totalBond)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgWrappedBeginRedelegate, "unable to generate positive amount"), nil, err
		}
		if redAmt.IsZero() {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgWrappedBeginRedelegate, "amount is zero"), nil, nil
		}

		// check if the shares truncate to zero
		shares, err := srcVal.SharesFromTokens(redAmt)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgWrappedBeginRedelegate, "invalid shares"), nil, err
		}

		if srcVal.TokensFromShares(shares).TruncateInt().IsZero() {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgWrappedBeginRedelegate, "shares truncate to zero"), nil, nil // skip
		}

		// need to retrieve the simulation account associated with delegation to retrieve PrivKey
		var simAccount simtypes.Account

		for _, simAcc := range accs {
			if simAcc.Address.Equals(delAddr) {
				simAccount = simAcc
				break
			}
		}

		// if simaccount.PrivKey == nil, delegation address does not exist in accs. Return error
		if simAccount.PrivKey == nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgWrappedBeginRedelegate, "account private key is nil"), nil, fmt.Errorf("delegation addr: %s does not exist in simulation accounts", delAddr)
		}

		account := ak.GetAccount(ctx, delAddr)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())

		msg := stakingtypes.NewMsgBeginRedelegate(
			delAddr, srcAddr, destAddr,
			sdk.NewCoin(stk.BondDenom(ctx), redAmt),
		)
		wmsg := types.NewMsgWrappedBeginRedelegate(msg)

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:             nil,
			Msg:             wmsg,
			MsgType:         wmsg.Type(),
			Context:         ctx,
			SimAccount:      simAccount,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: spendable,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}
