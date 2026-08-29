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
	head := "ab1274a1fd704328c06e4684554715003683a316"
	guard := GuardEvidence{
		Schema: guardSchema, FileSHA256: "sha256:f2f772943e57e74cc95a6e4caadee7b5b1c3175a423574876c07255703f91d29",
		ReportDigest: "sha256:9a70a19d2c4d05bf667e97dbd6209e2f54b2bd1abb45400880ecfd26c7b77012", HeadSHA: head,
		Decision: DecisionFailClosed, Reason: "GUARDED_PROMOTION_EVIDENCE_UNKNOWN", Resolution: "LOWER_RESOLUTION",
		Satisfied: 10, Total: 12, Unresolved: 2,
	}
	recovery := RecoveryEvidence{
		Schema: recoverySchema, FileSHA256: "sha256:8017ce469a71b796f096e67ec8c0f797076771d3305e75e5f9e440feff4a3483",
		ReportDigest: "sha256:e31daaef9bdda39981409595372b931a9c0427adfc8bf6b88bba6e2104fe8e6f", HeadSHA: head,
		Decision: DecisionFailClosed, Reason: "MIXED_REFUTED_NON_PROMOTABLE", Resolution: "EXACT",
		Mode: "MIXED_REFUTED_NON_PROMOTABLE", GuardDecision: DecisionFailClosed,
		GuardReason: "GUARDED_PROMOTION_EVIDENCE_UNKNOWN", GuardResolution: "LOWER_RESOLUTION",
		GuardSatisfied: 10, GuardTotal: 12, GuardUnresolved: 2,
		TransformationHeadSHA: head, TransformationDecision: "APPLIED",
		TransformationReason: "SANDBOX_EFFECTS_VERIFIED", TransformationWorkspaceMode: "DISPOSABLE_WORKTREE",
		TransformationEffects: 2, TransformationAppliedEffects: 1, TransformationRefutedEffects: 1,
		TransformationOperationOutcome: "MIXED_CLOSED_REFUTED", TransformationReceiptDecision: "REFUTED",
		TransformationReceiptCount: 1, TransformationFailureCount: 1, TransformationUnknownCount: 5,
		TransformationDirectUnknownCount: 0, TransformationDependencyBlockedUnknownCount: 5,
		TransformationUnknownCausalDigest: "003671c624aef06c7921c4032052c50a10ca1aaaa39304eb9953cc14ec4ecc40",
		Satisfied:                         8, Total: 10, Unresolved: 0, ReadinessBPS: 8000,
	}
	recovery.TransformationCausalBindingDigest = causalBindingDigest(recovery)
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
	if report.Summary.AuthorizedGuardReceipts != 0 || report.Summary.AuthorizedRecoveryRoutes != 0 {
		t.Fatalf("mixed report authorized indicators=%+v", report.Summary)
	}
	terminalIndicator := false
	authorizationIndicators := 0
	for _, indicator := range report.Indicators {
		switch indicator.MetricID {
		case "gooo.metric.language.promotion-continuity-authorized-guards.v1",
			"gooo.metric.language.promotion-continuity-authorized-routes.v1":
			authorizationIndicators++
			if indicator.Satisfied || indicator.Value != 0 {
				t.Fatalf("mixed authorization indicator=%+v", indicator)
			}
		case "gooo.metric.language.promotion-continuity-terminal-preserved.v1":
			terminalIndicator = indicator.Satisfied
		}
	}
	if authorizationIndicators != 2 || !terminalIndicator {
		t.Fatal("mixed terminal indicator is not satisfied")
	}
	for _, coordinate := range report.Coordinates {
		if coordinate.Status == "SATISFIED" && strings.Contains(coordinate.ID, "authoriz") {
			t.Fatalf("mixed report passed authorization coordinate %q", coordinate.ID)
		}
	}
	for _, proof := range report.Proofs {
		if proof.Passed && strings.Contains(proof.MetaOperation, "authoriz") {
			t.Fatalf("mixed report passed authorization proof %q", proof.MetaOperation)
		}
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
	recovery.TransformationUnknownCausalDigest = strings.Repeat("d", 64)
	report := Evaluate(head, guard, recovery)
	if report.Resolution != "LOWER_RESOLUTION" || IsKnownNonPromotingTerminal(report) {
		t.Fatalf("report=%+v", report)
	}
}

func TestEvaluateRejectsCausalDigestPrefixDrift(t *testing.T) {
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
