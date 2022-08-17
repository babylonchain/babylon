# Registration

To participate in the checkpointing, a validator needs to also register its BLS public key.

## Register a Validator

The original registration is done via a transaction that carries a `MsgCreateValidator` message.
To register a BLS public key, we need a wrapper message called `MsgWrappedCreateValidator` processed by the `Checkpointing` module.
This message wraps the original `MsgCreateValidator` message as well as a BLS public key and a `Proof-of-Possession` (PoP) for registering BLS public key.
The execution of `MsgWrappedCreateValidator` is as follows.

1. The `Checkpointing` module first processes `MsgWrappedCreateValidator` to register the validator's BLS key. If success, then
2. extract `MsgCreateValidator` and deliver `MsgCreateValidator` to the epoching module's message queue, which will be processed until the end of this epoch. If success, the registration is succeeded.
3. Otherwise, the registration fails and the validator should register again with the same keys.

## Genesis

Genesis validators are registered via the legacy `genutil` module from the Cosmos-SDK, which processes `MsgCreateValidator` messages contained in genesis transactions.
The BLS keys are registered as `GenesisState` in the checkpointing module.
The checkpointing module's `ValidateGenesis` should ensure that each genesis validator has both an Ed25519 key and BLS key which are bonded by PoP.

## Proof of Possession

The purpose of PoP is to prove that one validator owns:
1. the corresponding BLS private key;
2. the corresponding Ed25519 private key associated with the public key in the `MsgCreateValidator` message.

To achieve that, PoP is calculated as follows.

`PoP = sign(key = BLS_sk, data = sign(key = Ed25519_sk, data = BLS_pk)]`

Since the delegator already relates its account with the validator's Ed25519 key through the signatures in `MsgCreateValidator`, the adversary cannot do registration with the same PoP.

## Verification

To verify PoP, first we need to ensure that the BLS public key has never been registered by a different validator,
and that the current validator hasn't already registered a different BLS public key. Then, verify

```
MsgWrappedCreateValidator.BLS_pk ?= decrypt(key = Ed25519_pk, data = decrypt(key = BLS_pk, data = PoP))
```

If verification passes, the `Checkpointing` module stores the BLS public key and associates it with the validator.
