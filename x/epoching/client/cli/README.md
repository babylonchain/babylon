## CLI of the epoching module

### Delegating bbn

```shell
$BABYLON_PATH/build/babylond --home $TESTNET_PATH/node0/babylond --chain-id chain-test \
         --keyring-backend test --fees 1bbn \
         --from node0 --broadcast-mode block \
         tx epoching delegate <val_addr> <amount_of_bbn>
```

### Undelegating bbn

```shell
$BABYLON_PATH/build/babylond --home $TESTNET_PATH/node0/babylond --chain-id chain-test \
         --keyring-backend test --fees 3bbn \
         --from node0 --broadcast-mode block \
         tx epoching unbond <val_addr> <amount_of_bbn>
```

### Redelegating bbn

```shell
$BABYLON_PATH/build/babylond --home $TESTNET_PATH/node0/babylond --chain-id chain-test \
         --keyring-backend test --fees 3bbn \
         --from node0 --broadcast-mode block \
         tx epoching redelegate <from_val_addr> <to_val_addr> <amount_of_bbn>
```
