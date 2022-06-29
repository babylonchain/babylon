# Registration

To participate in the checkpointing, a validator needs to register its BLS public key.
We assume the validator is already registered its Ed25519 key to participate consensus.
The registration needs to ensure that the sender of the BLS public key owns:
1. the corresponding BLS private key;
2. the corresponding Ed25519 private key associated with the public key when it is registered as the validator.

To achieve that, the sender needs to include Proof-of-Possession (PoP) in the registration message as follows.
```
[BLS public key, proof of possession]
```

## Proof of Possession

To create a PoP associated with the BLS public key, the validator needs to do the following steps:
1. It signs a predefined message `m`, e.g, "hello Babylon", using its Ed25519 private key and obtains a signature `A`.
2. It signs `A` using its BLS private key and obtains a signature `B`.
3. PoP is composed of [`A`, `B`]

## Verification

To verify PoP:
1. Decrypt `B` using the BLS public key and get `A'`
2. If `A'` != `A`, verification fails. Otherwise,
3. Use the validator's Ed25519 public key to decrypt `A` and get `m'`
4. If `m'` != `m`, verification fails. Otherwise, verification passes.

If verification passes, the `checkpointing` module stores the BLS public key and associate it with the validator.


