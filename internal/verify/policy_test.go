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
