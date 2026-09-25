package shadow

const (
	CorpusSchema        = "gooo/selective-ci-shadow-corpus/v1"
	AnalyzerSchema      = "gooo/selective-ci-shadow-analyzer/v1"
	ManifestSchema      = "gooo/selective-ci-shadow-manifest/v1"
	PlannerSchema       = "gooo/selective-ci-shadow-planner/v1"
	ProofSchema         = "gooo/selective-ci-shadow-proof/v1"
	LaneSchema          = "gooo/selective-ci-shadow-lane/v1"
	ShadowSelective     = "SHADOW_SELECTIVE"
	FullSuiteFallback   = "FULL_SUITE_FALLBACK"
	StageInput          = "INPUT"
	StageSnapshot       = "SNAPSHOT_BINDING"
	StageRegistry       = "REGISTRY_BINDING"
	StagePlan           = "PLAN"
	StagePlanProof      = "PLAN_PROOF_BINDING"
	StageProofFail      = "PROOF_FAIL_CLOSED"
	StageProofUnknown   = "PROOF_UNKNOWN"
	StageLaneUnknown    = "LANE_UNKNOWN"
	StageLaneIneligible = "LANE_INELIGIBLE"
	StageSelective      = "SELECTIVE"
)

// Files are the five explicit JSON inputs consumed by the shadow command.
// Values remain raw strings so malformed and duplicate-key partitions cannot
// be normalized away by the corpus decoder.
type Files struct {
	AnalyzerBase string `json:"analyzer_base"`
	AnalyzerHead string `json:"analyzer_head"`
	Planner      string `json:"planner"`
	Proof        string `json:"proof"`
	Lane         string `json:"lane"`
}
type Case struct {
	Name      string `json:"name"`
	Partition string `json:"partition"`
	Files     Files  `json:"files"`
	Expected  Result `json:"expected"`
}
type Corpus struct {
	Schema string `json:"schema"`
	Cases  []Case `json:"cases"`
}

// Result is the complete observable decision vector. Fallback always carries
// empty selections and execution_authorized=false.
type Result struct {
	Status              string              `json:"status"`
	Stage               string              `json:"stage"`
	Reason              string              `json:"reason"`
	SelectedCommandIDs  []string            `json:"selected_command_ids"`
	SelectedGuardIDs    []string            `json:"selected_guard_command_ids"`
	SelectedWorkIDs     []string            `json:"selected_work_ids"`
	SelectedArgv        map[string][]string `json:"selected_argv"`
	ExecutionAuthorized bool                `json:"execution_authorized"`
	CanonicalDigest     string              `json:"canonical_digest"`
}
type analyzerSnapshot struct {
	Schema            string         `json:"schema"`
	Status            string         `json:"status"`
	FullSuiteFallback bool           `json:"full_suite_fallback"`
	RegistryDigest    string         `json:"registry_digest"`
	Files             []manifestFile `json:"files"`
	Digest            string         `json:"digest"`
}
type manifestFile struct {
	Path        string   `json:"path"`
	BlobDigest  string   `json:"blob_digest"`
	SemanticIDs []string `json:"semantic_ids"`
}
