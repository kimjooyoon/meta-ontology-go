package semanticdeltareceipt

const (
	ReceiptSchema = "gooo/semantic-delta-receipt/v2"
	ReportSchema  = "gooo/semantic-delta-receipt-report/v1"
	SuiteSchema   = "gooo/semantic-delta-receipt-conformance/v2"
	DenominatorID = "gooo/semantic-delta-receipt-denominator/v2"
	MetricID      = "gooo.metric.semantic.delta-receipt-totality.v1"
	MetaOperation = "separate-text-structural-semantic-deltas"
	Producer      = "semanticdeltareceipt.ProduceFiles"
	Consumer      = "semanticdeltareceiptconsumer.AdjudicateFiles"

	DecisionFixedPoint    = "FIXED_POINT"
	DecisionDelta         = "DELTA_OBSERVED"
	DecisionFailClosed    = "FAIL_CLOSED"
	ResolutionExact       = "EXACT"
	ResolutionLower       = "LOWER_RESOLUTION"
	ResolutionInvariant   = "INVARIANT_ONLY"
	ClassPreserved        = "SEMANTIC_PRESERVED"
	ClassChanged          = "SEMANTIC_CHANGED"
	ClassIndeterminate    = "INDETERMINATE"
	StatusOpen            = "OPEN"
	StatusDischarged      = "DISCHARGED"
	StatusRefuted         = "REFUTED"
	RawChanged            = "RAW_CHANGED"
	RawFixedPoint         = "RAW_FIXED_POINT"
	SemanticPreserved     = "SEMANTIC_PRESERVED"
	SemanticChanged       = "SEMANTIC_CHANGED"
	SemanticUnknown       = "SEMANTIC_UNKNOWN"
	ReasonTextualOnly     = "TEXTUAL_DELTA_WITH_SEMANTIC_FIXED_POINT"
	ReasonMeaning         = "SEMANTIC_CLAIM_DELTA_OBSERVED"
	ReasonUnavailable     = "SEMANTIC_TRANSLATION_VALIDATION_UNAVAILABLE"
	ReasonReceipt         = "SEMANTIC_DELTA_RECEIPT_MISMATCH"
	ReasonSubject         = "SEMANTIC_DELTA_SUBJECT_UNKNOWN"
	ReasonComponentDelta  = "SEMANTIC_COMPONENT_DELTA_OBSERVED"
	ReasonUnmodeled       = "UNMODELED_SEMANTIC_COMPONENT_CHANGED"
	ReasonAmbiguous       = "AMBIGUOUS_CLAIM_MATCH"
	ReasonMeta            = "META_CONTRACT_UNAVAILABLE"
	ReasonDenominatorZero = "DENOMINATOR_ZERO"
	SubjectStage          = "bind-subject"
	SubjectStep           = "resolve-subject"
	UnavailableStage      = "project-source"
	UnavailableStep       = "parse-lower"
	ClaimKindBounded      = "BOUNDED_SEMANTIC_EQUIVALENCE"
	ClaimKindObject       = "OBJECT_PROPOSITION"
	ClaimKindPreserve     = "BEFORE_CLAIM_PRESERVATION"
	MetaSourcePath        = "examples/semantic-delta-receipt/main.gooo"
	DenominatorVersion    = "v2"
	ModeledComponentCount = 5
	TotalComponentCount   = 5
	ComponentNode         = "node-semantic"
	ComponentField        = "entity-field"
	ComponentValue        = "activity-value-program"
	ComponentRelation     = "relation-fact"
	ComponentFingerprint  = "ir-semantic-fingerprint"
)

type CaseDefinition struct {
	ID                 string `json:"id"`
	BeforePath         string `json:"before_path"`
	AfterPath          string `json:"after_path"`
	ExpectedDecision   string `json:"expected_decision"`
	ExpectedResolution string `json:"expected_resolution"`
	ExpectedClass      string `json:"expected_class"`
	ExpectedReason     string `json:"expected_reason"`
	Kind               string `json:"kind"`
	ExpectedStage      string `json:"expected_stage"`
	ExpectedStep       string `json:"expected_step"`
}

func Denominator() []CaseDefinition {
	return []CaseDefinition{
		{ID: "equivalent", Kind: "text-only-preserved", BeforePath: "examples/semantic-delta-receipt/before.gooo", AfterPath: "examples/semantic-delta-receipt/equivalent-after.gooo", ExpectedDecision: DecisionFixedPoint, ExpectedResolution: ResolutionExact, ExpectedClass: ClassPreserved, ExpectedReason: ReasonTextualOnly, ExpectedStage: "produce", ExpectedStep: "classify"},
		{ID: "semantic-change", Kind: "topology-claim-changed", BeforePath: "examples/semantic-delta-receipt/before.gooo", AfterPath: "examples/semantic-delta-receipt/semantic-after.gooo", ExpectedDecision: DecisionDelta, ExpectedResolution: ResolutionExact, ExpectedClass: ClassChanged, ExpectedReason: ReasonMeaning, ExpectedStage: "produce", ExpectedStep: "classify"},
		{ID: "value-program-change", Kind: "value-program-only-changed", BeforePath: "examples/semantic-delta-receipt/value-program-before.gooo", AfterPath: "examples/semantic-delta-receipt/value-program-after.gooo", ExpectedDecision: DecisionDelta, ExpectedResolution: ResolutionExact, ExpectedClass: ClassChanged, ExpectedReason: ReasonComponentDelta, ExpectedStage: "produce", ExpectedStep: "classify"},
		{ID: "indeterminate", Kind: "unsupported-grammar-unknown", BeforePath: "examples/semantic-delta-receipt/before.gooo", AfterPath: "examples/semantic-delta-receipt/indeterminate-after.gooo", ExpectedDecision: DecisionFailClosed, ExpectedResolution: ResolutionLower, ExpectedClass: ClassIndeterminate, ExpectedReason: ReasonUnavailable, ExpectedStage: UnavailableStage, ExpectedStep: UnavailableStep},
		{ID: "ambiguous-match", Kind: "ambiguous-claim-matching-unknown", BeforePath: "examples/semantic-delta-receipt/ambiguous-before.gooo", AfterPath: "examples/semantic-delta-receipt/ambiguous-after.gooo", ExpectedDecision: DecisionFailClosed, ExpectedResolution: ResolutionLower, ExpectedClass: ClassIndeterminate, ExpectedReason: ReasonAmbiguous, ExpectedStage: "claim-delta", ExpectedStep: "match-claims"},
	}
}

func semanticCoverageBPS(modeled, total int) int {
	if total <= 0 {
		return 0
	}
	return modeled * 10000 / total
}
