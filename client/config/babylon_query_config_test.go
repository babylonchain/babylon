package config_test

import (
	"testing"

	"github.com/babylonchain/babylon/client/config"
	"github.com/stretchr/testify/require"
)

// TestBabylonQueryConfig ensures that the default Babylon query config is valid
func TestBabylonQueryConfig(t *testing.T) {
	defaultConfig := config.DefaultBabylonQueryConfig()
	err := defaultConfig.Validate()
	require.NoError(t, err)
}
