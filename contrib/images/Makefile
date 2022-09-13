all: babylond-env

babylond-env: babylond-rmi
	docker build --tag babylonchain/babylond -f babylond-env/Dockerfile \
    		$(shell git rev-parse --show-toplevel)

babylond-dlv: babylond-rmi
	docker build --tag babylonchain/babylond -f babylond-dlv/Dockerfile \
        	$(shell git rev-parse --show-toplevel)

babylond-rmi:
	docker rmi babylonchain/babylond 2>/dev/null; true

.PHONY: all babylond-env babylond-dlv babylond-rmi
