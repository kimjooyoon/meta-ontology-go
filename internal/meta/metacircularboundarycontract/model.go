package metacircularboundarycontract

type Coordinate struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type SourceObservation struct {
	Path                string            `json:"path"`
	SourceDigest        string            `json:"source_digest"`
	SemanticDigest      string            `json:"semantic_digest"`
	Package             string            `json:"package"`
	Namespace           string            `json:"namespace"`
	Entities            []string          `json:"entities"`
	Activities          []string          `json:"activities"`
	Computations        []Computation     `json:"computations"`
	Graph               GraphObservation  `json:"graph"`
	Effect              EffectObservation `json:"effect"`
	GrantArtifactDigest string            `json:"grant_artifact_digest"`
	DescriptionBound    bool              `json:"description_bound"`
	ReadOnly            bool              `json:"read_only"`
	RepositoryWrites    int               `json:"repository_writes"`
	MutationAuthority   bool              `json:"mutation_authority"`
}

type Computation struct {
	Activity string `json:"activity"`
	Program  string `json:"program"`
}

type GraphObservation struct {
	Schema    string          `json:"schema"`
	Relations []TypedRelation `json:"relations"`
	Path      []string        `json:"path"`
	Digest    string          `json:"digest"`
	Valid     bool            `json:"valid"`
	Reason    string          `json:"reason"`
}

type TypedRelation struct {
	Ordinal      int    `json:"ordinal"`
	FromActivity string `json:"from_activity"`
	FromType     string `json:"from_type"`
	Relation     string `json:"relation"`
	ThroughType  string `json:"through_type"`
	ToType       string `json:"to_type"`
	ToActivity   string `json:"to_activity"`
}

type EffectEvidence struct {
	Schema                  string `json:"schema"`
	Producer                string `json:"producer"`
	OutputPath              string `json:"output_path"`
	TrackedBeforeDigest     string `json:"tracked_before_digest"`
	TrackedAfterDigest      string `json:"tracked_after_digest"`
	UntrackedBeforeDigest   string `json:"untracked_before_digest"`
	UntrackedAfterDigest    string `json:"untracked_after_digest"`
	OutputOutsideRepository bool   `json:"output_outside_repository"`
	PermissionEvidence      string `json:"permission_evidence"`
	MutationAuthority       string `json:"mutation_authority"`
}

type EffectObservation struct {
	Known                   bool   `json:"known"`
	EvidenceDigest          string `json:"evidence_digest"`
	OutputPath              string `json:"output_path"`
	OutputOutsideRepository bool   `json:"output_outside_repository"`
	PermissionEvidence      string `json:"permission_evidence"`
	RepositoryWrites        int    `json:"repository_writes"`
	MutationAuthority       string `json:"mutation_authority"`
}

type ExternalGrantArtifact struct {
	Schema         string          `json:"schema"`
	Producer       string          `json:"producer"`
	Policy         string          `json:"policy"`
	Grants         []ExternalGrant `json:"grants"`
	ArtifactDigest string          `json:"artifact_digest"`
}

type ExternalGrant struct {
	CaseID        string `json:"case_id"`
	Decision      string `json:"decision"`
	Reason        string `json:"reason"`
	Issuer        string `json:"issuer"`
	SubjectDigest string `json:"subject_digest"`
	Operation     string `json:"operation"`
	Scope         string `json:"scope"`
	Handle        string `json:"handle"`
	GrantDigest   string `json:"grant_digest"`
}

type Attempt struct {
	FactActivity              string `json:"fact_activity"`
	Predicate                 string `json:"predicate"`
	DescriptionDigest         string `json:"description_digest"`
	RequestKind               string `json:"request_kind"`
	RequestExecution          bool   `json:"request_execution"`
	DescriptionAuthorityClaim bool   `json:"description_authority_claim"`
	Unknown                   bool   `json:"unknown"`
	Contradictory             bool   `json:"contradictory"`
}

