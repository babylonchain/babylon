# Finality

Babylon's BTC Staking protocol introduces an additional consensus round on
blocks produced by CometBFT, called the finality round. The participants of this
round are referred as finality providers and their voting power stems from
staked bitcoins delegated to them.

The Finality module is responsible for handling finality votes, maintaining the
finalization status of blocks, and identifying equivocating finality providers
in the finalization rounds. This includes:

- handling requests for submitting finality votes from finality providers;
- maintaining the finalization status of blocks; and
- maintaining equivocation evidences of culpable finality providers.

## Table of contents

- [Table of contents](#table-of-contents)
- [Concepts](#concepts)
- [States](#states)
  - [Finality votes](#finality-votes)
  - [Indexed blocks with finalization status](#indexed-blocks-with-finalization-status)
  - [Equivocation evidences](#equivocation-evidences)
- [Messages](#messages)
  - [MsgAddFinalitySig](#msgaddfinalitysig)
  - [MsgUpdateParams](#msgupdateparams)
- [EndBlocker](#endblocker)
- [Events](#events)
- [Queries](#queries)

## Concepts

<!-- summary of BTC staking protocol and BTC staking module -->
**Babylon Bitcoin Staking.** Babylon's Bitcoin Staking protocol allows bitcoin
holders to *trustlessly* stake their bitcoins, in order to provide economic
security to the Babylon chain and other Proof-of-Stake (PoS) blockchains. The
protocol composes a PoS blockchain with an off-the-shelf *finality voting round*
run by a set of [finality
providers](https://github.com/babylonchain/finality-provider) who receive *BTC
delegations* from [BTC stakers](https://github.com/babylonchain/btc-staker). The
finality providers and BTC delegations are maintained by Babylon's [BTC Staking
module](../btcstaking/README.md), and the Finality module is responsible for
maintaining the finality voting round.

<!-- introducing finality voting round, Finality module -->
**Finality voting round.**  In the finality voting round, a block committed in
the CometBFT ledger receives *finality votes* from a set of finality providers.
A finality vote is a signature under the [*Extractable One-Time Signature
(EOTS)*
primitive](https://docs.babylonchain.io/assets/files/btc_staking_litepaper-32bfea0c243773f0bfac63e148387aef.pdf).
A block is considered finalized if it receives a quorum, i.e., votes from
finality providers with more than 2/3 voting power at its height.

<!-- Babylon BTC staking security guarantee, i.e., slashable safety -->
**Slashable safety guarantee.** The finality voting round ensures the *slashable
safety* property of finalized blocks: upon a safety violation where a
conflicting block also receives a valid quorum, adversarial finality providers
with more than 1/3 total voting power will be provably identified by the
protocol and be slashed. The formal definition of slashable safety can be found
at [the S&P'23 paper](https://arxiv.org/pdf/2207.08392.pdf) and [the CCS'23
paper](https://arxiv.org/pdf/2305.07830.pdf). In Babylon's Bitcoin Staking
protocol, if a finality provider is slashed, then

- the secret key of the finality provider is revealed to the public,
- a parameterized amount of bitcoins of all BTC delegations under it will be
  burned *on the Bitcoin network*, and
- the finality provider's voting power will be zeroized.

In addition to the standard safety guarantee of CometBFT consensus, the
slashable safety guarantee disincentivizes safety offences launched by
adversarial finality providers.

<!-- user stories of finality provider and finality module -->
**Interaction between finality providers and the Finality module.** In order to
participate in the finality voting round, an active finality provider with BTC
delegations (as specified in the [BTC Staking module](../btcstaking/README.md))
needs to interact with Babylon as follows:

- **Committing EOTS master public randomness.** The finality provider needs to
  generate a pair of EOTS master secret/public randomness, and commit the master
  public randomness when registering itself to Babylon. The EOTS master
  secret/public randomness allows to derive a EOTS secret/public randomness
  deterministically for each given height, respectively. Babylon further
  requires the epoch of the finality provider registration to be finalized by
  BTC timestamping in order to submit finality signatures. This ensures that
  each finality provider has a unique public randomness for each height, and
  that if the finality provider submits two finality signatures over two
  conflicting blocks, anyone can extract the finality provider's secret key
  using EOTS.
- **Submitting EOTS signatures.** Upon a new block, the finality provider
  submits an EOTS signature w.r.t. the derived public randomness at that height.
  The Finality module will verify the EOTS signature, and check if there are
  known EOTS signatures on conflicting blocks from this finality provider. If
  yes, then this constitutes an equivocation, and the Finality module will save
  the equivocation evidence, such that anyone can extract the finality
  provider's secret key and slash it.

Babylon has implemented a [BTC staking
tracker](https://github.com/babylonchain/vigilante) daemon program that
subscribes to equivocation evidences in the Finality module, and slashes BTC
delegations under equivocating finality providers by sending their slashing
transactions to the Bitcoin network.

## States

The Finality module maintains the following KV stores.

### Finality votes

The [finality vote storage](./keeper/votes.go) maintains the finality votes of
finality providers on blocks. The key is the block height concatenated with the
finality provider's Bitcoin secp256k1 public key, and the value is a
`SchnorrEOTSSig` [object](../../types/btc_schnorr_eots.go) representing an EOTS
signature. Here, the EOTS signature is signed over a block's height and
`AppHash` by the finality provider, using the private randomness corresponding
to the EOTS public randomness derived using the block height. The EOTS signature
serves as a finality vote on this block from this finality provider. It is a
32-byte scalar and is defined as a 32-byte array in the implementation.

```go
type SchnorrEOTSSig []byte
const SchnorrEOTSSigLen = 32
```

### Indexed blocks with finalization status

The [indexed block storage](./keeper/indexed_blocks.go) maintains the necessary
metadata and finalization status of blocks. The key is the block height and the
value is an `IndexedBlock` object
[defined](../../proto/babylon/finality/v1/finality.proto) as follows.

```protobuf
// IndexedBlock is the necessary metadata and finalization status of a block
message IndexedBlock {
    // height is the height of the block
    uint64 height = 1;
    // app_hash is the AppHash of the block
    bytes app_hash = 2;
    // finalized indicates whether the IndexedBlock is finalised by 2/3
    // finality providers or not
    bool finalized = 3;
}
```

### Equivocation evidences

The [equivocation evidence storage](./keeper/evidence.go) maintains evidences of
equivocation offences committed by finality providers. The key is a finality
provider's Bitcoin secp256k1 public key concatenated with the block height, and
the value is an `Evidence`
[object](../../proto/babylon/finality/v1/finality.proto) representing the
evidence that this finality provider has equivocated at this height. Anyone
observing the `Evidence` object can extract the finality provider's Bitcoin
secp256k1 secret key, as per EOTS's extractability property.

```protobuf
// Evidence is the evidence that a finality provider has signed finality
// signatures with correct public randomness on two conflicting Babylon headers
message Evidence {
    // fp_btc_pk is the BTC PK of the finality provider that casts this vote
    bytes fp_btc_pk = 1 [ (gogoproto.customtype) = "github.com/babylonchain/babylon/types.BIP340PubKey" ];
    // block_height is the height of the conflicting blocks
    uint64 block_height = 2;
    // master_pub_rand is the master public randomness the finality provider has committed to
    // encoded as a base58 string
    string master_pub_rand = 3;
    // canonical_app_hash is the AppHash of the canonical block
    bytes canonical_app_hash = 4;
    // fork_app_hash is the AppHash of the fork block
    bytes fork_app_hash = 5;
    // canonical_finality_sig is the finality signature to the canonical block
    // where finality signature is an EOTS signature, i.e.,
    // the `s` in a Schnorr signature `(r, s)`
    // `r` is the public randomness that is already committed by the finality provider
    bytes canonical_finality_sig = 6 [ (gogoproto.customtype) = "github.com/babylonchain/babylon/types.SchnorrEOTSSig" ];
    // fork_finality_sig is the finality signature to the fork block
    // where finality signature is an EOTS signature
    bytes fork_finality_sig = 7 [ (gogoproto.customtype) = "github.com/babylonchain/babylon/types.SchnorrEOTSSig" ];
}
```

## Messages

The Finality module handles the following messages from finality providers. The
message formats are defined at
[proto/babylon/finality/v1/tx.proto](../../proto/babylon/finality/v1/tx.proto).
The message handlers are defined at
[x/finality/keeper/msg_server.go](./keeper/msg_server.go).

### MsgAddFinalitySig

The `MsgAddFinalitySig` message is used for submitting a finality vote, i.e., an
EOTS signature over a block signed by a finality provider. It is typically
submitted by a finality provider via the [finality
provider](https://github.com/babylonchain/finality-provider) program.

```protobuf
// MsgAddFinalitySig defines a message for adding a finality vote
message MsgAddFinalitySig {
    option (cosmos.msg.v1.signer) = "signer";

    string signer = 1;
    // fp_btc_pk is the BTC PK of the finality provider that casts this vote
    bytes fp_btc_pk = 2 [ (gogoproto.customtype) = "github.com/babylonchain/babylon/types.BIP340PubKey" ];
    // block_height is the height of the voted block
    uint64 block_height = 3;
    // block_app_hash is the AppHash of the voted block
    bytes block_app_hash = 4;
    // finality_sig is the finality signature to this block
    // where finality signature is an EOTS signature, i.e.,
    // the `s` in a Schnorr signature `(r, s)`
    // `r` is the public randomness that is already committed by the finality provider
    bytes finality_sig = 5 [ (gogoproto.customtype) = "github.com/babylonchain/babylon/types.SchnorrEOTSSig" ];
}
```

Upon `MsgAddFinalitySig`, a Babylon node will execute as follows:

1. Ensure the finality provider has been registered in Babylon and is not
   slashed.
2. Ensure the finality provider has voting power at this height.
3. Ensure the finality provider has not previously casted the same vote.
4. Derive the EOTS public randomness using the committed EOTS master public
   randomness and the block height.
5. Verify the EOTS signature w.r.t. the derived EOTS public randomness.
6. If the voted block's `AppHash` is different from the canonical block at the
   same height known by the Babylon node, then this means the finality provider
   has voted for a fork. Babylon node buffers this finality vote to the evidence
   storage. If the finality provider has also voted for the block at the same
   height, then this finality provider is slashed, i.e., its voting power is
   removed, equivocation evidence is recorded, and a slashing event is emitted.
7. If the voted block's `AppHash` is same as that of the canonical block at the
   same height, then this means the finality provider has voted for the
   canonical block, and the Babylon node will store this finality vote to the
   finality vote storage. If the finality provider has also voted for a fork
   block at the same height, then this finality provider will be slashed.

### MsgUpdateParams

The `MsgUpdateParams` message is used for updating the module parameters for the
Finality module. It can only be executed via a govenance proposal.

```protobuf
// MsgUpdateParams defines a message for updating finality module parameters.
message MsgUpdateParams {
    option (cosmos.msg.v1.signer) = "authority";
  
    // authority is the address of the governance account.
    // just FYI: cosmos.AddressString marks that this field should use type alias
    // for AddressString instead of string, but the functionality is not yet implemented
    // in cosmos-proto
    string authority = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  
    // params defines the finality parameters to update.
    //
    // NOTE: All parameters must be supplied.
    Params params = 2 [(gogoproto.nullable) = false];
}
```

## EndBlocker

Upon `EndBlocker`, the Finality module of each Babylon node will [execute the
following](./abci.go) *if the BTC staking protocol is activated (i.e., there has
been >=1 active BTC delegations)*:

1. Index the current block, i.e., extract its height and `AppHash`, construct an
   `IndexedBlock` object, and save it to the indexed block storage.
2. Tally all non-finalized blocks as follows:
   1. Find the starting height that the Babylon node should start to finalize.
      This is the earliest height that is not finalize yet since the activation
      of BTC staking.
   2. For each `IndexedBlock` between the starting height and the current
      height, tally this block as follows:
      1. Find the set of active finality providers at this height.
      2. If the finality provider set is empty, then this block is not
         finalizable and the Babylon node will skip this block.
      3. If the finality provider set is not empty, then find all finality votes
         on this `IndexedBlock`, and check whether this `IndexedBlock` has
         received votes of more than 2/3 voting power from the active finality
         provider set. If yes, then finalize this block, i.e., set this
         `IndexedBlock` to be finalized in the indexed block storage and
         distribute rewards to the voted finality providers and their BTC
         delegations. Otherwise, none of the subsequent blocks shall be
         finalized and the loop breaks here.

## Events

The Finality module defines the `EventSlashedFinalityProvider` event. It is
emitted when a finality provider is slashed due to equivocation.

```protobuf
// EventSlashedFinalityProvider is the event emitted when a finality provider is slashed
// due to signing two conflicting blocks
message EventSlashedFinalityProvider {
    // evidence is the evidence that the finality provider double signs
    Evidence evidence = 1;
}
```

## Queries

The Finality module provides a set of queries about finality signatures on each
block, listed at
[docs.babylonchain.io](https://docs.babylonchain.io/docs/developer-guides/grpcrestapi#tag/Finality).
<!-- TODO: update Babylon doc website -->
