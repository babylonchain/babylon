package cli

import (
	"fmt"
	"math"

	sdkmath "cosmossdk.io/math"
	"github.com/btcsuite/btcd/btcutil"
)

func parseLockTime(str string) (uint16, error) {
	num, ok := sdkmath.NewIntFromString(str)

	if !ok {
		return 0, fmt.Errorf("invalid staking time: %s", str)
	}

	if !num.IsUint64() {
		return 0, fmt.Errorf("staking time is not valid uint")
	}

	asUint64 := num.Uint64()

	if asUint64 > math.MaxUint16 {
		return 0, fmt.Errorf("staking time is too large. Max is %d", math.MaxUint16)
	}

	return uint16(asUint64), nil
}

func parseBtcAmount(str string) (btcutil.Amount, error) {
	num, ok := sdkmath.NewIntFromString(str)

	if !ok {
		return 0, fmt.Errorf("invalid staking value: %s", str)
	}

	if num.IsNegative() {
		return 0, fmt.Errorf("staking value is negative")
	}

	if !num.IsInt64() {
		return 0, fmt.Errorf("staking value is not valid uint")
	}

	asInt64 := num.Int64()

	return btcutil.Amount(asInt64), nil
}