type CaseDefinition struct {
	ID                    string `json:"id"`
	Predicate             string `json:"predicate"`
	ExpectedDecision      string `json:"expected_decision"`
	ExpectedAuthorization string `json:"expected_authorization"`
	ExpectedExecution     string `json:"expected_execution"`
	ExpectedReason        string `json:"expected_reason"`
	ProofChoice           string `json:"proof_choice"`
	MetaOperation         string `json:"meta_operation"`
}

type CaseObservation struct {
	Description              string `json:"description"`
	Authorization            string `json:"authorization"`
	Execution                string `json:"execution"`
	Decision                 string `json:"decision"`
	Reason                   string `json:"reason"`
	Predicate                string `json:"predicate"`
	DescriptionEscalated     bool   `json:"description_escalated"`
	GrantDigest              string `json:"grant_digest"`
	ExecutionArtifactPresent bool   `json:"execution_artifact_present"`
	ExecutionArtifactValid   bool   `json:"execution_artifact_valid"`
	OutputDigest             string `json:"output_digest"`
	RepositoryWrites         int    `json:"repository_writes"`
	MutationAuthority        bool   `json:"mutation_authority"`
}

type ClaimTransition struct {
	ClaimID           string     `json:"claim_id"`
	PropositionDigest string     `json:"proposition_digest"`
	Event             string     `json:"event"`
	Before            string     `json:"before"`
	After             string     `json:"after"`
	Coordinate        Coordinate `json:"coordinate"`
	EvidenceDigest    string     `json:"evidence_digest,omitempty"`
	DependsOnClaimID  string     `json:"depends_on_claim_id,omitempty"`
	DependsOnAfter    string     `json:"depends_on_after,omitempty"`
}

type ExecutionArtifact struct {
	Schema          string `json:"schema"`
	Producer        string `json:"producer"`
	CaseID          string `json:"case_id"`
	Path            string `json:"path"`
	OperationID     string `json:"operation_id"`
	GrantDigest     string `json:"grant_digest"`
	InputDigest     string `json:"input_digest"`
	OutputCanonical string `json:"output_canonical"`
	OutputDigest    string `json:"output_digest"`
	ArtifactDigest  string `json:"artifact_digest"`
}

type Receipt struct {
	Schema                      string             `json:"schema"`
	CaseID                      string             `json:"case_id"`
	Producer                    string             `json:"producer"`
	Consumer                    string             `json:"consumer"`
	MetaOperation               string             `json:"meta_operation"`
	ProofChoice                 string             `json:"proof_choice"`
	Coordinate                  Coordinate         `json:"coordinate"`
	SourceDigest                string             `json:"source_digest"`
	DescriptionDigest           string             `json:"description_digest"`
	AuthorizationEvidenceDigest string             `json:"authorization_evidence_digest"`
	GrantDigest                 string             `json:"grant_digest"`
	Decision                    string             `json:"decision"`
	Authorization               string             `json:"authorization"`
	Execution                   string             `json:"execution"`
	RepositoryWrites            int                `json:"repository_writes"`
	MutationAuthority           bool               `json:"mutation_authority"`
	ClaimTransitions            []ClaimTransition  `json:"claim_transitions"`
	ExecutionArtifact           *ExecutionArtifact `json:"execution_artifact,omitempty"`
	ReceiptDigest               string             `json:"receipt_digest"`
}

type CaseResult struct {
	Grant       ExternalGrant   `json:"grant"`
	Definition  CaseDefinition  `json:"definition"`
	Attempt     Attempt         `json:"attempt"`
	Observation CaseObservation `json:"observation"`
	Receipt     Receipt         `json:"receipt"`
	Passed      bool            `json:"passed"`
}

