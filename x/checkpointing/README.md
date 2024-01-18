# Checkpointing

The Checkpointing module is responsible for generating and maintaining
the status of Babylon's Bitcoin checkpoints. The technical core
of the Checkpointing module is the [BLS signature](https://en.wikipedia.org/wiki/BLS_digital_signature)
scheme, around which this module provides the following functionalities:

- handling requests for registering Babylon validators with their BLS keys,
- signing, verifying, and aggregating BLS signatures,
- constructing checkpoints out of the BLS signatures, and
- maintaining the status of the checkpoints.

## Table of contents

- [Concepts](#concepts)
- [States](#states)
  - [Validator With BLS Key](#validator-with-bls-key)
  - [Checkpoint](#checkpoint)
  - [Genesis](#genesis)
- [Messages](#messages)
  - [MsgWrappedCreateValidator](#msgwrappedcreatevalidator)
- [ABCI++](#abci)
  - [PrepareProposal](#prepareproposal)
  - [ProcessProposal](#processproposal)
  - [ExtendVote](#extendvote)
  - [VerifyVoteExtension](#verifyvoteextension)
  - [PreBlock](#preblock)
  - [BeginBlock](#beginblock)
- [Events](#events)
- [Queries](#queries)

## Concepts

Babylon checkpoints record the state of the Babylon chain at
the end of a particular [epoch](../../x/epoching/README.md).
They are created with the intention to be included in
the Bitcoin ledger to protect Babylon and
the chains that connect with it against long range attacks.
The confirmation of a Babylon checkpoint on Bitcoin
serves as an immutable record of Babylon's state
up to the checkpointed epoch and
determines the canonical branch of the Babylon chain.

At their core, checkpoints contain a unique identifier
of the state they commit to and BLS signatures
from the validator set that corresponds to that state.
The BLS signature scheme is chosen to keep the checkpoints
verifiable and succinct, as it enables the aggregation of signatures.
To that end, each validator needs to maintain a BLS key pair
and register the BLS public key on Babylon.
Validators use their BLS private key to sign over
the block ID of the last block of the epoch and
submit their signature through an ABCI++ vote extension interface.
Valid BLS signatures are then aggregated into a checkpoint
that is included in the next block proposal.

Once a valid checkpoint is generated,
it is checkpointed into the Bitcoin ledger through
an off-chain program
[Vigilante Submitter](https://docs.babylonchain.io/docs/developer-guides/modules/submitter).
It is responsible for constructing Bitcoin transactions that
contain outputs utilizing the
[`OP_RETURN`](https://en.bitcoin.it/wiki/OP_RETURN) script code
to include the checkpoint's data in the Bitcoin ledger.
Due to the data limitations of `OP_RETURN`,
two such transactions are constructed to contain
the whole checkpoint data.
After their inclusion,
an off-chain program called the
[Vigilante Reporter](https://docs.babylonchain.io/docs/developer-guides/modules/reporter)
submits inclusion proofs to the
[BTC Checkpoint module](../../x/btccheckpoint/README.md),
which is responsible for monitoring their confirmation status and
reporting it to the Checkpointing module.
The observation of two conflicting checkpoints with a valid BLS multi-signature
means that a fork exists and an alarm will be raised.
In this case, the Babylon chain's canonical chain is represented by
the state of the checkpoint that has been included first in the Bitcoin ledger.

## States

The Checkpointing module maintains the following KV stores.

### Checkpoint

The [checkpoint state](./keeper/ckpt_state.go) maintains all the checkpoints. 
The key is the epoch number and the value is a `RawCheckpointWithMeta`
[object](../../proto/babylon/checkpointing/v1/checkpoint.proto) representing a
raw checkpoint along with some metadata.

```protobuf
// RawCheckpoint wraps the BLS multi sig with metadata
message RawCheckpoint {
  // epoch_num defines the epoch number the raw checkpoint is for
  uint64 epoch_num = 1;
  // block_hash defines the 'BlockID.Hash', which is the hash of
  // the block that individual BLS sigs are signed on
  bytes block_hash = 2 [ (gogoproto.customtype) = "BlockHash" ];
  // bitmap defines the bitmap that indicates the signers of the BLS multi sig
  bytes bitmap = 3;
  // bls_multi_sig defines the multi sig that is aggregated from individual BLS
  // sigs
  bytes bls_multi_sig = 4
  [ (gogoproto.customtype) =
    "github.com/babylonchain/babylon/crypto/bls12381.Signature" ];
}

// RawCheckpointWithMeta wraps the raw checkpoint with metadata.
message RawCheckpointWithMeta {
  option (gogoproto.equal) = true;

  RawCheckpoint ckpt = 1;
  // status defines the status of the checkpoint
  CheckpointStatus status = 2;
  // bls_aggr_pk defines the aggregated BLS public key
  bytes bls_aggr_pk = 3
  [ (gogoproto.customtype) =
    "github.com/babylonchain/babylon/crypto/bls12381.PublicKey" ];
  // power_sum defines the accumulated voting power for the checkpoint
  uint64 power_sum = 4;
  // lifecycle defines the lifecycle of this checkpoint, i.e., each state
  // transition and the time (in both timestamp and block height) of this
  // transition.
  repeated CheckpointStateUpdate lifecycle = 5;
}
```

### Validator with BLS key

The [registration state](./keeper/registration_state.go) maintains
a two-way mapping between the validator address and its BLS public key.

The Checkpoint module also stores the [validator set](../../proto/babylon/checkpointing/v1/bls_key.proto)
of every epoch with their public BLS keys. The key of the storage is the epoch 
number.

```protobuf
// ValidatorWithBLSSet defines a set of validators with their BLS public keys
message ValidatorWithBlsKeySet { repeated ValidatorWithBlsKey val_set = 1; }

// ValidatorWithBlsKey couples validator address, voting power, and its bls
// public key
message ValidatorWithBlsKey {
  // validator_address is the address of the validator
  string validator_address = 1;
  // bls_pub_key is the BLS public key of the validator
  bytes bls_pub_key = 2;
  // voting_power is the voting power of the validator at the given epoch
  uint64 voting_power = 3;
}
```

### Genesis

The [genesis state](./keeper/genesis_bls.go) maintains the BLS keys of the 
genesis validators for the Checkpointing module.

```protobuf
// GenesisState defines the checkpointing module's genesis state.
message GenesisState {
  // genesis_keys defines the public keys for the genesis validators
  repeated GenesisKey genesis_keys = 1;
}

// GenesisKey defines public key information about the genesis validators
message GenesisKey {
  // validator_address is the address corresponding to a validator
  string validator_address = 1;
  // bls_key defines the BLS key of the validator at genesis
  BlsKey bls_key = 2;
  // val_pubkey defines the ed25519 public key of the validator at genesis
  cosmos.crypto.ed25519.PubKey val_pubkey = 3;
}
```

## Messages

The Checkpointing module handles requests of registering Babylon validators.
The request message type is defined at
[proto/babylon/checkpointing/v1/tx.proto](../../proto/babylon/checkpointing/v1/tx.proto).
The message handler is defined at
[x/checkpointing/keeper/msg_server.go](./keeper/msg_server.go).

### MsgWrappedCreateValidator

The `MsgWrappedCreateValidator` message wraps the [`MsgCreateValidator`](https://github.com/cosmos/cosmos-sdk/blob/9814f684b9dd7e384064ca86876688c05e685e54/proto/cosmos/staking/v1beta1/tx.proto#L51) 
defined in the staking module of the Cosmos SDK
in order to also include the BLS public key.
The message is used for registering a new validator and storing its BLS public
key.

```protobuf
// MsgWrappedCreateValidator defines a wrapped message to create a validator
message MsgWrappedCreateValidator {
  option (cosmos.msg.v1.signer) = "msg_create_validator";

  BlsKey key = 1;
  cosmos.staking.v1beta1.MsgCreateValidator msg_create_validator = 2;
}
```

Upon `MsgWrappedCreateValidator`, a Babylon node will execute as follows:

1. Extract `MsgCreateValidator` and check its validity.
2. Extract the BLS public key and save it to the `address->key` and
   `key->address` stores. We disallow a validator to register with different
   BLS public keys or the same BLS public key being used by different
   validators.
3. Enqueue the `MsgCreateValidator` to the [Epoching module](../../x/epoching/README.md)
   which will handle this message at the end of the epoch as validator set
   change happens per epoch.

## Checkpointing via ABCI++

[ABCI++](https://docs.cometbft.com/v0.38/spec/abci/) or ABCI 2.0 is the middle
layer that controls the communication between the underlying consensus and the 
application. We use ABCI++ interfaces to generate checkpoints a part
of the CometBFT consensus. Particularly, validators are responsible for
submitting a `VoteExtension` that includes their BLS signature at the end
of each epoch. The proposer of the next block builds a checkpoint by
aggregating these signatures and
injects it as a special transaction within the proposed block.
Through this, the checkpoint is stored within the application
when the block is committed to the CometBFT ledger.
The relevant handlers are defined in
[x/checkpointing/vote_ext.go](./vote_ext.go) and
[x/checkpointing/proposal.go](./proposal.go), respectively.

### PrepareProposal

The **PrepareProposal** method is utilized by the proposer
of the next block to construct a valid proposal.
It wraps the default `PrepareProposal` handler of the Cosmos SDK
to add a special condition that checks whether the next proposed block
will be the first block of a new epoch.
In that case, it builds a valid checkpoint using the
BLS signatures that were submitted as Vote Extensions
in the previous block by the validator set.
If a valid checkpoint cannot be built, this means
something critical is happening (e.g., invalid vote extensions or
insufficient valid BLS signatures) and the proposer should panic.

The checkpoint is encoded as a special transaction and injected as the first
transaction of the proposed block. The format of the injected checkpoint is
defined in [x/proto/babylon/checkpointing/checkpoint.proto](../../proto/babylon/checkpointing/checkpoint.proto).
Note that the extended commit info that contains previous vote extensions is
also part of the injected checkpoint, which is used for re-constructing the
checkpoint in `ProcessProposal`.

```protobuf
// InjectedCheckpoint wraps the checkpoint and the extended votes
message InjectedCheckpoint {
  RawCheckpointWithMeta ckpt = 1;
  // extended_commit_info is the commit info including the vote extensions
  // from the previous proposal
  tendermint.abci.ExtendedCommitInfo extended_commit_info = 2;
}
```

### ProcessProposal

The **ProcessProposal** method is utilized by validators
for verifying the validity of a proposed block and
acts as one of the first steps towards achieving consensus
for the new block.
It wraps the default `ProcessProposal` handler of the Cosmos SDK
with an extension for verifying the validity of the checkpoint injected
as a special transaction in the first block of an epoch. The
verification steps on the special transaction are:

1. extract the special transactions from the transaction set,
2. verify the vote extensions contained in the injected checkpoint, and
3. rebuild the checkpoint from the vote extensions and verify that it is
   compatible with the checkpoint contained within the injected transaction.

If any of the above steps fail, the proposal will be rejected.

### ExtendVote

The **ExtendVote** method is responsible for creating a BLS signature when the
validator votes for the last block of an epoch. It is invoked at the final
voting phase of a consensus round. It signs the block ID of the proposal and
constructs a vote extension which will be attached to the pre-commit vote as
opaque bytes. The format of the vote extension is defined in
[x/proto/babylon/checkpointing/bls_key.proto](../../proto/babylon/checkpointing/bls_key.proto).

```protobuf
// VoteExtension defines the structure used to create a BLS vote extension.
message VoteExtension {
  // signer is the address of the vote extension signer
  string signer = 1;
  // validator_address is the address of the validator
  string validator_address = 2;
  // block_hash is the hash of the block that the vote extension is signed over
  bytes block_hash = 3;
  // epoch_num is the epoch number of the vote extension
  uint64 epoch_num = 4;
  // height is the height of the vote extension
  uint64 height =5;
  // bls_sig is the BLS signature
  bytes bls_sig = 6
  [ (gogoproto.customtype) =
    "github.com/babylonchain/babylon/crypto/bls12381.Signature" ];
}
```

### VerifyVoteExtension

**VerifyVoteExtension** is responsible for verifying the vote extension if
the voting proposal is the last block of the current epoch. It is called
when a pre-commit vote is received. It extracts the BLS signature from
the vote extension attached to the pre-commit vote and verifies it using the
corresponding BLS public key.
If the verification fails, the relevant pre-commit vote will be rejected.

### PreBlock

**PreBlock** is responsible for persistently storing the checkpoint from the
special checkpoint transaction injected on the first block of the epoch.
It is called at the first step of finalizing a block.
Since the verification is already done in `ProcessProposal`,
the `PreBlock` will store the checkpoint to the application without further
checks.

### BeginBlock

**BeginBlock** is responsible for initiating the validator set with their
BLS public keys if the current proposal is the first block of the new epoch.
It is called right after `PreBlock` during block finalization.
It reads the validator set of the epoch from the Epoching module and
associates the validator set with their BLS public keys. The logic is defined
at [x/checkpointing/abci.go]((./abci.go)).

## Events

The Checkpointing module emits events when the status of checkpoints is
changed or a conflicting checkpoint is found. The events are
defined at [proto/babylon/checkpointing/v1/events.proto](../../proto/babylon/checkpointing/v1/events.proto).

```protobuf
// EventCheckpointAccumulating is emitted when a checkpoint reaches the
// `Accumulating` state.
message EventCheckpointAccumulating { RawCheckpointWithMeta checkpoint = 1; }
// EventCheckpointSealed is emitted when a checkpoint reaches the `Sealed`
// state.
message EventCheckpointSealed { RawCheckpointWithMeta checkpoint = 1; }
// EventCheckpointSubmitted is emitted when a checkpoint reaches the `Submitted`
// state.
message EventCheckpointSubmitted { RawCheckpointWithMeta checkpoint = 1; }
// EventCheckpointConfirmed is emitted when a checkpoint reaches the `Confirmed`
// state.
message EventCheckpointConfirmed { RawCheckpointWithMeta checkpoint = 1; }
// EventCheckpointFinalized is emitted when a checkpoint reaches the `Finalized`
// state.
message EventCheckpointFinalized { RawCheckpointWithMeta checkpoint = 1; }
// EventCheckpointForgotten is emitted when a checkpoint switches to a
// `Forgotten` state.
message EventCheckpointForgotten { RawCheckpointWithMeta checkpoint = 1; }
// EventConflictingCheckpoint is emitted when two conflicting checkpoints are
// found.
message EventConflictingCheckpoint {
  RawCheckpoint conflicting_checkpoint = 1;
  RawCheckpointWithMeta local_checkpoint = 2;
}
```

## Queries

The Checkpointing module provides a set of queries about BLS keys the status of
checkpoints, listed at
[docs.babylonchain.io](https://docs.babylonchain.io/docs/developer-guides/grpcrestapi#tag/Checkpointing).
