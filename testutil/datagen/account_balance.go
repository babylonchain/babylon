package datagen

import (
	sec256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func GenRandomAccount() *authtypes.BaseAccount {
	senderPrivKey := sec256k1.GenPrivKey()
	acc := authtypes.NewBaseAccount(senderPrivKey.PubKey().Address().Bytes(), senderPrivKey.PubKey(), 0, 0)
	return acc
}

func GenRandomAccWithBalance(n int) ([]authtypes.GenesisAccount, []banktypes.Balance) {
	accs := make([]authtypes.GenesisAccount, n)
	balances := make([]banktypes.Balance, n)
	for i := 0; i < n; i++ {
		senderPrivKey := sec256k1.GenPrivKey()
		acc := authtypes.NewBaseAccount(senderPrivKey.PubKey().Address().Bytes(), senderPrivKey.PubKey(), 0, 0)
		accs[i] = acc
		balance := banktypes.Balance{
			Address: acc.GetAddress().String(),
			Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100000000000000))),
		}
		balances[i] = balance
	}

	return accs, balances
}
