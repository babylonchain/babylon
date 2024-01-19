# Epoching

Babylon implements the *epoched staking* to reduce and parameterize the
frequency of updating the validator set in Babylon. In the epoched staking
design, the blockchain is divided into epochs, each of which contains a fixed
number of consecutive blocks. Staking-related messages that affect the validator
set's stake distribution are delayed to the end of each epoch for execution,
such that the stake distribution remains unchanged during each epoch.

The Epoching module is responsible for implementing the epoched staking design,
including:

- tracking the epoch number of the current blockchain;
- recording the metadata of each epoch;
- delaying the staking-related messages to the end of each epoch; and
- finishing all unbonding delegations till a Bitcoin-checkpointed epoch.

## Table of contents

- [Table of contents](#table-of-contents)
- [Concepts](#concepts)
  - [Problem statement](#problem-statement)
  - [Babylon's Epoching module design](#babylons-epoching-module-design)
- [States](#states)
  - [Parameters](#parameters)
  - [Epochs](#epochs)
  - [Epoch message queue](#epoch-message-queue)
  - [Epoch validator set](#epoch-validator-set)
- [Messages](#messages)
  - [Disabling Staking module messages via AnteHandler](#disabling-staking-module-messages-via-antehandler)
  - [Epoched staking messages](#epoched-staking-messages)
  - [MsgUpdateParams](#msgupdateparams)
- [BeginBlocker and EndBlocker](#beginblocker-and-endblocker)
  - [Disabling Staking module's EndBlocker](#disabling-staking-modules-endblocker)
- [BeginBlocker](#beginblocker)
- [EndBlocker](#endblocker)
- [Hooks](#hooks)
  - [Hooks in the Epoching module](#hooks-in-the-epoching-module)
  - [Bitcoin-assisted unbonding via `AfterRawCheckpointFinalized` hook](#bitcoin-assisted-unbonding-via-afterrawcheckpointfinalized-hook)
- [Events](#events)
- [Queries](#queries)

## Concepts

### Problem statement

In the Cosmos SDK, the validator set can change with every block, impacting
stake distribution through various staking-related actions (e.g., bond/unbond,
delegate/undelegate/redelegate, slash). This frequent updating poses challenges,
as 1) Babylon's Bitcoin Timestamping protocol requires checkpointing the
validator set to Bitcoin upon every validator set update, and 2) Bitcoin's
10-minute block interval makes checkpointing every new block impractical. In
addition, frequent validator set updates complicate the implementation of
threshold cryptography, light clients, fair leader election, and staking
derivatives, as highlighted in
[ADR-039](https://github.com/cosmos/cosmos-sdk/blob/main/docs/architecture/adr-039-epoched-staking.md).

The concept of epoched staking (aka epoching) aims to reduce the validator set
frequency, addressing these issues. In epoching design, the blockchain is
divided into epochs, and updating the validator set once per epoch. This
approach, detailed in
[ADR-039](https://github.com/cosmos/cosmos-sdk/blob/main/docs/architecture/adr-039-epoched-staking.md),
has been pursued by multiple efforts (e.g.,
[here](https://github.com/cosmos/cosmos-sdk/pull/8829),
[here](https://github.com/cosmos/cosmos-sdk/pull/10132) and
[here](https://github.com/cosmos/cosmos-sdk/pull/10173)) but was not fully
implemented. Babylon has implemented its own Epoching module, catering to
specific design goals such as checkpointing epochs. In addition, Babylon
implements *Bitcoin-assisted unbonding*, where unbonding requests in an epoch
will be finished upon this epoch is checkpointed on Bitcoin.

### Babylon's Epoching module design

Babylon implements the Epoching module in order to reduce the frequency of
validator set updates, and thus the frequency of checkpointing to Bitcoin.
Specifically, the Epoching module is responsible for the following tasks:

- Dividing the blockchain into epochs.
- Disabling some functionalities of the Staking module.
- Disabling messages of the Staking module.
- Delaying staking-related messages till the end of the epoch.
- Bitcoin-assisted unbonding.

**Dividing the blockchain into epochs.** The epoching mechanism introduces the
concept of epochs. The blockchain is divided into epochs, each consists of a
fixed number of consecutive blocks. The number of blocks in an epoch is called
epoch interval, which is a system parameter. At the moment, Babylon uses the
epoch interval of 900 blocks, which take about 30 minutes.

**Disabling functionalities of the Staking module.** Babylon disables two
functionalities of the Staking module, namely the validator set update mechanism
and the 21-day unbonding mechanism.

In Cosmos SDK, the Staking module handles staking-related messages and updates
the validator set upon every block. Consequently, the Staking module updates the
validator set upon every block. In order to reduce the frequency of validator
set updates to once per epoch, Babylon disables the validator set update
mechanism of the Staking module.

In addition, the Staking module enforces the "21-day unbonding rule": unbonding
validators and delegations will become unbonded after 21 days. Babylon departs
from Cosmos SDK by employing Bitcoin-assisted unbonding, where unbonding
validators and delegations become unbonded once the corresponding epoch has been
checkpointed on Bitcoin. Babylon disables the 21-day unbonding mechanism to this
end.

In order to disable the two functionalities, Babylon disables Staking module's
`EndBlocker` function that updates validator sets and unbonds mature validators
upon a block ends. Instead, upon an epoch has ended, the Epoching module will
invoke the Staking module's function that updates the validator sets. In
addition, upon an epoch has been checkpointed to Bitcoin, the Epoching module
will invoke the Staking module's function that unbonds mature validators.

**Disabling messages of the Staking module.** In order to keep the validator set
unchanged during each epoch, the Epoching module intercepts and rejects
staking-related messages that affect validator set's stake distribution via
[AnteHandler](https://docs.cosmos.network/main/learn/advanced/baseapp#antehandler),
but instead defines wrapped versions of them and forwards their unwrapped forms
to the Staking module upon an epoch ends. In the [Staking
module](https://github.com/cosmos/cosmos-sdk/blob/v0.50.3/proto/cosmos/staking/v1beta1/tx.proto),
these messages include

- `MsgCreateValidator` for creating a new validator
- `MsgDelegate` for delegating coins from a delegator to a validator
- `MsgBeginRedelegate` for redelegating coins from a delegator and source
  validator to a destination validator.
- `MsgUndelegate` for undelegating from a delegator and a validator.
- `MsgCancelUnbondingDelegation` for cancelling unbonding delegation for a
  delegator

Within these messages, `MsgCreateValidator`, `MsgDelegate`,
`MsgBeginRedelegate`, `MsgUndelegate`, and `MsgCancelUnbondingDelegation` affect
the validator set's stake distribution. The Epoching module implements an
`AnteHandler` to reject these messages, and implements wrapped versions for them
together with the Checkpointing module: `MsgWrappedCreateValidator`,
`MsgWrappedDelegate`, `MsgWrappedBeginRedelegate`, and `MsgWrappedUndelegate`.
The Epoching module receives these messages at any time, but will only process
them at the end of each epoch.

**Delaying wrapped messages to the end of epochs.** The Epoching module will
handle wrapped staking-related messages at the end of each epoch. Im particular,
the Epoching module maintains a message queue for each epoch. Upon each wrapped
message, the Epoching module performs basic sanity checks, then enqueue the
message to the message queue. When the epoch ends, the Epoching module will
forward queued messages to the Staking module. Consequently, the Staking module
receives and handles staking-related messages, thus performs validator set
updates, at the end of each epoch.

**Bitcoin-assisted Unbonding.** Babylon implements the Bitcoin-assisted
unbonding mechanism by invoking the Staking module upon an epoch becomes
checkpointed. Specifically, the Staking module's `ApplyMatureUnbondings` is
responsible for identifying and unbonding mature validators and delegations that
have been unbonding for 21 days, and is invoked upon every block. Babylon has
disabled the invocation of `ApplyMatureUnbondings` per block, and implements the
state management for epochs. Upon an epoch becomes finalized, i.e., its
checkpoint becomes deep enough in Bitcoin, the Epoching module will invoke
`ApplyMatureUnbondings` to finish all unbonding validators and delegations.

## States

The Epoching module maintains the following KV stores.

### Parameters

The [parameter storage](./keeper/params.go) maintains the Epoching module's
parameters. The Epoching module's parameters are represented as a `Params`
[object](../../proto/babylon/epoching/v1/params.proto) defined as follows:

```protobuf
// Params defines the parameters for the module.
message Params {
  option (gogoproto.equal) = true;

  // epoch_interval is the number of consecutive blocks to form an epoch
  uint64 epoch_interval = 1
      [ (gogoproto.moretags) = "yaml:\"epoch_interval\"" ];
}
```

### Epochs

The [epoch storage](./keeper/params.go) maintains the metadata of each epoch.
The key is the epoch number, and the value is an `Epoch`
[object](../../proto/babylon/epoching/v1/epoching.proto) representing the epoch
metadata.

```protobuf
// Epoch is a structure that contains the metadata of an epoch
message Epoch {
  // epoch_number is the number of this epoch
  uint64 epoch_number = 1;
  // current_epoch_interval is the epoch interval at the time of this epoch
  uint64 current_epoch_interval = 2;
  // first_block_height is the height of the first block in this epoch
  uint64 first_block_height = 3;
  // last_block_time is the time of the last block in this epoch.
  // Babylon needs to remember the last header's time of each epoch to complete
  // unbonding validators/delegations when a previous epoch's checkpoint is
  // finalised. The last_block_time field is nil in the epoch's beginning, and
  // is set upon the end of this epoch.
  google.protobuf.Timestamp last_block_time = 4 [ (gogoproto.stdtime) = true ];
  // app_hash_root is the Merkle root of all AppHashs in this epoch
  // It will be used for proving a block is in an epoch
  bytes app_hash_root = 5;
  // sealer is the last block of the sealed epoch
  // sealer_app_hash points to the sealer but stored in the 1st header
  // of the next epoch
  bytes sealer_app_hash = 6;
  // sealer_block_hash is the hash of the sealer
  // the validator set has generated a BLS multisig on the hash,
  // i.e., hash of the last block in the epoch
  bytes sealer_block_hash = 7;
}
```

### Epoch message queue

The Epoching module implements a message queue to delay the execution of
messages that affect the validator set's stake distribution to the end of each
epoch. This ensures that during an epoch, the validator set's stake distribution
remain unchanged, except for slashed validators. The [epoch message queue
storage](./keeper/epoch_msg_queue.go) maintains the queue of these
staking-related messages. The key is the epoch number concatenated with the
index of the queued message, and the value is a `QueuedMessage`
[object](../../proto/babylon/epoching/v1/epoching.proto) representing this
queued message.

```protobuf
// QueuedMessage is a message that can change the validator set and is delayed
// to the end of an epoch
message QueuedMessage {
  // tx_id is the ID of the tx that contains the message
  bytes tx_id = 1;
  // msg_id is the original message ID, i.e., hash of the marshaled message
  bytes msg_id = 2;
  // block_height is the height when this msg is submitted to Babylon
  uint64 block_height = 3;
  // block_time is the timestamp when this msg is submitted to Babylon
  google.protobuf.Timestamp block_time = 4 [ (gogoproto.stdtime) = true ];
  // msg is the actual message that is sent by a user and is queued by the
  // Epoching module
  oneof msg {
    cosmos.staking.v1beta1.MsgCreateValidator msg_create_validator = 5;
    cosmos.staking.v1beta1.MsgDelegate msg_delegate = 6;
    cosmos.staking.v1beta1.MsgUndelegate msg_undelegate = 7;
    cosmos.staking.v1beta1.MsgBeginRedelegate msg_begin_redelegate = 8;
    cosmos.staking.v1beta1.MsgCancelUnbondingDelegation msg_cancel_unbonding_delegation = 9;
  }
}
```

In Cosmos SDK, `MsgCreateValidator`, `MsgDelegate`, `MsgUndelegate`,
`MsgBeginRedelegate`, `MsgCancelUnbondingDelegation` in the Staking module might
affect the validator set, thus will be wrapped into `QueuedMessage` and be
delayed to the end of an epoch for execution.

### Epoch validator set

The [epoch validator set storage](./keeper/epoch_val_set.go) maintains the
validator set at the beginning of each epoch. The validator set will remain the
same throughout the epoch, unless some validators get slashed during this epoch.
The key is the epoch number concatenated with the validator's address, and the
value is this validator's voting power (in `sdk.Int`) at this epoch.

## Messages

The Epoching module implementing the epoched staking mechanism by using an
[AnteHandler](https://docs.cosmos.network/main/learn/advanced/baseapp#antehandler)
to intercept messages that affect the validator set's stake distribution, and
implemented their epoched counterparts for the validators and stakers.

### Disabling Staking module messages via AnteHandler

In Cosmos SDK,
[AnteHandler](https://docs.cosmos.network/main/learn/advanced/baseapp#antehandler)
is a component responsible for pre-processing transactions. It functions prior
to the execution of transaction logic, performing crucial tasks such as
validating signatures, ensuring sufficient account funds for transaction fees,
and setting up the necessary context for transaction processing. This offers
flexibility in the sense that developers can customize AnteHandlers to suit the
specific needs and rules of their applications.

The Epoching module implements an
[AnteHandler](./keeper/drop_validator_msg_decorator.go)
`DropValidatorMsgDecorator` in order to intercept messages that affect the
validator set's stake distribution in Cosmos SDK's Staking module. The messages
include `MsgCreateValidator`, `MsgDelegate`, `MsgUndelegate`,
`MsgBeginRedelegate`, `MsgCancelUnbondingDelegation`.

### Epoched staking messages

Babylon implements the epoched counterparts of the messages intercepted by
`DropValidatorMsgDecorator`, including `MsgWrappedDelegate`,
`MsgWrappedUndelegate`, `MsgWrappedBeginRedelegate`, and
`WrappedCancelUnbondingDelegation` in the Epoching module, and
`MsgWrappedCreateValidator` in the
[Checkpointing](../../proto/babylon/checkpointing/v1/tx.proto) module.

The epoched staking messages in the Epoching module are defined at
[proto/babylon/epoching/v1/tx.proto](../../proto/babylon/epoching/v1/tx.proto).
They are simply wrappers of the corresponding messages in Cosmos SDK's Staking
module.

```proto
// MsgWrappedDelegate is the message for delegating stakes
message MsgWrappedDelegate {
  option (gogoproto.equal) = false;
  option (gogoproto.goproto_getters) = false;
  option (cosmos.msg.v1.signer) = "msg";

  cosmos.staking.v1beta1.MsgDelegate msg = 1;
}
// MsgWrappedUndelegate is the message for undelegating stakes
message MsgWrappedUndelegate {
  option (gogoproto.equal) = false;
  option (gogoproto.goproto_getters) = false;
  option (cosmos.msg.v1.signer) = "msg";

  cosmos.staking.v1beta1.MsgUndelegate msg = 1;
}
// MsgWrappedDelegate is the message for moving bonded stakes from a
// validator to another validator
message MsgWrappedBeginRedelegate {
  option (gogoproto.equal) = false;
  option (gogoproto.goproto_getters) = false;
  option (cosmos.msg.v1.signer) = "msg";

  cosmos.staking.v1beta1.MsgBeginRedelegate msg = 1;
}
// MsgWrappedCancelUnbondingDelegation is the message for cancelling
// an unbonding delegation
message MsgWrappedCancelUnbondingDelegation {
  option (gogoproto.equal) = false;
  option (gogoproto.goproto_getters) = false;
  option (cosmos.msg.v1.signer) = "msg";

  cosmos.staking.v1beta1.MsgCancelUnbondingDelegation msg = 1;
}
```

The handlers of the epoched staking messages in the Epoching module are defined
at [x/epoching/keeper/msg_server.go](./keeper/msg_server.go). Each handler
performs the same [verification
logics](https://github.com/cosmos/cosmos-sdk/blob/v0.50.3/x/staking/keeper/msg_server.go)
of the corresponding message in Cosmos SDK's Staking module, then inserts the
message to the epoch message queue storage.

### MsgUpdateParams

The `MsgUpdateParams` message is used for updating the module parameters for the
Epoching module. It can only be executed via a govenance proposal.

```protobuf
// MsgUpdateParams defines a message for updating Epoching module parameters.
message MsgUpdateParams {
  option (cosmos.msg.v1.signer) = "authority";

  // authority is the address of the governance account.
  // just FYI: cosmos.AddressString marks that this field should use type alias
  // for AddressString instead of string, but the functionality is not yet implemented
  // in cosmos-proto
  string authority = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];

  // params defines the epoching parameters to update.
  //
  // NOTE: All parameters must be supplied.
  Params params = 2 [(gogoproto.nullable) = false];
}
```

## BeginBlocker and EndBlocker

Babylon disables Staking module's EndBlocker to avoid validator set updates upon
each block. The Epoching module implements `BeginBlocker` to initialize an epoch
upon the beginning of an epoch, and implements `EndBlocker` to execute all
messages and update the validator set upon the end of an epoch.

### Disabling Staking module's EndBlocker

Cosmos SDK's Staking module [updates the validator
set](https://github.com/cosmos/cosmos-sdk/blob/v0.50.3/x/staking/keeper/abci.go#L23C1-L24C1)
upon `EndBlocker` of every block. In order to implement the epoching mechanism,
Babylon disables Staking module's `EndBlocker` [as follows](../../app/app.go).

```go
// Babylon does not want EndBlock processing in staking
app.ModuleManager.OrderEndBlockers = append(app.ModuleManager.OrderEndBlockers[:2], app.ModuleManager.OrderEndBlockers[2+1:]...) // remove stakingtypes.ModuleName
```

## BeginBlocker

Upon `BeginBlocker`, the Epoching module of each Babylon node will [execute the
following](./abci.go):

1. If at the first block of the next epoch, then do the following:
   1. Enter a new epoch, i.e., create a new `Epoch` object and save it to the
      epoch metadata storage.
   2. Record the current `AppHash` as the *sealer Apphash* for the previous
      epoch. The entire `AppState` till the end of the last epoch commits to
      this `AppHash`, hence the name "sealer AppHash".
   3. Initialize the epoch message queue for the current epoch.
   4. Save the current validator set to the epoch validator set storage.
   5. Trigger hooks and emit events that the chain has entered a new epoch.
2. If at the last block of the current epoch, then record the current
   `BlockHash` as the *sealer BlockHash* for the current epoch. The entire
   blockchain so far commits to this `BlockHash` via a hash chain, hence the
   name "sealer BlockHash".

## EndBlocker

Upon `EndBlocker`, the Epoching module of each Babylon node will [execute the
following](./abci.go) *if at the last block of the current epoch*:

1. Get all queued messages of this epoch in the epoch message queue storage.
2. Forward each of the queued messages to the corresponding message handler in
   the Staking module.
3. Emit events about the execution results of the messages.
4. Invoke the Staking module to update the validator set.
5. Trigger hooks and emit events that the chain has ended the current epoch.

## Hooks

The Epoching module implements a set of hooks to notify other modules about
certain events, and utilizes the `AfterRawCheckpointFinalized`
[hook](../checkpointing/types/hooks.go) in the Checkpointing module for
Bitcoin-assisted unbonding.

### Hooks in the Epoching module

```go
// EpochingHooks event hooks for epoching validator object (noalias)
type EpochingHooks interface {
   AfterEpochBegins(ctx context.Context, epoch uint64)            // Must be called after an epoch begins
   AfterEpochEnds(ctx context.Context, epoch uint64)              // Must be called after an epoch ends
   BeforeSlashThreshold(ctx context.Context, valSet ValidatorSet) // Must be called before a certain threshold (1/3 or 2/3) of validators are slashed in a single epoch
}
```

### Bitcoin-assisted unbonding via `AfterRawCheckpointFinalized` hook

The Epoching module subscribes to the Checkpointing module's
`AfterRawCheckpointFinalized` [hook](../checkpointing/types/hooks.go) for
Bitcoin-assisted unbonding. The `AfterRawCheckpointFinalized` hook is triggered
upon a checkpoint becomes *finalized*, i.e., Bitcoin transactions of the
checkpoint become `w`-deep in Bitcoin's canonical chain, where `w` is the
`checkpoint_finalization_timeout`
[parameter](../../proto/babylon/btccheckpoint/v1/params.proto) in the
BTCCheckpoint module. Upon `AfterRawCheckpointFinalized`, the Epoching module
will finish all unbonding validators and delegations till the epoch associated
with the finalized checkpoint, including [the following](./keeper/hooks.go):

1. Find the metadata `Epoch` of the epoch associated with the finalized
   checkpoint.
2. Find the timestamp of the last block of this epoch from `Epoch`.
3. Notify the Staking module to finish all unbonding validators and delegations
   before this timestamp.

## Events

The Epoching module defines a set of events about the state updates of epochs,
validators, and delegations.

```protobuf
// EventBeginEpoch is the event emitted when an epoch has started
message EventBeginEpoch { uint64 epoch_number = 1; }
// EventEndEpoch is the event emitted when an epoch has ended
message EventEndEpoch { uint64 epoch_number = 1; }
// EventHandleQueuedMsg is the event emitted when a queued message has been handled
message EventHandleQueuedMsg {
  string original_event_type = 1;
  uint64 epoch_number = 2;
  uint64 height = 3;
  bytes tx_id = 4;
  bytes msg_id = 5;
  repeated bytes original_attributes = 6
      [ (gogoproto.customtype) =
            "github.com/cometbft/cometbft/abci/types.EventAttribute" ];
  string error = 7;
}
// EventSlashThreshold is the event emitted when a set of validators have been slashed
message EventSlashThreshold {
  int64 slashed_voting_power = 1;
  int64 total_voting_power = 2;
  repeated bytes slashed_validators = 3;
}
// EventWrappedDelegate is the event emitted when a MsgWrappedDelegate has been queued
message EventWrappedDelegate {
  string delegator_address = 1;
  string validator_address = 2;
  uint64 amount = 3;
  string denom = 4;
  uint64 epoch_boundary = 5;
}

// EventWrappedUndelegate is the event emitted when a MsgWrappedUndelegate has been queued
message EventWrappedUndelegate {
  string delegator_address = 1;
  string validator_address = 2;
  uint64 amount = 3;
  string denom = 4;
  uint64 epoch_boundary = 5;
}
// EventWrappedBeginRedelegate is the event emitted when a MsgWrappedBeginRedelegate has been queued
message EventWrappedBeginRedelegate {
  string delegator_address = 1;
  string source_validator_address = 2;
  string destination_validator_address = 3;
  uint64 amount = 4;
  string denom = 5;
  uint64 epoch_boundary = 6;
}
// EventWrappedCancelUnbondingDelegation is the event emitted when a MsgWrappedCancelUnbondingDelegation has been queued
message EventWrappedCancelUnbondingDelegation {
  string delegator_address = 1;
  string validator_address = 2;
  uint64 amount = 3;
  int64 creation_height = 4;
  uint64 epoch_boundary = 5;
}
```

## Queries

The Epoching module provides a set of queries about epochs, validators and
delegations, listed at
[docs.babylonchain.io](https://docs.babylonchain.io/docs/developer-guides/grpcrestapi#tag/Epoching).
<!-- TODO: update Babylon doc website -->
