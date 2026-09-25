package transformationeffectverification

import (
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

func TestProjectUpstreamCausePreservesReceiptFrontier(t *testing.T) {
	report := generation.ReceiptReport{
		ReportDigest: strings.Repeat("a", 64),
		Decision:     generation.ReceiptDecisionRefuted,
		Reason:       generation.ReceiptReason("REFUTED_OPERATION"),
		Failures: []generation.ObservationFailure{{
			ActionIndicatorID: "operation-1", Decision: "REFUTED", Stage: "execute",
			Step: "compile", Reason: "PROJECTED_COMPILE_OR_TEST_FAILED",
			NextOperation: "restore-operation-evidence", BlockedBy: []string{},
		}},
		Unknowns: []generation.ReceiptUnknown{{
			ActionIndicatorID: "operation-1", RequiredIndicatorID: "indicator-1",
			Stage: "execute", Step: "compile", Reason: generation.ReceiptReason("PROJECTED_COMPILE_OR_TEST_FAILED"),
			UnknownClass:  generation.ReceiptUnknownClassDependencyBlocked,
			NextOperation: "restore-operation-evidence",
			BlockedBy:     []string{"operation-failure:operation-1"},
		}},
	}
	var target Report
	projectUpstreamCause(&target, report)
	if target.Upstream == nil {
		t.Fatal("typed upstream cause was not projected")
	}
	if target.Upstream.ReceiptReportDigest != strings.Repeat("a", 64) ||
		target.Upstream.Decision != "REFUTED" ||
		target.Upstream.FailureCount != 1 || target.Upstream.UnknownCount != 1 {
		t.Fatalf("upstream identity or counts were not preserved: %+v", target.Upstream)
	}
	if target.Upstream.Failures[0].Reason != "PROJECTED_COMPILE_OR_TEST_FAILED" ||
		len(target.Upstream.Unknowns[0].BlockedBy) != 1 ||
		target.Upstream.Unknowns[0].BlockedBy[0] != "operation-failure:operation-1" {
		t.Fatalf("upstream causal frontier was not preserved: %+v", target.Upstream)
	}
	if target.Decision != "" {
		t.Fatalf("projection changed downstream decision: %q", target.Decision)
	}
}

func TestProjectUpstreamCauseRejectsUnboundReceipt(t *testing.T) {
	report := generation.ReceiptReport{
		Decision: generation.ReceiptDecisionRefuted,
		Reason:   generation.ReceiptReason("REFUTED_OPERATION"),
		Failures: []generation.ObservationFailure{{ActionIndicatorID: "operation-1", Decision: "REFUTED", Stage: "execute", Step: "compile", Reason: "failure", NextOperation: "restore", BlockedBy: []string{}}},
	}
	var target Report
	projectUpstreamCause(&target, report)
	if target.Upstream != nil {
		t.Fatal("receipt without an immutable report digest was projected")
	}
}
