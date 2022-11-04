package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcclientkeeper "github.com/cosmos/ibc-go/v5/modules/core/02-client/keeper"
	ibctmtypes "github.com/cosmos/ibc-go/v5/modules/light-clients/07-tendermint/types"
)

type Hooks struct {
	k Keeper
}

// ensures Hooks implements StakingHooks interfaces
var _ ibcclientkeeper.ClientHooks = Hooks{}

// Create new distribution hooks
func (k Keeper) Hooks() Hooks { return Hooks{k} }

func (h Hooks) AfterHeaderWithValidCommit(ctx sdk.Context, txHash []byte, header *ibctmtypes.Header) {
	panic("unimplemented")
}
