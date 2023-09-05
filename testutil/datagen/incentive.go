package datagen

import (
	"math/rand"

	itypes "github.com/babylonchain/babylon/x/incentive/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	characters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	denomLen   = 5
)

func GenRandomDenom(r *rand.Rand) string {
	var result string
	// Generate the random string
	for i := 0; i < denomLen; i++ {
		// Generate a random index within the range of the character set
		index := r.Intn(len(characters))
		// Add the randomly selected character to the result
		result += string(characters[index])
	}
	return result
}

func GenRandomStakeholderType(r *rand.Rand) itypes.StakeholderType {
	stBytes := []byte{byte(RandomInt(r, 4))}
	st, err := itypes.NewStakeHolderType(stBytes)
	if err != nil {
		panic(err) // only programming error is possible
	}
	return st
}

func GenRandomCoins(r *rand.Rand) sdk.Coins {
	numCoins := r.Int31n(10) + 10
	coins := sdk.NewCoins()
	for i := int32(0); i < numCoins; i++ {
		demon := GenRandomDenom(r)
		amount := r.Int63n(10000) + 1
		coin := sdk.NewInt64Coin(demon, amount)
		coins = coins.Add(coin)
	}
	return coins
}

func GenRandomRewardGauge(r *rand.Rand) *itypes.RewardGauge {
	coins := GenRandomCoins(r)
	return itypes.NewRewardGauge(coins)
}

func GenRandomWithdrawnCoins(r *rand.Rand, coins sdk.Coins) sdk.Coins {
	withdrawnCoins := sdk.NewCoins()
	for _, coin := range coins {
		// skip this coin with some probability
		if OneInN(r, 3) {
			continue
		}
		// a subset of the coin has been withdrawn
		amount := coin.Amount.Uint64()
		withdrawnAmount := RandomInt(r, int(amount)-1) + 1
		withdrawnCoin := sdk.NewCoin(coin.Denom, sdk.NewIntFromUint64(withdrawnAmount))
		withdrawnCoins = withdrawnCoins.Add(withdrawnCoin)
	}
	return withdrawnCoins
}

func GenRandomGauge(r *rand.Rand) *itypes.Gauge {
	coins := GenRandomCoins(r)
	return itypes.NewGauge(coins)
}
