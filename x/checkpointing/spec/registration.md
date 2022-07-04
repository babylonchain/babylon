# Registration

To participate in the checkpointing, a validator needs to also register its BLS public key.

## Register a Validator

The original registration is done via a transaction that carries a `MsgCreateValidator` message.
To register a BLS public key, we need a wrapper message called `MsgWrappedCreateValidator` processed by the `Checkpointing` module.
This message wraps the original `MsgCreateValidator` message as well as a new message called `MsgCreateBlsKey` specifically for registering BLS public key.
The execution of `MsgWrappedCreateValidator` is as follows.

1. The `Checkpointing` module first processes `MsgCreateBlsKey` to register the validator's BLS key. If success, then
2. delay the processing of `MsgCreateValidator` until the end of this epoch. If success, the registration is succeeded.
3. Otherwise, the corresponding BLS key registered before should be removed to ensure atomicity.
4. The validator should register again.

## Proof of Possession

A valid `MsgCreateBlsKey` needs to ensure that the sender of the BLS public key owns:
1. the corresponding BLS private key;
2. the corresponding Ed25519 private key associated with the public key in the `MsgCreateValidator` message.

To achieve that, the sender needs to include Proof-of-Possession (PoP) in the `MsgCreateBlsKey` message as follows.
```
MsgCreateBlsKey = [BLS_pk, PoP],
```
where `PoP = [m = Ed25519_pk, sig_BLS = sign(key = BLS_sk, data = m), sig_Ed25519 = sign(key = Ed25519_sk, data = sig_BLS)]`

## Verification

To verify PoP, first we need to ensure that the BLS public key has never been registered by a different validator,
and that the current validator hasn't already registered a different BLS public key. Then, verify

```
MsgCreateValidator.Ed25519_pk ?= decrypt(key = BLS_pk, data = decrypt(key = Ed25519_pk, data = PoP.sig_Ed25519))
```

If verification passes, the `Checkpointing` module stores the BLS public key and associates it with the validator.
