package rollbackfixedpoint

type GuardEvidence struct {
	FileSHA256                   string `json:"file_sha256"`
	ReportDigest                 string `json:"report_digest"`
	HeadSHA                      string `json:"head_sha"`
	Decision                     string `json:"decision"`
	Reason                       string `json:"reason"`
	Resolution                   string `json:"resolution"`
	Satisfied                    int    `json:"satisfied"`
	Total                        int    `json:"total"`
	Unresolved                   int    `json:"unresolved"`
	RepositoryWrites             int    `json:"repository_writes"`
	RepositoryMutationAuthorized bool   `json:"repository_mutation_authorized"`
}

type TransformationEvidence struct {
	FileSHA256               string `json:"file_sha256"`
	LedgerDigest             string `json:"ledger_digest"`
	HeadSHA                  string `json:"head_sha"`
	Decision                 string `json:"decision"`
	Reason                   string `json:"reason"`
	WorkspaceMode            string `json:"workspace_mode"`
	WriteBoundary            string `json:"write_boundary"`
	Effects                  int    `json:"effects"`
	SourceWorkspaceUnchanged bool   `json:"source_workspace_unchanged"`
	PromotionAuthorized      bool   `json:"promotion_authorized"`
}

type Source struct {
	ExpectedHeadSHA  string                 `json:"expected_head_sha"`
	Guard            GuardEvidence          `json:"guard"`
	Transformation   TransformationEvidence `json:"transformation"`
	CollectionError  string                 `json:"collection_error,omitempty"`
	RepositoryWrites int                    `json:"repository_writes"`
}
