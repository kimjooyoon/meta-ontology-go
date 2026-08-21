package shadow

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
