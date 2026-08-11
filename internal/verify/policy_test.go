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
	prefixes := []string{".github", "scripts", "internal/verify"}
	if err := CheckPathScopeForBranch([]string{"go.mod"}, prefixes, "agent/go-version"); err != nil {
		t.Fatal(err)
	}
	if err := CheckPathScopeForBranch([]string{"go.mod"}, prefixes, "agent/other"); err == nil {
		t.Fatal("go.mod was allowed outside agent/go-version")
	}
	if err := CheckPathScopeForBranch([]string{"go.mod", "go.sum"}, prefixes, "agent/go-version"); err == nil {
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
