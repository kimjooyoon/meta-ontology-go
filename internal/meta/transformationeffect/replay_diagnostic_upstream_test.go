package transformationeffect

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestReplayDiagnosticProjectsBoundUpstreamCause(t *testing.T) {
	directory := t.TempDir()
	output := filepath.Join(directory, "ledger.json")
	reason := string(generation.ReceiptReasonRefutedOperation)
	report := generation.ReceiptReport{
		BaseSHA:      "base-sha",
		HeadSHA:      "head-sha",
		PlanDigest:   "sha256:plan",
		InputDigest:  "sha256:input",
		Decision:     generation.ReceiptDecisionRefuted,
		Reason:       generation.ReceiptReasonRefutedOperation,
		ReportDigest: "sha256:report",
		Failures: []generation.ObservationFailure{{
			ActionIndicatorID: "action-1",
			Decision:          string(generation.ReceiptDecisionRefuted),
			Stage:             "input",
			Step:              "evaluate",
			Reason:            reason,
			NextOperation:     "restore-input",
			BlockedBy:         []string{"operation-failure:action-1"},
		}},
		Unknowns: []generation.ReceiptUnknown{{
			ActionIndicatorID:   "action-1",
			RequiredIndicatorID: "required-1",
			Operation:           sourcepolicy.Operation("operation"),
			Activity:            "activity",
			Output:              "output",
			Executor:            "executor",
			Evaluator:            "evaluator",
			Stage:               "input",
			Step:                "evaluate",
			Reason:              generation.ReceiptReasonRefutedOperation,
			UnknownClass:        generation.ReceiptUnknownClassDependencyBlocked,
			NextOperation:       "restore-input",
			BlockedBy:           []string{"operation-failure:action-1"},
		}},
	}
	if err := WriteReplayDiagnostic(output, &replayDiagnosticContext{
		Cause: errors.New("downstream was not executed"),
		Report: report,
	}); err != nil {
		t.Fatal(err)
	}
	diagnostic := readReplayDiagnostic(t, filepath.Join(directory, "replay-diagnostic.json"))
	if diagnostic.Decision != "UNKNOWN" ||
		diagnostic.UnknownClass != generation.ReceiptUnknownClassDependencyBlocked ||
		diagnostic.Reason != "UPSTREAM_RECEIPT_DEPENDENCY_BLOCKED" ||
		diagnostic.NextOperation != "restore-input" ||
		len(diagnostic.BlockedBy) != 1 ||
		diagnostic.Upstream == nil ||
		diagnostic.Upstream.HeadSHA != "head-sha" ||
		diagnostic.Upstream.ReportDigest != "sha256:report" ||
		len(diagnostic.Upstream.Failures) != 1 ||
		len(diagnostic.Upstream.Unknowns) != 1 ||
		diagnostic.Upstream.CausalDigest == "" {
		t.Fatalf("diagnostic = %#v", diagnostic)
	}
	if diagnostic.Upstream.PromotionAuthorized {
		t.Fatal("upstream receipt must not grant promotion authority")
	}
}

func TestReplayDiagnosticRejectsPromotionAuthority(t *testing.T) {
	directory := t.TempDir()
	output := filepath.Join(directory, "ledger.json")
	report := generation.ReceiptReport{
		Decision:            generation.ReceiptDecisionRefuted,
		Reason:              generation.ReceiptReasonRefutedOperation,
		PromotionAuthorized: true,
	}
	if err := WriteReplayDiagnostic(output, &replayDiagnosticContext{
		Cause: errors.New("downstream was not executed"),
		Report: report,
	}); err != nil {
		t.Fatal(err)
	}
	diagnostic := readReplayDiagnostic(t, filepath.Join(directory, "replay-diagnostic.json"))
	if diagnostic.Upstream != nil || diagnostic.UnknownClass != "UNCATALOGED_CAUSE" {
		t.Fatalf("diagnostic = %#v", diagnostic)
	}
}

var _ = json.Valid
var _ = os.ErrNotExist
