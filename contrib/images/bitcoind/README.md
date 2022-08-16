# bitcoind

`bitcoind` is a Docker image we build to be able to run a single Bitcoin node in the local devnet that produces blocks at regular intervals without consuming CPU.

To achieve this we use [btcd](https://github.com/btcsuite/btcd) in `--simnet` mode and make calls to the [generate](https://github.com/btcsuite/btcd/blob/v0.23.1/rpcserver.go#L886) API endpoint.

The image also runs a [btcwallet](https://github.com/btcsuite/btcwallet) process which can be used to make transfers. To talk to `btcd` programmatically we can use the [rpcclient](https://github.com/btcsuite/btcd/blob/v0.23.1/rpcclient/mining.go#L54-L62), and for the wallet the [chain](https://github.com/btcsuite/btcwallet/tree/master/chain) client.

See more at:
* https://gist.github.com/davecgh/2992ed85d41307e794f6
* http://piotrpasich.com/how-to-setup-own-bitcoin-simulation-network/
* https://blog.krybot.com/a?ID=00950-ef39d506-48ea-45df-81c5-2115e2f4a0f6
