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

func TestPromotionSeparatesReceiptAndArtifactDigests(t *testing.T) {
	input := fixtureInput()
	receipt, err := decodeStrict[receiptEnvelope](input.Receipt)
	if err != nil {
		t.Fatal(err)
	}
	producer, err := decodeStrict[producerEnvelope](input.Producer)
	if err != nil {
		t.Fatal(err)
	}
	producerCases, err := decodeView[[]producerCase](producer.Cases)
	if err != nil {
		t.Fatal(err)
	}
	oracle, err := decodeStrict[oracleEnvelope](input.Oracle)
	if err != nil {
		t.Fatal(err)
	}
	oracleCases, err := decodeView[[]oracleCase](oracle.Cases)
	if err != nil {
		t.Fatal(err)
	}
	artifactDigest := producerCases[0].EvidenceDigest
	if artifactDigest == receipt.Digest || oracleCases[0].ArtifactDigest != artifactDigest {
		t.Fatalf("receipt=%s producer-artifact=%s oracle-artifact=%s", receipt.Digest, artifactDigest, oracleCases[0].ArtifactDigest)
	}
	if err := Validate(Evaluate(input)); err != nil {
		t.Fatal(err)
	}
}

func TestDependencyBlockedMetricCountsPersistentClaims(t *testing.T) {
	report := Evaluate(fixtureInput())
	blockedClaims := 0
	for _, item := range report.Cases {
		for _, claim := range item.Claims {
			if claim.UnknownClass == "DEPENDENCY_BLOCKED" {
				blockedClaims++
			}
		}
	}
	if blockedClaims != 4 || report.Summary.DependencyBlocked != blockedClaims {
		t.Fatalf("blocked claims=%d summary=%d", blockedClaims, report.Summary.DependencyBlocked)
	}
}
