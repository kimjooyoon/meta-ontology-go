package lanefrontier

const SchemaV1 = "gooo/lane-frontier/v1"

type Decision string
type Reason string

const (
	Unknown    Decision = "UNKNOWN"
	Ineligible Decision = "INELIGIBLE"
	Eligible   Decision = "ELIGIBLE"
)
const (
	UnknownSchema  Reason = "UNKNOWN_SCHEMA"
	MissingInput   Reason = "MISSING_INPUT"
	InvalidCount   Reason = "INVALID_COUNT"
	AmbiguousOwner Reason = "AMBIGUOUS_OWNER"
	PathOutOfScope Reason = "PATH_OUT_OF_SCOPE"
	ActiveLease    Reason = "ACTIVE_LEASE"
	ActivePR       Reason = "ACTIVE_PR"
	DivergedBranch Reason = "DIVERGED_BRANCH"
	StaleBranch    Reason = "STALE_BRANCH"
	BranchAhead    Reason = "BRANCH_AHEAD"
	Clean          Reason = "CLEAN"
)

type Input struct {
	Schema            string   `json:"schema"`
	RegistryDigest    string   `json:"registry_digest"`
	BaseSHA           string   `json:"base_sha"`
	LaneHeadSHA       string   `json:"lane_head_sha"`
	LaneStableID      string   `json:"lane_stable_id"`
	RegisteredBranch  string   `json:"registered_branch"`
	OwnedPathPrefixes []string `json:"owned_path_prefixes"`
	ChangedPaths      []string `json:"changed_paths"`
	AheadCount        int64    `json:"ahead_count"`
	BehindCount       int64    `json:"behind_count"`
	OpenPRCount       int64    `json:"open_pr_count"`
	ActiveLeaseCount  int64    `json:"active_lease_count"`
}
type Result struct {
	Decision        Decision `json:"decision"`
	Reason          Reason   `json:"reason"`
	CanonicalDigest string   `json:"canonical_digest"`
}
type Case struct {
	Name            string `json:"name"`
	Input           Input  `json:"input"`
	Expected        Result `json:"expected"`
	CanonicalDigest string `json:"canonical_digest"`
}
type Corpus struct {
	Schema string `json:"schema"`
	Cases  []Case `json:"cases"`
}

func Evaluate(input Input) Result {
	decision, reason := decide(input)
	result := Result{Decision: decision, Reason: reason,
		CanonicalDigest: digestResult(input, decision, reason)}
	return result
}
