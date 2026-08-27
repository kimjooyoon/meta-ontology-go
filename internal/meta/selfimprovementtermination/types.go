package selfimprovementtermination

const (
	InputSchema        = "gooo/self-improvement-termination-input/v1"
	ReceiptSchema      = "gooo/self-improvement-termination-receipt/v1"
	Metaprogram        = "internal/meta/selfimprovementtermination"
	Producer           = "selfimprovementtermination.Evaluate"
	Consumer           = "self-improvement-cycle"
	MetaOperation      = "prove-self-improvement-termination"
	ProofChoice        = "TERMINATION"
	TraceStage         = "META_RUN"
	ClaimStage         = "CLAIM"
	DecisionFixedPoint = "FIXED_POINT"
	DecisionInProgress = "IN_PROGRESS"
	DecisionCycle      = "CYCLE"
	DecisionDivergence = "DIVERGENCE_POSSIBLE"
	ResolutionExact    = "EXACT_OBSERVED_PREFIX"
	ReceiptBound       = "BOUND"
	IndicatorTotal     = 10
	MaxTraceSteps      = 64
)

type Input struct {
	Schema        string        `json:"schema"`
	Repository    string        `json:"repository"`
	Subject       string        `json:"subject"`
	Producer      string        `json:"producer"`
	Consumer      string        `json:"consumer"`
	MetaOperation string        `json:"meta_operation"`
	ProofChoice   string        `json:"proof_choice"`
	Stage         string        `json:"stage"`
	MaxSteps      int           `json:"max_steps"`
	Trace         []Observation `json:"trace"`
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

type Authority struct {
	ReadOnly            bool `json:"read_only"`
	RepositoryWrites    int  `json:"repository_writes"`
	MutationAuthorized  bool `json:"mutation_authorized"`
	PromotionAuthorized bool `json:"promotion_authorized"`
}

type Summary struct {
	Satisfied         int    `json:"satisfied"`
	Total             int    `json:"total"`
	BasisPoints       int    `json:"basis_points"`
	ObservedSteps     int    `json:"observed_steps"`
	MaxSteps          int    `json:"max_steps"`
	StateCount        int    `json:"state_count"`
	DetectedPeriod    int    `json:"detected_period"`
	FinalState        string `json:"final_state"`
	TerminationProven bool   `json:"termination_proven"`
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
	Schema           string            `json:"schema"`
	Metaprogram      string            `json:"metaprogram"`
	Repository       string            `json:"repository"`
	Subject          string            `json:"subject"`
	Producer         string            `json:"producer"`
	Consumer         string            `json:"consumer"`
	MetaOperation    string            `json:"meta_operation"`
	ProofChoice      string            `json:"proof_choice"`
	Stage            string            `json:"stage"`
	Status           string            `json:"status"`
	Resolution       string            `json:"resolution"`
	Decision         string            `json:"decision"`
	Reason           string            `json:"reason"`
	InputDigest      string            `json:"input_digest"`
	TraceDigest      string            `json:"trace_digest"`
	Observations     []Observation     `json:"observations"`
	ClaimTransitions []ClaimTransition `json:"claim_transitions"`
	Summary          Summary           `json:"summary"`
	Authority        Authority         `json:"authority"`
	Indicators       []Indicator       `json:"indicators"`
	ReceiptDigest    string            `json:"receipt_digest"`
	ReplayDigest     string            `json:"replay_digest"`
}
