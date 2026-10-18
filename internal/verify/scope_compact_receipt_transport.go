package verify

func init() {
	branchScopeAllowlist["agent/compact-receipt-transport-20261018"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"cmd/gooo/package_construct.go",
		"cmd/gooo/package_construction_input.go",
		"cmd/gooo/package_execute.go",
		"cmd/gooo/package_execution_failure_test.go",
		"cmd/gooo/package_origins.go",
		"cmd/gooo/package_receipt_transport.go",
		"cmd/gooo/package_receipt_transport_test.go",
		"cmd/gooo/package_replay.go",
		"cmd/gooo/usage_package.go",
		"cmd/gooo/version_build.go",
		"docs/language/recorded-origin-queries.ko.md",
		"internal/verify/scope_compact_receipt_transport.go",
	}
}
