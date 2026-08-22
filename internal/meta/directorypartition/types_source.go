package directorypartition

type SourceMetrics struct {
	Repository  string            `json:"repository"`
	CommitSHA   string            `json:"commit_sha"`
	Files       []SourceFile      `json:"files"`
	Directories []SourceDirectory `json:"directories"`
	Meta        SourceMeta        `json:"meta"`
}

type SourceFile struct {
	Path     string `json:"path"`
	Language string `json:"language"`
	Lines    int    `json:"lines"`
}

type SourceDirectory struct {
	Path             string `json:"path"`
	SubjectKind      string `json:"subject_kind"`
	DirectFolders    int    `json:"direct_folders"`
	DirectFiles      int    `json:"direct_files"`
	RecursiveFolders int    `json:"recursive_folders"`
	RecursiveFiles   int    `json:"recursive_files"`
}

type SourceMeta struct {
	Schema     string            `json:"schema"`
	Policy     SourcePolicy      `json:"policy"`
	Indicators []SourceIndicator `json:"indicators"`
}

type SourcePolicy struct {
	Schema                    string `json:"schema"`
	MaxDirectDirectoryEntries int    `json:"max_direct_directory_entries"`
	ExemptProjectRootTopology bool   `json:"exempt_project_root_topology"`
	ExemptProjectRootREADME   bool   `json:"exempt_project_root_readme"`
}

type SourceIndicator struct {
	MetricID           string `json:"metric_id"`
	Subject            string `json:"subject"`
	SubjectKind        string `json:"subject_kind"`
	Value              int    `json:"value"`
	Limit              int    `json:"limit"`
	Applicability      string `json:"applicability"`
	ApplicabilityReason string `json:"applicability_reason"`
	Blocking           bool   `json:"blocking"`
	Satisfied          bool   `json:"satisfied"`
	ProofChoice        string `json:"proof_choice"`
	MetaOperation      string `json:"meta_operation"`
}