type Indicator struct {
	MetricID      string `json:"metric_id"`
	Class         string `json:"class"`
	ProofChoice   string `json:"proof_choice"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	Unit          string `json:"unit"`
	Relation      string `json:"relation"`
	Value         int    `json:"value"`
	Target        int    `json:"target"`
	Satisfied     bool   `json:"satisfied"`
}

type Summary struct {
	ReceiptSelfSealValid            int `json:"receipt_self_seal_valid"`
	ExecutionArtifactsObserved      int `json:"execution_artifacts_observed"`
	CasesTotal                      int `json:"cases_total"`
	CasesPassed                     int `json:"cases_passed"`
	CaseCoverageBPS                 int `json:"case_coverage_bps"`
	DescriptionBound                int `json:"description_bound"`
	ExplicitAuthorizations          int `json:"explicit_authorizations"`
	AllowedExecutions               int `json:"allowed_executions"`
	DescriptionOnlyBlocked          int `json:"description_only_blocked"`
	ForgedAuthorizationsBlocked     int `json:"forged_authorizations_blocked"`
	OutOfScopeAuthorizationsBlocked int `json:"out_of_scope_authorizations_blocked"`
	DescriptionEscalationPaths      int `json:"description_escalation_paths"`
	ReplayMatches                   int `json:"replay_matches"`
	RepositoryWrites                int `json:"repository_writes"`
	MutationAuthority               int `json:"mutation_authority"`
}

type Report struct {
	ExecutionArtifacts   []ExecutionArtifact `json:"execution_artifacts"`
	ReplayEvidenceDigest string              `json:"replay_evidence_digest"`
	Schema               string              `json:"schema"`
	Scope                string              `json:"scope"`
	HeadSHA              string              `json:"head_sha"`
	Source               SourceObservation   `json:"source"`
	Decision             string              `json:"decision"`
	Resolution           string              `json:"resolution"`
	Reason               string              `json:"reason"`
	Coordinate           Coordinate          `json:"coordinate"`
	Denominator          Denominator         `json:"denominator"`
	Cases                []CaseResult        `json:"cases"`
	Receipts             []Receipt           `json:"receipts"`
	Indicators           []Indicator         `json:"indicators"`
	MetaOperations       []MetaOperation     `json:"meta_operations"`
	Summary              Summary             `json:"summary"`
	RepositoryWrites     int                 `json:"repository_writes"`
	MutationAuthority    bool                `json:"mutation_authority"`
	MetaValue            string              `json:"meta_value"`
	NotClaimed           []string            `json:"not_claimed"`
	ReportDigest         string              `json:"report_digest"`
}

type MetaOperation struct {
	ID          string `json:"id"`
	Producer    string `json:"producer"`
	Consumer    string `json:"consumer"`
	ProofChoice string `json:"proof_choice"`
}

type Denominator struct {
	ID     string           `json:"id"`
	Cases  []CaseDefinition `json:"cases"`
	Digest string           `json:"digest"`
}

type Input struct {
	GrantEvidence      []byte
	EffectEvidence     []byte
	ReplayEvidence     []byte
	ExecutionArtifacts []ExecutionArtifact
	Path               string
	HeadSHA            string
	Source             []byte
}

type ReplayEvidence struct {
	Schema            string   `json:"schema"`
	Producer          string   `json:"producer"`
	ReceiptDigestsA   []string `json:"receipt_digests_a"`
	ReceiptDigestsB   []string `json:"receipt_digests_b"`
	ExecutionDigestsA []string `json:"execution_digests_a"`
	ExecutionDigestsB []string `json:"execution_digests_b"`
	Equal             bool     `json:"equal"`
	EvidenceDigest    string   `json:"evidence_digest"`
}

type JudgeReceipt struct {
	Schema            string   `json:"schema"`
	Producer          string   `json:"producer"`
	Consumer          string   `json:"consumer"`
	InputReportDigest string   `json:"input_report_digest"`
	ComparedCases     int      `json:"compared_cases"`
	Mismatches        int      `json:"mismatches"`
	Decision          string   `json:"decision"`
	Reason            string   `json:"reason"`
	ReceiptDigests    []string `json:"receipt_digests"`
	Digest            string   `json:"digest"`
}

type CausalityCase struct {
	ID                                       string `json:"id"`
	Kind                                     string `json:"kind"`
	TargetCaseID                             string `json:"target_case_id"`
	BaselineSourceDigest                     string `json:"baseline_source_digest"`
	IntervenedSourceDigest                   string `json:"intervened_source_digest"`
	BaselineSemanticDigest                   string `json:"baseline_semantic_digest"`
	IntervenedSemanticDigest                 string `json:"intervened_semantic_digest"`
	BaselineGrantDigest                      string `json:"baseline_grant_digest"`
	IntervenedGrantDigest                    string `json:"intervened_grant_digest"`
	BaselineOutputDigest                     string `json:"baseline_output_digest"`
	IntervenedOutputDigest                   string `json:"intervened_output_digest"`
	BaselineDescriptionPropositionDigest     string `json:"baseline_description_proposition_digest"`
	IntervenedDescriptionPropositionDigest   string `json:"intervened_description_proposition_digest"`
	BaselineAuthorizationPropositionDigest   string `json:"baseline_authorization_proposition_digest"`
	IntervenedAuthorizationPropositionDigest string `json:"intervened_authorization_proposition_digest"`
	BaselineExecutionPropositionDigest       string `json:"baseline_execution_proposition_digest"`
	IntervenedExecutionPropositionDigest     string `json:"intervened_execution_proposition_digest"`
	BaselineDescriptionEscalated             bool   `json:"baseline_description_escalated"`
	IntervenedDescriptionEscalated           bool   `json:"intervened_description_escalated"`
	SourceChanged                            bool   `json:"source_changed"`
	SemanticChanged                          bool   `json:"semantic_changed"`
	RawInputChanged                          bool   `json:"raw_input_changed"`
	GrantChanged                             bool   `json:"grant_changed"`
	GraphChanged                             bool   `json:"graph_changed"`
	ExpectedSemanticChange                   bool   `json:"expected_semantic_change"`
	ConsumerAccepted                         bool   `json:"consumer_accepted"`
	BaselineCaseDecision                     string `json:"baseline_case_decision"`
	IntervenedCaseDecision                   string `json:"intervened_case_decision"`
	BaselineAuthorization                    string `json:"baseline_authorization"`
	IntervenedAuthorization                  string `json:"intervened_authorization"`
	BaselineExecution                        string `json:"baseline_execution"`
	IntervenedExecution                      string `json:"intervened_execution"`
	BaselineReceiptDecision                  string `json:"baseline_receipt_decision"`
	IntervenedReceiptDecision                string `json:"intervened_receipt_decision"`
	BaselineAuthorizationEvidenceDigest      string `json:"baseline_authorization_evidence_digest"`
	IntervenedAuthorizationEvidenceDigest    string `json:"intervened_authorization_evidence_digest"`
	BaselineAuthorizationClaim               string `json:"baseline_authorization_claim"`
	IntervenedAuthorizationClaim             string `json:"intervened_authorization_claim"`
	BaselineExecutionClaim                   string `json:"baseline_execution_claim"`
	IntervenedExecutionClaim                 string `json:"intervened_execution_claim"`
	BaselineDescriptionClaim                 string `json:"baseline_description_claim"`
	IntervenedDescriptionClaim               string `json:"intervened_description_claim"`
	SemanticOutputsPreserved                 bool   `json:"semantic_outputs_preserved"`
	Passed                                   bool   `json:"passed"`
}

type CausalitySummary struct {
	Total       int `json:"total"`
	Passed      int `json:"passed"`
	CoverageBPS int `json:"coverage_bps"`
}

type CausalityReport struct {
	Schema  string           `json:"schema"`
	Cases   []CausalityCase  `json:"cases"`
	Summary CausalitySummary `json:"summary"`
}
