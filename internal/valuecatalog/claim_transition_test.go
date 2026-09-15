package valuecatalog

import (
	"strings"
	"testing"
	"testing/fstest"
)

func TestClaimTransitionLedgerIsFixedAndDigestChained(t *testing.T) {
	source := strings.Replace(catalogFixture, extensionDeclaration, extensionProgramLine, 1)
	report := Evaluate(fstest.MapFS{"main.gooo": {Data: []byte(source)}}, "main.gooo", strings.Repeat("d", 40))
	if err := validateClaimTransitionLedger(report); err != nil {
		t.Fatal(err)
	}
	if len(report.ClaimTransitions) != 18 || report.ClaimTransitions[8].Event != ClaimEventRegistered ||
		report.ClaimTransitions[9].Event != ClaimEventEvidenceAccepted {
		t.Fatalf("claim transition phases changed: %#v", report.ClaimTransitions)
	}
	tampered := report
	tampered.ClaimTransitions = append([]ClaimTransition(nil), report.ClaimTransitions...)
	tampered.ClaimTransitions[10].PreviousTransitionDigest = strings.Repeat("0", 71)
	if validateClaimTransitionLedger(tampered) == nil {
		t.Fatal("tampered transition chain was accepted")
	}
}

func TestUnknownClaimTransitionsPreserveExactFailureCoordinate(t *testing.T) {
	source := strings.Replace(catalogFixture, extensionDeclaration, extensionDeclaration+` computes "int.magic:2"`, 1)
	report := Evaluate(fstest.MapFS{"main.gooo": {Data: []byte(source)}}, "main.gooo", strings.Repeat("e", 40))
	for _, transition := range report.ClaimTransitions[OperationSpecAxisTotal:] {
		if transition.Event != ClaimEventEvidenceUnavailable || transition.After != ClaimStatusOpen ||
			transition.Coordinate != report.ProcessCoordinate || transition.EvidenceDigest != "" {
			t.Fatalf("unknown transition hid its coordinate: %#v", transition)
		}
	}
	if report.OperationSpecMetrics.EvidenceUnavailableTotal != 9 || report.OperationSpecMetrics.EvidenceAcceptedTotal != 0 {
		t.Fatalf("unknown transition counts changed: %#v", report.OperationSpecMetrics)
	}
}
