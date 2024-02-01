package types

import (
	"github.com/cosmos/cosmos-sdk/telemetry"
	"github.com/hashicorp/go-metrics"
)

// performance oriented metrics measuring the execution time of each message
const (
	MetricsKeyCreateFinalityProvider    = "create_finality_provider"
	MetricsKeyCreateBTCDelegation       = "create_btc_delegation"
	MetricsKeyAddCovenantSigs           = "add_covenant_sigs"
	MetricsKeyBTCUndelegate             = "btc_undelegate"
	MetricsKeySelectiveSlashingEvidence = "selective_slashing_evidence"
)

// Metrics for monitoring finality providers and BTC delegations
const (
	// MetricsKeyFinalityProviders is the key of the gauge recording the number
	// of {active, inactive} finality providers, and the key of the counter
	// recording the number of slashed finality providers
	MetricsKeyFinalityProviders = "finality_providers"
	// MetricsKeyBTCDelegations is the key of the gauge recording the number of
	// {pending, active, unbonded} BTC delegations, and the key of the counter
	// recording the number of slashed BTC delegations
	MetricsKeyBTCDelegations = "btc_delegations"
	// MetricsKeyStakedBitcoins is the key of the gauge recording the total
	// amount of Bitcoins staked under active finality providers
	MetricsKeyStakedBitcoins = "staked_bitcoins"
)

// RecordActiveFinalityProviders records the number of active finality providers.
// It is triggered upon recording voting power table.
func RecordActiveFinalityProviders(num int) {
	keys := []string{MetricsKeyFinalityProviders, "ACTIVE"}
	labels := []metrics.Label{telemetry.NewLabel(telemetry.MetricLabelNameModule, ModuleName)}
	telemetry.SetGaugeWithLabels(
		keys,
		float32(num),
		labels,
	)
}

// RecordInactiveFinalityProviders records the number of inactive finality providers.
// It is triggered upon recording voting power table.
func RecordInactiveFinalityProviders(num int) {
	keys := []string{MetricsKeyFinalityProviders, "INACTIVE"}
	labels := []metrics.Label{telemetry.NewLabel(telemetry.MetricLabelNameModule, ModuleName)}
	telemetry.SetGaugeWithLabels(
		keys,
		float32(num),
		labels,
	)
}

// RecordNewSlashedFinalityProvider increments the number slashed inactive finality providers.
// It is triggered upon a finality provider becomes slashed.
func RecordNewSlashedFinalityProvider() {
	keys := []string{MetricsKeyFinalityProviders, "SLASHED"}
	labels := []metrics.Label{telemetry.NewLabel(telemetry.MetricLabelNameModule, ModuleName)}
	telemetry.IncrCounterWithLabels(
		keys,
		1,
		labels,
	)
}

// RecordBTCDelegations records the number of BTC delegations under the given status.
// It is triggered upon recording voting power table.
func RecordBTCDelegations(num int, status BTCDelegationStatus) {
	keys := []string{MetricsKeyBTCDelegations, status.String()}
	labels := []metrics.Label{telemetry.NewLabel(telemetry.MetricLabelNameModule, ModuleName)}
	telemetry.SetGaugeWithLabels(
		keys,
		float32(num),
		labels,
	)
}

// RecordNewSlashedBTCDelegation increments the number of slashed BTC delegations.
// It is triggered upon the corresponding finality provider is slashed.
func RecordNewSlashedBTCDelegation() {
	keys := []string{MetricsKeyBTCDelegations, "SLASHED"}
	labels := []metrics.Label{telemetry.NewLabel(telemetry.MetricLabelNameModule, ModuleName)}
	telemetry.IncrCounterWithLabels(
		keys,
		1,
		labels,
	)
}

// RecordMetricsKeyStakedBitcoins records the amount of Bitcoins staked under
// all active finality providers.
// It is triggered upon recording voting power table.
func RecordMetricsKeyStakedBitcoins(amount float32) {
	keys := []string{MetricsKeyStakedBitcoins}
	labels := []metrics.Label{telemetry.NewLabel(telemetry.MetricLabelNameModule, ModuleName)}
	telemetry.SetGaugeWithLabels(
		keys,
		amount,
		labels,
	)
}
