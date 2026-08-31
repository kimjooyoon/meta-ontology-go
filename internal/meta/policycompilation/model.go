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
	Schema         string            `json:"schema"`
	PolicyID       string            `json:"policy_id"`
	Package        string            `json:"package"`
	Namespace      string            `json:"namespace"`
	SourceDigest   string            `json:"source_digest"`
	SemanticDigest string            `json:"semantic_digest"`
	Denominator    int               `json:"fixed_denominator"`
	Rules          []Rule            `json:"rules"`
	Reduction      DecisionReduction `json:"decision_reduction"`
}

type DecisionReduction struct {
	Schema string         `json:"schema"`
	Rules  []DecisionRule `json:"rules"`
}

type DecisionRule struct {
	Condition string `json:"condition"`
	Decision  string `json:"decision"`
	Stage     string `json:"stage"`
	Step      int    `json:"step"`
	Reason    string `json:"reason"`
}

type PolicyArtifact struct {
	Schema             string         `json:"schema"`
	Policy             CompiledPolicy `json:"policy"`
	GeneratedJudgeHash string         `json:"generated_judge_digest"`
}

type Case struct {
	ID                           string `json:"id"`
	ValidatorExpectation         string `json:"validator_expectation"`
	EvidenceClass                string `json:"evidence_class"`
	Provenance                   string `json:"provenance"`
	ProducerAvailable            bool   `json:"producer_available"`
	ConsumerAvailable            bool   `json:"consumer_available"`
	ObservedSourceDigest         string `json:"observed_source_digest"`
	ObservedArtifactSourceDigest string `json:"observed_artifact_source_digest"`
	ObservedGeneratedJudgeDigest string `json:"observed_generated_judge_digest"`
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
	Event             int    `json:"event"`
	ClaimID           string `json:"claim_id"`
	Predicate         string `json:"predicate"`
	From              string `json:"from"`
	To                string `json:"to"`
	Decision          string `json:"decision"`
	Stage             string `json:"stage"`
	Step              int    `json:"step"`
	Reason            string `json:"reason"`
	ObservationDigest string `json:"observation_digest"`
	Provenance        string `json:"provenance"`
	Observed          bool   `json:"observed"`
	PriorDigest       string `json:"prior_digest"`
	Digest            string `json:"digest"`
}

type ClaimPredicateObservation struct {
	ClaimID           string `json:"claim_id"`
	Predicate         string `json:"predicate"`
	Outcome           string `json:"outcome"`
	Observed          bool   `json:"observed"`
	Stage             string `json:"stage"`
	Step              int    `json:"step"`
	Reason            string `json:"reason"`
	ObservationDigest string `json:"observation_digest"`
	Provenance        string `json:"provenance"`
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
	ID                            string                      `json:"id"`
	ValidatorExpectation          string                      `json:"validator_expectation"`
	EvidenceClass                 string                      `json:"evidence_class"`
	ObservationDigest             string                      `json:"observation_digest"`
	Provenance                    string                      `json:"provenance"`
	Source                        DecisionResult              `json:"source"`
	Generated                     DecisionResult              `json:"generated"`
	Independent                   DecisionResult              `json:"independent"`
	AllDecisionsEquivalent        bool                        `json:"all_decisions_equivalent"`
	DecisionsEquivalent           bool                        `json:"decisions_equivalent"`
	ValidatorExpectationConfirmed bool                        `json:"validator_expectation_confirmed"`
	ClaimPredicates               []ClaimPredicateObservation `json:"claim_predicates"`
	ClaimStartDigest              string                      `json:"claim_start_digest"`
	ClaimEndDigest                string                      `json:"claim_end_digest"`
}

type CaseSummary struct {
	CaseCount                      int `json:"case_count"`
	PassCount                      int `json:"pass_count"`
	FailClosedCount                int `json:"fail_closed_count"`
	UnknownCount                   int `json:"unknown_count"`
	GeneratedIndependentEqual      int `json:"generated_independent_equivalent"`
	ValidatorExpectationsConfirmed int `json:"validator_expectations_confirmed"`
	SourceAllEquivalent            int `json:"source_all_equivalent"`
	ClaimPredicatesDischarged      int `json:"claim_predicates_discharged"`
	ClaimPredicatesRefuted         int `json:"claim_predicates_refuted"`
	ClaimPredicatesOpen            int `json:"claim_predicates_open"`
}

type Verification struct {
	Decision            string `json:"decision"`
	ConformanceDecision string `json:"conformance_decision"`
	SubjectResolution   string `json:"subject_resolution"`
	IndependentReplayed bool   `json:"independent_replayed"`
	GeneratedReplayed   bool   `json:"generated_replayed"`
	LedgerVerified      bool   `json:"ledger_verified"`
	FixedDenominator    int    `json:"fixed_denominator"`
	CaseDenominator     int    `json:"case_denominator"`
}

type WriteSetObservation struct {
	RepositoryBeforeDigest      string   `json:"repository_before_digest"`
	RepositoryAfterDigest       string   `json:"repository_after_digest"`
	RepositoryBeforeCount       int      `json:"repository_before_count"`
	RepositoryAfterCount        int      `json:"repository_after_count"`
	RepositoryNetChangeObserved bool     `json:"repository_net_change_observed"`
	GeneratedRootClass          string   `json:"generated_root_class"`
	GeneratedFiles              []string `json:"generated_files"`
	MutationAuthority           int      `json:"mutation_authority"`
	PromotionAuthority          int      `json:"promotion_authority"`
}

type EvidenceObservation struct {
	Class                           string `json:"class"`
	CaseID                          string `json:"case_id"`
	ProducerAvailable               bool   `json:"producer_available"`
	ConsumerAvailable               bool   `json:"consumer_available"`
	SourceDigest                    string `json:"source_digest"`
	ArtifactSourceDigest            string `json:"artifact_source_digest"`
	ArtifactDigest                  string `json:"artifact_digest"`
	GeneratedJudgeDigest            string `json:"generated_judge_digest"`
	IndependentDigest               string `json:"independent_digest"`
	IndependentReconstructionDigest string `json:"independent_reconstruction_digest"`
	SemanticDigest                  string `json:"semantic_digest"`
	ObservationDigest               string `json:"observation_digest"`
	Provenance                      string `json:"provenance"`
}

type Receipt struct {
	Schema          string                `json:"schema"`
	Policy          CompiledPolicy        `json:"policy"`
	Producer        ProducerEvidence      `json:"producer"`
	Consumer        ConsumerEvidence      `json:"consumer"`
	MetaOperation   string                `json:"meta_operation"`
	ProofChoice     string                `json:"proof_choice"`
	GeneratedDigest string                `json:"generated_judge_digest"`
	Cases           []CaseReceipt         `json:"cases"`
	Summary         CaseSummary           `json:"summary"`
	Claims          ClaimLedger           `json:"claims"`
	Verification    Verification          `json:"verification"`
	Evidence        []EvidenceObservation `json:"evidence"`
	WriteSet        WriteSetObservation   `json:"write_set"`
	ReceiptDigest   string                `json:"receipt_digest"`
}
