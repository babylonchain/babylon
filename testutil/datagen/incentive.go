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

func GenRandomRewardGauge(r *rand.Rand) *itypes.RewardGauge {
	numCoins := r.Int31n(10) + 10
	coins := sdk.NewCoins()
	for i := int32(0); i < numCoins; i++ {
		demon := GenRandomDenom(r)
		amount := r.Int63n(10000)
		coin := sdk.NewInt64Coin(demon, amount)
		coins = coins.Add(coin)
	}
	return itypes.NewRewardGauge(coins...)
}
