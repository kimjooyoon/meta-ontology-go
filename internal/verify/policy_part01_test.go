package verify

import (
	"testing"
)

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
