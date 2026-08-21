package main

type ArtifactRef struct {
	Kind           string `json:"kind"`
	Schema         string `json:"schema"`
	FileSHA256     string `json:"file_sha256"`
	SemanticDigest string `json:"semantic_digest"`
	Decision       string `json:"decision"`
}

type Indicator struct {
	ID       string `json:"id"`
	Route    string `json:"route"`
	Verdict  string `json:"verdict"`
	Relation string `json:"relation"`
	Value    string `json:"value"`
	Limit    string `json:"limit"`
}

type Envelope struct {
	Schema                 string        `json:"schema"`
	Metaprogram            string        `json:"metaprogram"`
	BaseSHA                string        `json:"base_sha"`
	HeadSHA                string        `json:"head_sha"`
	CIWorkflowRunID        int64         `json:"ci_workflow_run_id"`
	CIHeadBranch           string        `json:"ci_head_branch"`
	CIConclusion           string        `json:"ci_conclusion"`
	Status                 string        `json:"status"`
	Reason                 string        `json:"reason"`
	ContractSemanticHash   string        `json:"contract_semantic_hash"`
	ContractRegistryDigest string        `json:"contract_registry_digest"`
	SourceMetrics          MetricsBinding `json:"source_metrics"`
	IndicatorLedgerDigest  string        `json:"indicator_ledger_digest"`
	IndicatorLedgerCount   int           `json:"indicator_ledger_count"`
	Artifacts              []ArtifactRef `json:"artifacts"`
	ArtifactSetDigest      string        `json:"artifact_set_digest"`
	InputDigest            string        `json:"input_digest"`
	Indicators             []Indicator   `json:"indicators"`
	PromotionAuthorized    bool          `json:"promotion_authorized"`
	EnvelopeDigest         string        `json:"envelope_digest"`
	ReplayDigest           string        `json:"replay_digest"`
}

type MetricsSnapshot struct {
	Path             string `json:"path"`
	SubjectKind      string `json:"subject_kind"`
	DirectFolders    int    `json:"direct_folders"`
	DirectFiles      int    `json:"direct_files"`
	RecursiveFolders int    `json:"recursive_folders"`
	RecursiveFiles   int    `json:"recursive_files"`
	GoFiles          int    `json:"go_files"`
	GoooFiles        int    `json:"gooo_files"`
	GoLines          int    `json:"go_lines"`
	GoooLines        int    `json:"gooo_lines"`
}

type MetricsBinding struct {
	Schema         string          `json:"schema"`
	LogicalRoot    MetricsSnapshot `json:"logical_root"`
	StorageRoot    MetricsSnapshot `json:"storage_root"`
	SemanticDigest string          `json:"semantic_digest"`
}

type planDocument struct {
	SchemaVersion       string `json:"schema_version"`
	BaseSHA             string `json:"base_sha"`
	HeadSHA             string `json:"head_sha"`
	PlanDigest          string `json:"plan_digest"`
	Decision            string `json:"decision"`
	Reason              string `json:"reason"`
	PromotionAuthorized bool   `json:"promotion_authorized"`
}
