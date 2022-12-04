# End-to-end Tests

## Structure

### `e2e` Package

The `e2e` package defines an integration testing suite used for full
end-to-end testing functionality. The package is copy of Osmosis e2e testing
approach.


### Common Problems

Please note that if the tests are stopped mid-way, the e2e framework might fail to start again due to duplicated containers. Make sure that
containers are removed before running the tests again: `docker containers rm -f $(docker containers ls -a -q)`.

Additionally, Docker networks do not get auto-removed. Therefore, you can manually remove them by running `docker network prune`.
