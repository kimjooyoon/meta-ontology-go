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
	guard := "if: github.event_name != 'push' || !startsWith(github.ref, 'refs/heads/agent/')"
	if strings.Count(text, guard) != 5 {
		t.Fatalf("expected five full-check push guards, got %d", strings.Count(text, guard))
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
	for _, marker := range []string{
		"run-name: \"CI [${{",
		"types: [opened, synchronize, reopened, ready_for_review]",
		"startsWith(github.ref, 'refs/heads/agent/') && 'agent push cap-only'",
		"|| 'push full'",
		"Record CI event source",
		"source=\"PR authoritative\"",
		"source=\"agent push cap-only\"",
		"source=\"push full\"",
		"GITHUB_STEP_SUMMARY",
		"ref: ${{ github.event_name == 'pull_request' && github.event.pull_request.head.sha || github.sha }}",
		"Verify PR checkout identity",
		"EXPECTED_HEAD: ${{ github.event.pull_request.head.sha }}",
		"EXPECTED_BASE: ${{ github.event.pull_request.base.sha }}",
		"actual_head=\"$(git rev-parse HEAD)\"",
		"git rev-parse --verify \"$EXPECTED_BASE^{commit}\" >/dev/null",
		"GOOO_SCOPE_FROM: ${{ github.event_name == 'pull_request' && github.event.pull_request.base.sha || github.event.before }}",
		"GOOO_SCOPE_TO: ${{ github.event.pull_request.head.sha }}",
		"GOOO_EXPECTED_HEAD: ${{ github.event.pull_request.head.sha }}",
		"needs: [format, vet, test, race, semantic]",
		"if: ${{ always() }}",
		"actions/github-script@v7",
		"listJobsForWorkflowRun",
		"ci-jobs.json",
		"ci-evidence.json",
		"CI_SLOT_PRESERVATION: \"true\"",
		"CI_NO_WRITE_OUTSIDE_GENERATED: \"true\"",
		"actions/upload-artifact@v4",
		"if-no-files-found: error",
	} {
		if !strings.Contains(text, marker) {
			t.Fatalf("workflow lost event-source evidence marker %q", marker)
		}
	}
}

func TestCIWorkflowKeepsCanonicalJobsOnPullRequests(t *testing.T) {
	workflow, err := os.ReadFile(filepath.Join("..", "..", ".github", "workflows", "ci.yml"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(workflow)
	guard := "if: github.event_name != 'push' || !startsWith(github.ref, 'refs/heads/agent/')"
	if strings.Count(text, guard) != 5 {
		t.Fatalf("pull requests do not retain five guarded canonical jobs: %d", strings.Count(text, guard))
	}
	if !strings.Contains(text, "  policy:\n    name: CI policy") {
		t.Fatal("pull requests do not retain the unconditionally scheduled policy job")
	}
	if strings.Count(text, "    name: gofmt\n") != 1 || strings.Count(text, "    name: go vet\n") != 1 || strings.Count(text, "    name: go test\n") != 1 || strings.Count(text, "    name: go test -race\n") != 1 || strings.Count(text, "    name: Semantic conformance\n") != 1 || strings.Count(text, "    name: CI policy\n") != 1 {
		t.Fatal("canonical job names are missing or duplicated")
	}
}
