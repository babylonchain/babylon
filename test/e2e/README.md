# End-to-end Tests

### `e2e` Package

The `e2e` package defines an integration testing suite used for full
end-to-end testing functionality. The package is copy of Osmosis e2e testing
approach.


### Wasm contract used for e2e testing

Wasm contract located in `bytecode/babylon_contract.wasm` is compiled from most recent commit `main` branch - https://github.com/babylonchain/babylon-contract

This contract uses feature specific to Babylon, through Babylon bindings library.

### Common Problems

Please note that if the tests are stopped mid-way, the e2e framework might fail to start again due to duplicated containers. Make sure that
containers are removed before running the tests again: `docker containers rm -f $(docker containers ls -a -q)`.

Additionally, Docker networks do not get auto-removed. Therefore, you can manually remove them by running `docker network prune`.
