package partialknowledgecomposition

type Input struct {
	Repository   string           `json:"repository"`
	HeadSHA      string           `json:"head_sha"`
	SourcePath   string           `json:"source_path"`
	Source       []byte           `json:"-"`
	RawEvidence  []byte           `json:"-"`
	Intervention InterventionMode `json:"intervention"`
}

type RecipeOperand struct {
	Operation           string `json:"operation"`
	Required            string `json:"required"`
	ObservationRecipe   string `json:"observation_recipe"`
	DependencyRecipe    string `json:"dependency_recipe"`
	InvariantCapability string `json:"invariant_capability"`
}

type Case struct {
	ID               string        `json:"id"`
	SourceActivity   string        `json:"source_activity"`
	SourceActivityID string        `json:"source_activity_id"`
	Producer         string        `json:"producer"`
	Consumer         string        `json:"consumer"`
	MetaOperation    string        `json:"meta_operation"`
	ProofChoice      ProofChoice   `json:"proof_choice"`
	Left             RecipeOperand `json:"left"`
	Right            RecipeOperand `json:"right"`
}

type UpstreamClaim struct {
	ClaimID                 string `json:"claim_id"`
	Proposition             string `json:"proposition"`
	PropositionDigest       string `json:"proposition_digest"`
	Predicate               string `json:"predicate"`
	State                   string `json:"state"`
	Resolution              string `json:"resolution"`
	Stage                   string `json:"stage"`
	Step                    string `json:"step"`
	Reason                  string `json:"reason"`
	EvidenceDigest          string `json:"evidence_digest"`
	RawSourceDigest         string `json:"raw_source_digest"`
	SemanticDigest          string `json:"semantic_digest"`
	WorkspaceSnapshotDigest string `json:"workspace_snapshot_digest"`
	TargetOperation         string `json:"target_operation"`
	TargetOutput            string `json:"target_output"`
}

type EvidenceProvenance struct {
	Provider                string `json:"provider"`
	SourcePath              string `json:"source_path"`
	SourceDigest            string `json:"source_digest"`
	SemanticIRDigest        string `json:"semantic_ir_digest"`
	WorkspaceSnapshotDigest string `json:"workspace_snapshot_digest"`
	RawEvidenceDigest       string `json:"raw_evidence_digest"`
}

type Evidence struct {
	Operation         string             `json:"operation"`
	Required          string             `json:"required"`
	Observed          string             `json:"observed"`
	ObservedAvailable bool               `json:"observed_available"`
	Dependency        *UpstreamClaim     `json:"dependency,omitempty"`
	InvariantEvidence string             `json:"invariant_evidence,omitempty"`
	Stage             string             `json:"stage"`
	Step              string             `json:"step"`
	Reason            string             `json:"reason"`
	Provenance        EvidenceProvenance `json:"provenance"`
	EvidenceDigest    string             `json:"evidence_digest"`
}

type RawCase struct {
	ID               string      `json:"id"`
	SourceActivity   string      `json:"source_activity"`
	SourceActivityID string      `json:"source_activity_id"`
	Producer         string      `json:"producer"`
	Consumer         string      `json:"consumer"`
	MetaOperation    string      `json:"meta_operation"`
	ProofChoice      ProofChoice `json:"proof_choice"`
	Left             Evidence    `json:"left"`
	Right            Evidence    `json:"right"`
}

type Snapshot struct {
	Tracked   []string `json:"tracked"`
	Untracked []string `json:"untracked"`
	Status    []string `json:"status"`
	Digest    string   `json:"digest"`
}

type WorkspaceObservation struct {
	Before           Snapshot `json:"before"`
	After            Snapshot `json:"after"`
	ChangedPaths     []string `json:"changed_paths"`
	RepositoryWrites int      `json:"repository_writes"`
	Stage            string   `json:"stage"`
	Step             string   `json:"step"`
	Reason           string   `json:"reason"`
	EvidenceDigest   string   `json:"evidence_digest"`
}

