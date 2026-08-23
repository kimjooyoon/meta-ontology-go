package languageassurance

type Priority string

const (
	PriorityP0 Priority = "P0"
	PriorityP1 Priority = "P1"
	PriorityP2 Priority = "P2"
)

type IndicatorClass string

const (
	ClassOutcome   IndicatorClass = "OUTCOME"
	ClassDriver    IndicatorClass = "DRIVER"
	ClassGuardrail IndicatorClass = "GUARDRAIL"
)

type ProofChoice string

const (
	ProofFoundation ProofChoice = "FOUNDATION"
	ProofCoherence  ProofChoice = "COHERENCE"
	ProofRegression ProofChoice = "REGRESSION"
)

type Relation string

const (
	RelationGreaterOrEqual Relation = "GREATER_OR_EQUAL"
	RelationLessOrEqual    Relation = "LESS_OR_EQUAL"
)

type Resolution string

const (
	ResolutionExact         Resolution = "EXACT"
	ResolutionInvariantOnly Resolution = "INVARIANT_ONLY"
	ResolutionUnknown       Resolution = "UNKNOWN"
	ResolutionNone          Resolution = "NONE"
)

type Role string

const (
	RoleContractAuthor  Role = "CONTRACT_AUTHOR"
	RoleImplementer     Role = "IMPLEMENTER"
	RoleEvaluatorAuthor Role = "EVALUATOR_AUTHOR"
	RoleAdapterAuthor   Role = "ADAPTER_AUTHOR"
	RolePolicyAdopter   Role = "POLICY_ADOPTER"
	RolePromoter        Role = "PROMOTER"
	RoleAuditor         Role = "AUDITOR"
)

type Decision string

const (
	DecisionUnknown    Decision = "UNKNOWN"
	DecisionPass       Decision = "PASS"
	DecisionFail       Decision = "FAIL"
	DecisionFixedPoint Decision = "FIXED_POINT"
	DecisionAuthorized Decision = "AUTHORIZED"
	DecisionAllow      Decision = "ALLOW"
	DecisionBlock      Decision = "BLOCK"
)

type AuthorityRoute struct {
	RuleID     string `json:"rule_id"`
	AuthoredBy string `json:"authored_by"`
	PromotedBy string `json:"promoted_by"`
}

type RoleBinding struct {
	Principal string `json:"principal"`
	Roles     []Role `json:"roles"`
}

type DecisionTransition struct {
	ID     string   `json:"id"`
	Input  Decision `json:"input"`
	Output Decision `json:"output"`
}

type Transaction struct {
	Schema              string               `json:"schema"`
	TransactionID       string               `json:"transaction_id"`
	AuthorityRoutes     []AuthorityRoute     `json:"authority_routes"`
	RoleBindings        []RoleBinding        `json:"role_bindings"`
	DecisionTransitions []DecisionTransition `json:"decision_transitions"`
}

type ObligationDefinition struct {
	MetricID             string         `json:"metric_id"`
	Priority             Priority       `json:"priority"`
	Class                IndicatorClass `json:"class"`
	ProofChoice          ProofChoice    `json:"proof_choice"`
	RequiredMetaOperation string         `json:"required_meta_operation"`
}

type MetaOperation struct {
	ID          string      `json:"id"`
	Activity    string      `json:"activity"`
	ProofChoice ProofChoice `json:"proof_choice"`
}

type ObligationObservation struct {
	MetricID     string     `json:"metric_id"`
	Status       string     `json:"status"`
	Resolution   Resolution `json:"resolution"`
	MetaOperation string     `json:"meta_operation,omitempty"`
}

type RolePair struct {
	Left  Role `json:"left"`
	Right Role `json:"right"`
}

type Finding struct {
	MetricID  string     `json:"metric_id"`
	PathID    string     `json:"path_id"`
	Principal string     `json:"principal,omitempty"`
	RuleID    string     `json:"rule_id,omitempty"`
	Roles     []Role     `json:"roles,omitempty"`
	DecisionID string    `json:"decision_id,omitempty"`
	Input     Decision   `json:"input,omitempty"`
	Output    Decision   `json:"output,omitempty"`
}

type Indicator struct {
	MetricID     string         `json:"metric_id"`
	Class        IndicatorClass `json:"class"`
	ProofChoice  ProofChoice    `json:"proof_choice"`
	Producer     string         `json:"producer"`
	Consumer     string         `json:"consumer"`
	MetaOperation string         `json:"meta_operation"`
	Value        *int           `json:"value"`
	Target       int            `json:"target"`
	Unit         string         `json:"unit"`
	Relation     Relation       `json:"relation"`
	Resolution   Resolution     `json:"resolution"`
	Satisfied    bool           `json:"satisfied"`
}

type Summary struct {
	DenominatorTotal          int  `json:"denominator_total"`
	Operating                 int  `json:"operating"`
	NotImplemented            int  `json:"not_implemented"`
	ImplementationCoverageBPS int  `json:"implementation_coverage_bps"`
	EvidenceGroupsObserved    int  `json:"evidence_groups_observed"`
	EvidenceGroupsTotal       int  `json:"evidence_groups_total"`
	EvidenceCoverageBPS       int  `json:"evidence_coverage_bps"`
	SelfMintingPaths          *int `json:"self_minting_paths"`
	RoleConflictPaths         *int `json:"role_conflict_paths"`
	UnknownLaunderingPaths    *int `json:"unknown_laundering_paths"`
	UnknownTopDecisions       *int `json:"unknown_top_decisions"`
	UnresolvedIndicators      int  `json:"unresolved_indicators"`
	ViolatedGuardrails        int  `json:"violated_guardrails"`
	RepositoryWrites          int  `json:"repository_writes"`
}

type Report struct {
	Schema                     string                  `json:"schema"`
	SubjectSHA                 string                  `json:"subject_sha"`
	TransactionDigest          string                  `json:"transaction_digest"`
	DenominatorID              string                  `json:"denominator_id"`
	DenominatorDigest          string                  `json:"denominator_digest"`
	AssuranceDecision          string                  `json:"assurance_decision"`
	CandidateDecision          string                  `json:"candidate_decision"`
	CandidateReason            string                  `json:"candidate_reason"`
	CandidateResolution        Resolution              `json:"candidate_resolution"`
	Denominator                []ObligationDefinition  `json:"denominator"`
	Obligations                []ObligationObservation `json:"obligations"`
	MetaOperations             []MetaOperation         `json:"meta_operations"`
	RoleConflictPairs          []RolePair              `json:"role_conflict_pairs"`
	UnknownLaunderingOutputs   []Decision              `json:"unknown_laundering_outputs"`
	Transaction                Transaction             `json:"transaction"`
	Findings                   []Finding               `json:"findings"`
	Indicators                 []Indicator             `json:"indicators"`
	Summary                    Summary                 `json:"summary"`
	ReportDigest               string                  `json:"report_digest"`
}
