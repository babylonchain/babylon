# ZoneConcierge

The Zone Concierge module is responsible for generating BTC timestamps of
headers from other PoS blockchains. These BTC timestamps allow PoS blockchains
integrating with Babylon to achieve Bitcoin security, i.e., forking the PoS
blockchain is as hard as forking Bitcoin. The Zone Concierge module leverages
the IBC light client protocol to receive PoS blockchains' headers.

There are two phases of integration for a PoS blockchain:

- **Phase 1 integration:** Babylon receives PoS blockchain headers via standard
  `MsgUpdateClient` messages in IBC light client protocol, timestamps them, and
  functions as a canonical chain oracle for the PoS blockchain.
  [Babylonscan](https://babylonscan.io/) shows PoS blockchains with phase 1
  integration.
- **Phase 2 integration:** In addition to phase 1, phase 2 allows a consumer
  chain to receive BTC timestamps from Babylon via an IBC channel, such that the
  PoS blockchain can use BTC timestamps to detect and resolve forks, as well as
  other use cases such as Bitcoin-assisted fast unbonding.

## Table of contents

- [Table of contents](#table-of-contents)
- [Concepts](#concepts)
  - [Problem Statement](#problem-statement)
  - [Design](#design)
  - [Use cases](#use-cases)
- [State](#state)
  - [Parameters](#parameters)
  - [ChainInfo](#chaininfo)
  - [EpochChainInfo](#epochchaininfo)
  - [CanonicalChain](#canonicalchain)
  - [Fork](#fork)
  - [Params](#params)
- [PostHandler for intercepting IBC headers](#posthandler-for-intercepting-ibc-headers)
- [Hooks](#hooks)
  - [Indexing headers upon `AfterEpochEnds`](#indexing-headers-upon-afterepochends)
  - [Sending BTC timestamps upon `AfterRawCheckpointFinalized`](#sending-btc-timestamps-upon-afterrawcheckpointfinalized)
- [Interaction with PoS blockchains under phase 1 integration](#interaction-with-pos-blockchains-under-phase-1-integration)
- [Interaction with PoS blockchains under phase 2 integration](#interaction-with-pos-blockchains-under-phase-2-integration)
- [Messages and Queries](#messages-and-queries)

## Concepts

The Zone Concierge module is responsible for providing BTC timestamps of headers
from PoS blockchains connected to Babylon via the IBC protocol.  
These BTC timestamps allow PoS blockchains to achieve Bitcoin security, i.e.,
forking a PoS blockchain is as hard as forking Bitcoin. The Zone Concierge
module leverages the IBC light client protocol to receive headers with a valid
quorum certificate from PoS blockchains. These headers are then timestamped
together with the Babylon blockchain by Bitcoin, thereby achieving Bitcoin
security. The BTC timestamps can be propagated back to the PoS blockchains, such
that PoS blockchains can know their headers that have been checkpointed by
Bitcoin.

### Problem Statement

Babylon aims to provide Bitcoin security to other PoS blockchains. This involves
two functionalities: 1) checkpointing Babylon to Bitcoin, and 2) checkpointing
other PoS blockchains to Babylon. The {[Epoching](../epoching/),
[Checkpointing](../checkpointing/), [BTCCheckpoint](../btccheckpoint/),
[BTCLightclient](../btclightclient/)} modules jointly provide the functionality
of checkpointing Babylon to Bitcoin. The [Zone Concierge module](./) and the
[IBC modules](https://github.com/cosmos/ibc-go) jointly provide the
functionality of checkpointing PoS blockchains to Babylon.

In order to checkpoint PoS blockchains to Babylon, Babylon needs to receive
headers of PoS blockchains and maintain all headers that have a *quorum
certificate* (a set of signatures from validators with > 2/3 total voting
power). Checkpointing canonical headers allows Babylon to act as a canonical
chain oracle. Checkpointing fork headers allows Babylon to identify dishonest
majority attacks.

To summarize, the Zone Concierge module aims at providing the following
guarantees:

- **Timestamping headers:** Babylon checkpoints PoS blockchains' (canonical and
  fork) headers with a valid quorum certificate.
- **Verifiability of timestamps:** Babylon can provide a proof that a given
  header is checkpointed by Bitcoin, where the proof is publicly verifiable
  assuming access to a BTC light client.

Under the following assumptions:

- BTC is always secure with the [k-deep confirmation
  rule](https://en.bitcoin.it/wiki/Confirmation);
- There exists >=1 honest IBC relayer and vigilante {submitter, reporter}; and
- The network is synchronous (i.e., messages are delivered within a known and
  finite time bound).

Note that the Bitcoin timestamping protocol uses Bitcoin as a single source of
truth, and does not make any assumption on the fraction of adversarial
validators in Babylon or PoS blockchains. That is, the above statement shall
hold even if Babylon and a PoS blockchain have dishonest supermajority. The
formal security analysis of the Bitcoin timestamping protocol can be found at
the Bitcoin timestamping [reseaarch paper](https://arxiv.org/pdf/2207.08392.pdf)
published at [S\&P'23](https://sp2023.ieee-security.org/).

### Design

The Zone Concierge module is responsible for checkpointing headers from PoS
blockchains and propagating succinct and verifiable information about them back
to the PoS blockchains. Specifically, the Zone Concierge module  

- leverages IBC light clients for checkpointing PoS blockchains;
- intercepts and indexes headers from PoS blockchains; and
- provides BTC timestamps proving that a header is checkpointed by Babylon and
  Bitcoin (via queries or IBC packets).

**Leveraging IBC light clients for checkpointing PoS blockchains.** Babylon
leverages the [IBC light client
protocol](https://github.com/cosmos/ibc/tree/main/spec/client/ics-007-tendermint-client)
to receive and verify headers of PoS blockchains. The IBC light client protocol
allows a blockchain `A` to maintain a *light client* of another blockchain `B`.
The light client contains a subset of headers in the ledger of blockchain `B`,
securing the following properties when blockchain `B` has more than 2/3 honest
voting power and there exists at least 1 honest IBC relayer.

- **Safety:** The IBC light client in blockchain `A` is consistent with the
  ledger of blockchain `B`.
- **Liveness:** The IBC light client in blockchain `A` keeps growing.

Verifying a header is done by a special [quorum intersection
mechanism](https://arxiv.org/abs/2010.07031): upon a header from the relayer,
the light client checks whether the intersected voting power between the quorum
certificates of the current tip and the header is more than 1/3 of the voting
power in the current tip. If yes, then this ensures that there exists at least
one honest validator in the header's quorum certificate, and this header is
agreed by all honest validators.

Babylon leverages the IBC light client protocol to checkpoint PoS blockchains to
itself. In particular, each header with a valid quorum certificate can be viewed
as a timestamp, and Babylon can generate an inclusion proof that a given header
of a PoS blockchain is committed to Babylon's `AppHash`.

**Intercepting and Indexing Headers from PoS blockchains.** In order to further
checkpoint headers of PoS blockchains to Bitcoin, the Zone Concierge module
builds an index recording headers' positions on Babylon's ledger, which will
eventually be checkpointed by Bitcoin. To this end, the Zone Concierge module
intercepts headers from IBC light clients via a
[PostHandler](https://docs.cosmos.network/v0.50/learn/advanced/baseapp#runtx-antehandler-runmsgs-posthandler),
and indexes them.

Note that the Zone Concierge module intercepts all headers that have a valid
quorum certificate, including canonical headers and fork headers. A fork header
with a valid quorum certificate is a signal of the dishonest majority attack:
the majority of validators are dishonest and sign conflicted headers.

**Providing Proofs that a Header is Checkpointed by Bitcoin.** To support use
cases that need to verify BTC timestamps of headers, Zone Concierge can provide
proofs that the headers are indeed checkpointed to Bitcoin. The proof includes
the following:

- `ProofCzHeaderInEpoch`: Proof that the header of the PoS blockchain is
  included in an epoch of Babylon;
- `ProofEpochSealed`: Proof that the epoch has been agreed by > 2/3 voting power
  of the validator set; and
- `ProofEpochSubmitted`: Proof that the epoch's checkpoint has been submitted to
  Bitcoin.

The first proof is formed as a Merkle proof that the IBC header is committed to
the `AppHash` after the epoch. The second proof is formed as a BLS
multi-signature jointly generated by the epoch's validator set. The last proof
is formed as Merkle proofs of two transactions that constitute a BTC checkpoint,
same as in [BTCCheckpoint module](../btccheckpoint/README.md).

### Use cases

The Bitcoin-checkpointed PoS blockchain will enable several applications, such
as raising alarms upon dishonest majority attacks and reducing the unbonding
period. These use cases require new plugins in the PoS blockchains, and will be
developed by Babylon team in the future.

**Raising Alarms upon Dishonest Majority Attacks.** Zone Concierge timestamps
fork headers that have valid quorum certificates. Such fork header signals a
safety attack launched by the dishonest majority of validators. Babylon can send
the fork header back to the corresponding PoS blockchain, such that the PoS
blockchain will get notified with this dishonest majority attack, and can decide
to stall or initiate a social consensus.

**Reducing Unbonding Period.** Zone Concierge provides a Bitcoin-checkpointed
prefix for a PoS blockchain. Such Bitcoin-checkpointed prefix resists against
the long range attacks, thus unbonding requests in this prefix can be safely
finished, leading to much shorter unbonding period compared to that in existing
PoS blockchains (e.g., 21 days in Cosmos SDK chains).

## State

The Zone Concierge module keeps handling IBC headers of PoS blockchains, and
maintains the following KV stores.

### Parameters

The [parameter storage](./keeper/params.go) maintains Zone Concierge module's
parameters. The Zone Concierge module's parameters are represented as a `Params`
[object](../../proto/babylon/zoneconcierge/v1/params.proto) defined as follows:

```protobuf
// Params defines the parameters for the module.
message Params {
  option (gogoproto.equal) = true;
  
  // ibc_packet_timeout_seconds is the time period after which an unrelayed 
  // IBC packet becomes timeout, measured in seconds
  uint32 ibc_packet_timeout_seconds = 1
      [ (gogoproto.moretags) = "yaml:\"ibc_packet_timeout_seconds\"" ];
}
```

### ChainInfo

The [chain info storage](./keeper/chain_info_indexer.go) maintains `ChainInfo`
for each PoS blockchain. The key is the PoS blockchain's `ChainID`, and the
value is a `ChainInfo` object. The `ChainInfo` is a structure storing the
information of a PoS blockchain that checkpoints to Babylon.

```protobuf
// ChainInfo is the information of a CZ
message ChainInfo {
  // chain_id is the ID of the chain
  string chain_id = 1;
  // latest_header is the latest header in CZ's canonical chain
  IndexedHeader latest_header = 2;
  // latest_forks is the latest forks, formed as a series of IndexedHeader (from
  // low to high)
  Forks latest_forks = 3;
  // timestamped_headers_count is the number of timestamped headers in CZ's
  // canonical chain
  uint64 timestamped_headers_count = 4;
}
```

### EpochChainInfo

The [epoch chain info storage](./keeper/epoch_chain_info_indexer.go) maintains
`ChainInfo` at the end of each Babylon epoch for each PoS blockchain. The key is
the PoS blockchain's `ChainID` plus the epoch number, and the value is a
`ChainInfo` object.

### CanonicalChain

The [canonical chain storage](./keeper/canonical_chain_indexer.go) maintains the
metadata of canonical IBC headers of a PoS blockchain. The key is the consumer
chain's `ChainID` plus the height, and the value is a `IndexedHeader` object.
`IndexedHeader` is a structure storing IBC header's metadata.

```protobuf
// IndexedHeader is the metadata of a CZ header
message IndexedHeader {
  // chain_id is the unique ID of the chain
  string chain_id = 1;
  // hash is the hash of this header
  bytes hash = 2;
  // height is the height of this header on CZ ledger
  // (hash, height) jointly provides the position of the header on CZ ledger
  uint64 height = 3;
  // time is the timestamp of this header on CZ ledger
  // it is needed for CZ to unbond all mature validators/delegations
  // before this timestamp when this header is BTC-finalized
  google.protobuf.Timestamp time = 4 [ (gogoproto.stdtime) = true ];
  // babylon_header_hash is the hash of the babylon block that includes this CZ
  // header
  bytes babylon_header_hash = 5;
  // babylon_header_height is the height of the babylon block that includes this CZ
  // header
  uint64 babylon_header_height = 6;
  // epoch is the epoch number of this header on Babylon ledger
  uint64 babylon_epoch = 7;
  // babylon_tx_hash is the hash of the tx that includes this header
  // (babylon_block_height, babylon_tx_hash) jointly provides the position of
  // the header on Babylon ledger
  bytes babylon_tx_hash = 8;
}
```

### Fork

The [fork storage](./keeper/fork_indexer.go) maintains the metadata of canonical
IBC headers of a PoS blockchain. The key is the PoS blockchain's `ChainID` plus
the height, and the value is a list of `IndexedHeader` objects, which represent
fork headers at that height.

### Params

The [parameter storage](./keeper/params.go) maintains the parameters for the
Zone Concierge module.

```protobuf
// Params defines the parameters for the module.
message Params {
  option (gogoproto.equal) = true;
  
  // ibc_packet_timeout_seconds is the time period after which an unrelayed 
  // IBC packet becomes timeout, measured in seconds
  uint32 ibc_packet_timeout_seconds = 1
      [ (gogoproto.moretags) = "yaml:\"ibc_packet_timeout_seconds\"" ];
}
```

## PostHandler for intercepting IBC headers

The Zone Concierge module implements a
[PostHandler](https://docs.cosmos.network/v0.50/learn/advanced/baseapp#runtx-antehandler-runmsgs-posthandler)
`IBCHeaderDecorator` to intercept headers sent to the [IBC client
module](https://github.com/cosmos/ibc-go/tree/v8.0.0/modules/core/02-client).
The `IBCHeaderDecorator` PostHandler is defined at
[x/zoneconcierge/keeper/header_handler.go](./keeper/header_handler.go), and
works as follows.

1. If the PoS blockchain hosting the header is not known to Babylon, initialize
   `ChainInfo` storage for the PoS blockchain.
2. If the header is on a fork, insert the header to the fork storage and update
   `ChainInfo`.
3. If the header is canonical, insert the header to the canonical chain storage
   and update `ChainInfo`.

## Hooks

The Zone Concierge module subscribes to the Epoching module's `AfterEpochEnds`
[hook](../epoching/types/hooks.go) for indexing the epochs when receiving
headers from PoS blockchains, and the Checkpointing module's
`AfterRawCheckpointFinalized` [hook](../checkpointing/types/hooks.go) for phase
2 integration.

### Indexing headers upon `AfterEpochEnds`

The `AfterEpochEnds` hook is triggered upon an epoch is ended, i.e., the last
block in this epoch has been committed by CometBFT. Upon `AfterEpochEnds`, the
Zone Concierge will save the current `ChainInfo` to the `EpochChainInfo` storage
for each PoS blockchain.

### Sending BTC timestamps upon `AfterRawCheckpointFinalized`

The `AfterRawCheckpointFinalized` hook is triggered upon a checkpoint becoming
*finalized*, i.e., Bitcoin transactions of the checkpoint become `w`-deep in
Bitcoin's canonical chain, where `w` is the `checkpoint_finalization_timeout`
[parameter](../../proto/babylon/btccheckpoint/v1/params.proto) in the
[BTCCheckpoint](../btccheckpoint/) module.

Upon `AfterRawCheckpointFinalized`, the Zone Concierge module will prepare and
send a BTC timestamp to each PoS blockchain.  
The [BTCTimestamp](../../proto/babylon/zoneconcierge/v1/packet.proto) structure  
includes a header and a set of proofs that the header is checkpointed by
Bitcoin.

<!-- TODO: diagram depicting BTC timestamp -->

```protobuf
// BTCTimestamp is a BTC timestamp that carries information of a BTC-finalized epoch
// It includes a number of BTC headers, a raw checkpoint, an epoch metadata, and 
// a CZ header if there exists CZ headers checkpointed to this epoch.
// Upon a newly finalized epoch in Babylon, Babylon will send a BTC timestamp to each
// PoS blockchain that has phase-2 integration with Babylon via IBC.
message BTCTimestamp {
  // header is the last CZ header in the finalized Babylon epoch
  babylon.zoneconcierge.v1.IndexedHeader header = 1;

  /*
    Data for BTC light client
  */
  // btc_headers is BTC headers between
  // - the block AFTER the common ancestor of BTC tip at epoch `lastFinalizedEpoch-1` and BTC tip at epoch `lastFinalizedEpoch`
  // - BTC tip at epoch `lastFinalizedEpoch`
  // where `lastFinalizedEpoch` is the last finalized epoch in Babylon
  repeated babylon.btclightclient.v1.BTCHeaderInfo btc_headers = 2;

  /*
    Data for Babylon epoch chain
  */
  // epoch_info is the metadata of the sealed epoch
  babylon.epoching.v1.Epoch epoch_info = 3;
  // raw_checkpoint is the raw checkpoint that seals this epoch
  babylon.checkpointing.v1.RawCheckpoint raw_checkpoint = 4;
  // btc_submission_key is position of two BTC txs that include the raw checkpoint of this epoch
  babylon.btccheckpoint.v1.SubmissionKey btc_submission_key = 5;

  /* 
    Proofs that the header is finalized
  */
  babylon.zoneconcierge.v1.ProofFinalizedChainInfo proof = 6;
}

// ProofFinalizedChainInfo is a set of proofs that attest a chain info is
// BTC-finalized
message ProofFinalizedChainInfo {
  /*
    The following fields include proofs that attest the chain info is
    BTC-finalized
  */
  // proof_cz_header_in_epoch is the proof that the CZ header is timestamped
  // within a certain epoch
  tendermint.crypto.ProofOps proof_cz_header_in_epoch = 1;
  // proof_epoch_sealed is the proof that the epoch is sealed
  babylon.zoneconcierge.v1.ProofEpochSealed proof_epoch_sealed = 2;
  // proof_epoch_submitted is the proof that the epoch's checkpoint is included
  // in BTC ledger It is the two TransactionInfo in the best (i.e., earliest)
  // checkpoint submission
  repeated babylon.btccheckpoint.v1.TransactionInfo proof_epoch_submitted = 3;
}
```

Upon `AfterRawCheckpointFinalized` is triggered, the Zone Concierge module will
send an IBC packet including a `BTCTimestamp` to each PoS blockchain doing
[phase 2
integration](#interaction-with-pos-blockchains-under-phase-2-integration) with
Babylon. The logic is defined at
[x/zoneconcierge/keeper/hooks.go](./keeper/hooks.go) and works as follows.

1. Find all open IBC channels with Babylon's Zone Concierge module. The
   counterparty at each IBC channel is a PoS blockchain.
2. Get all BTC headers to be sent in BTC timestamps. Specifically,
   1. Find the segment of BTC headers sent upon the last time
      `AfterRawCheckpointFinalized` is triggered.
   2. If all BTC headers in the segment are no longer canonical, the BTC headers
      to be sent will be the last `w+1` ones in the BTC light client, where `w`
      is the `checkpoint_finalization_timeout`
      [parameter](../../proto/babylon/btccheckpoint/v1/params.proto) in the
      [BTCCheckpoint](../btccheckpoint/) module.
   3. Otherwise, the BTC headers to be sent will be from the latest header that
      is still canonical in the segment to the current tip of the BTC light
      client.
3. For each of these IBC channels:
   1. Find the `ChainID` of the counterparty chain (i.e., the PoS blockchain) in
      the IBC channel.
   2. Get the `ChainInfo` of the `ChainID` at the last finalized epoch.
   3. Get the metadata of the last finalized epoch and its corresponding raw
      checkpoint.
   4. Generate the proof that the last PoS blockchain's canonical header is
      committed to the epoch's metadata.
   5. Generate the proof that the epoch is sealed, i.e., receives a BLS
      multisignature generated by validators with >2/3 total voting power at the
      last finalized epoch.
   6. Generate the proof that the epoch's checkpoint is submitted, i.e., encoded
      in transactions on Bitcoin.
   7. Assemble all the above and the BTC headers obtained in step 2 as
      `BTCTimestamp`, and send it to the IBC channel in an IBC packet.

## Interaction with PoS blockchains under phase 1 integration

<!-- TODO: more technical details and connections with the spec section for phase 1/2 integration -->
<!-- TODO: mermaid flowchart for the interaction -->

In phase 1 integration, Babylon maintains headers for a PoS blockchain via the
IBC light client protocol. The IBC light client of the PoS blockchain is
checkpointed by Bitcoin via Babylon, thus achieves Bitcoin security.

Babylon utilizes the [IBC light client
protocol](https://github.com/cosmos/ibc/tree/main/spec/client/ics-007-tendermint-client)
for receiving headers from other PoS blockchains. The IBC headers are
encapsulated in the IBC protocol's `MsgUpdateClient`
[messages](https://github.com/cosmos/ibc-go/blob/v8.0.0/proto/ibc/core/client/v1/tx.proto#L20-L21),
and are sent to the [IBC client
module](https://github.com/cosmos/ibc-go/tree/v8.0.0/modules/core/02-client) by
an [IBC
relayer](https://github.com/cosmos/ibc/blob/main/spec/relayer/ics-018-relayer-algorithms/README.md).
The `IBCHeaderDecorator` PostHandler intercepts the headers and indexes their
positions in the `ChainInfo` storage, as per
[here](#indexing-headers-upon-afterepochends). This effectively checkpoints the
headers of PoS blockchains, completing the phase 1 integration.

## Interaction with PoS blockchains under phase 2 integration

<!-- TODO: mermaid flowchart for the interaction -->

In phase 2 integration, Babylon does everything in phase 1, and will send BTC
timestamps of headers back to each PoS blockchain. Each PoS blockchain can
verify the BTC timestamp and ensure that each header is finalized by Bitcoin,
thus obtaining Bitcoin security. The BTC timestamps can be used by the PoS
blockchain  
for different use cases, e.g., BTC-assisted unbonding.

The phase 2 integration does not require any change to the PoS blockchain's
code. Rather, it only needs to deploy a [Babylon
contract](https://github.com/babylonchain/babylon-contract) on the PoS
blockchain, and start an IBC relayer between Babylon and the Babylon contract on
the PoS blockchain. The Babylon contract can be deployed to a blockchain
supporting [CosmWasm](https://github.com/CosmWasm/cosmwasm) smart contracts,
connects with Babylon's Zone Concierge module via an IBC channel, and receives
BTC timestamps from Babylon to help the PoS blockchain get Bitcoin security.

Upon a Babylon epoch becoming finalized, i.e., upon
`AfterRawCheckpointFinalized` is triggered, Babylon will send an IBC packet
including a `BTCTimestamp` to each PoS blockchain doing phase 2/3 integration
with Babylon, as per
[here](#sending-btc-timestamps-upon-afterrawcheckpointfinalized).

Note that Zone Concierge provides 1-to-all connection, where the Zone Concierge
module establishes an IBC channel with each of multiple consumer chains. Zone
Concierge will send an BTC timestamp to each of these consumer chains upon an
epoch is finalized.

## Messages and Queries

The Zone Concierge module only has one message `MsgUpdateParams` for updating
the module parameters via a governance proposal.

It provides a set of queries about the status of checkpointed PoS blockchains,
listed at
[docs.babylonchain.io](https://docs.babylonchain.io/docs/developer-guides/grpcrestapi#tag/ZoneConcierge).
