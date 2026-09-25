package ciplanusecase

type Summary struct {
	CasesSatisfied       int   `json:"cases_satisfied"`
	PassDecisions        int   `json:"pass_decisions"`
	FailClosedDecisions  int   `json:"fail_closed_decisions"`
	UnknownDecisions     int   `json:"unknown_decisions"`
	DeterministicReplays int   `json:"deterministic_replays"`
	GoldenPlans          int   `json:"golden_plans"`
	RuleEvidenceRefs     int   `json:"rule_evidence_refs"`
	DirectUnknownClaims  int   `json:"direct_unknown_claims"`
	DependencyBlocked    int   `json:"dependency_blocked_claims"`
	RefutedClaims        int   `json:"refuted_claims"`
	PersistentClaims     int   `json:"persistent_claims"`
	GeneratedReplays     int   `json:"generated_source_replays"`
	ResourceSamples      int   `json:"resource_samples"`
	GoooFiles            int   `json:"gooo_files"`
	GoFiles              int   `json:"go_files"`
	GoooLines            int   `json:"gooo_lines"`
	GoLines              int   `json:"go_lines"`
	MaxWallMS            int64 `json:"max_wall_ms"`
	MaxPeakRSSKiB        int64 `json:"max_peak_rss_kib"`
	MaxReceiptBytes      int64 `json:"max_receipt_bytes"`
	RepositoryWrites     int   `json:"repository_writes"`
	MutationAuthority    int   `json:"mutation_authority"`
}
