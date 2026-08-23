package toolchainrelease

type Summary struct {
	CasesSatisfied              int `json:"cases_satisfied"`
	CasesTotal                  int `json:"cases_total"`
	ReadinessBPS                int `json:"readiness_bps"`
	PlatformReceipts            int `json:"platform_receipts"`
	OperatingSystems            int `json:"operating_systems"`
	Architectures               int `json:"architectures"`
	BinaryBuilds                int `json:"binary_builds"`
	ArchiveBuilds               int `json:"archive_builds"`
	NativeSmokes                int `json:"native_smokes"`
	BinaryReplays               int `json:"binary_replays"`
	ArchiveReplays              int `json:"archive_replays"`
	ChecksumEntries             int `json:"checksum_entries"`
	ToolchainBindings           int `json:"toolchain_bindings"`
	VCSBindings                 int `json:"vcs_bindings"`
	ConceptBindings             int `json:"concept_bindings"`
	CodeBindings                int `json:"code_bindings"`
	MetricBindings              int `json:"metric_bindings"`
	UseCaseBindings             int `json:"use_case_bindings"`
	MissingReceipts             int `json:"missing_receipts"`
	DuplicateReceipts           int `json:"duplicate_receipts"`
	UnexpectedReceipts          int `json:"unexpected_receipts"`
	CaseFailures                int `json:"case_failures"`
	PlatformMismatches          int `json:"platform_mismatches"`
	ToolchainMismatches         int `json:"toolchain_mismatches"`
	HeadMismatches              int `json:"head_mismatches"`
	DirtyBuilds                 int `json:"dirty_builds"`
	VCSRevisionMismatches       int `json:"vcs_revision_mismatches"`
	BinaryReplayMismatches      int `json:"binary_replay_mismatches"`
	ArchiveReplayMismatches     int `json:"archive_replay_mismatches"`
	SmokeFailures               int `json:"smoke_failures"`
	ChecksumDrift               int `json:"checksum_drift"`
	ReceiptDigestFailures       int `json:"receipt_digest_failures"`
	CorpusDrift                 int `json:"corpus_drift"`
	ConceptDrift                int `json:"concept_drift"`
	ProofFailures               int `json:"proof_failures"`
	Unresolved                  int `json:"unresolved"`
	RepositoryWrites            int `json:"repository_writes"`
	MutationAuthorities         int `json:"mutation_authorities"`
}
