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
	MetricsKeyFinalityProviders = "finality_providers"
	MetricsKeyBTCDelegations    = "btc_delegations"
	MetricsKeyStakedBitcoins    = "staked_bitcoins"
)

func RecordActiveFinalityProviders(num int) {
	keys := []string{MetricsKeyFinalityProviders, "ACTIVE"}
	labels := []metrics.Label{telemetry.NewLabel(telemetry.MetricLabelNameModule, ModuleName)}
	telemetry.SetGaugeWithLabels(
		keys,
		float32(num),
		labels,
	)
}

func RecordInactiveFinalityProviders(num int) {
	keys := []string{MetricsKeyFinalityProviders, "INACTIVE"}
	labels := []metrics.Label{telemetry.NewLabel(telemetry.MetricLabelNameModule, ModuleName)}
	telemetry.SetGaugeWithLabels(
		keys,
		float32(num),
		labels,
	)
}

func RecordNewSlashedFinalityProvider() {
	keys := []string{MetricsKeyFinalityProviders, "SLASHED"}
	labels := []metrics.Label{telemetry.NewLabel(telemetry.MetricLabelNameModule, ModuleName)}
	telemetry.IncrCounterWithLabels(
		keys,
		1,
		labels,
	)
}

func RecordBTCDelegations(num int, status BTCDelegationStatus) {
	keys := []string{MetricsKeyBTCDelegations, status.String()}
	labels := []metrics.Label{telemetry.NewLabel(telemetry.MetricLabelNameModule, ModuleName)}
	telemetry.SetGaugeWithLabels(
		keys,
		float32(num),
		labels,
	)
}

func RecordNewSlashedBTCDelegation() {
	keys := []string{MetricsKeyBTCDelegations, "SLASHED"}
	labels := []metrics.Label{telemetry.NewLabel(telemetry.MetricLabelNameModule, ModuleName)}
	telemetry.IncrCounterWithLabels(
		keys,
		1,
		labels,
	)
}

func RecordMetricsKeyStakedBitcoins(amount float32) {
	keys := []string{MetricsKeyStakedBitcoins}
	labels := []metrics.Label{telemetry.NewLabel(telemetry.MetricLabelNameModule, ModuleName)}
	telemetry.SetGaugeWithLabels(
		keys,
		amount,
		labels,
	)
}
