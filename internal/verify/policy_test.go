package verify

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPolicyChecksAreDeterministic(t *testing.T) {
	if err := CheckPathScope([]string{"scripts/verify.sh", "internal/verify/policy.go"}, []string{".github", "scripts", "internal/verify"}); err != nil {
		t.Fatal(err)
	}
	if err := CheckPathScope([]string{"internal/semantic/graph.go"}, []string{"internal/verify"}); err == nil {
		t.Fatal("core package path crossed CI ownership boundary")
	}
	if err := CheckIntegrationPullRequest("agent/ci-workflow", "integration"); err != nil {
		t.Fatal(err)
	}
	if err := CheckIntegrationPullRequest("feature/work", "main"); err == nil {
		t.Fatal("invalid branch policy was accepted")
	}
}

func TestGoVersionBranchAllowsOnlyToolchainGoModChanges(t *testing.T) {
	if err := CheckPathScopeForBranch([]string{"go.mod"}, "agent/go-version"); err != nil {
		t.Fatal(err)
	}
	if err := CheckPathScopeForBranch([]string{"go.mod"}, "agent/other"); err == nil {
		t.Fatal("go.mod was allowed outside agent/go-version")
	}
	if err := CheckPathScopeForBranch([]string{"go.mod", "go.sum"}, "agent/go-version"); err == nil {
		t.Fatal("dependency metadata crossed the go-version exception")
	}
	valid := "diff --git a/go.mod b/go.mod\n@@ -3 +3 @@\n-go 1.23\n+go 1.26.5\n"
	if err := CheckGoModToolchainDiff(valid); err != nil {
		t.Fatal(err)
	}
	toolchain := "diff --git a/go.mod b/go.mod\n@@ -4 +4 @@\n-toolchain go1.26.4\n+toolchain go1.26.5\n"
	if err := CheckGoModToolchainDiff(toolchain); err != nil {
		t.Fatal(err)
	}
	invalid := "diff --git a/go.mod b/go.mod\n@@ -1 +1,2 @@\n module example\n+require example.test v1.0.0\n"
	if err := CheckGoModToolchainDiff(invalid); err == nil {
		t.Fatal("non-toolchain go.mod change was accepted")
	}
}

func TestBranchScopeAllowlist(t *testing.T) {
	cases := []struct {
		branch string
		path   string
	}{
		{"agent/syntax", "internal/syntax/parser.go"},
		{"agent/semantic", "internal/semantic/graph.go"},
		{"agent/bidir", "internal/bidir/lens.go"},
		{"agent/bidir-research", "docs/research/bidirectional.md"},
		{"agent/generator", "internal/generator/render.go"},
		{"agent/analyzer", "internal/analyzer/analyzer.go"},
		{"agent/cache", "internal/cache/cache.go"},
		{"agent/lsp", "internal/lsp/server.go"},
		{"agent/cli", "cmd/gooo/main.go"},
		{"agent/docs", "docs/spec.md"},
		{"agent/docs", "examples/conformance/main.gooo"},
		{"agent/zerolang-research", "docs/research/zerolang.md"},
		{"agent/grammar-research", "docs/research/grammar.md"},
		{"agent/grammar-review", "docs/research/grammar.md"},
		{"agent/lsp-research", "docs/research/lsp.md"},
		{"agent/security", "docs/research/security.md"},
		{"agent/security-research", "docs/research/security.md"},
		{"agent/testing-research", "docs/research/testing.md"},
		{"agent/prov-o-research", "docs/research/prov-o.md"},
		{"agent/codegen-research", "docs/research/codegen.md"},
		{"agent/query-research", "docs/research/query.md"},
		{"agent/cache-research", "docs/research/cache.md"},
		{"agent/detection", "internal/detection/scan.go"},
		{"agent/dependency-cycle-detector", "internal/detection/cycles/detect.go"},
		{"agent/detection-cycles", "internal/detection/cycles/detect.go"},
		{"agent/provenance-freshness-detector", "internal/detection/freshness/check.go"},
		{"agent/freshness-detection", "internal/detection/freshness/check.go"},
		{"agent/roundtrip-detector", "internal/detection/roundtrip/compare.go"},
		{"agent/roundtrip-detection", "internal/detection/roundtrip/compare.go"},
		{"agent/linecaps", "internal/detection/linecaps/check.go"},
		{"agent/line-cap-detector", "internal/detection/linecaps/check.go"},
		{"agent/performance", "internal/detection/performance/benchmark.go"},
		{"agent/performance-regression", "internal/detection/performance/benchmark.go"},
		{"agent/semantic-delta-detector", "internal/detection/semanticdelta/detect.go"},
		{"agent/semanticdelta", "internal/detection/semanticdelta/detect.go"},
		{"agent/prototype-detection", "internal/detection/prototype.go"},
		{"agent/fuzz-conformance", "internal/conformance/fuzz/fuzz_test.go"},
		{"agent/conformance-fuzz", "internal/conformance/fuzz/fuzz_test.go"},
		{"agent/conformance-fuzz", "internal/syntax/parser.go"},
		{"agent/protected-regions", "internal/detection/protectedregions/markers.go"},
		{"agent/formatter", "internal/formatter/format.go"},
		{"agent/prototype-formatter", "internal/formatter/prototype.go"},
		{"agent/query-engine", "internal/query/engine.go"},
		{"agent/prototype-query", "internal/query/prototype.go"},
		{"agent/provenance-store", "internal/provenance/store.go"},
		{"agent/prototype-provenance", "internal/provenance/prototype.go"},
		{"agent/self-hosting-bootstrap", "docs/research/self-hosting.md"},
		{"agent/self-hosting-bootstrap", "internal/bootstrap/bootstrap.go"},
		{"agent/ci-workflow", ".github/workflows/ci.yml"},
		{"agent/ci-evidence-contract", "internal/verify/evidence.go"},
	}
	for _, test := range cases {
		if err := CheckPathScopeForBranch([]string{test.path}, test.branch); err != nil {
			t.Errorf("%s should allow %s: %v", test.branch, test.path, err)
		}
	}
	if err := CheckPathScopeForBranch([]string{".github/workflows/ci.yml"}, "agent/syntax"); err == nil {
		t.Fatal("shared CI file crossed syntax ownership")
	}
}

