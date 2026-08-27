package semanticdeltareceipt

type Input struct {
	CaseID     string `json:"case_id"`
	SubjectSHA string `json:"subject_sha"`
	Before     []byte `json:"-"`
	After      []byte `json:"-"`
}

type Node struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
}

type Fact struct {
	Subject   string `json:"subject"`
	Predicate string `json:"predicate"`
	Object    string `json:"object"`
}

type Claim struct {
	ID        string `json:"id"`
	Subject   string `json:"subject"`
	Predicate string `json:"predicate"`
	Object    string `json:"object"`
	Status    string `json:"status"`
	Stage     string `json:"stage"`
	Step      string `json:"step"`
	Reason    string `json:"reason"`
}

type Snapshot struct {
	SourceDigest     string  `json:"source_digest"`
	Bytes            int     `json:"bytes"`
	Lines            int     `json:"lines"`
	ParseStatus      string  `json:"parse_status"`
	ParseReason      string  `json:"parse_reason"`
	StructuralDigest string  `json:"structural_digest"`
	ClaimDigest      string  `json:"claim_digest"`
	Nodes            []Node  `json:"nodes,omitempty"`
	Facts            []Fact  `json:"facts,omitempty"`
	Claims           []Claim `json:"claims,omitempty"`
}

type TextualDelta struct {
	Changed      bool   `json:"changed"`
	BeforeBytes  int    `json:"before_bytes"`
	AfterBytes   int    `json:"after_bytes"`
	ChangedBytes int    `json:"changed_bytes"`
	BeforeDigest string `json:"before_digest"`
	AfterDigest  string `json:"after_digest"`
}

type StructuralDelta struct {
	Status       string `json:"status"`
	AddedNodes   []Node `json:"added_nodes,omitempty"`
	RemovedNodes []Node `json:"removed_nodes,omitempty"`
	AddedFacts   []Fact `json:"added_facts,omitempty"`
	RemovedFacts []Fact `json:"removed_facts,omitempty"`
}

type ClaimChange struct {
	ID     string `json:"id"`
	Before Claim  `json:"before"`
	After  Claim  `json:"after"`
}

type ClaimDelta struct {
	Status  string        `json:"status"`
	Added   []Claim       `json:"added,omitempty"`
	Removed []Claim       `json:"removed,omitempty"`
	Changed []ClaimChange `json:"changed,omitempty"`
}

type ClaimTransition struct {
	ClaimID    string `json:"claim_id"`
	FromStatus string `json:"from_status"`
	ToStatus   string `json:"to_status"`
	FromObject string `json:"from_object"`
	ToObject   string `json:"to_object"`
	Stage      string `json:"stage"`
	Step       string `json:"step"`
	Reason     string `json:"reason"`
}

type Receipt struct {
	Schema             string            `json:"schema"`
	CaseID             string            `json:"case_id"`
	SubjectSHA         string            `json:"subject_sha"`
	Producer           string            `json:"producer"`
	Consumer           string            `json:"consumer"`
	MetaOperation      string            `json:"meta_operation"`
	ProofChoice        string            `json:"proof_choice"`
	Stage              string            `json:"stage"`
	Step               string            `json:"step"`
	Reason             string            `json:"reason"`
	Decision           string            `json:"decision"`
	Resolution         string            `json:"resolution"`
	Classification     string            `json:"classification"`
	Before             Snapshot          `json:"before"`
	After              Snapshot          `json:"after"`
	TextualDelta       TextualDelta      `json:"textual_delta"`
	StructuralDelta    StructuralDelta   `json:"structural_delta"`
	SemanticClaimDelta ClaimDelta        `json:"semantic_claim_delta"`
	ClaimTransitions   []ClaimTransition `json:"claim_transitions"`
	RepositoryWrites   int               `json:"repository_writes"`
	ReceiptDigest      string            `json:"receipt_digest"`
}

type Verdict struct {
	Decision       string `json:"decision"`
	Resolution     string `json:"resolution"`
	Classification string `json:"classification"`
	Reason         string `json:"reason"`
	Passed         bool   `json:"passed"`
	Producer       string `json:"producer"`
	Consumer       string `json:"consumer"`
	Stage          string `json:"stage"`
	Step           string `json:"step"`
}

type Summary struct {
	CasesTotal             int `json:"cases_total"`
	CasesPassed            int `json:"cases_passed"`
	TextualChanges         int `json:"textual_changes"`
	StructuralObservations int `json:"structural_observations"`
	ClaimTransitionCases   int `json:"claim_transition_cases"`
	AdjudicatedCases       int `json:"adjudicated_cases"`
	SemanticPreserved      int `json:"semantic_preserved"`
	SemanticChanged        int `json:"semantic_changed"`
	Indeterminate          int `json:"indeterminate"`
	UnknownPaths           int `json:"unknown_paths"`
	RepositoryWrites       int `json:"repository_writes"`
}

type Report struct {
	Schema             string             `json:"schema"`
	CaseID             string             `json:"case_id"`
	SubjectSHA         string             `json:"subject_sha"`
	Receipt            Receipt            `json:"receipt"`
	IndependentVerdict Verdict            `json:"independent_verdict"`
	Indicators         []OperationBinding `json:"indicators"`
	RepositoryWrites   int                `json:"repository_writes"`
	ReportDigest       string             `json:"report_digest"`
}

type CaseResult struct {
	Definition CaseDefinition `json:"definition"`
	Passed     bool           `json:"passed"`
	Report     Report         `json:"report"`
}

type Suite struct {
	Schema            string       `json:"schema"`
	SubjectSHA        string       `json:"subject_sha"`
	DenominatorID     string       `json:"denominator_id"`
	DenominatorDigest string       `json:"denominator_digest"`
	Decision          string       `json:"decision"`
	Resolution        string       `json:"resolution"`
	Cases             []CaseResult `json:"cases"`
	Summary           Summary      `json:"summary"`
	CoverageBPS       int          `json:"coverage_bps"`
	SuiteDigest       string       `json:"suite_digest"`
}
