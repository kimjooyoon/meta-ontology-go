package languagesourcebindingpromotion

import "testing"

func TestPromotionRequiresIndependentSourceBinding(t *testing.T) {
	report := Evaluate(fixtureInput())
	if err := Validate(report); err != nil {
		t.Fatal(err)
	}
	wantReasons := []string{"SOURCE_BINDING_CLAIM_DISCHARGED", "ARTIFACT_ORACLE_EVIDENCE_MISSING",
		"ARTIFACT_ORACLE_DECISION_UNKNOWN", "SOURCE_BINDING_EVIDENCE_LINK_MISMATCH", "SOURCE_EXECUTION_DECISION_UNKNOWN"}
	for index, item := range report.Cases {
		if item.ObservedReason != wantReasons[index] || item.Status != "SATISFIED" || len(item.Claims) != 3 {
			t.Fatalf("case %d = %#v", index, item)
		}
	}
}

func TestPolicyReplayMismatchFailsClosed(t *testing.T) {
	input := fixtureInput()
	input.PolicyReplayArtifact = []byte("different")
	report := Evaluate(input)
	if report.Decision != DecisionClosed || report.Reason != "SOURCE_BINDING_PROMOTION_CONTRACT_MISMATCH" {
		t.Fatalf("report = %#v", report)
	}
}
