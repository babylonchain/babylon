# Babylon Architecture

The Babylon system is composed of a Babylon node
built using the Cosmos SDK as well as peripheral programs
that facilitate BTC staking, finality round participation, and
communication with Bitcoin and other Consumer Zones.
![Babylon Architecture](./static/arch.png)

## Babylon Node Modules

### [Epoching](../x/epoching)

The Babylon blockchain is divided into epochs
that consist of a parameterized number of blocks.
Within each epoch, the validator set does not change.
This way, Babylon needs a checkpoint per epoch rather than per block,
which reduces the checkpointing costs.
The epoching module achieves this by delaying the execution
of transactions that affect the validator set to the last block
of each epoch.

### [BTC Light Client](../x/btclightclient)

The BTC Light Client module receives Bitcoin headers
reported by the Vigilante Reporter and
maintains a BTC header chain based on the PoW rules of Bitcoin.
It exposes information about the canonical Bitcoin chain,
the depth of headers, and
whether the inclusion evidence for a Bitcoin transaction is valid.

### [BTC Checkpoint](../x/btccheckpoint)

The BTC Checkpoint module verifies Babylon’s BTC checkpoints
reported by the Vigilante Reporter, and
provides the confirmation status of these checkpoints to the Checkpointing
module based on their depth according to the BTC Light Client module.

### [Checkpointing](../x/checkpointing)