type CapabilityObservation struct {
	Name           string `json:"name"`
	Available      bool   `json:"available"`
	State          string `json:"state"`
	Resolution     string `json:"resolution"`
	Stage          string `json:"stage"`
	Step           string `json:"step"`
	Reason         string `json:"reason"`
	EvidenceDigest string `json:"evidence_digest"`
}

type RawEvidenceReceipt struct {
	Schema           string                `json:"schema"`
	Repository       string                `json:"repository"`
	HeadSHA          string                `json:"head_sha"`
	SourcePath       string                `json:"source_path"`
	SourceDigest     string                `json:"source_digest"`
	SemanticIRDigest string                `json:"semantic_ir_digest"`
	SourceCases      int                   `json:"source_cases"`
	SourceCasesTotal int                   `json:"source_cases_total"`
	Provider         string                `json:"provider"`
	Cases            []RawCase             `json:"cases"`
	Workspace        WorkspaceObservation  `json:"workspace"`
	Authority        CapabilityObservation `json:"authority"`
	Digest           string                `json:"digest"`
}

type Value struct {
	State               State    `json:"state"`
	Contributors        []string `json:"contributors"`
	DirectUnknowns      []string `json:"direct_unknowns,omitempty"`
	BlockedDependencies []string `json:"blocked_dependencies,omitempty"`
	PreservedInvariants []string `json:"preserved_invariants,omitempty"`
}

type Provenance struct {
	SourcePath        string      `json:"source_path"`
	SourceActivity    string      `json:"source_activity"`
	Producer          string      `json:"producer"`
	Consumer          string      `json:"consumer"`
	MetaOperation     string      `json:"meta_operation"`
	ProofChoice       ProofChoice `json:"proof_choice"`
	RawSourceDigest   string      `json:"raw_source_digest"`
	SemanticIRDigest  string      `json:"semantic_ir_digest"`
	RawEvidenceDigest string      `json:"raw_evidence_digest"`
	ObservationDigest string      `json:"observation_digest"`
}

type CaseResult struct {
	ID                string      `json:"id"`
	SourceActivity    string      `json:"source_activity"`
	SourceActivityID  string      `json:"source_activity_id"`
	Producer          string      `json:"producer"`
	Consumer          string      `json:"consumer"`
	MetaOperation     string      `json:"meta_operation"`
	ProofChoice       ProofChoice `json:"proof_choice"`
	Left              Evidence    `json:"left"`
	Right             Evidence    `json:"right"`
	Result            Value       `json:"result"`
	Predicate         string      `json:"predicate"`
	Proposition       string      `json:"proposition"`
	PropositionDigest string      `json:"proposition_digest"`
	TargetAddress     string      `json:"target_address"`
	TargetOperation   string      `json:"target_operation"`
	TargetOutput      string      `json:"target_output"`
	Decision          string      `json:"decision"`
	Resolution        string      `json:"resolution"`
	Stage             string      `json:"stage"`
	Step              string      `json:"step"`
	Reason            string      `json:"reason"`
	TopSuccess        bool        `json:"top_success"`
	Provenance        Provenance  `json:"provenance"`
	EvidenceDigest    string      `json:"evidence_digest"`
}

type Summary struct {
	FixedDenominator       int `json:"fixed_denominator"`
	CaseTotal              int `json:"case_total"`
	ExactCases             int `json:"exact_cases"`
	DirectUnknownCases     int `json:"direct_unknown_cases"`
	DependencyBlockedCases int `json:"dependency_blocked_cases"`
	InvariantOnlyCases     int `json:"invariant_only_cases"`
	MixedUnresolvedCases   int `json:"mixed_unresolved_cases"`
	TopSuccessCases        int `json:"top_success_cases"`
	NonExactCases          int `json:"non_exact_cases"`
	NonExactNotPromoted    int `json:"non_exact_not_promoted"`
	OpenClaims             int `json:"open_claims"`
	DischargedClaims       int `json:"discharged_claims"`
	ClaimTransitionTotal   int `json:"claim_transition_total"`
	DistinctPredicateCount int `json:"distinct_predicate_count"`
	PredicateDenominator   int `json:"predicate_denominator"`
	RepositoryWrites       int `json:"repository_writes"`
}

