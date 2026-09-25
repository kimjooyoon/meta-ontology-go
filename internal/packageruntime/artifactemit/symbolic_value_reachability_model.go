package artifactemit

const symbolicValueReachabilitySchema = "gooo/symbolic-invocation-value-reachability/v1"

type SymbolicValueReachability struct {
	Schema             string                           `json:"schema"`
	SubjectSHA         string                           `json:"subject_sha"`
	MetricID           string                           `json:"metric_id"`
	Decision           string                           `json:"decision"`
	Resolution         string                           `json:"resolution"`
	Reason             string                           `json:"reason"`
	Source             SymbolicValueReachabilitySource  `json:"source"`
	Summary            SymbolicValueReachabilitySummary `json:"summary"`
	Rules              []SymbolicValueRuleReachability  `json:"rules"`
	Default            SymbolicValueDefaultReachability `json:"default"`
	Coordinates        SymbolicValueContractCoordinates `json:"coordinates"`
	Classes            []SymbolicValueContractClass     `json:"classes"`
	Indicators         []SymbolicValueContractIndicator `json:"indicators"`
	Views              []SymbolicValueContractView      `json:"views"`
	Proofs             []SymbolicValueContractProof     `json:"proofs"`
	Effects            SymbolicValueContractEffects     `json:"effects"`
	PromotionCreditBPS int                              `json:"promotion_credit_bps"`
	NotClaimed         []string                         `json:"not_claimed"`
	Digest             string                           `json:"digest,omitempty"`
}

type SymbolicValueReachabilitySource struct {
	ArtifactDigest string `json:"artifact_digest"`
	ContractDigest string `json:"contract_digest"`
}

type SymbolicValueReachabilitySummary struct {
	PolicyBranches        int `json:"policy_branches"`
	ReachableRules        int `json:"reachable_rules"`
	DefenseOnlyRules      int `json:"defense_only_rules"`
	ReachableDefaults     int `json:"reachable_defaults"`
	DefenseOnlyDefaults   int `json:"defense_only_defaults"`
	UnknownPolicyBranches int `json:"unknown_policy_branches"`
}

type SymbolicValueRuleReachability struct {
	ID                           string `json:"id"`
	Reachability                 string `json:"reachability"`
	ReachableAfterStructuralGate bool   `json:"reachable_after_structural_gate"`
	Role                         string `json:"role"`
	Reason                       string `json:"reason"`
	ProofChoice                  string `json:"proof_choice"`
	MetaOperation                string `json:"meta_operation"`
}

type SymbolicValueDefaultReachability struct {
	Reachability                 string `json:"reachability"`
	ReachableAfterStructuralGate bool   `json:"reachable_after_structural_gate"`
	Role                         string `json:"role"`
	Reason                       string `json:"reason"`
	ProofChoice                  string `json:"proof_choice"`
	MetaOperation                string `json:"meta_operation"`
}

type symbolicValueReachabilityAnalysis struct {
	SchemaProfileSupported bool
	SchemaEntailsReady     bool
	Rules                  []SymbolicValueRuleReachability
	Default                SymbolicValueDefaultReachability
	Summary                SymbolicValueReachabilitySummary
}
