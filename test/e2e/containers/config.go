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
	// TODO: Replace with version tag once we have a working version
	hermesRelayerTag = "master"
	// Built using the `build-cosmos-relayer-docker` target on an Intel (amd64) machine and pushed to ECR
	cosmosRelayerRepository = "public.ecr.aws/t9e9i3h0/cosmos-relayer"
	// TODO: Replace with version tag once we have a working version
	cosmosRelayerTag = "v2.5.1"
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
