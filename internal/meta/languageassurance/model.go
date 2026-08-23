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
