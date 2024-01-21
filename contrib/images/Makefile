all: babylond

babylond: babylond-rmi
	docker build --tag babylonchain/babylond -f babylond/Dockerfile \
    		$(shell git rev-parse --show-toplevel)

babylond-rmi:
	docker rmi babylonchain/babylond 2>/dev/null; true

.PHONY: all babylond babylond-rmi
