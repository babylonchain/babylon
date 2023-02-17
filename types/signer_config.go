package types

import (
	"errors"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

func ParseKeyNameFromConfig(opts servertypes.AppOptions) string {
	valueInterface := opts.Get("signer-config.key-name")
	if valueInterface == nil {
		panic("Signer key name should be provided in options")
	}
	keyName, err := cast.ToStringE(valueInterface)
	if err != nil {
		panic("Signer key name should be valid string")
	}

	return keyName
}

func parseGasPriceFromConfig(opts servertypes.AppOptions) (string, error) {
	valueInterface := opts.Get("signer-config.gas-price")
	if valueInterface == nil {
		return "", errors.New("signer gas price should be provided in options")
	}
	gasPrice, err := cast.ToStringE(valueInterface)
	if err != nil {
		return "", errors.New("signer gas price should be valid string")
	}

	coin, err := sdk.ParseDecCoin(gasPrice)
	if err != nil {
		return "", errors.New("signer gas price is invalid")
	}

	if !coin.Amount.IsPositive() {
		return "", errors.New("gas price should be positive")
	}

	return gasPrice, nil
}

func parseGasAdjustmentFromConfig(opts servertypes.AppOptions) (float64, error) {
	valueInterface := opts.Get("signer-config.gas-adjustment")
	if valueInterface == nil {
		return 0, errors.New("signer gas adjustment should be provided in options")
	}
	gasAdjustment, err := cast.ToFloat64E(valueInterface)
	if err != nil {
		return 0, errors.New("signer gas adjustment should be valid float number")
	}

	if gasAdjustment <= 1 {
		return 0, errors.New("signer gas adjustment should be more than 1")
	}

	return gasAdjustment, nil
}

// MustGetGasSettings reads GasPrice and GasAdjustment from app.toml file
func MustGetGasSettings(configPath string, v *viper.Viper) (string, float64) {
	var (
		gasPrice      string
		gasAdjustment float64
		err           error
	)

	v.AddConfigPath(configPath)
	v.SetConfigName("app")
	v.SetConfigType("toml")

	if err := v.ReadInConfig(); err != nil {
		panic("failed to read app.toml")
	}

	gasPrice, err = parseGasPriceFromConfig(v)
	if err != nil {
		panic(err)
	}

	gasAdjustment, err = parseGasAdjustmentFromConfig(v)
	if err != nil {
		panic(err)
	}

	return gasPrice, gasAdjustment
}
