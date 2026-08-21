package verify

import (
	"strings"
	"testing"
)

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
		"actions/github-script@v8",
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
		"actions/upload-artifact@v6",
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
