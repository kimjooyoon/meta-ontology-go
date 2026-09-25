package lanefrontier

const SchemaVersion = "gooo/selective-ci-lane-frontier/v1"

type Decision string

const (
	DecisionEligible   Decision = "ELIGIBLE"
	DecisionIneligible Decision = "INELIGIBLE"
	DecisionUnknown    Decision = "UNKNOWN"
	ELIGIBLE                    = DecisionEligible
	INELIGIBLE                  = DecisionIneligible
	UNKNOWN                     = DecisionUnknown
)

type Reason string

const (
	ReasonUnknownSchema  Reason = "UNKNOWN_SCHEMA"
	ReasonMissingInput   Reason = "MISSING_INPUT"
	ReasonInvalidCount   Reason = "INVALID_COUNT"
	ReasonAmbiguousOwner Reason = "AMBIGUOUS_OWNER"
	ReasonPathOutOfScope Reason = "PATH_OUT_OF_SCOPE"
	ReasonActiveLease    Reason = "ACTIVE_LEASE"
	ReasonActivePR       Reason = "ACTIVE_PR"
	ReasonDivergedBranch Reason = "DIVERGED_BRANCH"
	ReasonStaleBranch    Reason = "STALE_BRANCH"
	ReasonBranchAhead    Reason = "BRANCH_AHEAD"
	ReasonEligible       Reason = "ELIGIBLE"
)

type Input struct {
	SchemaVersion     string   `json:"schema_version"`
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
}

type Output struct {
	SchemaVersion     string   `json:"schema_version"`
	Decision          Decision `json:"decision"`
	Reason            Reason   `json:"reason"`
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
