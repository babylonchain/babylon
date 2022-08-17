#!/usr/bin/env sh
set -euo pipefail
#set -x

# btcctl will be looking for this file, but the wallet doesn't create it.
mkdir -p /root/.btcwallet && touch /root/.btcwallet/btcwallet.conf
mkdir -p /root/.btcd      && touch /root/.btcd/btcd.conf

# Create a wallet with and a miner address, then mine enough blocks for the miner to have some initial balance.

echo "Starting btcd..."
btcd --simnet -u $RPCUSER -P $RPCPASS --rpclisten=0.0.0.0:18556 --listen=0.0.0.0:18555 2>&1 &
BTCD_PID=$!

echo "Creating a wallet..."
# Used autoexpect to create the wallet in the first instance.
# https://stackoverflow.com/questions/4857702/how-to-provide-password-to-a-command-that-prompts-for-one-in-bash
expect btcwallet_create.exp $RPCUSER $RPCPASS $PASSPHRASE

echo "Starting btcwallet..."
btcwallet --simnet -u $RPCUSER -P $RPCPASS --rpclisten=0.0.0.0:18554 2>&1 &
BTCWALLET_PID=$!

# Allow some time for the wallet to start
sleep 5

echo "Creating miner address..."
MINING_ADDR=$(btcctl --simnet --wallet -u $RPCUSER -P $RPCPASS getnewaddress)
echo $MINING_ADDR > mining.addr

echo "Restarting btcd with mining address $MINING_ADDR..."
kill -9 $BTCD_PID
sleep 1
btcd --simnet -u $RPCUSER -P $RPCPASS --rpclisten=0.0.0.0:18556 --listen=0.0.0.0:18555 --miningaddr=$MINING_ADDR 2>&1 &
BTCD_PID=$!

# Allow btcd to start
sleep 5

echo "Generating enough blocks for the first coinbase to mature..."
btcctl --simnet -u $RPCUSER -P $RPCPASS generate 100

# Allow some time for the wallet to catch up.
sleep 5

echo "Checking balance..."
btcctl --simnet --wallet -u $RPCUSER -P $RPCPASS getbalance

echo "Generating a block every ${GENERATE_INTERVAL_SECS} seconds."
echo "Press [CTRL+C] to stop..."
while true
do
  btcctl --simnet -u $RPCUSER -P $RPCPASS generate 1

  sleep ${GENERATE_INTERVAL_SECS}
done
