package semanticdeltareceipt

const (
	ReceiptSchema = "gooo/semantic-delta-receipt/v1"
	ReportSchema  = "gooo/semantic-delta-receipt-report/v1"
	SuiteSchema   = "gooo/semantic-delta-receipt-conformance/v1"
	DenominatorID = "gooo/semantic-delta-receipt-denominator/v1"
	MetricID      = "gooo.metric.semantic.delta-receipt-totality.v1"
	MetaOperation = "separate-text-structural-semantic-deltas"
	Producer      = "semanticdeltareceipt.ProduceFiles"
	Consumer      = "semanticdeltareceiptconsumer.AdjudicateFiles"

	DecisionFixedPoint  = "FIXED_POINT"
	DecisionDelta       = "DELTA_OBSERVED"
	DecisionFailClosed  = "FAIL_CLOSED"
	ResolutionExact     = "EXACT"
	ResolutionLower     = "LOWER_RESOLUTION"
	ResolutionInvariant = "INVARIANT_ONLY"
	ClassPreserved      = "SEMANTIC_PRESERVED"
	ClassChanged        = "SEMANTIC_CHANGED"
	ClassIndeterminate  = "INDETERMINATE"
	StatusOpen          = "OPEN"
	StatusDischarged    = "DISCHARGED"
	StatusRefuted       = "REFUTED"
	RawChanged          = "RAW_CHANGED"
	RawFixedPoint       = "RAW_FIXED_POINT"
	SemanticPreserved   = "SEMANTIC_PRESERVED"
	SemanticChanged     = "SEMANTIC_CHANGED"
	SemanticUnknown     = "SEMANTIC_UNKNOWN"
	ReasonTextualOnly   = "TEXTUAL_DELTA_WITH_SEMANTIC_FIXED_POINT"
	ReasonMeaning       = "SEMANTIC_CLAIM_DELTA_OBSERVED"
	ReasonUnavailable   = "SEMANTIC_TRANSLATION_VALIDATION_UNAVAILABLE"
	ReasonReceipt       = "SEMANTIC_DELTA_RECEIPT_MISMATCH"
	ReasonSubject       = "SEMANTIC_DELTA_SUBJECT_UNKNOWN"
	SubjectStage        = "bind-subject"
	SubjectStep         = "resolve-subject"
	UnavailableStage    = "project-source"
	UnavailableStep     = "parse-lower"
	ClaimKindBounded    = "BOUNDED_SEMANTIC_EQUIVALENCE"
	ClaimKindObject     = "OBJECT_PROPOSITION"
	ClaimKindPreserve   = "BEFORE_CLAIM_PRESERVATION"
)

type CaseDefinition struct {
	ID                 string `json:"id"`
	BeforePath         string `json:"before_path"`
	AfterPath          string `json:"after_path"`
	ExpectedDecision   string `json:"expected_decision"`
	ExpectedResolution string `json:"expected_resolution"`
	ExpectedClass      string `json:"expected_class"`
	ExpectedReason     string `json:"expected_reason"`
}

func Denominator() []CaseDefinition {
	return []CaseDefinition{
		{"equivalent", "examples/semantic-delta-receipt/before.gooo", "examples/semantic-delta-receipt/equivalent-after.gooo", DecisionFixedPoint, ResolutionExact, ClassPreserved, ReasonTextualOnly},
		{"semantic-change", "examples/semantic-delta-receipt/before.gooo", "examples/semantic-delta-receipt/semantic-after.gooo", DecisionDelta, ResolutionExact, ClassChanged, ReasonMeaning},
		{"indeterminate", "examples/semantic-delta-receipt/before.gooo", "examples/semantic-delta-receipt/indeterminate-after.gooo", DecisionFailClosed, ResolutionLower, ClassIndeterminate, ReasonUnavailable},
	}
}
