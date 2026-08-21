package lanefrontier

type semanticVector struct {
	SchemaVersion     string   `json:"schema_version"`
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
}
type pairedReceipt struct {
	CaseID                     string         `json:"case_id"`
	SchemaIdentical            bool           `json:"schema_identical"`
	OracleCanonicalDigest      string         `json:"oracle_canonical_digest"`
	ProductionDigest           string         `json:"production_canonical_digest"`
	OracleVector               semanticVector `json:"oracle_vector"`
	ProductionVector           semanticVector `json:"production_vector"`
	OraclePermutationEqual     bool           `json:"oracle_permutation_equal"`
	ProductionPermutationEqual bool           `json:"production_permutation_equal"`
}
