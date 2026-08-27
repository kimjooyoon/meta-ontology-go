package semanticdeltareceipt

const (
	ReceiptSchema = "gooo/semantic-delta-receipt/v1"
	ReportSchema  = "gooo/semantic-delta-receipt-report/v1"
	SuiteSchema   = "gooo/semantic-delta-receipt-conformance/v1"
	DenominatorID = "gooo/semantic-delta-receipt-denominator/v1"
	MetricID      = "gooo.metric.semantic.delta-receipt-totality.v1"
	MetaOperation = "separate-text-structural-semantic-deltas"

	Producer = "semanticdeltareceipt.Produce"
	Consumer = "semanticdeltareceipt.Adjudicate"

	DecisionFixedPoint = "FIXED_POINT"
	DecisionDelta      = "DELTA_OBSERVED"
	DecisionFailClosed = "FAIL_CLOSED"

	ResolutionExact     = "EXACT"
	ResolutionUnknown   = "UNKNOWN"
	ResolutionInvariant = "INVARIANT_ONLY"

	ClassPreserved     = "SEMANTIC_PRESERVED"
	ClassChanged       = "SEMANTIC_CHANGED"
	ClassIndeterminate = "INDETERMINATE"

	ReasonTextualOnly = "TEXTUAL_DELTA_WITH_SEMANTIC_FIXED_POINT"
	ReasonMeaning     = "SEMANTIC_CLAIM_DELTA_OBSERVED"
	ReasonUnavailable = "SEMANTIC_TRANSLATION_VALIDATION_UNAVAILABLE"
	ReasonReceipt     = "SEMANTIC_DELTA_RECEIPT_MISMATCH"
	ReasonSubject     = "SEMANTIC_DELTA_SUBJECT_UNKNOWN"
)

type CaseDefinition struct {
	ID                 string `json:"id"`
	ExpectedDecision   string `json:"expected_decision"`
	ExpectedResolution string `json:"expected_resolution"`
	ExpectedClass      string `json:"expected_class"`
	ExpectedReason     string `json:"expected_reason"`
}

func Denominator() []CaseDefinition {
	return []CaseDefinition{
		{ID: "equivalent", ExpectedDecision: DecisionFixedPoint, ExpectedResolution: ResolutionExact, ExpectedClass: ClassPreserved, ExpectedReason: ReasonTextualOnly},
		{ID: "semantic-change", ExpectedDecision: DecisionDelta, ExpectedResolution: ResolutionExact, ExpectedClass: ClassChanged, ExpectedReason: ReasonMeaning},
		{ID: "indeterminate", ExpectedDecision: DecisionFailClosed, ExpectedResolution: ResolutionUnknown, ExpectedClass: ClassIndeterminate, ExpectedReason: ReasonUnavailable},
	}
}

type OperationBinding struct {
	MetricID      string `json:"metric_id"`
	Class         string `json:"class"`
	ProofChoice   string `json:"proof_choice"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	Stage         string `json:"stage"`
	Step          string `json:"step"`
	Reason        string `json:"reason"`
	Unit          string `json:"unit"`
	Relation      string `json:"relation"`
	Value         int    `json:"value"`
	Target        int    `json:"target"`
	Satisfied     bool   `json:"satisfied"`
}

func operationBindings(summary Summary) []OperationBinding {
	return []OperationBinding{
		binding(MetricID, "OUTCOME", "COHERENCE", "cases", "GREATER_OR_EQUAL", summary.CasesPassed, summary.CasesTotal, "reduce", "case-suite", "ALL_FIXED_DENOMINATOR_CASES_REPLAYED"),
		binding("gooo.metric.evidence.textual-delta-observation.v1", "DRIVER", "FOUNDATION", "cases", "GREATER_OR_EQUAL", summary.TextualChanges, summary.CasesTotal, "observe", "raw-bytes", "TEXTUAL_BYTES_BOUND"),
		binding("gooo.metric.evidence.structural-delta-separation.v1", "DRIVER", "FOUNDATION", "cases", "GREATER_OR_EQUAL", summary.StructuralObservations, summary.CasesTotal, "derive", "canonical-graph", "STRUCTURAL_GRAPH_BOUND_SEPARATELY"),
		binding("gooo.metric.semantic.claim-transition-totality.v1", "DRIVER", "FOUNDATION", "cases", "GREATER_OR_EQUAL", summary.ClaimTransitionCases, summary.CasesTotal, "derive", "claim-transition", "CLAIM_TRANSITIONS_EXPLICIT"),
		binding("gooo.metric.epistemic.delta-receipt-adjudication.v1", "GUARDRAIL", "COHERENCE", "cases", "GREATER_OR_EQUAL", summary.AdjudicatedCases, summary.CasesTotal, "adjudicate", "independent-replay", "INDEPENDENT_JUDGE_REPLAYED"),
		binding("gooo.metric.effects.delta-receipt-writes.v1", "GUARDRAIL", "REGRESSION", "writes", "LESS_OR_EQUAL", summary.RepositoryWrites, 0, "observe", "read-only-boundary", "NO_REPOSITORY_WRITES"),
	}
}

func binding(metricID, class, proof, unit, relation string, value, target int, stage, step, reason string) OperationBinding {
	satisfied := value >= target
	if relation == "LESS_OR_EQUAL" {
		satisfied = value <= target
	}
	return OperationBinding{MetricID: metricID, Class: class, ProofChoice: proof,
		Producer: Producer, Consumer: Consumer, MetaOperation: MetaOperation,
		Stage: stage, Step: step, Reason: reason, Unit: unit, Relation: relation,
		Value: value, Target: target, Satisfied: satisfied}
}
