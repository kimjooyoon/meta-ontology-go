package verify

import (
	"testing"
)

func TestFollowUpBranchScopeAliases(t *testing.T) {
	cases := []struct {
		branch string
		path   string
	}{
		{"agent/analyzer-contract", "internal/analyzer/hosting.go"},
		{"agent/bidir-followup", "internal/bidir/hosting_contract.go"},
		{"agent/bidirectional-experiment-contract", "docs/research/bidirectional.md"},
		{"agent/bidirectional-property-matrix", "docs/research/bidirectional.md"},
		{"agent/cache-experiment-followup", "docs/research/cache.md"},
		{"agent/ci-ownership-audit", ".github/agent-scope-table.md"},
		{"agent/ci-ownership-audit-current", "internal/verify/scope.go"},
		{"agent/ci-ownership-audit-current2", "internal/verify/scope.go"},
		{"agent/ci-alias-refresh", ".github/agent-scope-table.md"},
		{"agent/ci-generator-current7", "internal/verify/scope.go"},
		{"agent/ci-scope-triage", "internal/verify/scope.go"},
		{"agent/cli-bootstrap-contract", "cmd/gooo/evidence_adapter.go"},
		{"agent/cli-check", "cmd/gooo/main.go"},
		{"agent/codegen-followup", "docs/research/codegen-reproducibility.md"},
		{"agent/codegen-hypotheses", "docs/research/codegen-experiments.md"},
		{"agent/codegen-fixture-adapter", "docs/research/codegen-fixture-adapter.md"},
		{"agent/generator-fixtures-current", "internal/generator/fixture_contract_test.go"},
		{"agent/generator-fixtures-current2", "internal/generator/fixture_contract_test.go"},
		{"agent/generator-fixtures-current5", "internal/generator/fixture_followup_test.go"},
		{"agent/generator-fixtures-current7", "internal/generator/fixture_deep_followup_test.go"},
		{"agent/freshness-research", "internal/research/freshness/contract.go"},
		{"agent/grammar-followup", "docs/research/grammar.md"},
		{"agent/integration-governance", "docs/governance/integration-promotion.md"},
		{"agent/integration-governance-followup", "docs/governance/integration-promotion.md"},
		{"agent/lsp-contracts", "docs/research/lsp.md"},
		{"agent/lsp-experiments", "docs/research/lsp.md"},
		{"agent/testing-research-contracts", "docs/research/testing.md"},
		{"agent/testing-research-followup", "docs/research/testing.md"},
		{"agent/zerolang-experiments", "docs/research/zerolang.md"},
	}
	for _, test := range cases {
		if err := CheckPathScopeForBranch([]string{test.path}, test.branch); err != nil {
			t.Errorf("%s should allow %s: %v", test.branch, test.path, err)
		}
	}
}
