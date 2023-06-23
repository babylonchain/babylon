package types

import (
	"math/big"

	bbn "github.com/babylonchain/babylon/types"
	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	btclctypes "github.com/babylonchain/babylon/x/btclightclient/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// AccountKeeper defines the expected account keeper used for simulations (noalias)
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) types.AccountI
	// Methods imported from account should be defined here
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	// Methods imported from bank should be defined here
}

type BTCLightClientKeeper interface {
	GetTipInfo(ctx sdk.Context) *btclctypes.BTCHeaderInfo
	GetHeaderByHash(ctx sdk.Context, hash *bbn.BTCHeaderHashBytes) *btclctypes.BTCHeaderInfo
}

type BtcCheckpointKeeper interface {
	GetPowLimit() *big.Int
	GetParams(ctx sdk.Context) (p btcctypes.Params)
}
