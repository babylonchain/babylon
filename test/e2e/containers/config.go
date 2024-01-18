package containers

// ImageConfig contains all images and their respective tags
// needed for running e2e tests.
type ImageConfig struct {
	RelayerRepository string
	RelayerTag        string
}

//nolint:deadcode
const (
	// name of babylon container produced by running `make localnet-build-env`
	BabylonContainerName = "babylonchain/babylond"

	hermesRelayerRepository = "informalsystems/hermes"
	hermesRelayerTag        = "master"
	cosmosRelayerRepository = "babylonchain/cosmos-relayer"
	cosmosRelayerTag        = "v2.4.2"
)

// NewImageConfig returns ImageConfig needed for running e2e test.
// If isUpgrade is true, returns images for running the upgrade
// If isFork is true, utilizes provided fork height to initiate fork logic
func NewImageConfig(isCosmosRelayer bool) ImageConfig {
	if isCosmosRelayer {
		config := ImageConfig{
			RelayerRepository: cosmosRelayerRepository,
			RelayerTag:        cosmosRelayerTag,
		}
		return config
	} else {
		config := ImageConfig{
			RelayerRepository: hermesRelayerRepository,
			RelayerTag:        hermesRelayerTag,
		}
		return config
	}
}
