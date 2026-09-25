package externalconformanceactivation

type mergeProof struct {
	Schema         string `json:"schema"`
	Repository     string `json:"repository"`
	PullRequest    int    `json:"pull_request"`
	State          string `json:"state"`
	BaseSHA        string `json:"base_sha"`
	HeadSHA        string `json:"head_sha"`
	MergeCommitSHA string `json:"merge_commit_sha"`
	MergedAt       string `json:"merged_at"`
}
