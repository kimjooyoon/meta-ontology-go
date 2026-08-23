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
