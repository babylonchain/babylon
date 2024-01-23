# BTC light client

Babylon chain needs to know about different events happening on Bitcoin chain
To make it possible in secure way, Babylon needs to known what is the current
state of Bitcoin chain i.e what is best chain recognized by Bitcoin network.

## Table of contents

- [Table of contents](#table-of-contents)
- [Concepts](#concepts)
  - [Problem statement](#problem-statement)
  - [Babylon's BTC light client design](#babylons-btc-light-client-design)
- [States](#states)
  - [Parameters](#parameters)
  - [Headers storage](#headers-storage)
  - [HashToHeight storage](#hashtoheight-storage)
- [Messages](#messages)
  - [MsgInsertHeaders](#msginsertheaders)
  - [MsgUpdateParams](#msgupdateparams)
- [Hooks](#hooks)
  - [Hooks exposed by BTC light client](#hooks-exposed-by-btc-light-client)
- [Events](#events)

### Problem statement

In the Babylon chain, the are few events which needs to be validated against
current state of BTC chain.

Those events are:

1. Bitcoin Timestamping protocol requires to report back checkpoints which were
send to BTC chain. To do it securely, each checkpoint must be reported back along
with BTC inclusion proof of transaction which carried given checkpoint.
2. Bitcoin staking protocol requires that staking transaction reported to
Babylon must have proper depth in BTC chain and also must ba accompanied by BTC
inclusion proof.

To properly validate those inclusion proofs, the Babylon chain needs to know what
is current state of BTC chain i.e what is current canonical chain recognized by BTC
network.


### Babylon's BTC light client design

In high level overview, Babylon BTC light client starts from some initial BTC
header existing on BTC network and allow to extend this header by applying the
same rules as normal BTC node.

Initial BTC header must:
- be deep enough in BTC chain so that it will never be reverted by BTC network.
- be at height which is at BTC difficulty adjustment boundary. This required, to
properly validate all future difficulty adjustments.

The Babylon BTC light client module stores only BTC canonical header chain and
does not store the headers which end of forks. The BTC canonical chain can only
be extended by processing valid [MsgInsertHeaders](#msginsertheaders) message.

If the better fork is encountered:
1. current chain is rollbacked up to header which is parent of received fork
2. chain is extend with new header from the fork


## States

The BTC light client module maintains the following KV stores.

### Parameters

The [parameter storage](./keeper/params.go) maintains the BTC light client module's
parameters. The BTC light client module's parameters are represented as a `Params`
[object](../../proto/babylon/btclightclient/v1/params.proto) defined as follows:

```protobuf
// Params defines the parameters for the module.
message Params {
  option (gogoproto.equal) = true;

  // List of addresses which are allowed to insert headers to btc light client
  // if the list is empty, any address can insert headers
  repeated string insert_headers_allow_list = 1;
}
```

In nutshell, `insert_headers_allow_list` makes it possible to set up
restrictions about who is able to update BTC light client module state.

If `insert_headers_allow_list` is not empty, only addresses in the list can send
`MsgInsertHeaders` message.

### Headers storage

The [Headers storage](./keeper/state.go) maintains information about each header.
The key is the header height in BTC chain, and the value is an `BTCHeaderInfo`
[object](../../proto/babylon/btclightclient/v1/btclightclient.proto)
which contains BTC header along with some metadata.

```protobuf
// BTCHeaderInfo is a structure that contains all relevant information about a
// BTC header
//  - Full header bytes
//  - Header hash for easy retrieval
//  - Height of the header in the BTC chain
//  - Total work spent on the header. This is the sum of the work corresponding
//  to the header Bits field
//    and the total work of the header.
message BTCHeaderInfo {
  bytes header = 1
      [ (gogoproto.customtype) =
            "github.com/babylonchain/babylon/types.BTCHeaderBytes" ];
  bytes hash = 2
      [ (gogoproto.customtype) =
            "github.com/babylonchain/babylon/types.BTCHeaderHashBytes" ];
  uint64 height = 3;
  bytes work = 4
      [ (gogoproto.customtype) = "cosmossdk.io/math.Uint" ];
}
```

### HashToHeight storage

The [HashToHeight storage](./keeper/state.go) maintains index in which key is
BTC header hash and value is BTC header height.

This index enables efficient lookup of BTC headers by their hash. This useful
in many situations, but mostly when receiving potential chain extension which
does not points to current BTC chain tip.

## Messages

### MsgInsertHeaders

`MsgInsertHeaders` is the main message processed by BTC light client module.
Its purpose is to update the state of the BTC chain as viewed by Babylon chain.

```proto
message MsgInsertHeaders {
  option (cosmos.msg.v1.signer) = "signer";

  string signer = 1;
  repeated bytes headers = 2
      [ (gogoproto.customtype) =
            "github.com/babylonchain/babylon/types.BTCHeaderBytes" ];
}
```

The message contains a list of BTC headers encoded in Bitcoin format.

Rules applied when processing the message:
- `headers` list must not be empty
- headers in the list must be connected by parent-child relationships i.e
header at position `i + 1`, must have its `PrevBlock` field set to header `i`
hash
- first header of the list must point to the header already maintained by BTC
light client module
- each header must be correctly encoded
- each header in the list must have valid proof of work and difficulty
- each header in the list must have `Timestamp` which happened after median
of last 11 ancestors
- if the first header of the list does not point to the current tip of the
chain maintained by BTC light client, it means that message contains a fork. For
fork to be valid, forked chain must be better than current chain maintained by
BTC light client. The fork is better when its total work is larger that the work
of current chain

All those rules are the same rules which are applied by BTC nodes when receiving
headers from the BTC network.

Processing of the message is atomic, so even if one header in the list is
invalid, the state of the BTC light client module won't be updated.

In case of receiving valid chain extension i.e first header of the `headers` list
points to current BTC chain tip, the chain maintained by BTC light client module
will be extended and received headers will be saved into the storage.

In case of receiving valid and better fork i.e first header of the `headers` list
does not point to to current BTC chain tip and fork total work is larger than
current chain BTC chain total work, chain maintained by BTC light client will
be roll backed to header which is a fork header, and then it will be
extended with headers received in `headers`.



### MsgUpdateParams

The `MsgUpdateParams` message is used for updating the module parameters for the
BTC light client module. It can only be executed via a governance proposal.

```protobuf
// MsgUpdateParams defines a message for updating btc light client module parameters.
message MsgUpdateParams {
  option (cosmos.msg.v1.signer) = "authority";

  // authority is the address of the governance account.
  // just FYI: cosmos.AddressString marks that this field should use type alias
  // for AddressString instead of string, but the functionality is not yet implemented
  // in cosmos-proto
  string authority = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];

  // params defines the btc light client parameters to update.
  //
  // NOTE: All parameters must be supplied.
  Params params = 2 [(gogoproto.nullable) = false];
}
```

## Hooks

The BTC light client module exposes a set of hooks to inform other modules
about the updates to the maintained Bitcoin best chain:

### Hooks exposed by BTC light client

```go
type BTCLightClientHooks interface {
	AfterBTCRollBack(ctx context.Context, headerInfo *BTCHeaderInfo)       // Must be called after the chain is rolled back
	AfterBTCRollForward(ctx context.Context, headerInfo *BTCHeaderInfo)    // Must be called after the chain is rolled forward
	AfterBTCHeaderInserted(ctx context.Context, headerInfo *BTCHeaderInfo) // Must be called after a header is inserted
}

```

## Events

The BTC light client module exposes a set of events about the updates to the
maintained Bitcoin best chain:

```protobuf

// The header included in the event is the block in the history
// of the current mainchain to which we are rolling back to.
// In other words, there is one rollback event emitted per re-org, to the
// greatest common ancestor of the old and the new fork.
message EventBTCRollBack { BTCHeaderInfo header = 1; }

// EventBTCRollForward is emitted on Msg/InsertHeader
// The header included in the event is the one the main chain is extended with.
// In the event of a reorg, each block on the new fork that comes after
// the greatest common ancestor will have a corresponding roll forward event.
message EventBTCRollForward { BTCHeaderInfo header = 1; }

// EventBTCHeaderInserted is emitted on Msg/InsertHeader
// The header included in the event is the one that was added to the
// on chain BTC storage.
message EventBTCHeaderInserted { BTCHeaderInfo header = 1; }

```

