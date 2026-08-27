package selfimprovementtermination

const (
	InputSchema            = "gooo/self-improvement-termination-input/v2"
	ReceiptSchema          = "gooo/self-improvement-termination-receipt/v2"
	Metaprogram            = "internal/meta/selfimprovementtermination"
	Producer               = "selfimprovementtermination.Evaluate"
	Consumer               = "self-improvement-cycle"
	MetaOperation          = "prove-self-improvement-termination"
	ProofChoice            = "TERMINATION"
	TraceStage             = "META_RUN"
	ClaimStage             = "CLAIM"
	InterventionStage      = "INTERVENTION"
	DecisionFixedPoint     = "FIXED_POINT"
	DecisionInProgress     = "IN_PROGRESS"
	DecisionCycle          = "CYCLE"
	DecisionDivergence     = "DIVERGENCE_POSSIBLE"
	DecisionFailClosed     = "FAIL_CLOSED"
	ResolutionExact        = "EXACT"
	ResolutionLower        = "LOWER_RESOLUTION"
	ReceiptBound           = "BOUND"
	ReceiptFailClosed      = "FAIL_CLOSED"
	ClaimOpen              = "OPEN"
	ClaimDischarged        = "DISCHARGED"
	ClaimRefuted           = "REFUTED"
	UpstreamNoChange       = "NO_CHANGE"
	UpstreamChanged        = "CHANGED"
	ReasonNoChange         = "NO_CHANGE_FIXED_POINT_OBSERVED"
	ReasonStateChanged     = "METAPROGRAM_STATE_CHANGED"
	ReasonCycle            = "REPEATED_STATE_CYCLE_OBSERVED"
	ReasonInProgress       = "TRACE_ENDED_BEFORE_TERMINATION"
	ReasonDivergence       = "STRICTLY_GROWING_BOUNDARY_NO_FIXED_POINT"
	ReasonDecisionUnknown  = "FEEDBACK_COVERAGE_DECISION_UNKNOWN"
	SourcePath             = "examples/self-improvement-termination/main.gooo"
	SourceProgramSchema    = "termination-case/v2"
	InterventionSchema     = "termination-intervention/v1"
	ConformanceAggregation = "NONE"
	IndicatorTotal         = 2
	MaxTraceSteps          = 64
)

type Input struct {
	Schema           string          `json:"schema"`
	Repository       string          `json:"repository"`
	Subject          string          `json:"subject"`
	Producer         string          `json:"producer"`
	Consumer         string          `json:"consumer"`
	MetaOperation    string          `json:"meta_operation"`
	ProofChoice      string          `json:"proof_choice"`
	Stage            string          `json:"stage"`
	Source           SourceCausality `json:"source"`
	UpstreamDecision string          `json:"upstream_decision"`
	MaxSteps         int             `json:"max_steps"`
	Trace            []Observation   `json:"trace"`
	Interventions    []Intervention  `json:"interventions"`
}

type SourceCausality struct {
	Path              string `json:"path"`
	SourceDigest      string `json:"source_digest"`
	SemanticDigest    string `json:"semantic_digest"`
	CaseID            string `json:"case_id"`
	CaseProgramDigest string `json:"case_program_digest"`
}

type Observation struct {
	Stage       string `json:"stage"`
	Step        int    `json:"step"`
	BeforeState string `json:"before_state"`
	AfterState  string `json:"after_state"`
	BeforeRank  int    `json:"before_rank"`
	AfterRank   int    `json:"after_rank"`
	Decision    string `json:"decision"`
	Reason      string `json:"reason"`
}

type Intervention struct {
	ID                   string `json:"id"`
	Schema               string `json:"schema"`
	Stage                string `json:"stage"`
	Step                 int    `json:"step"`
	Reason               string `json:"reason"`
	SourceBeforeDigest   string `json:"source_before_digest"`
	SourceAfterDigest    string `json:"source_after_digest"`
	SemanticBeforeDigest string `json:"semantic_before_digest"`
	SemanticAfterDigest  string `json:"semantic_after_digest"`
	SourceChanged        bool   `json:"source_changed"`
	SemanticChanged      bool   `json:"semantic_changed"`
}

type Authority struct {
	ReadOnly            bool `json:"read_only"`
	RepositoryWrites    int  `json:"repository_writes"`
	MutationAuthorized  bool `json:"mutation_authorized"`
	PromotionAuthorized bool `json:"promotion_authorized"`
}

type OutcomeSummary struct {
	ObservedSteps     int    `json:"observed_steps"`
	MaxSteps          int    `json:"max_steps"`
	StateCount        int    `json:"state_count"`
	RepeatedStates    int    `json:"repeated_states"`
	DetectedPeriod    int    `json:"detected_period"`
	FinalState        string `json:"final_state"`
	TerminationProven bool   `json:"termination_proven"`
	ClaimState        string `json:"claim_state"`
}

type ConformanceSummary struct {
	Satisfied   int    `json:"satisfied"`
	Total       int    `json:"total"`
	BasisPoints int    `json:"basis_points"`
	Aggregation string `json:"aggregation"`
}

type Indicator struct {
	ID            string `json:"id"`
	Route         string `json:"route"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	ProofChoice   string `json:"proof_choice"`
	Stage         string `json:"stage"`
	Step          int    `json:"step"`
	Reason        string `json:"reason"`
	Value         string `json:"value"`
	Limit         string `json:"limit"`
	Satisfied     bool   `json:"satisfied"`
}

type ClaimTransition struct {
	Stage  string `json:"stage"`
	Step   int    `json:"step"`
	From   string `json:"from"`
	To     string `json:"to"`
	Reason string `json:"reason"`
}

type Receipt struct {
	Schema           string             `json:"schema"`
	Metaprogram      string             `json:"metaprogram"`
	Repository       string             `json:"repository"`
	Subject          string             `json:"subject"`
	Producer         string             `json:"producer"`
	Consumer         string             `json:"consumer"`
	MetaOperation    string             `json:"meta_operation"`
	ProofChoice      string             `json:"proof_choice"`
	Stage            string             `json:"stage"`
	Status           string             `json:"status"`
	Resolution       string             `json:"resolution"`
	Decision         string             `json:"decision"`
	Reason           string             `json:"reason"`
	Source           SourceCausality    `json:"source"`
	UpstreamDecision string             `json:"upstream_decision"`
	InputDigest      string             `json:"input_digest"`
	TraceDigest      string             `json:"trace_digest"`
	Observations     []Observation      `json:"observations"`
	Interventions    []Intervention     `json:"interventions"`
	ClaimTransitions []ClaimTransition  `json:"claim_transitions"`
	Outcome          OutcomeSummary     `json:"outcome"`
	Conformance      ConformanceSummary `json:"conformance"`
	Authority        Authority          `json:"authority"`
	Indicators       []Indicator        `json:"indicators"`
	ReceiptDigest    string             `json:"receipt_digest"`
	ReplayDigest     string             `json:"replay_digest"`
}
