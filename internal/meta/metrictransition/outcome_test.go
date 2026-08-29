package metrictransition

import (
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/transformationeffect"
)

func mixedOutcomeInput() inputSet {
	unknowns := make([]generation.ReceiptUnknown, 0, 5)
	for _, required := range []string{"indicator-1", "indicator-2", "indicator-3", "indicator-4", "indicator-5"} {
		unknowns = append(unknowns, generation.ReceiptUnknown{
			ActionIndicatorID: "action-b", RequiredIndicatorID: required,
			Operation: "extract-function", Activity: "extract", Output: "receipt",
			Executor: "bootstrap/function-extractor", Evaluator: "extract-evaluator",
			Stage: "derive-recipe", Step: "select-declaration",
			Reason:        "NO_SAFE_DECLARATION_CAPACITY",
			UnknownClass:  generation.ReceiptUnknownClassDependencyBlocked,
			NextOperation: "report-counterexample", BlockedBy: []string{"operation-failure:action-b"}})
	}
	failure := generation.ObservationFailure{ActionIndicatorID: "action-b", Decision: "REFUTED",
		Stage: "derive-recipe", Step: "select-declaration", Reason: "NO_SAFE_DECLARATION_CAPACITY",
		NextOperation: "report-counterexample", BlockedBy: []string{}}
	ledger := transformationeffect.Ledger{Status: "BOUND", Decision: "APPLIED",
		Reason: "SANDBOX_EFFECTS_VERIFIED", HeadSHA: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		SourceWorkspaceUnchanged: true, SelectedPlanOperations: 2,
		BoundExecutorOperations: 2, OperationOutcome: "MIXED_CLOSED_REFUTED",
		ReceiptDecision: "REFUTED", ReceiptCount: 1, FailureCount: 1, UnknownCount: 5,
		Effects: []transformationeffect.Effect{{ActionIndicatorID: "action-a", Status: "APPLIED"},
			{ActionIndicatorID: "action-b", Status: "REFUTED"}}}
	report := generation.ReceiptReport{Decision: generation.ReceiptDecisionRefuted,
		Reason:       generation.ReceiptReasonRefutedOperation,
		ReportDigest: "sha256:report", Receipts: []generation.OperationReceipt{{ActionIndicatorID: "action-a"}},
		Failures: []generation.ObservationFailure{failure}, Unknowns: unknowns}
	causal, err := deriveCausalUnknowns(report)
	if err != nil {
		panic(err)
	}
	ledger.DirectUnknownCount = causal.DirectUnknownCount
	ledger.DependencyBlockedUnknownCount = causal.DependencyBlockedUnknownCount
	ledger.UnknownCausalDigest = causal.Digest
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

func TestMixedOutcomeRejectsMalformedDependencyFrontier(t *testing.T) {
	inputs := mixedOutcomeInput()
	inputs.receiptReport.Unknowns[0].BlockedBy = []string{"operation-failure:other"}
	if _, err := validateEffectOutcome(inputs); err == nil {
		t.Fatal("malformed dependency frontier was accepted")
	}
}

func TestMixedOutcomeRequiresAllSelectedExecutorsBound(t *testing.T) {
	inputs := mixedOutcomeInput()
	inputs.effectLedger.BoundExecutorOperations = 1
	if _, err := validateEffectOutcome(inputs); err == nil {
		t.Fatal("partially bound selected operations were accepted")
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
	inputs.receiptReport.Unknowns = nil
	inputs.effectLedger.UnknownCount = 0
	inputs.effectLedger.DirectUnknownCount = 0
	inputs.effectLedger.DependencyBlockedUnknownCount = 0
	causal, err := deriveCausalUnknowns(inputs.receiptReport)
	if err != nil {
		t.Fatal(err)
	}
	inputs.effectLedger.UnknownCausalDigest = causal.Digest
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
	firstCausal, err := deriveCausalUnknowns(inputs.receiptReport)
	if err != nil {
		t.Fatal(err)
	}
	secondCausal, err := deriveCausalUnknowns(inputs.receiptReport)
	if err != nil || !reflect.DeepEqual(firstCausal, secondCausal) {
		t.Fatalf("causal projection was not deterministic: %#v %#v", firstCausal, secondCausal)
	}
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
