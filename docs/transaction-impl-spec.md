# Lock only network transactions implementation spec

## Introduction
In lock only network there is no Babylon chain. Users lock their bitcoins on BTC
network using special BTC transactions. The purpose of this doc is to precisely
define those transactions.


## Prerequisites
- [Scripts doc](staking-script.md) - document which defines how different
Babylon scripts look like
- [BIP341](https://github.com/bitcoin/bips/blob/master/bip-0341.mediawiki)-
a document specifying how to spend taproot outputs

## System parameters

At different times, as perceived by BTC height, there is a different set of
parameters which must be followed when constructing following transactions.
In this document, those parameters will be referred to as `global_parameters.

More details about parameters can be found [here](https://github.com/babylonchain/phase1-devnet/tree/main/parameters)

## Taproot outputs

Taproot outputs are outputs whose locking script is elliptic curve point `Q`
created as follows:
```
Q = P + hash(P||m)G
```
where:
- `P` is internal public key
- `m` is the root of a Merkle tree whose leaves consist of a version number and a
script

In Babylon transactions, internal public key is chosen as:

```
P = lift_x(0x50929b74c1a04954b78b4b6035e97a5e078a5a0f28ec96d547bfee9ace803ac0)
```

This key is described in [BIP341](https://github.com/bitcoin/bips/blob/master/bip-0341.mediawiki#constructing-and-spending-taproot-outputs)
specification.

Using this key as an internal public key disables spending from taproot output
through key spending path.

Construction of this key can be found [here](../btcstaking/types.go?plain=1#L27)

## Locking network special transactions

### Staking transaction

A staking transaction is a transaction that allows staker entry into the system.

#### Requirements

For the transaction to be considered a valid staking transaction, it must:
- have output which is taproot output. This taproot output must have disabled
key spending path, and commit to a script tree composed of three scripts:
timelock script, unbonding script, slashing script. This output is henceforth known as
`staking_output` and value in this output is known as `staking_amount`
- have OP_RETURN output which contains: `global_parameters.tag`,
 `version`, `staker_pk`,`finality_provider_pk`, `staking_time`
- all the values must be valid for `global_parameters` which are applicable at
height when the staking transaction was included in the BTC ledger.


#### OP_RETURN output description

Data in op_return output can be described by the following struct:
```go
type V0OpReturnData struct {
	MagicBytes                []byte
	Version                   byte
	StakerPublicKey           []byte
	FinalityProviderPublicKey []byte
	StakingTime               []byte
}
```
and its implementation can be found [here](../btcstaking/identifiable_staking.go?pain=1#L52)

Fields description:
`MagicBytes` - 4 byte, tag that will be used to identify output among other
outputs on the BTC ledger. It can be retrieved from `global_parameters.tag`.
`Version` - 1 byte, current version of the op_return output
`StakerPublicKey` - 32 byte, staker public key. The same key must be used in
scripts used to create the taproot output in the staking transaction.
`FinalityProviderPublicKey` - 32 byte, finality provider public key. The same key
must be used in scripts used to create the taproot output in the
staking transaction.
`StakingTime` - 2 byte big-endian unsigned number, staking time.
The same timelock time be mus be used in scripts used to create the taproot
output in the staking transaction.


This data is serialized as follows:
```
SerializedStakingData = MagicBytes || Version || StakerPublicKey || FinalityProviderPublicKey || StakingTime
```

To transform this data into op_return data:
```
StakingDataPkScript = 0x6a || 0x47 || SerializedStakingData
where:
0x6a - is byte marker representing OP_RETURN op code
0x47 - is byte marker representing OP_DATA_71 op code, which pushed 71 bytes onto the stack
```

The final op_return output will have the following shape:
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
will be used in every script. This key needs to be put in the op_return output
in the staking transaction.
- `finality_provider_public_key` - chosen by the user sending the staking
transaction. It will be used as `<FinalityPk>` in `slashing_script`. In lock
lock-only network there is no slashing, so this key has mostly informative purposes.
This key needs to be put in the op_return output in the staking transaction.
- `staking_time` - chosen by the user sending the staking transaction. It will
used as locking time in the `time_lock` script. It must be valid `uint16` number,
in range `global_parameters.min_staking_time <= staking_time <= global_parameters.max_staking_time`.
It needs to be put in the op_return output in the staking transaction.
- `covenant_committee_public_keys` - it can be retrieved from
`global_parameters.covenant_pks`.
- `covenant_committee_quorum` - it can be retrieved from
`global_parameters.covenant_quorum`
- `staking_amout` - chosen by the user, it will be placed in `staking_output.value`
- `btc_network` - btc network on which staking transactions will take place

#### Building OP_RETRUN and staking output implementation
Babylon staking library exposes [BuildV0IdentifiableStakingOutputsAndTx](../btcstaking/identifiable_staking.go?plain=1#L231)
function.

This functions has following signature:

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

and enables the caller to create valid outputs as well as unfunded and not-signed
valid staking transaction.

Suggested way of creating and sending staking transaction using bitcoind is:
1. create `staker_key` in the bitcoind wallet
2. create unfunded and not signed staking transaction using `BuildV0IdentifiableStakingOutputsAndTx`
function
3. serialize unfunded and not signed staking transaction, to `staking_transaction_hex`
4. call `bitcoin-cli fundrawtransaction "staking_transaction_hex"`. Bitcoind wallet
will automatically choose unspent outputs to fund this transaction. This call will
return `funded_staking_transaction_hex`
5. call `bitcoin-cli signrawtransactionwithwallet "funded_staking_transaction_hex"`.
This call will sign all inputs to transaction and return `signed_staking_transaction_hex`
6. call `bitcoin-cli sendrawtransaction "signed_staking_transaction_hex"`

### Unbonding transaction

An unbonding transaction is a transaction which allows the staker early exit
from the system.

#### Requirements

For the transaction to be considered a valid unbonding transaction, it must:
- have exactly one input and one output
- input must be valid staking output
- output must be taproot output. This taproot output must have disabled
the key spending path, and committed to script tree composed of two scripts:
timelock script, slashing script. This output is henceforth known as
`unbonding_output`
- timelock in time lock script must be equal to `global_parameters.unbonding_time`
- value in unbonding output must be equal to `staking_output.value - global_parameters.unbonding_fee`

#### Building Unbonding output

Babylon staking library exposes [BuildUnbondingInfo](../btcstaking/types.go?plain=1#416)
function which build valid unbonding output

This function has following signature:

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
- `stakerKey`-must be the same key as the staker key in `staking_transaction`
- `fpKeys` - must contain one key, which is the same finality provider key used
in `staking_transaction`
- `covenantKeys`- are the same covenant keys as used in `staking_transaction`
- `covenantQuorum` - is the same quorum as used in `staking_transaction`
- `unbondingTime` - is equal to `global_parameters.unbonding_time`
- `unbondingAmount` - is equal to `staking_amount - global_parameters.unbonding_fee`

## Spending taproot outputs

To create transactions which spends from taproot outputs, either staking output
or unbonding output, providing signatures satisfying the script is not enough.

Spender must also provide:
- whole script which is being spend
- control block which contains: leaf version, internal public key and proof of
inclusion of given script in script tree

Given that creating scripts is deterministic for given data, it is possible to
avoid storing scripts by re-building scripts when need arises.

### Re-creating script and control block

To build script and control block necessary to spend from staking output through
timelock script following function could be implemented

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

Returned script and control block can be used either to build witness directly
or to put them in PSBT which can be used to create witness by Bitcoind.

### Creating PSBT to get signature for given taproot path from Bitcoind

To avoid creating signatures/witness manually Bitcoind [walletprocesspsbt](https://developer.bitcoin.org/reference/rpc/walletprocesspsbt.html)
can be used. To use this Bitcoind endpoint to get signature/witness wallet must
maintain one of the keys used in script.

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

Given that to spend through unbonding script requires more than staker signature,
`walletprocesspsbt` endpoint will produce new psbt with staker signature attached.

If timelock path, which requires only staker signature, would be used
`walletprocesspsbt` would produce whole witness required to send transaction to
the BTC network.



