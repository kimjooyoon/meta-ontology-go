package promotioncontinuity

import (
	"strings"
	"testing"
)

func exactEvidence() (string, GuardEvidence, RecoveryEvidence) {
	head := strings.Repeat("a", 40)
	digest := "sha256:" + strings.Repeat("b", 64)
	guard := GuardEvidence{
		Schema: guardSchema, FileSHA256: digest, ReportDigest: digest, HeadSHA: head,
		Decision: "AUTHORIZED", Reason: "MERGED_PUSH_PROMOTION_AUTHORIZED", Resolution: "EXACT",
		Satisfied: 12, Total: 12, PromotionAuthorized: true,
	}
	recovery := RecoveryEvidence{
		Schema: recoverySchema, FileSHA256: digest, ReportDigest: digest, HeadSHA: head,
		Decision: "PASS", Reason: "PROMOTION_AUTHORIZATION_PRESERVED", Resolution: "EXACT",
		Mode: "PROMOTION_AUTHORIZED", GuardDecision: "AUTHORIZED", GuardResolution: "EXACT",
		Satisfied: 10, Total: 10, ReadinessBPS: 10000, AuthorizedPromotions: 1,
		TransformationDecision: "FIXED_POINT", WriteBoundary: "SANDBOX_ONLY",
		SourceWorkspaceUnchanged: true,
	}
	return head, guard, recovery
}

func mixedEvidence() (string, GuardEvidence, RecoveryEvidence) {
	head, guard, recovery := exactEvidence()
	guard.Decision = DecisionFailClosed
	guard.Reason = "GUARDED_PROMOTION_EVIDENCE_UNKNOWN"
	guard.Resolution = "LOWER_RESOLUTION"
	guard.Satisfied, guard.Total, guard.Unresolved = 10, 12, 2
	guard.PromotionAuthorized = false
	recovery.Decision = DecisionFailClosed
	recovery.Reason = "MIXED_REFUTED_NON_PROMOTABLE"
	recovery.Resolution = "EXACT"
	recovery.Mode = "MIXED_REFUTED_NON_PROMOTABLE"
	recovery.GuardDecision = DecisionFailClosed
	recovery.GuardReason = guard.Reason
	recovery.GuardResolution = guard.Resolution
	recovery.GuardSatisfied, recovery.GuardTotal, recovery.GuardUnresolved = 10, 12, 2
	recovery.TransformationHeadSHA = head
	recovery.TransformationDecision = "APPLIED"
	recovery.TransformationReason = "SANDBOX_EFFECTS_VERIFIED"
	recovery.TransformationWorkspaceMode = "DISPOSABLE_WORKTREE"
	recovery.TransformationEffects = 2
	recovery.TransformationAppliedEffects = 1
	recovery.TransformationRefutedEffects = 1
	recovery.TransformationOperationOutcome = "MIXED_CLOSED_REFUTED"
	recovery.TransformationReceiptDecision = "REFUTED"
	recovery.TransformationReceiptCount = 1
	recovery.TransformationFailureCount = 1
	recovery.TransformationUnknownCount = 5
	recovery.TransformationDirectUnknownCount = 0
	recovery.TransformationDependencyBlockedUnknownCount = 5
	recovery.TransformationUnknownCausalDigest = "sha256:" + strings.Repeat("c", 64)
	recovery.Satisfied, recovery.Total, recovery.Unresolved = 8, 10, 0
	recovery.ReadinessBPS = 8000
	recovery.RecoveredFixedPoints, recovery.AuthorizedPromotions = 0, 0
	return head, guard, recovery
}

func TestEvaluateProvesAuthorizedContinuity(t *testing.T) {
	head, guard, recovery := exactEvidence()
	report := Evaluate(head, guard, recovery)
	if report.Decision != "PASS" || report.Resolution != "EXACT" {
		t.Fatalf("decision=%s resolution=%s", report.Decision, report.Resolution)
	}
	if report.Summary.Satisfied != 8 || report.Summary.Total != 8 || report.Summary.ReadinessBPS != 10000 {
		t.Fatalf("summary=%+v", report.Summary)
	}
	if err := Validate(report); err != nil {
		t.Fatal(err)
	}
}

func TestEvaluateLowersResolutionForUnknownGuard(t *testing.T) {
	head, guard, recovery := exactEvidence()
	guard.Decision = "UNKNOWN"
	report := Evaluate(head, guard, recovery)
	if report.Decision != "FAIL_CLOSED" || report.Resolution != "LOWER_RESOLUTION" {
		t.Fatalf("decision=%s resolution=%s", report.Decision, report.Resolution)
	}
	if report.Summary.Unresolved == 0 || report.Summary.AuthorizedGuardReceipts != 0 {
		t.Fatalf("summary=%+v", report.Summary)
	}
}

func TestEvaluateRejectsTransformationEffects(t *testing.T) {
	head, guard, recovery := exactEvidence()
	recovery.TransformationEffects = 1
	report := Evaluate(head, guard, recovery)
	if report.Decision != "FAIL_CLOSED" || report.RepositoryWrites != 0 {
		t.Fatalf("decision=%s writes=%d", report.Decision, report.RepositoryWrites)
	}
}

func TestEvaluatePreservesExactMixedNonPromotingTerminal(t *testing.T) {
	head, guard, recovery := mixedEvidence()
	report := Evaluate(head, guard, recovery)
	if !IsKnownNonPromotingTerminal(report) || report.Decision != DecisionFailClosed ||
		report.Reason != ReasonMixed || report.Resolution != "EXACT" ||
		report.Mode != ModeMixed || report.MetaOperation != OperationMixed ||
		report.Summary.Satisfied != 8 || report.Summary.Total != 8 {
		t.Fatalf("report=%+v", report)
	}
	if err := Validate(report); err != nil {
		t.Fatal(err)
	}
}

func TestEvaluateRejectsUnknownTopLevelMixedDecision(t *testing.T) {
	head, guard, recovery := mixedEvidence()
	recovery.Decision = "UNKNOWN"
	report := Evaluate(head, guard, recovery)
	if report.Resolution != "LOWER_RESOLUTION" || IsKnownNonPromotingTerminal(report) {
		t.Fatalf("report=%+v", report)
	}
}

func TestEvaluateRejectsMixedCausalDigestMismatch(t *testing.T) {
	head, guard, recovery := mixedEvidence()
	recovery.TransformationUnknownCausalDigest = "sha256:" + strings.Repeat("d", 64)
	report := Evaluate(head, guard, recovery)
	if report.Resolution != "LOWER_RESOLUTION" || IsKnownNonPromotingTerminal(report) {
		t.Fatalf("report=%+v", report)
	}
}

func TestEvaluateMixedTerminalReplaysExactly(t *testing.T) {
	head, guard, recovery := mixedEvidence()
	first := Evaluate(head, guard, recovery)
	second := Evaluate(head, guard, recovery)
	if first.ReportDigest != second.ReportDigest || first.Mode != ModeMixed ||
		first.MetaOperation != OperationMixed {
		t.Fatalf("first=%+v second=%+v", first, second)
	}
}