func TestBranchScopeBoundaries(t *testing.T) {
	boundaries := []struct {
		branch string
		path   string
	}{
		{"agent/zerolang-research", "docs/research/grammar.md"},
		{"agent/bidir-research", "docs/research/prov-o.md"},
		{"agent/lsp-research", "docs/research/security.md"},
		{"agent/semantic-delta-detector", "internal/detection/freshness/check.go"},
		{"agent/provenance-freshness-detector", "internal/detection/roundtrip/compare.go"},
		{"agent/roundtrip-detector", "internal/detection/cycles/detect.go"},
		{"agent/dependency-cycle-detector", "internal/detection/linecaps/check.go"},
		{"agent/line-cap-detector", "internal/detection/performance/benchmark.go"},
		{"agent/linecaps", "internal/detection/performance/benchmark.go"},
		{"agent/fuzz-conformance", "internal/conformance/markers.go"},
		{"agent/provenance-store", "internal/query/engine.go"},
		{"agent/performance-regression", "internal/detection/linecaps/check.go"},
		{"agent/performance", "internal/detection/linecaps/check.go"},
		{"agent/protected-regions", "internal/conformance/markers.go"},
		{"agent/semanticdelta", "internal/provenance/store.go"},
		{"agent/self-hosting-bootstrap", "docs/research/other.md"},
		{"agent/self-hosting-bootstrap", "internal/semantic/graph.go"},
	}
	for _, boundary := range boundaries {
		if err := CheckPathScopeForBranch([]string{boundary.path}, boundary.branch); err == nil {
			t.Errorf("%s incorrectly allowed %s", boundary.branch, boundary.path)
		}
	}
	for _, branch := range []string{"agent/unknown", "agent/unknown-research", "agent/prototype-unknown"} {
		if err := CheckPathScopeForBranch([]string{"internal/syntax/parser.go"}, branch); err == nil {
			t.Errorf("unknown agent branch %s was not rejected", branch)
		}
	}
}

