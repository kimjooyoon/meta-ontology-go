package consumer

type Observation struct {
	Schema       string                   `json:"schema"`
	Repository   string                   `json:"repository"`
	BaseSHA      string                   `json:"base_sha"`
	HeadSHA      string                   `json:"head_sha"`
	SourcePath   string                   `json:"source_path"`
	ChangedFiles []ChangedFileObservation `json:"changed_files"`
	PriorClaims  []PriorClaimObservation  `json:"prior_claims"`
	Isolation    IsolationObservation     `json:"isolation"`
}

type ChangedFileObservation struct {
	Path         string `json:"path"`
	Status       string `json:"status"`
	BeforeObject string `json:"before_object,omitempty"`
	AfterObject  string `json:"after_object,omitempty"`
}

type PriorClaimObservation struct {
	ClaimID     string `json:"claim_id"`
	SubjectPath string `json:"subject_path"`
	State       string `json:"state"`
	Provenance  string `json:"provenance"`
}

type IsolationObservation struct {
	Before RepositorySnapshot `json:"before"`
	After  RepositorySnapshot `json:"after"`
}

type RepositorySnapshot struct {
	StatusLines  []string `json:"status_lines"`
	StatusDigest string   `json:"status_digest"`
}

type SourceEvidence struct {
	Path           string `json:"path"`
	RawDigest      string `json:"raw_digest"`
	ParsedDigest   string `json:"parsed_digest"`
	SemanticDigest string `json:"semantic_digest"`
}

type Check struct {
	ID         string `json:"id"`
	Ordinal    int    `json:"ordinal"`
	SemanticID string `json:"semantic_id"`
}

type PolicyEdge struct {
	ID           string `json:"id"`
	Kind         string `json:"kind"`
	From         string `json:"from"`
	To           string `json:"to"`
	ActivityID   string `json:"activity_id"`
	ValueProgram string `json:"value_program"`
}

type PriorStateRule struct {
	State        string `json:"state"`
	ActivityID   string `json:"activity_id"`
	ValueProgram string `json:"value_program"`
}

type PolicyContradiction struct {
	Stage  string   `json:"stage"`
	Step   string   `json:"step"`
	Reason string   `json:"reason"`
	Edges  []string `json:"edges"`
}

type PolicyGraph struct {
	Source          SourceEvidence
	ChangedFileID   string
	ClaimID         string
	SurfaceID       string
	Checks          []Check
	Edges           []PolicyEdge
	PriorStates     []PriorStateRule
	Contradictions  []PolicyContradiction
	ClaimStateRules map[string]string
}

type Coordinate struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type Operation struct {
	Producer          string `json:"producer"`
	Consumer          string `json:"consumer"`
	MetaOperation     string `json:"meta_operation"`
	ProofChoice       string `json:"proof_choice"`
	ReadOnly          bool   `json:"read_only"`
	RepositoryWrites  int    `json:"repository_writes"`
	MutationAuthority bool   `json:"mutation_authority"`
}

type PathEvidence struct {
	SubjectPath    string   `json:"subject_path"`
	ClaimIDs       []string `json:"claim_ids"`
	SurfaceID      string   `json:"surface_id"`
	CheckID        string   `json:"check_id"`
	PolicyEdgeIDs  []string `json:"policy_edge_ids"`
	SemanticDigest string   `json:"semantic_digest"`
	Explanation    string   `json:"explanation"`
	ProofChoice    string   `json:"proof_choice"`
}

type UnknownCause struct {
	SubjectPath string     `json:"subject_path"`
	Coordinate  Coordinate `json:"coordinate"`
	Provenance  string     `json:"provenance"`
}

type CheckChoice struct {
	CheckID     string   `json:"check_id"`
	ProofChoice string   `json:"proof_choice"`
	Reason      string   `json:"reason"`
	ClaimIDs    []string `json:"claim_ids,omitempty"`
	PathIDs     []string `json:"path_ids,omitempty"`
}

type SubjectResolution struct {
	Path           string         `json:"path"`
	Resolution     string         `json:"resolution"`
	Coordinate     Coordinate     `json:"coordinate"`
	Paths          []PathEvidence `json:"paths,omitempty"`
	UnknownCauses  []UnknownCause `json:"unknown_causes,omitempty"`
	SelectedChecks []CheckChoice  `json:"selected_checks"`
}

type ClaimTransition struct {
	Sequence       int    `json:"sequence"`
	ClaimID        string `json:"claim_id"`
	SubjectPath    string `json:"subject_path"`
	Before         string `json:"before"`
	After          string `json:"after"`
	Resolution     string `json:"resolution"`
	Stage          string `json:"stage"`
	Step           string `json:"step"`
	Reason         string `json:"reason"`
	EvidenceDigest string `json:"evidence_digest"`
	Provenance     string `json:"provenance"`
	PreviousDigest string `json:"previous_digest"`
	Digest         string `json:"digest"`
}

type Conformance struct {
	Decision   string     `json:"decision"`
	Coordinate Coordinate `json:"coordinate"`
}

type Metrics struct {
	ChangedFileNumerator      int `json:"changed_file_numerator"`
	ChangedFileDenominator    int `json:"changed_file_denominator"`
	SubjectTotal              int `json:"subject_total"`
	SelectedSubjectTotal      int `json:"selected_subject_total"`
	UnknownSubjectTotal       int `json:"unknown_subject_total"`
	FailClosedSubjectTotal    int `json:"fail_closed_subject_total"`
	SelectedCheckTotal        int `json:"selected_check_total"`
	FullSuiteCheckDenominator int `json:"full_suite_check_denominator"`
	ClaimTransitionTotal      int `json:"claim_transition_total"`
	DischargedClaimTotal      int `json:"discharged_claim_total"`
	LowerResolutionClaimTotal int `json:"lower_resolution_claim_total"`
	RefutedClaimTotal         int `json:"refuted_claim_total"`
	FixedIndicatorSatisfied   int `json:"fixed_indicator_satisfied"`
	FixedIndicatorDenominator int `json:"fixed_indicator_denominator"`
	SourceReconstructionNumer int `json:"source_reconstruction_numerator"`
	SourceReconstructionDenom int `json:"source_reconstruction_denominator"`
}

type Indicator struct {
	ID          string `json:"id"`
	Observed    int    `json:"observed"`
	Denominator int    `json:"denominator"`
	Satisfied   bool   `json:"satisfied"`
}

type IndependentVerifier struct {
	ID       string `json:"id"`
	Mode     string `json:"mode"`
	Required bool   `json:"required"`
	ReadOnly bool   `json:"read_only"`
}

type Receipt struct {
	Schema              string              `json:"schema"`
	Scope               string              `json:"scope"`
	Source              SourceEvidence      `json:"source"`
	ObservationDigest   string              `json:"observation_digest"`
	Operation           Operation           `json:"operation"`
	ExecutionMode       string              `json:"execution_mode"`
	Conformance         Conformance         `json:"conformance"`
	Subjects            []SubjectResolution `json:"subjects"`
	ClaimTransitions    []ClaimTransition   `json:"claim_transitions"`
	Metrics             Metrics             `json:"metrics"`
	Indicators          []Indicator         `json:"indicators"`
	IndependentVerifier IndependentVerifier `json:"independent_verifier"`
	PlanDigest          string              `json:"plan_digest"`
	Digest              string              `json:"digest"`
}
