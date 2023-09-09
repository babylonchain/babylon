package datagen

import (
	"math/rand"

	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	bstypes "github.com/babylonchain/babylon/x/btcstaking/types"
	itypes "github.com/babylonchain/babylon/x/incentive/types"
	secp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
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
	return itypes.NewRewardGauge(coins...)
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
	return itypes.NewGauge(coins...)
}

func GenRandomBTCDelDistInfo(r *rand.Rand) *bstypes.BTCDelDistInfo {
	return &bstypes.BTCDelDistInfo{
		BabylonPk:   GenRandomAccount().GetPubKey().(*secp256k1.PubKey),
		VotingPower: RandomInt(r, 1000) + 1,
	}
}

func GenRandomBTCValDistInfo(r *rand.Rand) (*bstypes.BTCValDistInfo, error) {
	// create BTC validator with random commission
	btcVal, err := GenRandomBTCValidator(r)
	if err != nil {
		return nil, err
	}
	// create BTC validator distribution info
	btcValDistInfo := bstypes.NewBTCValDistInfo(btcVal)
	// add a random number of BTC delegation distribution info
	numBTCDels := RandomInt(r, 100) + 1
	for i := uint64(0); i < numBTCDels; i++ {
		btcDelDistInfo := GenRandomBTCDelDistInfo(r)
		btcValDistInfo.BtcDels = append(btcValDistInfo.BtcDels, btcDelDistInfo)
		btcValDistInfo.TotalVotingPower += btcDelDistInfo.VotingPower
	}
	return btcValDistInfo, nil
}

func GenRandomBTCStakingRewardDistCache(r *rand.Rand) (*bstypes.RewardDistCache, error) {
	rdc := bstypes.NewRewardDistCache()
	// a random number of BTC validators
	numBTCVals := RandomInt(r, 10) + 1
	for i := uint64(0); i < numBTCVals; i++ {
		v, err := GenRandomBTCValDistInfo(r)
		if err != nil {
			return nil, err
		}
		rdc.AddBTCValDistInfo(v)
	}
	return rdc, nil
}

func GenRandomCheckpointAddressPair(r *rand.Rand) *btcctypes.CheckpointAddressPair {
	return &btcctypes.CheckpointAddressPair{
		Submitter: GenRandomAccount().GetAddress(),
		Reporter:  GenRandomAccount().GetAddress(),
	}
}

func GenRandomBTCTimestampingRewardDistInfo(r *rand.Rand) *btcctypes.RewardDistInfo {
	best := GenRandomCheckpointAddressPair(r)
	numOthers := RandomInt(r, 10)
	others := []*btcctypes.CheckpointAddressPair{}
	for i := uint64(0); i < numOthers; i++ {
		others = append(others, GenRandomCheckpointAddressPair(r))
	}
	return btcctypes.NewRewardDistInfo(best, others...)
}
