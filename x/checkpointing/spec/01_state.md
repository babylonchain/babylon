# State

The `x/checkpointing` module manages the bls-sig collections and checkpoints for each epoch.

Upon entering a new epoch (notified by the `x/epoching` module), a bls signature signed on a commit proof is generated spread to other validators via a BLS signer.
A BLS signer is a goroutine that is initiated by the `checkpointing` module.
Then the `checkpointing` module processes `AddBlsSig` messages for the new epoch and generates a raw checkpoint when sufficient `bls-sig`s are collected for the new epoch.
Raw checkpoints are exposed to BTC relayers for submission to BTC.

Raw checkpoints have three states described as follows.

- UNCHECKPOINTED: a newly generated checkpoint is first set to UNCHECKPOINTED, waiting for BTC relayers to submit it to BTC.
- CHECKPOINTED_NOT_CONFIRMED: when a checkpoint tx is first processed by the `btccheckpoint` module, the state of the relevant raw checkpoint is set to CHECKPOINTED_NOT_CONFIRMED 
- CONFIRMED: when a btc-header tx that confirms a checkpoint (e.g., 6-block-depth) is first processed by the `btclightclient` module, the state of the relevant raw checkpoint is set to CONFIRMED

