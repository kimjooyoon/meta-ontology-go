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
		"actions/checkout@11d5960a326750d5838078e36cf38b85af677262", "ref: ${{ github.workflow_sha }}",
		"persist-credentials: false", "actions/github-script@f28e40c7f34bde8b3046d885e986cb6290c5673b", "listFiles",
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

func TestCIGuardianAppTokenUsesClientIDAndOneMintSecret(t *testing.T) {
	workflow, err := os.ReadFile(filepath.Join("..", "..", ".github", "workflows", "ci-guardian.yml"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(workflow)
	if strings.Contains(text, "app-id:") || !strings.Contains(text, "client-id: ${{ env.GUARDIAN_APP_CLIENT_ID }}") {
		t.Fatal("Guardian App token action is not using the pinned client-id input")
	}
	mintStart := strings.Index(text, "id: guardian-app-token")
	if mintStart < 0 {
		t.Fatal("Guardian App mint step is missing")
	}
	mintEnd := strings.Index(text[mintStart:], "- name: Inspect changed paths from default authority")
	if mintEnd < 0 {
		t.Fatal("Guardian App mint step has no bounded end")
	}
	mintStep := text[mintStart : mintStart+mintEnd]
	if strings.Count(text, "${{ secrets.") != 1 || strings.Count(mintStep, "${{ secrets.") != 1 || !strings.Contains(mintStep, "GUARDIAN_APP_PRIVATE_KEY: ${{ secrets.GUARDIAN_APP_PRIVATE_KEY }}") {
		t.Fatal("Guardian mint step does not use exactly GUARDIAN_APP_PRIVATE_KEY")
	}
}

func assertWorkflowMarkers(t *testing.T, text string) {
	t.Helper()
	for _, marker := range []string{
		"run-name: \"CI [${{ github.event_name == 'pull_request' && 'PR authoritative' || 'push full' }}]\"",
		"types: [opened, synchronize, reopened, ready_for_review]",
		"Record CI event source",
		"source=\"PR authoritative\"",
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
		"ci-final-jobs.json",
		"ci-evidence.json",
		"Capture CLI domain evidence",
		"go run ./cmd/gooo check examples/billing/main.gooo",
		"go run ./cmd/gooo graph-dump examples/billing/main.gooo",
		"ci-domain-evidence.json",
		"domain_evidence",
		"CI_SLOT_PRESERVATION: \"true\"",
		"CI_NO_WRITE_OUTSIDE_GENERATED: \"true\"",
		"actions/upload-artifact@v4",
		"administration: read must not be added here",
		"read_status: 'unavailable'",
		"event_ref: context.ref",
		"checkout_ref: headSha",
		"trusted_guardian_required",
		"branch_protection",
		"digest_sha256",
		"ci-proof.json",
		"provenance-receipt.jsonl",
		"if-no-files-found: error",
		"scripts/ci-proof/artifacts_test.js",
		"listWorkflowArtifacts",
		"selectCurrentEvidenceArtifact",
		"normalizeBaseRef",
		"./scripts/ci-proof/refs",
	} {
		if !strings.Contains(text, marker) {
			t.Fatalf("workflow lost event-source evidence marker %q", marker)
		}
	}
	assertWorkflowIdentityMarkers(t, text)
	if strings.Contains(text, "BRANCH_PROTECTION_TOKEN") || strings.Contains(text, "getBranchProtection") {
		t.Fatal("pull_request CI must not read branch protection or receive its observer credential")
	}
}

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
	if strings.Count(string(workflow), marker) != 8 {
		t.Fatalf("expected eight immutable checkout refs, got %d", strings.Count(string(workflow), marker))
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
	if strings.Count(text, "    name: gofmt\n") != 1 || strings.Count(text, "    name: go vet\n") != 1 || strings.Count(text, "    name: go test\n") != 1 || strings.Count(text, "    name: go test -race\n") != 1 || strings.Count(text, "    name: Semantic conformance\n") != 1 || strings.Count(text, "    name: CI policy\n") != 1 {
		t.Fatal("canonical job names are missing or duplicated")
	}
}

func readCacheSlowObservationWorkflow(t *testing.T) string {
	t.Helper()
	workflow, err := os.ReadFile(filepath.Join("..", "..", ".github", "workflows", "cache-slow-observation.yml"))
	if err != nil {
		t.Fatal(err)
	}
	return string(workflow)
}

func TestCacheSlowObservationHasOnlyOpenLoopTriggers(t *testing.T) {
	text := readCacheSlowObservationWorkflow(t)
	if !strings.Contains(text, "schedule:\n    - cron: '*/30 * * * *'") {
		t.Fatal("cache observation lost its exact 30-minute schedule")
	}
	if !strings.Contains(text, "  workflow_dispatch:\n") {
		t.Fatal("cache observation lost workflow_dispatch")
	}
	for _, forbidden := range []string{
		"  pull_request:", "  pull_request_target:", "  push:",
		"pull_request:", "pull_request_target:", "push:",
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("cache observation has forbidden PR or push trigger %q", forbidden)
		}
	}
	if strings.Contains(text, "required: true") || strings.Contains(text, "needs: [") {
		t.Fatal("cache observation is coupled to required CI gating")
	}
}

func TestCacheSlowObservationBindsRegisteredClassAndCommands(t *testing.T) {
	text := readCacheSlowObservationWorkflow(t)
	for _, marker := range []string{
		"\"class_id\": \"slow-observation\"",
		"\"class_flag\": \"-cache-test-class=slow-observation\"",
		"CACHE_CONTRACT_BRANCH: agent/cache",
		"CACHE_CONTRACT_HEAD: 3a7154c2c021c581bb7b442bd1e81a1cb3a3061e",
		"Dependency: PR #232 supplies this exact cache contract and must merge first.",
		"git cat-file -e \"$CACHE_CONTRACT_HEAD^{commit}\"",
		"git merge-base --is-ancestor \"$CACHE_CONTRACT_HEAD\" \"$actual_sha\"",
		"TestCacheLatencyEvidenceMatrix",
		"TestCacheSameKeyCrossProcessStampede",
		"TestIncrementalCacheMutationMatrix",
		"go test -count=1 -json ./internal/cache -args -cache-test-class=slow-observation",
		"go test -count=1 -race -json ./internal/cache -args -cache-test-class=slow-observation",
		"Run slow observation (normal durability)",
		"Run slow observation (race durability)",
		"\"checkout_ref\": \"refs/heads/dev\"",
		"git fetch --no-tags origin refs/heads/dev",
		"git checkout --detach \"$dev_sha\"",
		"test \"$actual_sha\" = \"$dev_sha\"",
		"event\": os.environ[\"EVENT_NAME\"]",
	} {
		if !strings.Contains(text, marker) {
			t.Fatalf("cache observation lost binding marker %q", marker)
		}
	}
	if strings.Count(text, "-cache-test-class=slow-observation") < 6 {
		t.Fatal("cache observation does not record and execute the exact class flag in both modes")
	}
	if strings.Contains(text, "-run ") || strings.Contains(text, "grep") || strings.Contains(text, "duration") {
		t.Fatal("cache observation selects tests by name or elapsed-time inference")
	}
}

func TestCacheSlowObservationRetainsFailureEvidence(t *testing.T) {
	text := readCacheSlowObservationWorkflow(t)
	for _, marker := range []string{
		"if: ${{ always() }}",
		"normal-go-test.json",
		"race-go-test.json",
		"normal-exit-status.txt",
		"race-exit-status.txt",
		"observation-result.json",
		"human-summary.md",
		"artifact-manifest.json",
		"actions/upload-artifact@v4",
		"name: cache-slow-observation-${{ github.run_id }}-${{ github.run_attempt }}",
		"if-no-files-found: error",
		"retention-days: 90",
		"overwrite: false",
		"overall_status",
		"UNKNOWN",
		"RED",
		"if not json_file.exists():",
		"cancel-in-progress: false",
	} {
		if !strings.Contains(text, marker) {
			t.Fatalf("cache observation lost failure-retention marker %q", marker)
		}
	}
}

func TestCacheSlowObservationCannotGatePRsOrPromotion(t *testing.T) {
	text := readCacheSlowObservationWorkflow(t)
	for _, marker := range []string{
		"\"merge_gate\": \"non-required\"",
		"\"promotion_authority\": \"none\"",
		"merge gate: non-required open-loop evidence",
		"this observation cannot authorize promotion",
		"name: cache slow observation (non-required)",
	} {
		if !strings.Contains(text, marker) {
			t.Fatalf("cache observation lost no-gating marker %q", marker)
		}
	}
	for _, forbidden := range []string{
		"pull_request", "pull_request_target", "github.event.pull_request",
		"CI policy", "Semantic conformance", "CI guardian",
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("cache observation contains PR/protection coupling %q", forbidden)
		}
	}
}
