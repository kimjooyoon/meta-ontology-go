// Package shadow contains an independent, read-only oracle for the selective
// CI shadow decision. It deliberately does not import analyzer, planner,
// proof, lane, or command packages.
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

type plannerManifest struct {
	Schema string         `json:"schema"`
	Files  []manifestFile `json:"files"`
	Digest string         `json:"digest"`
}

type command struct {
	ID   string   `json:"id"`
	Argv []string `json:"argv"`
}

type plannerInput struct {
	Schema                  string          `json:"schema"`
	Status                  string          `json:"status"`
	RegistryDigest          string          `json:"registry_digest"`
	BaseManifest            plannerManifest `json:"base_manifest"`
	HeadManifest            plannerManifest `json:"head_manifest"`
	PlanDigest              string          `json:"plan_digest"`
	ChangedRootIDs          []string        `json:"changed_root_ids"`
	SelectedCommandIDs      []string        `json:"selected_command_ids"`
	SelectedGuardCommandIDs []string        `json:"selected_guard_command_ids"`
	SelectedWorkIDs         []string        `json:"selected_work_ids"`
	Commands                []command       `json:"commands"`
	GuardCommands           []command       `json:"guard_commands"`
}

type snapshotBinding struct {
	Source   string `json:"source"`
	Semantic string `json:"semantic"`
}

type proofInput struct {
	Schema             string         `json:"schema"`
	Status             string         `json:"status"`
	Fallback           string         `json:"fallback"`
	RegistryDigest     string         `json:"registry_digest"`
	PlanDigest         string         `json:"plan_digest"`
	Snapshots          proofSnapshots `json:"snapshots"`
	ChangedRootIDs     []string       `json:"changed_root_ids"`
	SelectedCommandIDs []string       `json:"selected_command_ids"`
	VerifiedCommandIDs []string       `json:"verified_command_ids"`
	ProofDigest        string         `json:"proof_digest"`
}

type proofSnapshots struct {
	Base snapshotBinding `json:"base"`
	Head snapshotBinding `json:"head"`
}

type laneInput struct {
	Schema            string   `json:"schema"`
	Decision          string   `json:"decision"`
	Reason            string   `json:"reason"`
	RegistryDigest    string   `json:"registry_digest"`
	BaseSHA           string   `json:"base_sha"`
	LaneHeadSHA       string   `json:"lane_head_sha"`
	LaneID            string   `json:"lane_id"`
	RegisteredBranch  string   `json:"registered_branch"`
	OwnedPathPrefixes []string `json:"owned_path_prefixes"`
	ChangedPaths      []string `json:"changed_paths"`
	AheadCount        int64    `json:"ahead_count"`
	BehindCount       int64    `json:"behind_count"`
	OpenPRCount       int64    `json:"open_pr_count"`
	ActiveLeaseCount  int64    `json:"active_lease_count"`
	CanonicalDigest   string   `json:"canonical_digest"`
}

type decodedInputs struct {
	base, head analyzerSnapshot
	planner    plannerInput
	proof      proofInput
	lane       laneInput
}

// Evaluate applies the contract's fixed precedence. It performs no process,
// filesystem, network, or argv execution side effect.
func Evaluate(c Case) Result {
	digest := caseDigest(c)
	inputs, err := decodeFiles(c.Files)
	if err != nil {
		return fallback(StageInput, err.Error(), digest)
	}
	if err := validateSnapshots(inputs.base, inputs.head, inputs.planner); err != nil {
		return fallback(StageSnapshot, err.Error(), digest)
	}
	if err := validateRegistry(inputs); err != nil {
		return fallback(StageRegistry, err.Error(), digest)
	}
	if err := validatePlan(inputs.base, inputs.head, inputs.planner); err != nil {
		return fallback(StagePlan, err.Error(), digest)
	}
	if err := validatePlanProofBinding(inputs); err != nil {
		return fallback(StagePlanProof, err.Error(), digest)
	}
	if inputs.proof.Status != "VERIFIED" || inputs.proof.Fallback != "NONE" || inputs.proof.ProofDigest != proofDigest(inputs.proof) {
		if inputs.proof.Status == "UNKNOWN" {
			return fallback(StageProofUnknown, "proof is UNKNOWN", digest)
		}
		return fallback(StageProofFail, "proof is not verified", digest)
	}
	if !validLaneFacts(inputs.lane) || inputs.lane.Schema != LaneSchema || inputs.lane.CanonicalDigest != laneDigest(inputs.lane) || inputs.lane.Decision == "UNKNOWN" {
		return fallback(StageLaneUnknown, "lane is UNKNOWN or stale", digest)
	}
	if inputs.lane.Decision == "INELIGIBLE" {
		return fallback(StageLaneIneligible, "lane is INELIGIBLE", digest)
	}
	if inputs.lane.Decision != "ELIGIBLE" || inputs.lane.Reason != "ELIGIBLE" {
		return fallback(StageLaneUnknown, "lane decision is malformed", digest)
	}

	selected, guards, work, argv := normalizedSelection(inputs.planner)
	return Result{
		Status: ShadowSelective, Stage: StageSelective, Reason: "all bindings verified",
		SelectedCommandIDs: selected, SelectedGuardIDs: guards, SelectedWorkIDs: work,
		SelectedArgv: argv, ExecutionAuthorized: true, CanonicalDigest: digest,
	}
}

func fallback(stage, reason, digest string) Result {
	return Result{Status: FullSuiteFallback, Stage: stage, Reason: reason,
		SelectedCommandIDs: []string{}, SelectedGuardIDs: []string{}, SelectedWorkIDs: []string{},
		SelectedArgv: map[string][]string{}, ExecutionAuthorized: false, CanonicalDigest: digest}
}