The checkpointing module is responsible for creating Babylon checkpoints
to be submitted to Bitcoin and maintaining their confirmation status.
It collects the validator's
[BLS signatures](https://en.wikipedia.org/wiki/BLS_digital_signature)
for each block to be checkpointed and aggregates them
into a BLS multisignature to include in the Bitcoin checkpoint.
The confirmation status of each checkpoint is determined by
Bitcoin checkpoint inclusion information retrieved from the
BTC checkpoint module.

### [ZoneConcierge](../x/zoneconcierge)

The Zone Concierge module
extracts verified Consumer Zone headers from
connected [IBC light clients](https://github.com/cosmos/ibc-go) and
maintains their Bitcoin confirmation status based on the
Bitcoin confirmation status of the
Babylon transactions that carry them.
It communicates the Bitcoin confirmation status to the Consumer Zone
using verifiable proofs through an
[IBC](https://github.com/cosmos/ibc-go) connection.

### [BTC Staking](../x/btcstaking)

The BTC Staking module
is the bookkeeper for the BTC staking protocol.
It is responsible for verifying and activating
BTC staking requests and
maintaining the active finality provider set.
It communicates with the BTC Light Client module
to extract the confirmation status of staking requests and
receives notifications about on-demand unlocked stake from the
BTC Staking Monitor.

### [Finality](../x/finality)

The Finality module is responsible for finalizing blocks
produced by the CometBFT consensus.
It receives and verifies finality round votes
from finality providers and
a block is considered finalized if sufficient
voting power is cast on it.
The voting power of each finality provider is based on
its Bitcoin stake retrieved from the BTC Staking module.
Finality votes are performed using
[Extractable-One-Time-Signatures (EOTS)](https://docs.babylonchain.io/assets/files/btc_staking_litepaper-32bfea0c243773f0bfac63e148387aef.pdf)
and verified using
the finality providers' committed public randomness.

### [Incentive](../x/incentive)

The incentive module consumes a percentage
of the rewards intended for Babylon stakers and
distributes it as rewards to Bitcoin stakers and
vigilantes.

## Vigilantes

The vigilante suite of programs acts as a
relayer of data between Babylon and Bitcoin.
Babylon's secure operation requires
that at least one honest
operator of each of the programs exist.
Otherwise,
an alarm will be raised by the monitor program.

### [Vigilante Submitter](https://github.com/babylonchain/vigilante)

A standalone program that submits
Babylon checkpoints to Bitcoin as
Bitcoin transactions embedding data
utilizing the `OP_RETURN` Bitcoin script code.

### [Vigilante Reporter](https://github.com/babylonchain/vigilante)

A standalone program that scans
the Bitcoin ledger for Bitcoin headers and Babylon checkpoints,
and reports them back to Babylon using Babylon transactions.

## Monitors

The monitor programs suite is responsible for
monitoring the consistency between Babylon's state and
Bitcoin.

### [Checkpointing Monitor](https://github.com/babylonchain/vigilante)

A standalone program that monitors:

- The consistency between the Bitcoin canonical chain and
  the Bitcoin header chain maintained by
  Babylon's BTC Light client module.
- The timely inclusion of Babylon's Bitcoin checkpoints
  information in the Babylon ledger.

### [BTC Staking Monitor](https://github.com/babylonchain/vigilante)

A standalone program that monitors:

- The execution of BTC Staking on-demand unbonding transactions
  on the Bitcoin ledger to inform Babylon about them.
- The execution of BTC Staking slashing transactions in the case
  of a finality provider double voting.
  In the case of non-execution the monitor extracts the finality provider's
  private key and executes the slashing.
- The execution of a selective slashing attack launched
  by a finality provider. In this case,
  the monitor extracts the finality provider's private key
  and slashes them.

## BTC Staking Programs

The BTC Staking programs suite
involves components that enable the function
Bitcoin Stakers and Finality Providers
while also ensuring their adherence to the protocol.

### BTC Staker

Bitcoin holders can stake their Bitcoin
by creating a set of Bitcoin transactions,
including them to the Bitcoin ledger, and
then informing Babylon about their staking.
Later, they can also on-demand unlock or
withdraw their funds when their stake expires.
The following set of standalone programs
has been developed to enable these functionalities:

- [BTC Staker Daemon](https://github.com/babylonchain/btc-staker):
  Daemon program connecting to a Bitcoin wallet and Babylon.
- [BTC Staker Dashboard](https://github.com/babylonchain/btc-staking-dashboard):
  Web application connecting to a Bitcoin wallet extension and the Babylon API.
  Should only be used for testing purposes.
- Wallet Integrations (TBD)

### [Finality Provider](https://github.com/babylonchain/finality-provider)

A standalone program that allows for the registration and
maintenance of a finality provider.
It monitors for a finality provider's inclusion in the active set, commits
[Extractable One Time Signature (EOTS)](https://docs.babylonchain.io/assets/files/btc_staking_litepaper-32bfea0c243773f0bfac63e148387aef.pdf)
public randomness, and
submits finality votes for blocks.
Finality votes are created through a connection to a standalone
[EOTS manager daemon](https://github.com/babylonchain/finality-provider)
responsible for securely maintaining the
finality provider's private keys.

### [Covenant Emulator](https://github.com/babylonchain/covenant-emulator)

A standalone program utilized by the covenant emulation committee members.
It emulates [covenant](https://covenants.info) functionality by monitoring
for pending staking requests,
verifying their contents, and
submitting necessary signatures.

## Consumer Zones

### IBC Relayer

The IBC Relayer maintains the
[IBC protocol](https://cosmos.network/ibc/) connection
between Babylon and other Consumer Zones (CZs).
It is responsible for updating the CZ's light client
inside the Babylon ledger to enable checkpointing and
propagating checkpoint information to the Babylon smart contract
deployed within the CZ.

There are different IBC relayer implementations that can achieve
this function. Most notably:

- [Cosmos Relayer](https://github.com/cosmos/relayer):
  A fully functional relayer written in Go.
- [Babylon Relayer](https://github.com/babylonchain/babylon-relayer/):
  A wrapper of the Cosmos Relayer that can maintain a one-way IBC connection.
  It is recommended to be used when the Consumer Zone does not deploy the
  Babylon smart contract.
- [Hermes Relayer](https://github.com/informalsystems/hermes):
  A fully functional relayer written in Rust.

### [Babylon Contract](https://github.com/babylonchain/babylon-contract)

A [CosmWasm](https://cosmwasm.com/) smart contract intended for
deployment in a Consumer Zone.
It enables Bitcoin Checkpointing functionality without introducing
invasive changes in the codebase of the Consumer Zone.
Based on the Bitcoin Checkpointing functionality,
the Consumer Zone can make decisions based on the inclusion
of its checkpoints in the Bitcoin ledger
(e.g. execute BTC-assisted unbonding requests).
