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
		Reason: generation.ReceiptReasonRefutedOperation,
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

func TestUnknownLedgerDecisionOrReasonDoesNotBecomeKnownOutcome(t *testing.T) {
	for _, mutate := range []func(*inputSet){
		func(inputs *inputSet) { inputs.effectLedger.Decision = "UNRECOGNIZED" },
		func(inputs *inputSet) { inputs.effectLedger.Reason = "UNRECOGNIZED" },
	} {
		inputs := mixedOutcomeInput()
		mutate(&inputs)
		if _, err := validateEffectOutcome(inputs); err == nil {
			t.Fatal("unknown ledger tuple was accepted")
		}
	}
}

func TestMixedOutcomeRejectsCrossKindDuplicateIdentity(t *testing.T) {
	inputs := mixedOutcomeInput()
	inputs.effectLedger.Effects[1].Status = "APPLIED"
	inputs.receiptReport.Receipts[0].ActionIndicatorID = "action-b"
	if _, err := validateEffectOutcome(inputs); err == nil {
		t.Fatal("receipt/failure duplicate identity was accepted")
	}
}

func TestClosedOutcomeUsesExactNonPromotingTuple(t *testing.T) {
	inputs := mixedOutcomeInput()
	inputs.effectLedger.Effects = inputs.effectLedger.Effects[:1]
	inputs.effectLedger.SelectedPlanOperations = 1
	inputs.effectLedger.Effects[0].Status = "APPLIED"
	inputs.effectLedger.OperationOutcome = effectOutcomeClosed
	inputs.effectLedger.ReceiptDecision = string(generation.ReceiptDecisionConformant)
	inputs.effectLedger.ReceiptCount = 1
	inputs.effectLedger.FailureCount = 0
	inputs.receiptReport.Decision = generation.ReceiptDecisionConformant
	inputs.receiptReport.Reason = generation.ReceiptReasonVerified
	inputs.receiptReport.Failures = nil
	if outcome, err := validateEffectOutcome(inputs); err != nil || outcome != effectOutcomeClosed {
		t.Fatalf("outcome=%q err=%v, want closed", outcome, err)
	}
}

func TestClosedOutcomeDoesNotUseFixedPointMetaOperation(t *testing.T) {
	effect := EffectEvidence{Outcome: effectOutcomeClosed,
		Artifacts: []ArtifactEvidence{{Role: "effect-ledger", Digest: "sha256:effect"}},
		SetDigest: "sha256:set"}
	indicators := transitionIndicators(RepositoryState{}, effect, "head")
	for _, indicator := range indicators {
		if indicator.MetaOperation == "terminate-at-fixed-point" || indicator.ID == "regression.zero-metric-delta" {
			t.Fatalf("closed outcome used fixed-point indicator: %#v", indicator)
		}
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
