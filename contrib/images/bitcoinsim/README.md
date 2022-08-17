# bitcoinsim

`bitcoinsim` is a Docker image we build to be able to run a single Bitcoin node in the local devnet that produces blocks at regular intervals without consuming CPU.

To achieve this we use [btcd](https://github.com/btcsuite/btcd) in `--simnet` mode and make calls to the [generate](https://github.com/btcsuite/btcd/blob/v0.23.1/rpcserver.go#L886) API endpoint.

The image also runs a [btcwallet](https://github.com/btcsuite/btcwallet) process which can be used to make transfers. To talk to `btcd` programmatically we can use the [rpcclient](https://github.com/btcsuite/btcd/blob/v0.23.1/rpcclient/mining.go#L54-L62), and for the wallet the [chain](https://github.com/btcsuite/btcwallet/tree/master/chain) client.

See more at:
* https://gist.github.com/davecgh/2992ed85d41307e794f6
* http://piotrpasich.com/how-to-setup-own-bitcoin-simulation-network/
* https://blog.krybot.com/a?ID=00950-ef39d506-48ea-45df-81c5-2115e2f4a0f6


## Build

You can build the image with the following command:

```bash
make bitcoinsim
```

## Test

One way to see if it works is to run the container interactively:

```bash
docker run -it --rm --name bitcoinsim babylonchain/bitcoinsim
```

The logs should show that a wallet is created with and mining is started:

```console
$ docker run -it --rm babylonchain/bitcoinsim
Starting btcd...
Creating a wallet...
spawn btcwallet --simnet -u rpcuser -P rpcpass --create
2022-08-16 23:39:33.091 [INF] BTCD: Version 0.23.1-beta
2022-08-16 23:39:33.091 [INF] BTCD: Loading block database from '/root/.btcd/data/simnet/blocks_ffldb'
Enter the private passphrase for your new wallet:
Confirm passphrase:
Do you want to add an additional layer of encryption for public data? (n/no/y/yes) [no]: n
Do you have an existing wallet seed you want to use? (n/no/y/yes) [no]: n
Your wallet generation seed is:
b46944f9dd96792481ec71c4d945a3e79959e429c5b77fcda14e587cabebc672
IMPORTANT: Keep the seed in a safe place as you
will NOT be able to restore your wallet without it.
Please keep in mind that anyone who has access
to the seed can also restore your wallet thereby
giving them access to all your funds, so it is
imperative that you keep it in a secure location.
Once you have stored the seed in a safe and secure location, enter "OK" to continue: OK
Creating the wallet...
2022-08-16 23:39:33.114 [INF] BTCD: Block database loaded
2022-08-16 23:39:33.116 [INF] INDX: Committed filter index is enabled
2022-08-16 23:39:33.116 [INF] INDX: Catching up indexes from height -1 to 0
2022-08-16 23:39:33.116 [INF] INDX: Indexes caught up to height 0
2022-08-16 23:39:33.116 [INF] CHAN: Chain state (height 0, hash 683e86bd5c6d110d91b94b97137ba6bfe02dbbdb8e3dff722a669b5d69d77af6, totaltx 1, work 2)
2022-08-16 23:39:33.117 [INF] RPCS: Generating TLS certificates...
2022-08-16 23:39:33.136 [INF] RPCS: Done generating TLS certificates
2022-08-16 23:39:33.141 [INF] AMGR: Loaded 0 addresses from file '/root/.btcd/data/simnet/peers.json'
2022-08-16 23:39:33.141 [INF] RPCS: RPC server listening on 0.0.0.0:18556
2022-08-16 23:39:33.141 [INF] CMGR: Server listening on 0.0.0.0:18555
2022-08-16 23:39:35.814 [INF] WLLT: Opened wallet
The wallet has been created successfully.
Starting btcwallet...
2022-08-16 23:39:35.822 [INF] BTCW: Version 0.15.1-alpha
2022-08-16 23:39:35.822 [INF] BTCW: Generating TLS certificates...
2022-08-16 23:39:35.846 [INF] BTCW: Done generating TLS certificates
2022-08-16 23:39:35.846 [INF] RPCS: Listening on 0.0.0.0:18554
2022-08-16 23:39:35.846 [INF] BTCW: Attempting RPC client connection to localhost:18556
2022-08-16 23:39:35.872 [INF] RPCS: New websocket client 127.0.0.1:49538
2022-08-16 23:39:35.872 [INF] CHNS: Established connection to RPC server localhost:18556
2022-08-16 23:39:36.768 [INF] WLLT: Opened wallet
2022-08-16 23:39:36.771 [INF] WLLT: RECOVERY MODE ENABLED -- rescanning for used addresses with recovery_window=250
2022-08-16 23:39:36.790 [INF] WLLT: Started rescan from block 683e86bd5c6d110d91b94b97137ba6bfe02dbbdb8e3dff722a669b5d69d77af6 (height 0) for 0 addresses
2022-08-16 23:39:36.791 [INF] RPCS: Beginning rescan for 0 addresses
2022-08-16 23:39:36.791 [INF] RPCS: Skipping rescan as client has no addrs/utxos
2022-08-16 23:39:36.791 [INF] RPCS: Finished rescan
2022-08-16 23:39:36.791 [INF] WLLT: Catching up block hashes to height 0, this might take a while
2022-08-16 23:39:36.792 [INF] WLLT: Done catching up block hashes
2022-08-16 23:39:36.792 [INF] WLLT: Finished rescan for 0 addresses (synced to block 683e86bd5c6d110d91b94b97137ba6bfe02dbbdb8e3dff722a669b5d69d77af6, height 0)
Creating miner address...
Restarting btcd with mining address Sjfg9uCaTqSR7UvGRuuMRJtUtWRhoBA7YX...
2022-08-16 23:39:40.861 [ERR] CHNS: Websocket receive error from localhost:18556: websocket: close 1006 unexpected EOF
2022-08-16 23:39:40.862 [INF] CHNS: Failed to connect to localhost:18556: dial tcp 127.0.0.1:18556: connect: connection refused
2022-08-16 23:39:40.862 [INF] CHNS: Retrying connection to localhost:18556 in 5s
2022-08-16 23:39:40.873 [INF] BTCD: Version 0.23.1-beta
2022-08-16 23:39:40.873 [INF] BTCD: Loading block database from '/root/.btcd/data/simnet/blocks_ffldb'
2022-08-16 23:39:40.897 [INF] BCDB: Detected unclean shutdown - Repairing...
2022-08-16 23:39:40.900 [INF] BCDB: Database sync complete
2022-08-16 23:39:40.901 [INF] BTCD: Block database loaded
2022-08-16 23:39:40.910 [INF] INDX: Committed filter index is enabled
2022-08-16 23:39:40.913 [INF] INDX: Catching up indexes from height -1 to 0
2022-08-16 23:39:40.913 [INF] INDX: Indexes caught up to height 0
2022-08-16 23:39:40.913 [INF] CHAN: Chain state (height 0, hash 683e86bd5c6d110d91b94b97137ba6bfe02dbbdb8e3dff722a669b5d69d77af6, totaltx 1, work 2)
2022-08-16 23:39:40.927 [INF] AMGR: Loaded 0 addresses from file '/root/.btcd/data/simnet/peers.json'
2022-08-16 23:39:40.927 [INF] RPCS: RPC server listening on 0.0.0.0:18556
2022-08-16 23:39:40.927 [INF] CMGR: Server listening on 0.0.0.0:18555
Generating enought blocks for the first coinbase to mature...
2022-08-16 23:39:45.890 [INF] RPCS: New websocket client 127.0.0.1:49552
2022-08-16 23:39:45.890 [INF] CHNS: Reestablished connection to RPC server localhost:18556
2022-08-16 23:39:45.890 [INF] WLLT: RECOVERY MODE ENABLED -- rescanning for used addresses with recovery_window=250
2022-08-16 23:39:45.891 [WRN] CHNS: Received unexpected reply: {"hash":"683e86bd5c6d110d91b94b97137ba6bfe02dbbdb8e3dff722a669b5d69d77af6","height":0} (id 16)
2022-08-16 23:39:45.896 [INF] MINR: Block submitted via CPU miner accepted (hash 3650f1363cc7fc4c3f01d1d904caa9f9687ef08927498658b6c50b585edb7872, amount 50 BTC)
...
2022-08-16 23:39:45.920 [INF] MINR: Block submitted via CPU miner accepted (hash 7a6dc189c7ffc38db73304e2993dacb241a932fc3980ca840da466b58e2d26d8, amount 50 BTC)
[
  "3650f1363cc7fc4c3f01d1d904caa9f9687ef08927498658b6c50b585edb7872",
  "243d889cc77cbabcb4d28a3c5224dc3c0ef6970d0ea9f7e143e1216abb87c0d4",
  "1577ee934812dace99bbab757ce6ccbed443da0863baab3b8ba7efe1b4453917",
  ...
  "63cb51a672732ec1175b1ed288d20cd8f1cb8653fdd7dc31f90fbdb9844c7998",
  "7a6dc189c7ffc38db73304e2993dacb241a932fc3980ca840da466b58e2d26d8"
]
2022-08-16 23:39:46.148 [INF] WLLT: Catching up block hashes to height 32, this might take a while
2022-08-16 23:39:46.151 [INF] WLLT: Done catching up block hashes
2022-08-16 23:39:46.151 [INF] WLLT: Finished rescan for 1 address (synced to block 42ad6a44786a2c6edcfcca7098b3d18cc13c668a004c494ca87566c404f9ae34, height 32)
Checking balance...
50
Generating a block every 30 seconds.
Press [CTRL+C] to stop...
2022-08-16 23:39:50.975 [INF] MINR: Block submitted via CPU miner accepted (hash 1e4931b5b590e297f736da282ab1367f1704ff5a21943c3883161d4f0a60dda2, amount 50 BTC)
[
  "1e4931b5b590e297f736da282ab1367f1704ff5a21943c3883161d4f0a60dda2"
]

```

Then we can connect to this from another container to query the balance:

```console
$ docker ps
CONTAINER ID        IMAGE                    COMMAND                  CREATED             STATUS              PORTS               NAMES
93e58681637d        babylonchain/bitcoinsim   "/bitcoinsim/wrapper.sh"   5 minutes ago       Up 5 minutes        18554-18556/tcp     bitcoinsim
$ docker exec -it bitcoinsim sh
/bitcoinsim # btcctl --simnet --wallet -u $RPCUSER -P $RPCPASS getbalance
600
/bitcoinsim #

```

The balance will go up as more blocks mature. The default coinbase maturity is 100 blocks.

## Use in docker-compose

The image has been added to the main `docker-compose.yml` file. It will build it if it's not built already and start running it as part of the local testnet. The other containers can use the default ports to connect to it, which are also exposed on the host.

It can be started on its own like so:

```bash
docker-compose up -d bitcoinsim
```

Currently the image doesn't support restarting, the container has to be completely removed and recreated if it's stopped. It can be removed with the following command:

```bash
docker-compose stop bitcoinsim
docker-compose rm bitcoinsim
```

The ports are mapped to the same default ports that `btcd` and `btcwallet` would use, so if the host already runs these services they might clash.
