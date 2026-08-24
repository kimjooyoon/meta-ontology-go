package changedsurfacereceipt

const (
	InputSchema         = "gooo/changed-surface-receipt-input/v1"
	ReportSchema        = "gooo/changed-surface-receipt-report/v1"
	SuiteSchema         = "gooo/changed-surface-receipt-conformance/v1"
	DenominatorID       = "gooo/changed-surface-receipt-denominator/v1"
	MetricID            = "gooo.metric.semantic.changed-surface-receipt-totality.v1"
	MetaOperation       = "totalize-changed-surface-receipts"
	DecisionFixedPoint  = "FIXED_POINT"
	DecisionFailClosed  = "FAIL_CLOSED"
	ResolutionExact     = "EXACT"
	ResolutionUnknown   = "UNKNOWN"
	ResolutionInvariant = "INVARIANT_ONLY"
	EffectObserve       = "OBSERVE_ONLY"
	EffectBlock         = "BLOCK"
	ReasonTotal         = "CHANGED_SURFACE_RECEIPTS_TOTAL"
	ReasonUnavailable   = "CHANGED_SURFACE_EVIDENCE_UNAVAILABLE"
	ReasonSchema        = "CHANGED_SURFACE_SCHEMA_MISMATCH"
	ReasonMissing       = "CHANGED_SURFACE_RECEIPT_MISSING"
	ReasonOrphan        = "CHANGED_SURFACE_RECEIPT_ORPHAN"
	ReasonDuplicate     = "CHANGED_SURFACE_RECEIPT_DUPLICATE"
	ReasonUnknown       = "CHANGED_SURFACE_RECEIPT_DECISION_UNKNOWN"
	ReasonMalformed     = "CHANGED_SURFACE_PATH_MALFORMED"
)

type MetaOperationDefinition struct {
	ID          string `json:"id"`
	ProofChoice string `json:"proof_choice"`
}

func MetaOperations() []MetaOperationDefinition {
	return []MetaOperationDefinition{
		{ID: "freeze-changed-surface-denominator", ProofChoice: "COHERENCE"},
		{ID: "observe-changed-surfaces", ProofChoice: "FOUNDATION"},
		{ID: "bind-surface-receipts", ProofChoice: "FOUNDATION"},
		{ID: MetaOperation, ProofChoice: "COHERENCE"},
		{ID: "preserve-receipt-unknown", ProofChoice: "REGRESSION"},
		{ID: "observe-shadow-writes", ProofChoice: "FOUNDATION"},
	}
}
