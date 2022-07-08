package types

import (
	bbl "github.com/babylonchain/babylon/types"
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

type BTCLightClientHooks interface {
	AfterBTCRollBack(ctx sdk.Context, hash bbl.BTCHeaderHashBytes, height uint64)    // Must be called after the chain is rolled back
	AfterBTCRollForward(ctx sdk.Context, hash bbl.BTCHeaderHashBytes, height uint64) // Must be called after the chain is rolled forward
}
