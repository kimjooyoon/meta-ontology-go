package policycompilation

type Rule struct {
	ActivityID    string `json:"activity_id"`
	ActivityName  string `json:"activity_name"`
	Role          string `json:"role"`
	MetaOperation string `json:"meta_operation"`
	ProofChoice   string `json:"proof_choice"`
	Stage         string `json:"stage"`
	Step          int    `json:"step"`
	Reason        string `json:"reason"`
	Claim         string `json:"claim"`
}

type CompiledPolicy struct {
	Schema         string `json:"schema"`
	PolicyID       string `json:"policy_id"`
	Package        string `json:"package"`
	Namespace      string `json:"namespace"`
	SourceDigest   string `json:"source_digest"`
	SemanticDigest string `json:"semantic_digest"`
	Denominator    int    `json:"fixed_denominator"`
	Rules          []Rule `json:"rules"`
}

type PolicyArtifact struct {
	Schema             string         `json:"schema"`
	Policy             CompiledPolicy `json:"policy"`
	GeneratedJudgeHash string         `json:"generated_judge_digest"`
}

type Case struct {
	ID                           string `json:"id"`
	Expected                     string `json:"expected"`
	ProducerAvailable            bool   `json:"producer_available"`
	ConsumerAvailable            bool   `json:"consumer_available"`
	ObservedSourceDigest         string `json:"observed_source_digest"`
	ObservedArtifactSourceDigest string `json:"observed_artifact_source_digest"`
	ObservedIndependentDigest    string `json:"observed_independent_digest"`
}

type DecisionResult struct {
	CaseID         string `json:"case_id"`
	Decision       string `json:"decision"`
	Stage          string `json:"stage"`
	Step           int    `json:"step"`
	Reason         string `json:"reason"`
	PolicyDigest   string `json:"policy_digest"`
	SemanticDigest string `json:"semantic_digest"`
	Denominator    int    `json:"fixed_denominator"`
}

type ClaimTransition struct {
	Event       int    `json:"event"`
	ClaimID     string `json:"claim_id"`
	From        string `json:"from"`
	To          string `json:"to"`
	Decision    string `json:"decision"`
	Stage       string `json:"stage"`
	Step        int    `json:"step"`
	Reason      string `json:"reason"`
	PriorDigest string `json:"prior_digest"`
	Digest      string `json:"digest"`
}

type ClaimLedger struct {
	Schema     string            `json:"schema"`
	EventCount int               `json:"event_count"`
	Events     []ClaimTransition `json:"events"`
	HeadDigest string            `json:"head_digest"`
}

type ProducerEvidence struct {
	Role           string `json:"role"`
	Stage          string `json:"stage"`
	Step           int    `json:"step"`
	Reason         string `json:"reason"`
	SourceDigest   string `json:"source_digest"`
	SemanticDigest string `json:"semantic_digest"`
	Denominator    int    `json:"fixed_denominator"`
}

type ConsumerEvidence struct {
	Role                 string `json:"role"`
	Stage                string `json:"stage"`
	Step                 int    `json:"step"`
	Reason               string `json:"reason"`
	ArtifactSourceDigest string `json:"artifact_source_digest"`
	ArtifactDigest       string `json:"artifact_digest"`
	SourceMatches        bool   `json:"source_matches"`
	RulesMatch           bool   `json:"rules_match"`
}

type CaseReceipt struct {
	ID                        string         `json:"id"`
	Expected                  string         `json:"expected"`
	Generated                 DecisionResult `json:"generated"`
	Independent               DecisionResult `json:"independent"`
	DecisionsEquivalent       bool           `json:"decisions_equivalent"`
	ExpectedDecisionConfirmed bool           `json:"expected_decision_confirmed"`
	ClaimStartDigest          string         `json:"claim_start_digest"`
	ClaimEndDigest            string         `json:"claim_end_digest"`
}

type CaseSummary struct {
	CaseCount                  int `json:"case_count"`
	PassCount                  int `json:"pass_count"`
	FailClosedCount            int `json:"fail_closed_count"`
	UnknownCount               int `json:"unknown_count"`
	GeneratedIndependentEqual  int `json:"generated_independent_equivalent"`
	ExpectedDecisionsConfirmed int `json:"expected_decisions_confirmed"`
}

type Verification struct {
	Decision            string `json:"decision"`
	IndependentReplayed bool   `json:"independent_replayed"`
	GeneratedReplayed   bool   `json:"generated_replayed"`
	LedgerVerified      bool   `json:"ledger_verified"`
	FixedDenominator    int    `json:"fixed_denominator"`
	CaseDenominator     int    `json:"case_denominator"`
}

type Receipt struct {
	Schema          string           `json:"schema"`
	Policy          CompiledPolicy   `json:"policy"`
	Producer        ProducerEvidence `json:"producer"`
	Consumer        ConsumerEvidence `json:"consumer"`
	MetaOperation   string           `json:"meta_operation"`
	ProofChoice     string           `json:"proof_choice"`
	GeneratedDigest string           `json:"generated_judge_digest"`
	Cases           []CaseReceipt    `json:"cases"`
	Summary         CaseSummary      `json:"summary"`
	Claims          ClaimLedger      `json:"claims"`
	Verification    Verification     `json:"verification"`
	ReceiptDigest   string           `json:"receipt_digest"`
}
