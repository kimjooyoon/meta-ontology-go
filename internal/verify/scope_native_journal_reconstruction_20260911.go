package verify

func init() {
	branchScopeAllowlist["agent/native-journal-reconstruction-20260911"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"internal/verify/scope_native_journal_reconstruction_20260911.go",
		"scripts/meta-cost-report/native_journal_fixture_test.go",
		"scripts/meta-cost-report/native_journal_rejection_test.go",
		"scripts/meta-cost-report/native_journal_test.go",
		"scripts/meta-cost-report/testdata/native-interruption-34510267647/README.md",
		"scripts/meta-cost-report/testdata/native-interruption-34510267647/capture.json",
		"scripts/meta-cost-report/testdata/native-interruption-34510267647/first.ndjson",
		"scripts/meta-cost-report/testdata/native-interruption-34510267647/replay.ndjson",
	}
}
