package main

type promotionTopology struct {
	Status       string `json:"status"`
	AheadBy      int    `json:"ahead_by"`
	BehindBy     int    `json:"behind_by"`
	MergeBaseSHA string `json:"merge_base_sha"`
}

type promotionObservation struct {
	Repository     string            `json:"repository"`
	PRNumber       int64             `json:"pr_number"`
	Action         string            `json:"action"`
	Mode           string            `json:"promotion_mode"`
	State          string            `json:"state"`
	Draft          bool              `json:"draft"`
	Merged         bool              `json:"merged"`
	Mergeable      bool              `json:"mergeable"`
	MergeableState string            `json:"mergeable_state"`
	BaseRepo       string            `json:"base_repo"`
	BaseRef        string            `json:"base_ref"`
	BaseSHA        string            `json:"base_sha"`
	HeadRepo       string            `json:"head_repo"`
	HeadRef        string            `json:"head_ref"`
	HeadSHA        string            `json:"head_sha"`
	HeadParentSHA  string            `json:"head_parent_sha"`
	HeadTreeSHA    string            `json:"head_tree_sha"`
	LiveDevSHA     string            `json:"live_dev_sha"`
	LiveDevTreeSHA string            `json:"live_dev_tree_sha"`
	LiveMainSHA    string            `json:"live_main_sha"`
	Topology       promotionTopology `json:"topology"`
}

type promotionAuthorization struct {
	Decision    string  `json:"decision"`
	Code        *string `json:"code"`
	Operation   string  `json:"operation"`
	Source      string  `json:"source"`
	Target      string  `json:"target"`
	BaseSHA     string  `json:"base_sha"`
	HeadSHA     string  `json:"head_sha"`
	ProofDigest string  `json:"proof_digest"`
}
