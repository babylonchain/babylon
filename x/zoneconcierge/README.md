# ZoneConcierge

The Zone Concierge module is responsible for generating BTC timestamps of
headers from other PoS blockchains. These BTC timestamps allow PoS blockchains
integrating with Babylon to achieve Bitcoin security, i.e., forking the PoS blockchain is as hard as forking Bitcoin. The Zone Concierge module leverages the IBC
light client protocol to receive PoS blockchains' headers.

There are two phases of integration for a consumer chain:

- **Phase 1 integration:** Babylon receives consumer chain headers via standard
  `MsgUpdateClient` messages in IBC light client protocol, timestamps them, and
  functions as a canonical chain oracle for the consumer chain.
  [Babylonscan](https://babylonscan.io/) shows PoS blockchains with phase 1
  integration.
- **Phase 2 integration:** In addition to phase 1, phase 2 allows a consumer
  chain to receive BTC timestamps from Babylon via an IBC channel, such that the
  consumer chain can use BTC timestamps to detect and resolve forks, as well as
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
- [Interaction with consumer chains under phase 1 integration](#interaction-with-consumer-chains-under-phase-1-integration)
- [Interaction with consumer chains under phase 2 integration](#interaction-with-consumer-chains-under-phase-2-integration)
- [Queries](#queries)

## Concepts

The Zone Concierge module is responsible for providing BTC timestamps of headers from other PoS blockchains.
These BTC timestamps allow PoS blockchains to achieve Bitcoin security, i.e., forking a PoS blockchain is as hard as forking Bitcoin.
The Zone Concierge module leverages the IBC light client protocol to maintain headers with a valid quorum certificate from PoS blockchains.
These headers receive Bitcoin timestamps together with the Babylon blockchain, thereby achieving Bitcoin security.

### Problem Statement

Babylon aims at providing Bitcoin security to other PoS blockchains.
To this end, Babylon needs to checkpoint itself to Bitcoin, and checkpoint other PoS blockchains to itself.
The {Epoching, Checkpointing, BTCCheckpoint, BTCLightclient} modules jointly provide the functionality of checkpointing Babylon to Bitcoin.
The Zone Concierge module and the IBC modules jointly provide the functionality of checkpointing PoS blockchains to Babylon.

In order to checkpoint PoS blockchains to Babylon, Babylon needs to receive and verify headers of PoS blockchains, in particular, the canonical and fork headers that have a quorum certificate, i.e., a set of signatures from validators with > 2/3 voting power.
Checkpointing canonical headers allows Babylon to act as a canonical chain oracle.
Checkpointing fork headers allows Babylon to identify dishonest majority attacks and slash equivocating validators.

To summarize, the Zone Concierge module aims at providing the following guarantees:

- **Timestamping headers:** Babylon timestamps all PoS blockchains' headers with a valid quorum certificate from the IBC relayer, regardless whether they are on canonical chains or not.
- **Verifiability of timestamps:** For any header, Babylon can provide a proof that the header is checkpointed by Bitcoin, where the proof is publicly verifiable assuming access to a BTC light client.

under the following assumptions:

- BTC is always secure with the k-deep confirmation rule;
- Babylon might have dishonest majority;
- PoS blockchains might have dishonest majority;
- There exists >=1 honest IBC relayer and >=1 vigilante {submitter, reporter}; and
- The network is synchronous (i.e., messages are delivered within a known time bound).

### Design

The Zone Concierge module is responsible for checkpointing headers of PoS blockchains.
Specifically, the Zone Concierge module

- leverages IBC light clients for receiving and verifying headers from PoS blockchains;
- intercepts and indexes headers from PoS blockchains; and
- provides proofs that a header is finalized by Bitcoin.

**IBC Light Client for Checkpointing PoS blockchains.** Babylon leverages the IBC light client protocol to receive and verify headers of PoS blockchains, ensuring the following security properties when the PoS blockchain has more than 2/3 honest voting power and there exists at least 1 honest relayer.

- **Safety:** The IBC light client is consistent with the PoS blockchain.
- **Liveness:** The IBC light client keeps growing.

Verifying a header is done by a special [quorum intersection mechanism](https://arxiv.org/abs/2010.07031): upon a header from the relayer, the light client checks whether the intersected voting power bewteen the quorum certificates of the current tip and the header is more than 1/3 of the voting power in the current tip.
If yes, then this ensures that there exists at least one honest validator in the header's quorum certificate, and this header is agreed by all honest validators.
Each header of a PoS blockchain carries `AppHash`, which is the root of the Merkle IAVL tree for the PoS blockchain's database.
The `AppHash` allows a light client to verify whether an IBC packet is included in the PoS blockchain's blockchain.

Babylon leverages the IBC light client protocol to checkpoint PoS blockchains to itself.
In particular, each header with a valid quorum certificate can be viewed as a timestamp, and Babylon can generate an inclusion proof that a given IBC header of a PoS blockchain is committed to Babylon's `AppHash`.

**Intercepting and Indexing Headers from PoS blockchains.** In order to further timestamp headers of PoS blockchains to Babylon, the Zone Concierge module builds an index that maps each header to the current epoch, which will be eventually finalized by Bitcoin.
To this end, the Zone Concierge module intercepts headers from IBC light clients via a `PostHandler` and indexes it, including the header's positions on the PoS blockchain and Babylon.

Note that the Zone Concierge module intercepts all headers that have a valid quorum certificate, including both canonical headers and fork headers.
A fork header with a valid quorum certificate is a signal of the dishonest majority attack: the majority of validators are honest and sign conflicted headers.
<!-- TODO: describe PostHandler -->

**Providing Proofs that a Header is Finalized by Bitcoin.** To supports applications that demand a BTC-finalized PoS chain, Zone Concierge will provide proofs that the headers are indeed finalized by Bitcoin.
The proof of a BTC-finalized header includes the following:

- `ProofCzHeaderInEpoch`: Proof that the header of the PoS blockchain is included in an epoch of Babylon;
- `ProofEpochSealed`: Proof that the epoch has been agreed by > 2/3 voting power of the validator set; and
- `ProofEpochSubmitted`: Proof that the epoch's checkpoint has been submitted to Bitcoin.

The first proof is formed as a Merkle proof that the IBC header is committed to the `AppHash` after the epoch.
The second proof is formed as a BLS multi-signature jointly generated by the epoch's validator set.
The last proof is formed as Merkle proofs of two transactions that constitute a BTC checkpoint, same as in [BTCCheckpoint module](../btccheckpoint/README.md).

### Use cases

The BTC-finalized PoS chain will enable a number of applications, such as raising alarms upon dishonest majority attacks and reducing the unbonding time period.
These use cases require new plugins in the PoS blockchains, and will be developed by Babylon team in the future.

**Raising Alarms upon Dishonest Majority Attacks.** Zone Concierge timestamps fork headers that have valid quorum certificates.
Such fork header signals an dishonest majority attack.
Babylon can send the fork header back to the corresponding PoS blockchain, such that the PoS blockchain will get notified with this dishonest majority attack, and can decide to stall and initiate a social consensus.

**Reducing Unbonding Time.** Zone Concierge provides a Bitcoin-checkpointed prefix for a PoS blockchain.
Such Bitcoin-checkpointed prefix resists against the long range attacks, thus unbonding requests in this prefix can be safely finished, leading to much shorter unbonding period compared to 21 days in the current design.

## State

The Zone Concierge module keeps handling IBC headers of consumer chains, and
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
for each consumer chain. The key is the consumer chain's `ChainID`, and the
value is a `ChainInfo` object. The `ChainInfo` is a structure storing the
information of a consumer chain that checkpoints to Babylon.

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
`ChainInfo` at the end of each Babylon epoch for each consumer chain. The key is
the consumer chain's `ChainID` plus the epoch number, and the value is a
`ChainInfo` object.

### CanonicalChain

The [canonical chain storage](./keeper/canonical_chain_indexer.go) maintains the
metadata of canonical IBC headers of a consumer chain. The key is the consumer
chain's `ChainID` plus the height, and the value is a `IndexedHeader` object.
`IndexedHeader` is a structure storing an IBC header's metadata.

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
  // before this timestamp when this header is BTC-finalised
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
IBC headers of a consumer chain. The key is the consumer chain's `ChainID` plus
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


<!-- TODO: describe the PostHandler -->

<!-- TODO: hooks -->

## Interaction with consumer chains under phase 1 integration

<!-- TODO: mermaid flowchart for the interaction -->

In phase 1 integration, Babylon maintains headers for a consumer chain via the
IBC light client protocol. The IBC header chain of the consumer chain is
checkpointed by Bitcoin via Babylon, thus achieves Bitcoin security.

Babylon utilizes IBC light client protocol for the phase 1 integration.
Specifically, the IBC relayer keeps relaying the consumer chain's headers to
Babylon via IBC protocol's `MsgUpdateClient`
[endpoint](https://github.com/cosmos/ibc-go/blob/v8.0.0/proto/ibc/core/client/v1/tx.proto#L20-L21),
and the Zone Concierge module uses a `PostHandler` to handle the IBC header
timely. This does not involve any IBC connection or channel between Babylon and
a consumer chain.

The `PostHandler` is defined at
[x/zoneconcierge/keeper/header_handler.go](./keeper/header_handler.go), and
works as follows.

1. If the consumer chain hosting the header is not known to Babylon, initialize
   `ChainInfo` storage for the consumer chain.
2. If the header is on a fork, insert the header to the fork storage and update
   `ChainInfo`.
3. If the header is canonical, insert the header to the canonical chain storage
   and update `ChainInfo`.

## Interaction with consumer chains under phase 2 integration

<!-- TODO: mermaid flowchart for the interaction -->

In phase 2 integration, Babylon does everything in phase 1, and will send BTC
timestamps of headers back to each consumer chain. Each consumer chain can
verify the BTC timestamp and ensure that each header is finalized by Bitcoin,
thus obtaining Bitcoin security. To do phase 2 integration, one needs to deploy
a Babylon smart contract on the consumer chain, and start an IBC relayer between
Babylon and the Babylon contract on the consumer chain. It does not require any
change to the consumer chain's code.

The BTC timestamps will allow a consumer chain to make use of BTC timestamps for
different use cases, such as BTC-assisted fast unbonding.

The BTC timestamp is defined in the structure `BTCTimestamp`. It includes a
header and a set of proofs that the header is finalized by Bitcoin.

<!-- TODO: diagram depicting BTC timestamp -->

```protobuf
// BTCTimestamp is a BTC timestamp that carries information of a BTC-finalised epoch
// It includes a number of BTC headers, a raw checkpoint, an epoch metadata, and 
// a CZ header if there exists CZ headers checkpointed to this epoch.
// Upon a newly finalised epoch in Babylon, Babylon will send a BTC timestamp to each
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
  // where `lastFinalizedEpoch` is the last finalised epoch in Babylon
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
// BTC-finalised
message ProofFinalizedChainInfo {
  /*
    The following fields include proofs that attest the chain info is
    BTC-finalised
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

Upon a Babylon epoch is finalized, Babylon will send an IBC packet including a
`BTCTimestamp` to each consumer chain doing phase 2/3 integration with Babylon.
The logic upon each finalized epoch is defined at
[x/zoneconcierge/keeper/ibc_packet_btc_timestamp.go](./keeper/ibc_packet_btc_timestamp.go)
and works as follows.

1. Find all open IBC channels with Babylon's Zone Concierge module. The
   counterparty at each IBC channel is a consumer chain.
2. Get all canonical BTC headers received during the time of finalizing the last
  epoch. Specifically, 2.1. Find the tip `h'` of BTC light client when the
  second last epoch becomes finalized. 2.2. Find the tip `h` of BTC light client
  when the last epoch becomes finalized. 2.3. If `h'` and `h` are on the same
  chain, then the canonical BTC headers are headers from `h'` to `h`. 2.4. If
  `h'` and `h` are on different forks, then the canonical BTC headers start from
  their last common ancestor to `h`.
3. For each of these IBC channels: 3.1. Find the `ChainID` of the counterparty
  chain (i.e., the consumer chain) in the IBC channel 3.2. Get the `ChainInfo`
  of the `ChainID` at the last finalized epoch. 3.3. Get the metadata of the
  last finalized epoch and its corresponding raw checkpoint. 3.4. Generate the
  proof that the last consumer chain's canonical header is committed to the
  epoch's metadata. 3.5. Generate the proof that the epoch is sealed, i.e.,
  receives a BLS multisignature generated by validators with >2/3 total voting
  power at the last finalized epoch. 3.6. Generate the proof that the epoch's
  checkpoint is submitted, i.e., encoded in transactions on Bitcoin. 3.7.
  Assemble all the above as `BTCTimestamp`, and send it to the IBC channel in an
  IBC packet.

[Babylon contract](https://github.com/babylonchain/babylon-contract) is a
CosmWasm smart contract for phase 2 integration. It can be deployed to a
blockchain supporting CosmWasm smart contracts, connects with Babylon's Zone
Concierge module via an IBC channel, and receives BTC timestamps from Babylon to
help the consumer chain get Bitcoin security.

Note that Zone Concierge provides 1-to-all connection, where where the Zone
Concierge module establishes an IBC channel with each of multiple consumer
chains. Zone Concierge will send an BTC timestamp to each of these consumer
chains upon an epoch is finalised.

## Queries

The Zone Concierge module only has one message `MsgUpdateParams` for updating
the module parameters via a governance proposal. It provides a set of queries
about the status of checkpointed consumer chains, listed at
[docs.babylonchain.io](https://docs.babylonchain.io/docs/developer-guides/grpcrestapi#tag/ZoneConcierge).