type Indicator struct {
	ID                string      `json:"id"`
	Class             string      `json:"class"`
	ProofChoice       ProofChoice `json:"proof_choice"`
	MetaOperation     string      `json:"meta_operation"`
	Producer          string      `json:"producer"`
	Consumer          string      `json:"consumer"`
	Observed          int         `json:"observed"`
	Denominator       int         `json:"denominator"`
	BasisPoints       int         `json:"basis_points"`
	TargetBasisPoints int         `json:"target_basis_points"`
	Satisfied         bool        `json:"satisfied"`
}

type ClaimTransition struct {
	Sequence          int        `json:"sequence"`
	ClaimID           string     `json:"claim_id"`
	From              string     `json:"from"`
	To                string     `json:"to"`
	Predicate         string     `json:"predicate"`
	Proposition       string     `json:"proposition"`
	PropositionDigest string     `json:"proposition_digest"`
	TargetAddress     string     `json:"target_address"`
	TargetOperation   string     `json:"target_operation"`
	TargetOutput      string     `json:"target_output"`
	Stage             string     `json:"stage"`
	Step              string     `json:"step"`
	Reason            string     `json:"reason"`
	RawSourceDigest   string     `json:"raw_source_digest"`
	SemanticDigest    string     `json:"semantic_digest"`
	RawEvidenceDigest string     `json:"raw_evidence_digest"`
	EvidenceDigest    string     `json:"evidence_digest"`
	Provenance        Provenance `json:"provenance"`
	PreviousDigest    string     `json:"previous_digest,omitempty"`
	Digest            string     `json:"digest"`
}

type Intervention struct {
	Mode             InterventionMode `json:"mode"`
	Semantic         bool             `json:"semantic"`
	SourceDigest     string           `json:"source_digest"`
	SemanticIRDigest string           `json:"semantic_ir_digest"`
	Target           string           `json:"target,omitempty"`
	From             string           `json:"from,omitempty"`
	To               string           `json:"to,omitempty"`
	Comment          string           `json:"comment,omitempty"`
}

type Receipt struct {
	Schema                   string            `json:"schema"`
	Repository               string            `json:"repository"`
	HeadSHA                  string            `json:"head_sha"`
	SourcePath               string            `json:"source_path"`
	SourceDigest             string            `json:"source_digest"`
	SemanticIRDigest         string            `json:"semantic_ir_digest"`
	RawEvidenceDigest        string            `json:"raw_evidence_digest"`
	SourceCases              int               `json:"source_cases"`
	SourceCasesTotal         int               `json:"source_cases_total"`
	Producer                 string            `json:"producer"`
	Consumer                 string            `json:"consumer"`
	MetaOperation            string            `json:"meta_operation"`
	ProofChoice              ProofChoice       `json:"proof_choice"`
	Resolution               string            `json:"resolution"`
	SubjectResolution        string            `json:"subject_resolution"`
	EvidenceCoverage         string            `json:"evidence_coverage"`
	AuthorityState           string            `json:"authority_state"`
	AuthorityResolution      string            `json:"authority_resolution"`
	AuthorityStage           string            `json:"authority_stage"`
	AuthorityStep            string            `json:"authority_step"`
	AuthorityReason          string            `json:"authority_reason"`
	Decision                 string            `json:"decision"`
	Reason                   string            `json:"reason"`
	FixedDenominator         int               `json:"fixed_denominator"`
	Cases                    []CaseResult      `json:"cases"`
	Claims                   []ClaimTransition `json:"claims"`
	Indicators               []Indicator       `json:"indicators"`
	Summary                  Summary           `json:"summary"`
	Intervention             Intervention      `json:"intervention"`
	RepositoryWrites         int               `json:"repository_writes"`
	PromotionAuthorized      bool              `json:"promotion_authorized"`
	SemanticProjectionDigest string            `json:"semantic_projection_digest"`
	Digest                   string            `json:"digest"`
}
