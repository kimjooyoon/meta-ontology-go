package ambiguitybudgetjudge

const (
	ContractSchema        = "gooo/ambiguity-budget-contract/v2"
	ReceiptSchema         = "gooo/ambiguity-budget-receipt/v2"
	JudgeSchema           = "gooo/ambiguity-budget-judge/v2"
	Producer              = "gooo://meta/ambiguity-budget/producer"
	Consumer              = "gooo://meta/ambiguity-budget/independent-verifier"
	MetaOperation         = "measure-deterministic-ambiguity-budget"
	CaseTotal             = 4
	InterventionTotal     = 2
	IntegerDimensionTotal = 3
	FixedDenominator      = 2
)

type integerSet struct {
	InterpretationCandidates int `json:"interpretation_candidates"`
	UnresolvedBranches       int `json:"unresolved_branches"`
	EvidencePaths            int `json:"evidence_paths"`
}

type coordinate struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type transition struct {
	CaseID         string `json:"case_id"`
	From           string `json:"from"`
	To             string `json:"to"`
	Stage          string `json:"stage"`
	Step           string `json:"step"`
	Reason         string `json:"reason"`
	EvidenceDigest string `json:"evidence_digest"`
}

type caseContract struct {
	ID       string `json:"id"`
	Activity string `json:"activity"`
}

type interventionContract struct {
	ID             string `json:"id"`
	Kind           string `json:"kind"`
	TargetActivity string `json:"target_activity"`
}

type contract struct {
	Schema           string                 `json:"schema"`
	ID               string                 `json:"id"`
	SourcePath       string                 `json:"source_path"`
	SourcePackage    string                 `json:"source_package"`
	SourceNamespace  string                 `json:"source_namespace"`
	BudgetActivity   string                 `json:"budget_activity"`
	FixedDenominator int                    `json:"fixed_denominator"`
	Cases            []caseContract         `json:"cases"`
	Interventions    []interventionContract `json:"interventions"`
	NotClaimed       []string               `json:"not_claimed"`
}

type programObservation struct {
	Activity    string     `json:"activity"`
	Program     string     `json:"program"`
	ProgramKind string     `json:"program_kind"`
	ID          string     `json:"id"`
	Class       string     `json:"class,omitempty"`
	InputState  string     `json:"input_state,omitempty"`
	Counts      integerSet `json:"counts"`
	Digest      string     `json:"digest"`
}

type sourceObservation struct {
	Path           string               `json:"path"`
	Digest         string               `json:"digest"`
	SemanticDigest string               `json:"semantic_digest"`
	Lowering       string               `json:"lowering"`
	Package        string               `json:"package"`
	Namespace      string               `json:"namespace"`
	Entities       int                  `json:"entities"`
	Activities     int                  `json:"activities"`
	Programs       []programObservation `json:"programs"`
}

type caseReceipt struct {
	ID             string     `json:"id"`
	Activity       string     `json:"activity"`
	Class          string     `json:"class"`
	InputState     string     `json:"input_state"`
	Program        string     `json:"program"`
	ProgramDigest  string     `json:"program_digest"`
	Counts         integerSet `json:"counts"`
	Decision       string     `json:"decision"`
	Resolution     string     `json:"resolution"`
	Reason         string     `json:"reason"`
	Coordinate     coordinate `json:"coordinate"`
	Claim          transition `json:"claim"`
	EvidenceDigest string     `json:"evidence_digest"`
	Conformance    string     `json:"conformance"`
}

type indicator struct {
	MetricID       string `json:"metric_id"`
	CaseID         string `json:"case_id"`
	Dimension      string `json:"dimension"`
	ProofChoice    string `json:"proof_choice"`
	Producer       string `json:"producer"`
	Consumer       string `json:"consumer"`
	MetaOperation  string `json:"meta_operation"`
	Observed       int    `json:"observed"`
	Budget         int    `json:"budget"`
	Relation       string `json:"relation"`
	Evaluation     string `json:"evaluation"`
	EvidenceDigest string `json:"evidence_digest"`
}

