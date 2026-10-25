package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestCollapseFailureReleasesItsReadOnlyGitSnapshot(t *testing.T) {
	root, temporary := t.TempDir(), t.TempDir()
	for _, key := range []string{"TMPDIR", "TMP", "TEMP"} {
		t.Setenv(key, temporary)
	}
	// The old consumer leaks a sealed snapshot; also clean that counterexample.
	t.Cleanup(func() { cleanupGitSnapshot(temporary) })
	registrationTestCommand(t, root, "git", "init", "--quiet")
	registrationTestCommand(t, root, "git", "-c", "user.name=Gooo test", "-c", "user.email=test@example.invalid",
		"-c", "commit.gpgsign=false", "commit", "--allow-empty", "--quiet", "-m", "fixture")
	head := strings.TrimSpace(string(registrationTestCommand(t, root, "git", "rev-parse", "HEAD")))
	// CI's Git context belongs to the caller, not this newly created repository.
	t.Setenv("GIT_DIR", filepath.Join(temporary, "caller.git"))
	gitDir := strings.TrimSpace(string(registrationTestCommand(t, root, "git", "rev-parse", "--absolute-git-dir")))
	plan := generation.Plan{HeadSHA: head}
	subject := sourcepolicy.SourceSubject{Path: "missing.go", Line: 1, Name: "Missing"}
	_, failure := materializeCollapse(root, gitDir, "", plan, generation.Action{}, subject, metaExecutionTrace{}, "first")
	if failure == nil || failure.reason != "SOURCE_UNAVAILABLE" {
		t.Fatal("did not exercise post-snapshot failure", failure)
	}
	entries, err := os.ReadDir(temporary)
	if err != nil || len(entries) != 0 {
		t.Fatal("completed collapse retained disposable files", entries, err)
	}
	if _, err := os.Stat(filepath.Join(root, ".git", "HEAD")); err != nil {
		t.Fatal("source repository changed", err)
	}
}
