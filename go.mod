go 1.17

module github.com/babylonchain/babylon

require (
	github.com/cosmos/cosmos-sdk v0.45.4
	github.com/gogo/protobuf v1.3.3
	github.com/golang/mock v1.6.0
	github.com/golang/protobuf v1.5.2
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/rakyll/statik v0.1.7
	github.com/spf13/cast v1.4.1
	github.com/spf13/cobra v1.4.0
	github.com/spf13/viper v1.10.1
	github.com/stretchr/testify v1.7.1
	github.com/tendermint/tendermint v0.34.19
	github.com/tendermint/tm-db v0.6.6
	google.golang.org/genproto v0.0.0-20220519153652-3a47de7e79bd
	google.golang.org/grpc v1.46.2
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/google/go-cmp v0.5.8 // indirect
	golang.org/x/net v0.0.0-20220127200216-cd36cc0744dd // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/99designs/keyring => github.com/cosmos/keyring v1.1.7-0.20210622111912-ef00f8ac3d76

	// Fix upstream GHSA-h395-qcrw-5vmq vulnerability.
	// TODO Remove it: https://github.com/cosmos/cosmos-sdk/issues/10409
	github.com/gin-gonic/gin => github.com/gin-gonic/gin v1.7.0
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1

	// latest grpc doesn't work with with our modified proto compiler, so we need to enforce
	// the following version across all dependencies.
	google.golang.org/grpc => google.golang.org/grpc v1.33.2
)
