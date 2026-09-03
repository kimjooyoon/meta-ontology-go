package verify

func init() {
	protocolPaths := []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/main-history-reconciliation-v2.json",
		".github/workflows/ci-guardian.yml",
		".github/workflows/ci.yml",
		".github/workflows/compiler-self-improvement.yml",
		"internal/verify/scope_main_history_reconciliation_protocol_20260903.go",
		"scripts/ci-proof/guardian.js",
		"scripts/ci-proof/guardian_evidence_part01.go",
		"scripts/ci-proof/guardian_evidence_types.go",
		"scripts/ci-proof/build_part01.go",
		"scripts/ci-proof/linear_tree_reconciliation.js",
		"scripts/ci-proof/linear_tree_reconciliation_test.js",
		"scripts/ci-proof/promotion_part01.go",
		"scripts/ci-proof/promotion_part02.go",
		"scripts/ci-proof/proof_types_part01.go",
		"scripts/ci-proof/route.go",
		"scripts/ci-proof/route.js",
		"scripts/ci-proof/route_test.js",
		"scripts/ci-proof/types_part02.go",
	}
	branchScopeAllowlist["agent/main-history-reconciliation-protocol-20260903"] = append([]string(nil), protocolPaths...)
	branchScopeAllowlist["agent/main-history-reconciliation-promotion-20260903"] = []string{
		".github",
		"README.md",
		"docs",
		"examples",
		"internal",
		"scripts",
		"cmd",
		"go.mod",
		"go.sum",
		"AGENTS.md",
		"CONTRIBUTING.md",
	}
	branchScopeAllowlist["agent/main-history-reconciliation-owner-authorization-20260903"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/main-history-reconciliation-v2.json",
		"internal/verify/scope_main_history_reconciliation_protocol_20260903.go",
		"scripts/ci-proof/linear_tree_reconciliation.js",
		"scripts/ci-proof/linear_tree_reconciliation_test.js",
	}
}
