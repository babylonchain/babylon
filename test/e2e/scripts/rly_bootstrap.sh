#!/bin/bash

set -ex

RELAYER_CONF_DIR=/root/.rly

# Initialize Cosmos relayer configuration
mkdir -p $RELAYER_CONF_DIR
rly --home $RELAYER_CONF_DIR config init
RELAYER_CONF=$RELAYER_CONF_DIR/config/config.yaml

#echo $BBN_A_E2E_VAL_MNEMONIC >$RELAYER_CONF_DIR/BBN_A_MNEMONIC.txt
#echo $BBN_B_E2E_VAL_MNEMONIC >$RELAYER_CONF_DIR/BBN_B_MNEMONIC.txt

# Setup Cosmos relayer configuration
cat <<EOF >$RELAYER_CONF
global:
    api-listen-addr: :5183
    timeout: 20s
    memo: ""
    light-cache-size: 10
chains:
    $BBN_A_E2E_CHAIN_ID:
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
            max-gas-amount: 0
            debug: true
            timeout: 10s
            block-timeout: ""
            output-format: json
            sign-mode: direct
            extra-codecs: []
            coin-type: null
            signing-algorithm: ""
            broadcast-mode: batch
            min-loop-duration: 0s
            extension-options: []
            feegrants: null
    $BBN_B_E2E_CHAIN_ID:
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
            max-gas-amount: 0
            debug: true
            timeout: 10s
            block-timeout: ""
            output-format: json
            sign-mode: direct
            extra-codecs: []
            coin-type: null
            signing-algorithm: ""
            broadcast-mode: batch
            min-loop-duration: 0s
            extension-options: []
            feegrants: null
paths:
    babylond:
        src:
            chain-id: $BBN_A_E2E_CHAIN_ID
            # client-id: 07-tendermint-0
            # connection-id: connection-0
            # port-id: $CHAIN_A_IBC_PORT
            # order: ordered
            # version: zoneconcierge-1
        dst:
            chain-id: $BBN_B_E2E_CHAIN_ID
            # client-id: 07-tendermint-0
            # connection-id: connection-0
            # port-id: $CHAIN_B_IBC_PORT
            # order: ordered
            # version: zoneconcierge-1
        src-channel-filter:
            rule: ""
            channel-list: []
EOF

# Import keys
rly --home $RELAYER_CONF_DIR keys restore ${BBN_A_E2E_CHAIN_ID} val01-bbn-a "$BBN_A_E2E_VAL_MNEMONIC"
rly --home $RELAYER_CONF_DIR keys restore ${BBN_B_E2E_CHAIN_ID} val01-bbn-b "$BBN_B_E2E_VAL_MNEMONIC"
sleep 10

# Start Cosmos relayer
rly --home $RELAYER_CONF_DIR tx link babylond --src-port ${CHAIN_A_IBC_PORT} --dst-port ${CHAIN_B_IBC_PORT} --order ordered --version zoneconcierge-1
sleep 10

rly --home $RELAYER_CONF_DIR start babylond --debug-addr ""
