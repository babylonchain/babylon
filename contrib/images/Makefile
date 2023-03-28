all: babylond-env

babylond-env: babylond-rmi
	docker build --tag babylonchain/babylond -f babylond/Dockerfile \
    		$(shell git rev-parse --show-toplevel)

babylond-dlv: babylond-rmi
	docker build --tag babylonchain/babylond \
		--build-arg DLV_ENABLED=true --build-arg COSMOS_BUILD_OPTIONS="debug,nostrip" \
		-f babylond/Dockerfile \
        	$(shell git rev-parse --show-toplevel)

babylond-rmi:
	docker rmi babylonchain/babylond 2>/dev/null; true

.PHONY: all babylond-env babylond-dlv babylond-rmi
