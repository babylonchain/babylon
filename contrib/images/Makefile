COSMOS_RELAYER_TAG := $(shell grep '^ENV RELAYER_TAG' cosmos-relayer/Dockerfile | cut -f3 -d\ )
HERMES_RELAYER_TAG := $(shell grep '^ENV RELAYER_TAG' hermes/Dockerfile | cut -f3 -d\ | cut -f2 -d\/)

all: babylond cosmos-relayer hermes-relayer

babylond: babylond-rmi
	docker build --tag babylonchain/babylond -f babylond/Dockerfile \
    		$(shell git rev-parse --show-toplevel)

babylond-rmi:
	docker rmi babylonchain/babylond 2>/dev/null; true

hermes-relayer: hermes-relayer-rmi
	docker build --tag babylonchain/hermes-relayer:${HERMES_RELAYER_TAG} -f hermes/Dockerfile \
				$(shell git rev-parse --show-toplevel)/contrib/images/hermes
	docker tag babylonchain/hermes-relayer:${HERMES_RELAYER_TAG} babylonchain/hermes-relayer:latest

hermes-relayer-rmi:
	docker rmi babylonchain/hermes-relayer 2>/dev/null; true

cosmos-relayer: cosmos-relayer-rmi
	docker build --tag babylonchain/cosmos-relayer:${COSMOS_RELAYER_TAG} -f cosmos-relayer/Dockerfile \
				$(shell git rev-parse --show-toplevel)/contrib/images/cosmos-relayer
	docker tag babylonchain/cosmos-relayer:${COSMOS_RELAYER_TAG} babylonchain/cosmos-relayer:latest

cosmos-relayer-rmi:
	docker rmi babylonchain/cosmos-relayer 2>/dev/null; true

.PHONY: all babylond cosmos-relayer babylond-rmi cosmos-relayer-rmi
