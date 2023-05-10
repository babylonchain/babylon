package checkpointing_test

import (
	"testing"

	"github.com/babylonchain/babylon/crypto/bls12381"
	"github.com/babylonchain/babylon/privval"
	"github.com/babylonchain/babylon/x/checkpointing"
	"github.com/cometbft/cometbft/crypto/ed25519"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cosmosed "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	simapp "github.com/babylonchain/babylon/app"
	"github.com/babylonchain/babylon/x/checkpointing/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
)

func TestInitGenesis(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ckptKeeper := app.CheckpointingKeeper

	valNum := 10
	genKeys := make([]*types.GenesisKey, valNum)
	for i := 0; i < valNum; i++ {
		valKeys, err := privval.NewValidatorKeys(ed25519.GenPrivKey(), bls12381.GenPrivKey())
		require.NoError(t, err)
		valPubkey, err := cryptocodec.FromTmPubKeyInterface(valKeys.ValPubkey)
		require.NoError(t, err)
		genKey, err := types.NewGenesisKey(
			sdk.ValAddress(valKeys.ValPubkey.Address()),
			&valKeys.BlsPubkey,
			valKeys.PoP,
			&cosmosed.PubKey{Key: valPubkey.Bytes()},
		)
		require.NoError(t, err)
		genKeys[i] = genKey
	}
	genesisState := types.GenesisState{
		GenesisKeys: genKeys,
	}

	checkpointing.InitGenesis(ctx, ckptKeeper, genesisState)
	for i := 0; i < valNum; i++ {
		addr, err := sdk.ValAddressFromBech32(genKeys[i].ValidatorAddress)
		require.NoError(t, err)
		blsKey, err := ckptKeeper.GetBlsPubKey(ctx, addr)
		require.NoError(t, err)
		require.True(t, genKeys[i].BlsKey.Pubkey.Equal(blsKey))
	}
}
