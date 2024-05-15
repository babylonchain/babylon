# Observable Staking Transactions Specification

## Introduction

A lock-only network involves users locking their Bitcoin using the self-custodial
Bitcoin Staking script without a Babylon chain operating.
In this document, we precisely define how one can construct
the Bitcoin transactions specified by the Bitcoin Staking protocol.

## Prerequisites

- [Scripts doc](staking-script.md) - document which defines how different
Babylon scripts look like
- [BIP341](https://github.com/bitcoin/bips/blob/master/bip-0341.mediawiki)-
a document specifying how to spend taproot outputs

## System parameters

The lock-only staking system is governed by a set of parameters that specify
what constitutes a valid staking transaction. Based on those,
an observer of the Bitcoin ledger can precisely identify which transactions
are valid staking transactions and whether they should be considered active stake.
These parameters are different depending on the Bitcoin height a transaction is
included in and a constructor of a Bitcoin Staking transaction should take them into
account before propagating a transaction to Bitcoin.
For the rest of the document, we will refer to those parameters as `global_parameters`.

More details about parameters can be found in the
[parameters spec](https://github.com/babylonchain/networks/tree/main/bbn-test-4/parameters).

## Taproot outputs

Taproot outputs are outputs whose locking script is an elliptic curve point `Q`
created as follows:
```
Q = P + hash(P||m)G
```
where:
- `P` is the internal public key
- `m` is the root of a Merkle tree whose leaves consist of a version number and a
script

For Bitcoin Staking transactions, the internal public key is chosen as:

```
P = lift_x(0x50929b74c1a04954b78b4b6035e97a5e078a5a0f28ec96d547bfee9ace803ac0)
```

This key is described in the
[BIP341](https://github.com/bitcoin/bips/blob/master/bip-0341.mediawiki#constructing-and-spending-taproot-outputs)
specification.

Using this key as an internal public key disables spending from taproot output
through the key spending path.
The construction of this key can be found [here](../btcstaking/types.go?plain=1#L27).

## Observable Staking Transactions

### Staking transaction

A staking transaction is a transaction that allows staker entry into the system.

#### Requirements

For the transaction to be considered a valid staking transaction, it must:
- have a taproot output which has the key spending path disabled
and commit to a script tree composed of three scripts:
timelock script, unbonding script, slashing script.
This output is henceforth known as the `staking_output` and
the value in this output is known as `staking_amount`
- have `OP_RETURN` output which contains: `global_parameters.tag`,
 `version`, `staker_pk`,`finality_provider_pk`, `staking_time`
- all the values must be valid for the `global_parameters` which are applicable at
the height in which the staking transaction is included in the BTC ledger.


#### OP_RETURN output description

Data in the OP_RETURN output is described by the following struct:

```go
type V0OpReturnData struct {
	MagicBytes                []byte
	Version                   byte
	StakerPublicKey           []byte
	FinalityProviderPublicKey []byte
	StakingTime               []byte
}
```
The implementation of the struct can be found [here](../btcstaking/identifiable_staking.go?pain=1#L52)

Fields description:
- `MagicBytes` - 4 bytes, tag which is used to identify the staking transaction
among other transactions in the Bitcoin ledger.
It is specified in the `global_parameters.Tag` field.
- `Version` - 1 byte, current version of the OP_RETURN output
- `StakerPublicKey` - 32 bytes, staker public key. The same key must be used in
the scripts used to create the taproot output in the staking transaction.
- `FinalityProviderPublicKey` - 32 bytes, finality provider public key. The same key
must be used in the scripts used to create the taproot output in the
staking transaction.
- `StakingTime` - 2 bytes big-endian unsigned number, staking time.
The same timelock time must be used in scripts used to create the taproot
output in the staking transaction.


This data is serialized as follows:
```
SerializedStakingData = MagicBytes || Version || StakerPublicKey || FinalityProviderPublicKey || StakingTime
```

To transform this data into OP_RETURN data:

```
StakingDataPkScript = 0x6a || 0x47 || SerializedStakingData
```

where:
- 0x6a - is byte marker representing OP_RETURN op code
- 0x47 - is byte marker representing OP_DATA_71 op code, which pushed 71 bytes onto the stack

The final OP_RETURN output will have the following shape:
```
TxOut {
 Value: 0,
 PkScript: StakingDataPkScript
}
```

Logic creating output from data can be found [here](../btcstaking/identifiable_staking.go?pain=1#L175)


#### Staking output description

Staking output should commit to three scripts:
- `timelock_script`
- `unbonding_script`
- `slashing_script`

Data needed to create `staking_output`:
- `staker_public_key` - chosen by the user sending the staking transaction. It
will be used in every script. This key needs to be put in the OP_RETURN output
in the staking transaction.
- `finality_provider_public_key` - chosen by the user sending the staking
transaction. It will be used as `<FinalityPk>` in `slashing_script`. In the
lock-only network there is no slashing, so this key has mostly informative purposes.
This key needs to be put in the OP_RETURN output of the staking transaction.
- `staking_time` - chosen by the user sending the staking transaction. It will
be used as locking time in the `timelock_script`. It must be a valid `uint16` number,
in the range `global_parameters.min_staking_time <= staking_time <= global_parameters.max_staking_time`.
It needs to be put in the OP_RETURN output of the staking transaction.
- `covenant_committee_public_keys` - it can be retrieved from
`global_parameters.covenant_pks`. It is set of covenant committee public keys which
will be put in `unbonding_script` and `slashing_script`.
- `covenant_committee_quorum` - it can be retrieved from
`global_parameters.covenant_quorum`. It is quorum of covenant committee
member required to authorize spending using `unbonding_script` or `slashing_script`
- `staking_amout` - chosen by the user, it will be placed in `staking_output.value`
- `btc_network` - btc network on which staking transactions will take place

#### Building OP_RETRUN and staking output implementation

Babylon staking library exposes [BuildV0IdentifiableStakingOutputsAndTx](../btcstaking/identifiable_staking.go?plain=1#L231)
function with the following signature:

```go
func BuildV0IdentifiableStakingOutputsAndTx(
	magicBytes []byte,
	stakerKey *btcec.PublicKey,
	fpKey *btcec.PublicKey,
	covenantKeys []*btcec.PublicKey,
	covenantQuorum uint32,
	stakingTime uint16,
	stakingAmount btcutil.Amount,
	net *chaincfg.Params,
) (*IdentifiableStakingInfo, *wire.MsgTx, error)
```

It enables the caller to create valid outputs to put inside an unfunded and not-signed
staking transaction.

The suggested way of creating and sending a staking transaction using
[bitcoind](https://github.com/bitcoin/bitcoin) is:
1. create `staker_key` in the bitcoind wallet
2. create unfunded and not signed staking transaction using
the `BuildV0IdentifiableStakingOutputsAndTx` function
3. serialize the unfunded and not signed staking transaction to `staking_transaction_hex`
4. call `bitcoin-cli fundrawtransaction "staking_transaction_hex"` to
retrieve `funded_staking_transaction_hex`.
The bitcoind wallet will automatically choose unspent outputs to fund this transaction.
5. call `bitcoin-cli signrawtransactionwithwallet "funded_staking_transaction_hex"`.
This call will sign all inputs of the transaction and return `signed_staking_transaction_hex`.
6. call `bitcoin-cli sendrawtransaction "signed_staking_transaction_hex"`

### Unbonding transaction

The unbonding transaction allows the staker to on-demand unbond their
locked Bitcoin stake prior to its original timelock expiration.

#### Requirements

For the transaction to be considered a valid unbonding transaction, it must:
- have exactly one input and one output
- input must be valid a staking output
- output must be a taproot output. This taproot output must have disabled
the key spending path, and committed to script tree composed of two scripts:
the timelock script and the slashing script. This output is henceforth known
as the `unbonding_output`
- timelock in the time lock script must be equal to `global_parameters.unbonding_time`
- value in the unbonding output must be equal to `staking_output.value - global_parameters.unbonding_fee`

#### Building Unbonding output

The Babylon Bitcoin staking library exposes
the [BuildUnbondingInfo](../btcstaking/types.go?plain=1#416)
function which builds a valid unbonding output.
It has the following signature:

```go
func BuildUnbondingInfo(
	stakerKey *btcec.PublicKey,
	fpKeys []*btcec.PublicKey,
	covenantKeys []*btcec.PublicKey,
	covenantQuorum uint32,
	unbondingTime uint16,
	unbondingAmount btcutil.Amount,
	net *chaincfg.Params,
) (*UnbondingInfo, error)
```

where:
- `stakerKey`- must be the same key as the staker key in `staking_transaction`
- `fpKeys` - must contain one key, which is the same finality provider key used
in `staking_transaction`
- `covenantKeys`- are the same covenant keys as used in `staking_transaction`
- `covenantQuorum` - is the same quorum as used in `staking_transaction`
- `unbondingTime` - is equal to `global_parameters.unbonding_time`
- `unbondingAmount` - is equal to `staking_amount - global_parameters.unbonding_fee`

## Spending taproot outputs

To create transactions which spend from taproot outputs, either staking output
or unbonding output, providing signatures satisfying the script is not enough.

The spender must also provide:
- the whole script which is being spent
- the control block which contains: leaf version, internal public key, and proof of
inclusion of the given script in the script tree

Given that creating scripts is deterministic for given data, it is possible to
avoid storing scripts by re-building scripts when the need arises.

### Re-creating script and control block

To build the script and control block necessary to spend from a staking output through the
timelock script, the following function could be implemented

```go
import (
	// Babylon btc staking library
	"github.com/babylonchain/babylon/btcstaking"
)

func buildTimelockScriptAndControlBlock(
	stakerKey *btcec.PublicKey,
	finalityProviderKey *btcec.PublicKey,
	covenantKeys []*btcec.PublicKey,
	covenantQuorum uint32,
	stakingTime uint16,
	stakingAmount btcutil.Amount,
	netParams *chaincfg.Params,
) ([]byte, []byte, error) {

	stakingInfo, err := btcstaking.BuildStakingInfo(
		stakerKey,
		[]*btcec.PublicKey{finalityProviderKey},
		covenantKeys,
		covenantQuorum,
		stakingTime,
		stakingAmount,
		netParams,
	)

	if err != nil {
		return nil, nil, err
	}

	si, err := stakingInfo.TimeLockPathSpendInfo()

	if err != nil {
		return nil, nil, err
	}

	scriptBytes := si.RevealedLeaf.Script

	controlBlock := si.ControlBlock

	controlBlockBytes, err := controlBlock.ToBytes()
	if err != nil {
		return nil, nil, err
	}

	return scriptBytes, controlBlockBytes, nil
}

```

The returned script and control block can be used to either build the witness directly
or to put them in a PSBT which can be used by bitcoind to create the witness.

### Creating PSBT to get signature for given taproot path from Bitcoind

To avoid creating signatures/witness manually,
Bitcoind's [walletprocesspsbt](https://developer.bitcoin.org/reference/rpc/walletprocesspsbt.html)
can be used. To use this Bitcoind endpoint to get signature/witness the wallet must
maintain one of the keys used in the script.

Example of creating psbt to sign unbonding transaction using unbonding script from
staking output:

```go
import (
	"github.com/btcsuite/btcd/btcutil/psbt"
)

func BuildPsbtForSigningUnbondingTransaciton(
	unbondingTx *wire.MsgTx,
	stakingOutput *wire.TxOut,
	stakerKey *btcec.PublicKey,
	spentLeaf *txscript.TapLeaf,
	controlBlockBytes []byte,
) (string, error) {
	psbtPacket, err := psbt.New(
		[]*wire.OutPoint{&unbondingTx.TxIn[0].PreviousOutPoint},
		unbondingTx.TxOut,
		unbondingTx.Version,
		unbondingTx.LockTime,
		[]uint32{unbondingTx.TxIn[0].Sequence},
	)

	if err != nil {
		return "", fmt.Errorf("failed to create PSBT packet with unbonding transaction: %w", err)
	}

	psbtPacket.Inputs[0].SighashType = txscript.SigHashDefault
	psbtPacket.Inputs[0].WitnessUtxo = stakingOutput
	psbtPacket.Inputs[0].Bip32Derivation = []*psbt.Bip32Derivation{
		{
			PubKey: stakerKey.SerializeCompressed(),
		},
	}

	psbtPacket.Inputs[0].TaprootLeafScript = []*psbt.TaprootTapLeafScript{
		{
			ControlBlock: controlBlockBytes,
			Script:       spentLeaf.Script,
			LeafVersion:  spentLeaf.LeafVersion,
		},
	}

	return psbtPacket.B64Encode()
}

```

Given that to spend through the unbonding script requires more than the
staker's signature, the `walletprocesspsbt` endpoint will produce a new psbt
with the staker signature attached.

In the case of a timelock path which requires only the staker's signature,
`walletprocesspsbt` would produce the whole witness required to send the
transaction to the BTC network.
