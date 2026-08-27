package operationprovenance

type Lineage struct {
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	EvidencePath  string `json:"evidence_path"`
}

type Provenance struct {
	SourceDigest     string `json:"source_digest"`
	SemanticDigest   string `json:"semantic_digest"`
	Producer         string `json:"producer"`
	Consumer         string `json:"consumer"`
	MetaOperation    string `json:"meta_operation"`
	EvidencePath     string `json:"evidence_path"`
	ScenarioMutation string `json:"scenario_mutation"`
}

type Issue struct {
	Stage     string   `json:"stage"`
	Step      string   `json:"step"`
	Reason    string   `json:"reason"`
	Cause     string   `json:"cause"`
	BlockedBy []string `json:"blocked_by,omitempty"`
}

type ClaimTransition struct {
	PriorClaim          string     `json:"prior_claim"`
	NextClaim           string     `json:"next_claim"`
	ConformanceDecision string     `json:"conformance_decision"`
	SubjectResolution   string     `json:"subject_resolution"`
	Transition          string     `json:"transition"`
	Stage               string     `json:"stage"`
	Step                string     `json:"step"`
	Reason              string     `json:"reason"`
	EvidenceDigest      string     `json:"evidence_digest"`
	Provenance          Provenance `json:"provenance"`
}

type MetricResult struct {
	ID                string          `json:"id"`
	Family            string          `json:"family"`
	Claim             string          `json:"claim"`
	Numerator         int             `json:"numerator"`
	Denominator       int             `json:"denominator"`
	Decision          string          `json:"decision"`
	SubjectResolution string          `json:"subject_resolution"`
	EvaluationState   string          `json:"evaluation_state"`
	Lineage           Lineage         `json:"lineage"`
	Issue             *Issue          `json:"issue,omitempty"`
	Transition        ClaimTransition `json:"claim_transition"`
}

type GraphSummary struct {
	Nodes     int            `json:"nodes"`
	Edges     int            `json:"edges"`
	EdgeKinds map[string]int `json:"edge_kinds"`
}

type ScenarioResult struct {
	ID                  string         `json:"id"`
	Mutation            string         `json:"mutation"`
	Graph               GraphSummary   `json:"graph"`
	Numerator           int            `json:"numerator"`
	Denominator         int            `json:"denominator"`
	ConformanceDecision string         `json:"conformance_decision"`
	SubjectResolution   string         `json:"subject_resolution"`
	Decisions           map[string]int `json:"decisions"`
	Transitions         map[string]int `json:"transitions"`
	Metrics             []MetricResult `json:"metrics"`
}

type SourceReconstruction struct {
	Numerator               int `json:"numerator"`
	Denominator             int `json:"denominator"`
	MetricFieldsNumerator   int `json:"metric_fields_numerator"`
	MetricFieldsDenominator int `json:"metric_fields_denominator"`
	ScenarioNumerator       int `json:"scenario_numerator"`
	ScenarioDenominator     int `json:"scenario_denominator"`
}

type WorkspaceObservation struct {
	BeforeDigest              string   `json:"before_digest"`
	AfterDigest               string   `json:"after_digest"`
	ChangedPaths              []string `json:"changed_paths,omitempty"`
	RepositoryWorkspaceWrites bool     `json:"repository_workspace_writes"`
	MutationAuthority         bool     `json:"mutation_authority"`
}

type Receipt struct {
	Schema                  string               `json:"schema"`
	Toolchain               string               `json:"toolchain"`
	SourceDigest            string               `json:"source_digest"`
	CanonicalSemanticDigest string               `json:"canonical_semantic_digest"`
	SourceReconstruction    SourceReconstruction `json:"source_reconstruction"`
	WorkspaceObservation    WorkspaceObservation `json:"workspace_observation"`
	Scenarios               []ScenarioResult     `json:"scenarios"`
	Digest                  string               `json:"digest"`
}

type ImportCheck struct {
	Numerator   int    `json:"numerator"`
	Denominator int    `json:"denominator"`
	Status      string `json:"status"`
}

type Report struct {
	Schema                  string               `json:"schema"`
	Status                  string               `json:"status"`
	ConformanceDecision     string               `json:"conformance_decision"`
	SubjectResolution       string               `json:"subject_resolution"`
	SourceDigest            string               `json:"source_digest,omitempty"`
	CanonicalSemanticDigest string               `json:"canonical_semantic_digest,omitempty"`
	ReceiptDigest           string               `json:"receipt_digest,omitempty"`
	ScenarioCount           int                  `json:"scenario_count"`
	MetricCount             int                  `json:"metric_count"`
	FailClosedCount         int                  `json:"fail_closed_count"`
	DirectUnknowns          int                  `json:"direct_unknowns"`
	DependencyBlocks        int                  `json:"dependency_blocks"`
	TransitionCounts        map[string]int       `json:"transition_counts"`
	SourceReconstruction    SourceReconstruction `json:"source_reconstruction"`
	ProducerImport          ImportCheck          `json:"producer_import"`
	Issue                   *Issue               `json:"issue,omitempty"`
	Digest                  string               `json:"digest"`
}

type InterventionResult struct {
	ID                      string `json:"id"`
	Kind                    string `json:"kind"`
	RawSourceDigest         string `json:"raw_source_digest"`
	CanonicalSemanticDigest string `json:"canonical_semantic_digest"`
	ReceiptDigest           string `json:"receipt_digest"`
	DecisionFingerprint     string `json:"decision_fingerprint"`
	TransitionFingerprint   string `json:"transition_fingerprint"`
	SemanticDigestChanged   bool   `json:"semantic_digest_changed"`
	DecisionChanged         bool   `json:"decision_changed"`
	TransitionChanged       bool   `json:"transition_changed"`
	Status                  string `json:"status"`
}

type InterventionReport struct {
	Schema      string             `json:"schema"`
	Base        InterventionResult `json:"base"`
	Semantic    InterventionResult `json:"semantic"`
	Nonsemantic InterventionResult `json:"nonsemantic"`
	Digest      string             `json:"digest"`
}
