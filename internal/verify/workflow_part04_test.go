package verify

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func assertWorkflowIdentityMarkers(t *testing.T, text string) {
	t.Helper()
	for _, forbidden := range []string{"updateBranchProtection", "replaceBranchProtection", "PUT /repos/", "repos.updateBranchProtection"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("workflow contains forbidden protection write %q", forbidden)
		}
	}
	if strings.Contains(text, "context.ref_name") {
		t.Fatal("workflow uses the unavailable github-script context.ref_name field")
	}
	for _, marker := range []string{"- dev", "normalizeOwnerBranch", "ci-base-ref.txt", "ci-owner-branch.txt"} {
		if !strings.Contains(text, marker) {
			t.Fatalf("workflow lost protected-push identity marker %q", marker)
		}
	}
	for _, forbidden := range []string{"pulls.listReviews", "approval_api_unavailable", "independent_approval_missing_or_overlapping", "approvals_status"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("workflow still treats human approval as CI proof authority: %q", forbidden)
		}
	}
	if strings.Contains(text, "\nadministration: read\n") {
		t.Fatal("workflow declares unsupported administration permission key")
	}
}
func TestCIWorkflowUsesImmutableCheckoutForEveryJob(t *testing.T) {
	workflow, err := os.ReadFile(filepath.Join("..", "..", ".github", "workflows", "ci.yml"))
	if err != nil {
		t.Fatal(err)
	}
	marker := "ref: ${{ github.event_name == 'pull_request' && github.event.pull_request.head.sha || github.sha }}"
	if strings.Count(string(workflow), marker) != 9 {
		t.Fatalf("expected nine immutable checkout refs, got %d", strings.Count(string(workflow), marker))
	}
}
func TestCISCOPE008WorkflowKeepsCanonicalJobsOnPullRequests(t *testing.T) {
	workflow, err := os.ReadFile(filepath.Join("..", "..", ".github", "workflows", "ci.yml"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(workflow)
	if !strings.Contains(text, "pull_request:\n    branches:\n      - dev\n      - main") {
		t.Fatal("pull-request trigger does not include the dev/main steady-state targets")
	}
	if strings.Contains(text, "'agent/**'") {
		t.Fatal("push trigger includes an unprotected agent branch")
	}
	if !strings.Contains(text, "  policy:\n    name: CI policy") {
		t.Fatal("pull requests do not retain the unconditionally scheduled policy job")
	}
	if strings.Count(text, "    name: gofmt\n") != 1 || strings.Count(text, "    name: go vet\n") != 1 || strings.Count(text, "    name: go test\n") != 1 || strings.Count(text, "    name: go test -race\n") != 1 || strings.Count(text, "    name: Semantic conformance\n") != 1 || strings.Count(text, "    name: CI policy\n") != 1 || strings.Count(text, "    name: Feedback predecessor\n") != 1 {
		t.Fatal("canonical job names are missing or duplicated")
	}
}
