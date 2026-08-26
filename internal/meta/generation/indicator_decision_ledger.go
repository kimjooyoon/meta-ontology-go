package generation

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"

const IndicatorDecisionLedgerSchemaVersion = "gooo/indicator-decision-ledger/v2"

type TrilemmaRoute string

const (
	TrilemmaRouteFoundation TrilemmaRoute = "FOUNDATION"
	TrilemmaRouteCoherence  TrilemmaRoute = "COHERENCE"
	TrilemmaRouteRegression TrilemmaRoute = "REGRESSION"
)

type IndicatorDisposition string

const (
	IndicatorDispositionExempt         IndicatorDisposition = "EXEMPT"
	IndicatorDispositionConforming     IndicatorDisposition = "CONFORMING"
	IndicatorDispositionRepairSelected IndicatorDisposition = "REPAIR_SELECTED"
	IndicatorDispositionRepairDeferred IndicatorDisposition = "REPAIR_DEFERRED"
)

// IndicatorDecisionLedgerEntry binds a metric to its proof and meta operation.
type IndicatorDecisionLedgerEntry struct {
	IndicatorID      string                        `json:"indicator_id"`
	SourceIndicator  sourcepolicy.Indicator        `json:"source_indicator"`
	IndicatorOutcome sourcepolicy.IndicatorOutcome `json:"indicator_outcome"`
	TrilemmaRoute    TrilemmaRoute                 `json:"trilemma_route"`
	Disposition      IndicatorDisposition          `json:"disposition"`
	Action           *Action                       `json:"action,omitempty"`
}

// IndicatorDecisionLedger is the replayable proof for all source indicators.
type IndicatorDecisionLedger struct {
	SchemaVersion     string                         `json:"schema_version"`
	IndicatorCount    int                            `json:"indicator_count"`
	SelectedCount     int                            `json:"selected_count"`
	DeferredCount     int                            `json:"deferred_count"`
	FoundationalCount int                            `json:"foundational_count"`
	CoherenceCount    int                            `json:"coherence_count"`
	RegressiveCount   int                            `json:"regressive_count"`
	Entries           []IndicatorDecisionLedgerEntry `json:"entries"`
	Digest            string                         `json:"digest"`
}
