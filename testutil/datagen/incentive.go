package datagen

import (
	"math/rand"

	sdkmath "cosmossdk.io/math"
	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	bstypes "github.com/babylonchain/babylon/x/btcstaking/types"
	itypes "github.com/babylonchain/babylon/x/incentive/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
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
		withdrawnAmount := coin.Amount.Uint64()
		if withdrawnAmount > 1 {
			withdrawnAmount = RandomInt(r, int(withdrawnAmount)-1) + 1
		}
		withdrawnCoin := sdk.NewCoin(coin.Denom, sdkmath.NewIntFromUint64(withdrawnAmount))
		withdrawnCoins = withdrawnCoins.Add(withdrawnCoin)
	}
	return withdrawnCoins
}

func GenRandomGauge(r *rand.Rand) *itypes.Gauge {
	coins := GenRandomCoins(r)
	return itypes.NewGauge(coins...)
}

func GenRandomBTCDelDistInfo(r *rand.Rand) (*bstypes.BTCDelDistInfo, error) {
	btcPK, err := GenRandomBIP340PubKey(r)
	if err != nil {
		return nil, err
	}
	return &bstypes.BTCDelDistInfo{
		BtcPk:       btcPK,
		BabylonPk:   GenRandomAccount().GetPubKey().(*secp256k1.PubKey),
		VotingPower: RandomInt(r, 1000) + 1,
	}, nil
}

func GenRandomFinalityProviderDistInfo(r *rand.Rand) (*bstypes.FinalityProviderDistInfo, error) {
	// create finality provider with random commission
	fp, err := GenRandomFinalityProvider(r)
	if err != nil {
		return nil, err
	}
	// create finality provider distribution info
	fpDistInfo := bstypes.NewFinalityProviderDistInfo(fp)
	// add a random number of BTC delegation distribution info
	numBTCDels := RandomInt(r, 100) + 1
	for i := uint64(0); i < numBTCDels; i++ {
		btcDelDistInfo, err := GenRandomBTCDelDistInfo(r)
		if err != nil {
			return nil, err
		}
		fpDistInfo.BtcDels = append(fpDistInfo.BtcDels, btcDelDistInfo)
		fpDistInfo.TotalVotingPower += btcDelDistInfo.VotingPower
	}
	return fpDistInfo, nil
}

func GenRandomVotingPowerDistCache(r *rand.Rand, maxFPs uint32) (*bstypes.VotingPowerDistCache, error) {
	dc := bstypes.NewVotingPowerDistCache()
	// a random number of finality providers
	numFps := RandomInt(r, 10) + 1
	for i := uint64(0); i < numFps; i++ {
		v, err := GenRandomFinalityProviderDistInfo(r)
		if err != nil {
			return nil, err
		}
		dc.AddFinalityProviderDistInfo(v)
	}
	dc.ApplyActiveFinalityProviders(maxFPs)
	return dc, nil
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
