package types

import (
	"github.com/cosmos/cosmos-sdk/telemetry"
	"github.com/hashicorp/go-metrics"
)

// performance oriented metrics measuring the execution time of each message
const (
	MetricsKeyCommitPubRandList = "commit_pub_rand_list"
	MetricsKeyAddFinalitySig    = "add_finality_sig"
)

// Metrics for monitoring block finalization status
const (
	// MetricsKeyLastHeight is the key of the gauge recording the last height
	// of the ledger
	MetricsKeyLastHeight = "last_height"
	// MetricsKeyLastFinalizedHeight is the key of the gauge recording the
	// last height finalized by finality providers
	MetricsKeyLastFinalizedHeight = "last_finalized_height"
)

// RecordLastHeight records the last height. It is triggered upon `IndexBlock`
func RecordLastHeight(height int) {
	keys := []string{MetricsKeyLastHeight}
	labels := []metrics.Label{telemetry.NewLabel(telemetry.MetricLabelNameModule, ModuleName)}
	telemetry.SetGaugeWithLabels(
		keys,
		float32(height),
		labels,
	)
}

// RecordLastHeight records the last finalized height. It is triggered upon
// finalizing a block becomes finalized
func RecordLastFinalizedHeight(height int) {
	keys := []string{MetricsKeyLastFinalizedHeight}
	labels := []metrics.Label{telemetry.NewLabel(telemetry.MetricLabelNameModule, ModuleName)}
	telemetry.SetGaugeWithLabels(
		keys,
		float32(height),
		labels,
	)
}
