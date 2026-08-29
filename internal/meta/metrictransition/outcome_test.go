package metrictransition

import (
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/transformationeffect"
)

func mixedOutcomeInput() inputSet {
	ledger := transformationeffect.Ledger{Status: "BOUND", Decision: "APPLIED",
		Reason: "SANDBOX_EFFECTS_VERIFIED", HeadSHA: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		SourceWorkspaceUnchanged: true, SelectedPlanOperations: 2,
		BoundExecutorOperations: 2, OperationOutcome: "MIXED_CLOSED_REFUTED",
		ReceiptDecision: "REFUTED", ReceiptCount: 1, FailureCount: 1,
		Effects: []transformationeffect.Effect{{ActionIndicatorID: "action-a", Status: "APPLIED"},
			{ActionIndicatorID: "action-b", Status: "REFUTED"}}}
	report := generation.ReceiptReport{Decision: generation.ReceiptDecisionRefuted,
		ReportDigest: "sha256:report", Receipts: []generation.OperationReceipt{{ActionIndicatorID: "action-a"}},
		Failures: []generation.ObservationFailure{{ActionIndicatorID: "action-b", Decision: "REFUTED"}}}
	return inputSet{effectLedger: ledger, receiptReport: report,
		provenanceReport: generation.ArtifactProvenance{Decision: generation.ArtifactProvenanceDecisionBound,
			HeadSHA: ledger.HeadSHA, ReceiptReportDigest: report.ReportDigest}}
}

func TestMixedOutcomeIsTerminalWithoutFixedPoint(t *testing.T) {
	inputs := mixedOutcomeInput()
	outcome, err := validateEffectOutcome(inputs)
	if err != nil || outcome != effectOutcomeMixedRefuted {
		t.Fatalf("outcome=%q err=%v, want mixed refuted", outcome, err)
	}
}

func TestUnknownOutcomeDoesNotBecomeFixedPoint(t *testing.T) {
	inputs := mixedOutcomeInput()
	inputs.receiptReport.Decision = generation.ReceiptDecisionUnknown
	if _, err := validateEffectOutcome(inputs); err == nil {
		t.Fatal("unknown top-level outcome was accepted")
	}
}

func TestMixedOutcomeRejectsIdentityDrift(t *testing.T) {
	inputs := mixedOutcomeInput()
	inputs.effectLedger.Effects[1].ActionIndicatorID = "action-c"
	if _, err := validateEffectOutcome(inputs); err == nil {
		t.Fatal("unbound effect identity was accepted")
	}
}

func TestMixedEffectProjectionIsDeterministic(t *testing.T) {
	inputs := mixedOutcomeInput()
	first, err := buildEffectEvidence(inputs, effectOutcomeMixedRefuted)
	if err != nil {
		t.Fatal(err)
	}
	second, err := buildEffectEvidence(inputs, effectOutcomeMixedRefuted)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatal("mixed effect projection is not deterministic")
	}
}
