## CLI of the epoching module

### Delegating stakes

```shell
$BABYLON_PATH/build/babylond --home $TESTNET_PATH/node0/babylond --chain-id chain-test \
         --keyring-backend test --fees 1stake \
         --from node0 --broadcast-mode block \
         tx epoching delegate <val_addr> <amount_of_stake>
```

### Undelegating stakes

```shell
$BABYLON_PATH/build/babylond --home $TESTNET_PATH/node0/babylond --chain-id chain-test \
         --keyring-backend test --fees 3stake \
         --from node0 --broadcast-mode block \
         tx epoching unbond <val_addr> <amount_of_stake>
```

### Redelegating stakes

```shell
$BABYLON_PATH/build/babylond --home $TESTNET_PATH/node0/babylond --chain-id chain-test \
         --keyring-backend test --fees 3stake \
         --from node0 --broadcast-mode block \
         tx epoching redelegate <from_val_addr> <to_val_addr> <amount_of_stake>
```
