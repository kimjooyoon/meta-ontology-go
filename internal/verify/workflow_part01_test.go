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
	for _, condition := range []string{"if: github.event_name == 'pull_request'", "if: github.event_name == 'push'"} {
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
func TestCIGuardianIsBasePinnedAndReadOnly(t *testing.T) {
	workflow, err := os.ReadFile(filepath.Join("..", "..", ".github", "workflows", "ci-guardian.yml"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(workflow)
	for _, marker := range []string{
		"name: CI guardian", "pull_request_target:", "- dev\n      - main",
		"environment: ${{ github.base_ref == 'main' && 'guardian-observer'", "actions: read", "actions/create-github-app-token@bcd2ba49218906704ab6c1aa796996da409d3eb1", "permission-administration: read", "GUARDIAN_APP_PRIVATE_KEY", "getBranchProtection", "observer_environment",
		"actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1", "ref: ${{ github.workflow_sha }}",
		"persist-credentials: false", "actions/github-script@3a2844b7e9c422d3c10d287c895573f7108da1b3", "listFiles",
		"github.rest.pulls.get", "github.workflow_ref", "github.workflow_sha", "github.sha",
		"actions/upload-artifact@ea165f8d65b6e75b540449e92b4886f43607fa02", "head_binding_status", "ci-guardian.json",
		"contents: read", "pull-requests: read",
	} {
		if !strings.Contains(text, marker) {
			t.Fatalf("guardian workflow lost marker %q", marker)
		}
	}
	for _, forbidden := range []string{
		"github.event.pull_request.head.sha", "refs/pull/", "BRANCH_PROTECTION_TOKEN", "contents: write",
		"pull-requests: write", "agent/ci-workflow", "ref: ${{ github.event.pull_request.base.sha }}", "\n        run:", "\n    pull_request:",
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("guardian workflow contains unsafe marker %q", forbidden)
		}
	}
}
