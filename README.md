# Babylon

Bringing Bitcoin security to Cosmos and beyond.

[![Website](https://badgen.net/badge/icon/website?label=)](https://babylonchain.io)
[![Whitepaper](https://badgen.net/badge/icon/whitepaper?label=)](https://arxiv.org/abs/2207.08392)
[![Twitter](https://badgen.net/badge/icon/twitter?icon=twitter&label)](https://twitter.com/babylon_chain)
[![Discord](https://badgen.net/badge/icon/discord?icon=discord&label)](https://discord.gg/babylonchain)

## Build and install

The babylond application based on the [Cosmos SDK](https://github.com/cosmos/cosmos-sdk) is the main application of the Babylon network. 
This repository is used to build the Babylon core application to join the Babylon network.

### Requirements
To build and install, you need to have Go 1.19 available.
Follow the instructions on the [Golang page](https://go.dev/doc/install) to do that.

To build the binary:
```console
make build
```

The binary will then be available at `./build/babylond` .

To install:
```console
make install
```

## Documentation

For the most up-to-date documentation please visit [docs.babylonchain.io](https://docs.babylonchain.io)

## Joining the testnet

Please follow the instructions on the [Joining the Testnet documentation page](https://docs.babylonchain.io/docs/testnet/overview).

## Contributing

The [docs](./docs) directory contains the necessary information on how to get started using the babylond executable for development purposes.
