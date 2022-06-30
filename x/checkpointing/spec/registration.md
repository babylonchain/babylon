# Registration

To participate in the checkpointing, a validator needs to also register its BLS public key.

## Register BLS Key

The original registration is done via a transaction that carries a `MsgCreateValidator` message.
To register a BLS public key, we need a new message called `MsgCreateBlsKey` which is included in the same transaction with `MsgCreateValidator`.
The execution of the transaction ensures atomicity.

1. The `Staking` module processes `MsgCreateValidator` first as usual. If success, then
2. the `Checkpointing` module process `MsgCreateBlsKey`. If success, the registration succeeds.

A valid `MsgCreateBlsKey` needs to ensure that the sender of the BLS public key owns:
1. the corresponding BLS private key;
2. the corresponding Ed25519 private key associated with the public key in `MsgCreateValidator` when it is registered as the validator.

To achieve that, the sender needs to include Proof-of-Possession (PoP) in the `MsgCreateBlsKey` as follows.
```
[BLS public key, proof of possession]
```

## Proof of Possession

To create a PoP associated with the BLS public key, the validator needs to do the following steps:
1. It signs a message `m=Ed25519 public key` using its BLS private key and obtains a signature `A`.
Note that the Ed25519 public used as `m` must be the same one used in `MsgCreateValidator`.
2. It signs `A` using its Ed25519 private key and obtains a signature `B`.
3. PoP is composed of [`m`, `A`, `B`].

## Verification

To verify PoP, first we need to ensure that the BLS public key has never been registered. Then,
1. if `m` != Ed25519 public key of the validator from `MsgCreateValidator`, verification fails. Otherwise,
2. decrypt `B` using `m`.
3. If `A'` != `A`, verification fails. Otherwise,
4. Use the validator's BLS public key to decrypt `A` and get `m'`.
5. If `m'` != `m`, verification fails. Otherwise, verification passes.

If verification passes, the `Checkpointing` module stores the BLS public key and associates it with the validator.


