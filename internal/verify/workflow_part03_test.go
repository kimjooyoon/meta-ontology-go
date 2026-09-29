package verify

import (
	"strings"
	"testing"
)

func assertWorkflowMarkers(t *testing.T, text string) {
	t.Helper()
	for _, marker := range []string{
		"run-name: \"CI [${{ github.event_name == 'pull_request' && 'PR authoritative' || (github.event_name == 'workflow_dispatch' && 'promotion dispatch' || 'push full') }}]\"",
		"types: [opened, synchronize, reopened, ready_for_review]",
		"Record CI event source",
		"source=\"PR authoritative\"",
		"source=\"promotion dispatch\"",
		"source=\"push full\"",
		"GITHUB_STEP_SUMMARY",
		"ref: ${{ github.event_name == 'pull_request' && github.event.pull_request.head.sha || github.sha }}",
		"Verify PR checkout identity",
		"EXPECTED_HEAD: ${{ github.event.pull_request.head.sha || github.event.inputs.promotion_head_sha }}",
		"EXPECTED_BASE: ${{ github.event.pull_request.base.sha || github.event.inputs.promotion_base_sha }}",
		"actual_head=\"$(git rev-parse HEAD)\"",
		"git rev-parse --verify \"$EXPECTED_BASE^{commit}\" >/dev/null",
		"GOOO_SCOPE_FROM: ${{ github.event_name == 'pull_request' && github.event.pull_request.base.sha || github.event.inputs.promotion_base_sha || github.event.before }}",
		"GOOO_SCOPE_TO: ${{ github.event.pull_request.head.sha || github.event.inputs.promotion_head_sha || github.sha }}",
		"GOOO_EXPECTED_HEAD: ${{ github.event.pull_request.head.sha || github.event.inputs.promotion_head_sha || github.sha }}",
		"needs: [format, vet, test, race, semantic]",
		"if: ${{ always() }}",
		"actions/github-script@v9.0.0",
		"listJobsForWorkflowRun",
		"ci-jobs.json",
		"ci-final-jobs.json",
		"ci-evidence.json",
		"Capture CLI domain evidence",
		"go run ./cmd/gooo check examples/billing/main.gooo",
		"go run ./cmd/gooo graph dump examples/billing/main.gooo",
		"reason: 'GRAPH_OBSERVER_NOT_RUN'",
		"observation: {state: 'UNKNOWN', stage: 'DOMAIN_EVIDENCE', step: 'GRAPH_DUMP'",
		"unknown_class: 'DIRECT_MISSING', next_operation: 'RUN_GRAPH_DUMP_OBSERVER', blocked_by: []",
		"ci-domain-evidence.json",
		"domain_evidence",
		"CI_SLOT_PRESERVATION: \"true\"",
		"CI_NO_WRITE_OUTSIDE_GENERATED: \"true\"",
		"actions/upload-artifact@v7",
		"event_ref: context.ref",
		"checkout_ref: headSha",
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
	if strings.Contains(text, "graph-dump is not implemented") {
		t.Fatal("an unexecuted observer must not claim language capability is absent")
	}
	if strings.Contains(text, "BRANCH_PROTECTION_TOKEN") || strings.Contains(text, "getBranchProtection") || strings.Contains(strings.ToLower(text), "guardian") || strings.Contains(text, "branch_protection") || strings.Contains(text, "GOOO_HUMAN_DECISION") || strings.Contains(text, "ALLOW_ONE_TIME_FOUNDATION_PROMOTION") {
		t.Fatal("main CI must not depend on a separate Guardian or branch-protection observer")
	}
}
