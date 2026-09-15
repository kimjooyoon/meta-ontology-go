package verify

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
func TestCheckSourcePolicyRejectsOversizedGoAndGooo(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "oversized.go"), []byte("package fixture\n"+strings.Repeat("// line\n", 80)), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "oversized.gooo"), []byte("intent: example\n"+strings.Repeat("detail\n", 80)), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := CheckSourcePolicy(root, nil, DefaultLinePolicy()); err == nil {
		t.Fatal("oversized go/gooo files were accepted")
	}
}
