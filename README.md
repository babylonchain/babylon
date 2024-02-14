# Babylon

Unlocking 21 Million ₿ to Secure the Decentralized Economy

[![Website](https://badgen.net/badge/icon/Website?label=)](https://babylonchain.io)
[![Twitter](https://badgen.net/badge/icon/twitter?icon=twitter&label)](https://twitter.com/babylon_chain)
[![Discord](https://badgen.net/badge/icon/discord?icon=discord&label)](https://discord.com/invite/babylonglobal)
[![Medium](https://badgen.net/badge/icon/medium?icon=medium&label)](https://medium.com/babylonchain-io)

[Babylon](https://babylonchain.io) provides a suite of security-sharing
protocols between Bitcoin and the PoS world. It provides two inter-connected
protocols:

- **Bitcoin timestamping:** Submits succinct and verifiable timestamps of any
  data (such as PoS blockchains) to Bitcoin.
- **Bitcoin staking:** Enables Bitcoin holders to provide economic security to
  any decentralized system through trustless (and self-custodian) staking.

[![BTC staking
litepaper](https://badgen.net/badge/icon/BTC%20staking%20litepaper?label=)](https://docs.babylonchain.io/assets/files/btc_staking_litepaper-32bfea0c243773f0bfac63e148387aef.pdf)
[![BTC timestamping
whitepaper](https://badgen.net/badge/icon/BTC%20timestamping%20whitepaper?label=)](https://arxiv.org/abs/2207.08392)

## System requirements

This system spec has been tested by validators and found to be comfortable:

- Quad Core or larger AMD or Intel (amd64) CPU
- 32GB RAM
- 1TB NVMe Storage
- 100MBps bidirectional internet connection

You can run Babylon on lower-spec hardware for each component, but you may find
that it is not highly performant or prone to crashing.

## Build and install

To build and install, you need to have Go 1.21 available. Follow the
instructions on the [Golang page](https://go.dev/doc/install) to do that.

To build the binary:

```console
make build
```

The binary will then be available at `./build/babylond` .

To install the binary to system directories:

```console
make install
```

## Documentation

For user-facing documents, please visit
[docs.babylonchain.io](https://docs.babylonchain.io). For technical documents
about high-level designs of Babylon, please visit
[docs/README.md](./docs/README.md). Each module under `x/` also contains a
document about its design and implementation.

## Joining the testnet

Please follow the instructions on the [User
Guides](https://docs.babylonchain.io/docs/user-guides/overview).

## Contributing

The [docs](./docs) directory contains the necessary information on how to get
started using the babylond executable for development purposes.
