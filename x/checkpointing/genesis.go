package checkpointing

import (
	"github.com/babylonchain/babylon/x/checkpointing/keeper"
	"github.com/babylonchain/babylon/x/checkpointing/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)

	valPubkeys := make([]cryptotypes.PubKey, len(genState.ValPubkeys))
	for i := 0; i < len(genState.ValPubkeys); i++ {
		valPubkeys[i] = &ed25519.PubKey{Key: genState.ValPubkeys[i]}
	}
	k.SetGenBlsKeys(ctx, genState.BlsKeys, valPubkeys)
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

	return genesis
}
