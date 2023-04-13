package wasmbinding

import (
	"encoding/json"

	errorsmod "cosmossdk.io/errors"
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	"github.com/babylonchain/babylon/wasmbinding/bindings"
	epochingkeeper "github.com/babylonchain/babylon/x/epoching/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type QueryPlugin struct {
	epochingKeeper *epochingkeeper.Keeper
}

// NewQueryPlugin returns a reference to a new QueryPlugin.
func NewQueryPlugin(ek *epochingkeeper.Keeper) *QueryPlugin {
	return &QueryPlugin{
		epochingKeeper: ek,
	}
}

// CustomQuerier dispatches custom CosmWasm bindings queries.
func CustomQuerier(qp *QueryPlugin) func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
	return func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
		var contractQuery bindings.BabylonQuery
		if err := json.Unmarshal(request, &contractQuery); err != nil {
			return nil, errorsmod.Wrap(err, "failed to unarshall request ")
		}

		switch {
		case contractQuery.Epoch != nil:
			epoch := qp.epochingKeeper.GetEpoch(ctx)
			res := bindings.CurrentEpochResponse{
				Epoch: epoch.EpochNumber,
			}

			ctx.Logger().Debug("Marshalled custom response")

			bz, err := json.Marshal(res)
			if err != nil {
				return nil, errorsmod.Wrap(err, "failed marshaling")
			}

			return bz, nil

		default:
			return nil, wasmvmtypes.UnsupportedRequest{Kind: "unknown babylon query variant"}
		}
	}
}

func RegisterCustomPlugins(
	ek *epochingkeeper.Keeper,
) []wasmkeeper.Option {
	wasmQueryPlugin := NewQueryPlugin(ek)

	queryPluginOpt := wasmkeeper.WithQueryPlugins(&wasmkeeper.QueryPlugins{
		Custom: CustomQuerier(wasmQueryPlugin),
	})

	return []wasm.Option{
		queryPluginOpt,
	}
}
