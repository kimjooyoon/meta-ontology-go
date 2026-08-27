package provider

type Input struct {
	Repository      string
	HeadSHA         string
	SourcePath      string
	Source          []byte
	BeforeTracked   []byte
	BeforeUntracked []byte
	BeforeStatus    []byte
	AfterTracked    []byte
	AfterUntracked  []byte
	AfterStatus     []byte
}

type RecipeOperand struct {
	Operation           string `json:"operation"`
	Required            string `json:"required"`
	ObservationRecipe   string `json:"observation_recipe"`
	DependencyRecipe    string `json:"dependency_recipe"`
	InvariantCapability string `json:"invariant_capability"`
}

type Recipe struct {
	ID               string        `json:"id"`
	SourceActivity   string        `json:"source_activity"`
	SourceActivityID string        `json:"source_activity_id"`
	Producer         string        `json:"producer"`
	Consumer         string        `json:"consumer"`
	MetaOperation    string        `json:"meta_operation"`
	ProofChoice      string        `json:"proof_choice"`
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
	ID               string   `json:"id"`
	SourceActivity   string   `json:"source_activity"`
	SourceActivityID string   `json:"source_activity_id"`
	Producer         string   `json:"producer"`
	Consumer         string   `json:"consumer"`
	MetaOperation    string   `json:"meta_operation"`
	ProofChoice      string   `json:"proof_choice"`
	Left             Evidence `json:"left"`
	Right            Evidence `json:"right"`
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
