#!/bin/bash

set -ex

RELAYER_CONF_DIR=/root/.rly

# Initialize Cosmos relayer configuration
mkdir -p $RELAYER_CONF_DIR
rly --home $RELAYER_CONF_DIR config init
RELAYER_CONF=$RELAYER_CONF_DIR/config/config.yaml

# Setup Cosmos relayer configuration
cat <<EOF >$RELAYER_CONF
global:
    api-listen-addr: :5183
    timeout: 20s
    memo: ""
    light-cache-size: 10
chains:
    bbn-a:
        type: cosmos
        value:
            key-directory: $RELAYER_CONF_DIR/keys/$BBN_A_E2E_CHAIN_ID
            key: val01-bbn-a
            chain-id: $BBN_A_E2E_CHAIN_ID
            rpc-addr: http://$BBN_A_E2E_VAL_HOST:26657
            account-prefix: bbn
            keyring-backend: test
            gas-adjustment: 1.5
            gas-prices: 0.002ubbn
            min-gas-amount: 1
            debug: true
            timeout: 10s
            output-format: json
            sign-mode: direct
            extra-codecs: []
    bbn-b:
        type: cosmos
        value:
            key-directory: $RELAYER_CONF_DIR/keys/$BBN_B_E2E_CHAIN_ID
            key: val01-bbn-b
            chain-id: $BBN_B_E2E_CHAIN_ID
            rpc-addr: http://$BBN_B_E2E_VAL_HOST:26657
            account-prefix: bbn
            keyring-backend: test
            gas-adjustment: 1.5
            gas-prices: 0.002ubbn
            min-gas-amount: 1
            debug: true
            timeout: 10s
            output-format: json
            sign-mode: direct
            extra-codecs: []
paths:
    bbna-bbnb:
        src:
            chain-id: $BBN_A_E2E_CHAIN_ID
        dst:
            chain-id: $BBN_B_E2E_CHAIN_ID
EOF

# Import keys
rly -d --home $RELAYER_CONF_DIR keys restore bbn-a val01-bbn-a "$BBN_A_E2E_VAL_MNEMONIC"
rly -d --home $RELAYER_CONF_DIR keys restore bbn-b val01-bbn-b "$BBN_B_E2E_VAL_MNEMONIC"
sleep 3

# Start Cosmos relayer
echo "Creating IBC light clients, connection, and channel between the two CZs"
rly -d --home $RELAYER_CONF_DIR tx link bbna-bbnb --src-port ${CHAIN_A_IBC_PORT} --dst-port ${CHAIN_B_IBC_PORT} --order ordered --version zoneconcierge-1
echo "Created IBC channel successfully!"
sleep 10

rly -d --home $RELAYER_CONF_DIR start bbna-bbnb --debug-addr ""