func TestGoCapsRejectOversizedFileAndFunction(t *testing.T) {
	root := t.TempDir()
	source := "package fixture\n\nfunc TooLong() {\n" + strings.Repeat("\t_ = 1\n", 76) + "}\n"
	path := "fixture.go"
	if err := os.WriteFile(filepath.Join(root, path), []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := CheckGoCaps(root, []string{path}, 300, 75); err == nil {
		t.Fatal("oversized function was accepted")
	}
}

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
		{"agent/ci-scope-triage", "internal/verify/scope.go"},
		{"agent/cli-bootstrap-contract", "cmd/gooo/evidence_adapter.go"},
		{"agent/cli-check", "cmd/gooo/main.go"},
		{"agent/codegen-followup", "docs/research/codegen-reproducibility.md"},
		{"agent/codegen-hypotheses", "docs/research/codegen-experiments.md"},
		{"agent/codegen-fixture-adapter", "docs/research/codegen-fixture-adapter.md"},
		{"agent/generator-fixtures-current", "internal/generator/fixture_contract_test.go"},
		{"agent/generator-fixtures-current2", "internal/generator/fixture_contract_test.go"},
		{"agent/generator-fixtures-current5", "internal/generator/fixture_followup_test.go"},
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

func TestFollowUpScopeBoundariesAndUnknowns(t *testing.T) {
	cases := []struct {
		branch string
		path   string
	}{
		{"agent/codegen-hypotheses", "docs/research/codegen.md"},
		{"agent/codegen-fixture-adapter", "docs/research/codegen-experiments.md"},
		{"agent/bidir-followup", "internal/semantic/graph.go"},
		{"agent/freshness-research", "internal/research/other/contract.go"},
		{"agent/integration-governance", "docs/research/integration-promotion.md"},
		{"agent/integration-governance-followup", "docs/governance/other.md"},
		{"agent/lsp-contracts", "docs/research/grammar.md"},
		{"agent/testing-research-contracts", "docs/research/security.md"},
		{"agent/ci-ownership-audit", "internal/semantic/graph.go"},
		{"agent/ci-ownership-audit-current", "internal/semantic/graph.go"},
		{"agent/ci-ownership-audit-current2", "internal/semantic/graph.go"},
		{"agent/ci-alias-refresh", "internal/semantic/graph.go"},
		{"agent/cli-check", "internal/semantic/graph.go"},
		{"agent/generator-fixtures-current", "internal/semantic/graph.go"},
		{"agent/generator-fixtures-current2", "internal/semantic/graph.go"},
		{"agent/generator-fixtures-current5", "internal/semantic/graph.go"},
	}
	for _, test := range cases {
		if err := CheckPathScopeForBranch([]string{test.path}, test.branch); err == nil {
			t.Errorf("%s incorrectly allowed %s", test.branch, test.path)
		}
	}
	for _, branch := range []string{"agent/unknown-followup", "agent/freshness-unknown"} {
		if err := CheckPathScopeForBranch([]string{"docs/research/unknown.md"}, branch); err == nil {
			t.Errorf("unknown branch %s was not rejected", branch)
		}
	}
}

func TestBidirFollowUpAliasIsExact(t *testing.T) {
	paths, ok := BranchScope("agent/bidir-followup")
	if !ok || len(paths) != 1 || paths[0] != "internal/bidir" {
		t.Fatalf("unexpected bidir-followup ownership: %#v", paths)
	}
	if err := CheckPathScopeForBranch([]string{"internal/bidir/hosting_contract.go"}, "agent/bidir-followup"); err != nil {
		t.Fatal(err)
	}
}

func TestAllowlistKeysAndSelfHostingPathsAreUnique(t *testing.T) {
	branches := ConfiguredBranches()
	if len(branches) != len(sortedUnique(branches)) {
		t.Fatalf("duplicate branch keys detected: %#v", branches)
	}
	for _, branch := range branches {
		if strings.ContainsAny(branch, "*?") {
			t.Fatalf("wildcard branch key weakens fail-closed policy: %q", branch)
		}
	}
	paths, ok := BranchScope("agent/self-hosting-bootstrap")
	if !ok || len(paths) != len(sortedUnique(paths)) {
		t.Fatalf("duplicate self-hosting alias paths detected: %#v", paths)
	}
}

func TestScopeTableMatchesAllowlist(t *testing.T) {
	table, err := os.ReadFile(filepath.Join("..", "..", ".github", "agent-scope-table.md"))
	if err != nil {
		t.Fatal(err)
	}
	registered := make(map[string]bool)
	for _, line := range strings.Split(string(table), "\n") {
		cells := strings.Split(line, "|")
		if len(cells) < 3 {
			continue
		}
		branch := strings.Trim(strings.TrimSpace(cells[1]), "`")
		if !strings.HasPrefix(branch, "agent/") {
			continue
		}
		if registered[branch] {
			t.Fatalf("duplicate branch row in scope table: %q", branch)
		}
		registered[branch] = true
		if _, ok := BranchScope(branch); !ok {
			t.Fatalf("scope table contains unconfigured branch: %q", branch)
		}
	}
	for _, branch := range ConfiguredBranches() {
		if !registered[branch] {
			t.Errorf("scope table is missing configured branch: %q", branch)
		}
	}
}
