# Babylon

## Requirements

- Go 1.17
- Docker (for running a testnet)

## Building

To build the chain, simply:
```bash
>>> make build
```

This will lead to the creation of a `babylond` executable under the `build`
directory.

## Testing

```bash
>>> make test
```

## Running a testnet

### Single node testnet

First, generate the required testnet files under the `.testnet` directory
```bash
>>> ./build/babylond testnet --v 1 --output-dir ./.testnet --starting-ip-address 192.168.10.2 --keyring-backend test
```

This will lead to the creation of a `.testnet` directory that contains the
following:

```bash
>>> ls .testnet
gentxs node0
```

The `node0` directory contains the configuration for the single node. To start
running it, execute
```bash
>>> ./build/babylond start --home ./.testnet/node0/babylond
[logs]
```

### Multi node testnet

We provide support for running a multi-node testnet using Docker.
```bash
>>> make localnet-start
```

This will lead to the generation of a testnet with 4 nodes. The corresponding
node directories can be found under `.testnets`
```bash
>>> ls .testnets
gentxs node0 node1 node2 node3
```

The logs for a particular node can be found under
`.testnets/node{id}/babylond/babylond.log`.
