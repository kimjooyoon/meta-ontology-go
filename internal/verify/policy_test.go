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
