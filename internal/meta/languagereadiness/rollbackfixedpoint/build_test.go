package rollbackfixedpoint

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/guardedpromotion"
)

func fixtureSource() Source {
	head := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	return Source{ExpectedHeadSHA: head, Guard: GuardEvidence{
		FileSHA256: digestJSON("guard-file"), ReportDigest: digestJSON("guard-report"),
		HeadSHA: head, Decision: guardedpromotion.DecisionFailClosed,
		Reason:     guardedpromotion.ReasonEvidenceUnknown,
		Resolution: guardedpromotion.ResolutionLower, Satisfied: 10, Total: 12, Unresolved: 2},
		Transformation: TransformationEvidence{FileSHA256: digestJSON("ledger-file"),
			LedgerDigest: digestJSON("ledger"), HeadSHA: head, Decision: "FIXED_POINT",
			Reason: "EXACT_FIXED_POINT", WorkspaceMode: "DISPOSABLE_WORKTREE",
			WriteBoundary: "SANDBOX_ONLY", SourceWorkspaceUnchanged: true}}
}

func TestFailClosedGuardRecoversAtFixedPoint(t *testing.T) {
	report := Build(fixtureSource())
	if err := Validate(report); err != nil {
		t.Fatal(err)
	}
	if report.Decision != DecisionPass || report.Mode != ModeRecovered ||
		report.Summary.Satisfied != 10 || report.Summary.RecoveredFixedPoints != 1 ||
		report.RepositoryWrites != 0 {
		t.Fatalf("report = %#v", report)
	}
}

func TestUnknownGuardDecisionLowersResolution(t *testing.T) {
	source := fixtureSource()
	source.Guard.Decision = "FUTURE_DECISION"
	report := Build(source)
	if report.Decision != DecisionFailClosed || report.Resolution != ResolutionLower ||
		report.Summary.Unresolved != 3 {
		t.Fatalf("report = %#v", report)
	}
}

func TestRecoveryRejectsSandboxEffects(t *testing.T) {
	source := fixtureSource()
	source.Transformation.Effects = 1
	report := Build(source)
	if report.Decision != DecisionFailClosed || report.Summary.NotSatisfied != 1 {
		t.Fatalf("report = %#v", report)
	}
}

func TestMixedRefutedOutcomeIsExactNonPromotingTerminal(t *testing.T) {
	source := fixtureSource()
	source.Transformation.Decision = "APPLIED"
	source.Transformation.Reason = "SANDBOX_EFFECTS_VERIFIED"
	source.Transformation.Effects = 2
	source.Transformation.AppliedEffects = 1
	source.Transformation.RefutedEffects = 1
	source.Transformation.OperationOutcome = "MIXED_CLOSED_REFUTED"
	source.Transformation.ReceiptDecision = "REFUTED"
	source.Transformation.ReceiptCount = 1
	source.Transformation.FailureCount = 1
	source.Transformation.UnknownCount = 5
	source.Transformation.DependencyBlockedUnknownCount = 5
	source.Transformation.UnknownCausalDigest = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	report := Build(source)
	if !IsKnownMixedTerminal(report) || report.Decision != DecisionFailClosed ||
		report.Resolution != ResolutionExact || report.Mode != ModeMixedTerminal {
		t.Fatalf("report = %#v", report)
	}
	if err := Validate(report); err != nil {
		t.Fatal(err)
	}
}

func TestMalformedMixedOutcomeIsNotKnownTerminal(t *testing.T) {
	source := fixtureSource()
	source.Transformation.Decision = "APPLIED"
	source.Transformation.Reason = "SANDBOX_EFFECTS_VERIFIED"
	source.Transformation.Effects = 2
	source.Transformation.AppliedEffects = 1
	source.Transformation.RefutedEffects = 1
	source.Transformation.OperationOutcome = "MIXED_CLOSED_REFUTED"
	source.Transformation.ReceiptDecision = "REFUTED"
	source.Transformation.ReceiptCount = 0
	source.Transformation.FailureCount = 1
	if IsKnownMixedTerminal(Build(source)) {
		t.Fatal("malformed mixed evidence was accepted as known")
	}
}

func TestMixedOutcomeWithDirectUnknownIsNotKnownTerminal(t *testing.T) {
	source := fixtureSource()
	source.Transformation.Decision = "APPLIED"
	source.Transformation.Reason = "SANDBOX_EFFECTS_VERIFIED"
	source.Transformation.Effects = 2
	source.Transformation.AppliedEffects = 1
	source.Transformation.RefutedEffects = 1
	source.Transformation.OperationOutcome = "MIXED_CLOSED_REFUTED"
	source.Transformation.ReceiptDecision = "REFUTED"
	source.Transformation.ReceiptCount = 1
	source.Transformation.FailureCount = 1
	source.Transformation.UnknownCount = 5
	source.Transformation.DirectUnknownCount = 1
	source.Transformation.DependencyBlockedUnknownCount = 4
	source.Transformation.UnknownCausalDigest = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	if IsKnownMixedTerminal(Build(source)) {
		t.Fatal("mixed evidence with unknown effects was accepted as known")
	}
}

func TestAuthorizedPromotionIsTerminal(t *testing.T) {
	source := fixtureSource()
	source.Guard.Decision = guardedpromotion.DecisionAuthorized
	source.Guard.Reason = guardedpromotion.ReasonAuthorized
	source.Guard.Resolution = guardedpromotion.ResolutionExact
	source.Guard.Satisfied, source.Guard.Unresolved = 12, 0
	report := Build(source)
	if report.Decision != DecisionPass || report.Mode != ModeAuthorized ||
		report.Summary.AuthorizedPromotions != 1 {
		t.Fatalf("report = %#v", report)
	}
}
