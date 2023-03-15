DOCKER := $(shell which docker)
CUR_DIR := $(shell pwd)
CUR_BASENAME := $(shell basename $(CUR_DIR))

# TODO should use correct path to bindings library
build-optimized:
	$(DOCKER) run --rm -v "$(CUR_DIR)":/code \
		--mount type=bind,source="/Users/konradstaniec/Work/babylon/babylon/contracts/babylon-bindings",target=/babylon-bindings \
		--mount type=volume,source="$(CUR_BASENAME)_cache",target=/code/target \
		--mount type=volume,source=registry_cache,target=/usr/local/cargo/registry \
		cosmwasm/rust-optimizer-arm64:0.12.12
