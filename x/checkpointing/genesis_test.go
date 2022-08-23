package checkpointing_test

import (
	"github.com/babylonchain/babylon/crypto/bls12381"
	"github.com/babylonchain/babylon/privval"
	"github.com/babylonchain/babylon/x/checkpointing"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"testing"

	simapp "github.com/babylonchain/babylon/app"
	"github.com/babylonchain/babylon/x/checkpointing/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestInitGenesis(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ckptKeeper := app.CheckpointingKeeper

	valNum := 10
	blsKeys := make([]*types.BlsKey, valNum)
	valPubkeys := make([][]byte, valNum)
	for i := 0; i < valNum; i++ {
		valKeys, err := privval.NewValidatorKeys(ed25519.GenPrivKey(), bls12381.GenPrivKey())
		require.NoError(t, err)
		blsKey := &types.BlsKey{
			ValidatorAddress: valKeys.ValPubkey.Address().String(),
			Pubkey:           &valKeys.BlsPubkey,
			Pop:              valKeys.PoP,
		}
		blsKeys[i] = blsKey
		valPubkeys[i] = valKeys.ValPubkey.Bytes()
	}
	genesisState := types.GenesisState{
		Params:     types.Params{},
		BlsKeys:    blsKeys,
		ValPubkeys: valPubkeys,
	}

	checkpointing.InitGenesis(ctx, ckptKeeper, genesisState)
	for i := 0; i < valNum; i++ {
		addr, err := sdk.ValAddressFromHex(blsKeys[i].ValidatorAddress)
		require.NoError(t, err)
		blsKey, err := ckptKeeper.GetBlsPubKey(ctx, addr)
		require.True(t, blsKeys[i].Pubkey.Equal(blsKey))
	}
}
