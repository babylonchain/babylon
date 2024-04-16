# Finality Provider

Finality providers are responsible for voting at a finality round on top of CometBFT. Similar to any native PoS validator, a finality provider can receive voting power delegations from BTC stakers, and can earn commission from the staking rewards denominated in Babylon tokens.

## Creation

- First of all the finality providers will need a key on the consumer chain, in this case `babylon`, for creating a key use `babylond keys add [name]`.

By example.:

```shell
$~ babylond keys add finality-provider
```

- Since the finality provider will receive delegations he also will need...