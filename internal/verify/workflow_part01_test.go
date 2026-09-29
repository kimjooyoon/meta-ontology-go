package verify

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCIWorkflowSeparatesPushCapsFromPullRequestChecks(t *testing.T) {
	workflow, err := os.ReadFile(filepath.Join("..", "..", ".github", "workflows", "ci.yml"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(workflow)
	if strings.Contains(text, "'agent/**'") || strings.Contains(text, "agent push cap-only") {
		t.Fatal("CI workflow retained a non-protected agent push trigger")
	}
	for _, condition := range []string{"if: ${{ github.event_name == 'pull_request' || github.event_name == 'workflow_dispatch' }}", "if: github.event_name == 'push'"} {
		if !strings.Contains(text, condition) {
			t.Fatalf("workflow lost event condition %q", condition)
		}
	}
	fullJobs := []string{"format:", "vet:", "test:", "race:", "semantic:", "policy:"}
	for _, job := range fullJobs {
		if !strings.Contains(text, "  "+job) {
			t.Fatalf("workflow lost required full job %q", job)
		}
	}
	for _, name := range []string{
		"name: gofmt",
		"name: go vet",
		"name: go test",
		"name: go test -race",
		"name: Semantic conformance",
		"name: CI policy",
	} {
		if !strings.Contains(text, name) {
			t.Fatalf("workflow lost canonical required check name %q", name)
		}
	}
	if strings.Contains(text, "name: \"gofmt [") || strings.Contains(text, "name: \"CI policy [") {
		t.Fatal("required check names were changed instead of using run metadata")
	}
	assertWorkflowMarkers(t, text)
}
func TestCIGuardianWorkflowAndCredentialAreRemoved(t *testing.T) {
	guardianPath := filepath.Join("..", "..", ".github", "workflows", "ci-guardian.yml")
	if _, err := os.Stat(guardianPath); !os.IsNotExist(err) {
		t.Fatalf("separate Guardian workflow still exists: %v", err)
	}
	workflow, err := os.ReadFile(filepath.Join("..", "..", ".github", "workflows", "ci.yml"))
	if err != nil {
		t.Fatal(err)
	}
	text := strings.ToLower(string(workflow))
	for _, forbidden := range []string{"ci guardian", "guardian_app_private_key", "guardian-observer", "pull_request_target:"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("main CI workflow retains Guardian dependency %q", forbidden)
		}
	}
}
