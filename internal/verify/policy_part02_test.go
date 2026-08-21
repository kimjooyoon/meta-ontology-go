package verify

import (
	"testing"
)

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