type interventionReceipt struct {
	ID                   string     `json:"id"`
	Kind                 string     `json:"kind"`
	TargetActivity       string     `json:"target_activity"`
	SourceDigestBefore   string     `json:"source_digest_before"`
	SourceDigestAfter    string     `json:"source_digest_after"`
	SemanticDigestBefore string     `json:"semantic_digest_before"`
	SemanticDigestAfter  string     `json:"semantic_digest_after"`
	CountsBefore         integerSet `json:"counts_before"`
	CountsAfter          integerSet `json:"counts_after"`
	DecisionBefore       string     `json:"decision_before"`
	ResolutionBefore     string     `json:"resolution_before"`
	ReasonBefore         string     `json:"reason_before"`
	DecisionAfter        string     `json:"decision_after"`
	ResolutionAfter      string     `json:"resolution_after"`
	ReasonAfter          string     `json:"reason_after"`
	Satisfied            bool       `json:"satisfied"`
	EvidenceDigest       string     `json:"evidence_digest"`
}

type proof struct {
	Choice         string `json:"choice"`
	Claim          string `json:"claim"`
	Producer       string `json:"producer"`
	Consumer       string `json:"consumer"`
	MetaOperation  string `json:"meta_operation"`
	EvidenceDigest string `json:"evidence_digest"`
	Passed         bool   `json:"passed"`
}

type summary struct {
	CasesTotal           int `json:"cases_total"`
	KnownCases           int `json:"known_cases"`
	ZeroAmbiguityCases   int `json:"zero_ambiguity_cases"`
	BoundaryCases        int `json:"boundary_cases"`
	OverBudgetCases      int `json:"over_budget_cases"`
	UnknownCases         int `json:"unknown_cases"`
	LowerResolutionCases int `json:"lower_resolution_cases"`
	OpenClaims           int `json:"open_claims"`
	IntegerDimensions    int `json:"integer_dimensions"`
	InterventionsTotal   int `json:"interventions_total"`
	FixedDenominator     int `json:"fixed_denominator"`
}

type effects struct {
	RepositoryWrites  int  `json:"repository_writes"`
	MutationAuthority bool `json:"mutation_authority"`
}

type receipt struct {
	Schema                string                `json:"schema"`
	SubjectSHA            string                `json:"subject_sha"`
	ContractID            string                `json:"contract_id"`
	Source                sourceObservation     `json:"source"`
	Budget                integerSet            `json:"budget"`
	Producer              string                `json:"producer"`
	Consumer              string                `json:"consumer"`
	MetaOperation         string                `json:"meta_operation"`
	ProofChoice           string                `json:"proof_choice"`
	ConformanceDecision   string                `json:"conformance_decision"`
	ConformanceResolution string                `json:"conformance_resolution"`
	ConformanceReason     string                `json:"conformance_reason"`
	SubjectDecision       string                `json:"subject_decision"`
	SubjectResolution     string                `json:"subject_resolution"`
	SubjectReason         string                `json:"subject_reason"`
	SubjectCoordinate     coordinate            `json:"subject_coordinate"`
	Cases                 []caseReceipt         `json:"cases"`
	Claims                []transition          `json:"claims"`
	Indicators            []indicator           `json:"indicators"`
	Interventions         []interventionReceipt `json:"interventions"`
	Proofs                []proof               `json:"proofs"`
	Summary               summary               `json:"summary"`
	NotClaimed            []string              `json:"not_claimed"`
	Effects               effects               `json:"effects"`
	FactsDigest           string                `json:"facts_digest"`
	Digest                string                `json:"digest"`
}

type check struct {
	ID         string     `json:"id"`
	Status     string     `json:"status"`
	Coordinate coordinate `json:"coordinate"`
}

type Result struct {
	Schema                   string  `json:"schema"`
	SubjectSHA               string  `json:"subject_sha"`
	ContractID               string  `json:"contract_id"`
	Producer                 string  `json:"producer"`
	Consumer                 string  `json:"consumer"`
	MetaOperation            string  `json:"meta_operation"`
	ConformanceDecision      string  `json:"conformance_decision"`
	ConformanceResolution    string  `json:"conformance_resolution"`
	ConformanceReason        string  `json:"conformance_reason"`
	SubjectDecision          string  `json:"subject_decision"`
	SubjectResolution        string  `json:"subject_resolution"`
	ReportDigest             string  `json:"report_digest"`
	SourceDigest             string  `json:"source_digest"`
	FixedDenominator         int     `json:"fixed_denominator"`
	CasesTotal               int     `json:"cases_total"`
	InterventionsTotal       int     `json:"interventions_total"`
	ForbiddenProducerImports int     `json:"forbidden_producer_imports"`
	AllowedProducerImports   int     `json:"allowed_producer_imports"`
	RepositoryWrites         int     `json:"repository_writes"`
	MutationAuthority        bool    `json:"mutation_authority"`
	Checks                   []check `json:"checks"`
	Digest                   string  `json:"digest"`
}
