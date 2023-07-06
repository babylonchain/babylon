package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

type BTCStakingKeeper interface {
	HasBTCValidator(ctx sdk.Context, valBTCPK []byte) bool
	GetVotingPower(ctx sdk.Context, valBTCPK []byte, height uint64) uint64
	GetVotingPowerTable(ctx sdk.Context, height uint64) map[string]uint64
	GetBTCStakingActivatedHeight(ctx sdk.Context) (uint64, error)
}

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
