# Babylon Chain

## Workflows

The following are the most prominent workflows. The diagrams depict cross-module communication, which hopefully helps us build a common picture of the high level interactions of the system.

### Validator Registration and Staking

In order to support regular checkpointing, Babylon has two extensions over the regular Tendermint consensus:
* the use of epochs, during which the validator set in stable
* the use of BLS keys for signature aggregation

In order to keep changes to the Cosmos SDK to a minimum and maximize code reuse, the `epoching` module _wraps_ the `staking` module: the regular staking transactions are still used, but enveloped in a Babylon transaction that allows us to attach extra data as well as to control when these transactions are executed.

![Validator Registration](diagrams/validator_registration.png)
