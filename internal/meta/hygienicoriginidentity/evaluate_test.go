package hygienicoriginidentity

import (
	"strings"
	"testing"
	"testing/fstest"
)

const testSource = `package hygienicoriginidentity
namespace hygienicoriginidentity
entity OriginIdentity id "origin"
entity ScopeProvenance id "scope"
entity GeneratedName id "generated"
entity ConsumerBinding id "consumer"
entity CapturedResult id "captured-result"
entity HygienicResult id "hygienic-result"
entity CapturedOriginClaim id "captured-origin"
entity CapturedScopeClaim id "captured-scope"
entity HygienicOriginClaim id "hygienic-origin"
entity HygienicScopeClaim id "hygienic-scope"
entity ProofReceipt id "receipt"
activity ProduceSameSpelling(OriginIdentity) -> GeneratedName
activity ConsumeName(GeneratedName) -> ConsumerBinding
activity ObserveCapturedResult(ConsumerBinding) -> CapturedResult
activity ObserveHygienicResult(ConsumerBinding) -> HygienicResult
activity PreserveOriginIdentity(GeneratedName) -> OriginIdentity
activity PreserveScopeProvenance(GeneratedName) -> ScopeProvenance
activity EmitProofReceipt(CapturedResult) -> ProofReceipt
# experiment.case id=captured spelling=tmp origin=consumer-binding scope=consumer-call-site resolves=consumer-binding expected=captured
# experiment.case id=hygienic spelling=tmp origin=producer-expansion-1 scope=fresh-producer-expansion-1 resolves=producer-expansion-1 expected=not-captured
`

func TestEvaluateClassifiesSameSpellingByIdentity(t *testing.T) {
	report, err := Evaluate(fstest.MapFS{"main.gooo": {Data: []byte(testSource)}}, "main.gooo", strings.Repeat("a", 40))
	if err != nil {
		t.Fatal(err)
	}
	if err := Validate(report, false, strings.Repeat("a", 40)); err != nil {
		t.Fatal(err)
	}
	if report.Cases[0].Captured != true || report.Cases[1].Captured != false {
		t.Fatalf("capture classification = %#v", report.Cases)
	}
	if report.Metrics.FixedClaimDenominator != 4 || report.Metrics.PreservationSatisfactionBPS != 5000 {
		t.Fatalf("metrics = %#v", report.Metrics)
	}
}

func TestEvaluateRetainsUnknownCoordinates(t *testing.T) {
	source := testSource + "# experiment.unknown stage=scope-resolution step=resolve-generated-binding reason=scope-provenance-absent\n"
	report, err := Evaluate(fstest.MapFS{"unknown.gooo": {Data: []byte(source)}}, "unknown.gooo", strings.Repeat("b", 40))
	if err != nil {
		t.Fatal(err)
	}
	if err := Validate(report, true, strings.Repeat("b", 40)); err != nil {
		t.Fatal(err)
	}
	if report.Decision != DecisionUnknown || report.Unknowns[0].Stage != "scope-resolution" {
		t.Fatalf("unknown report = %#v", report)
	}
	if report.Metrics.OpenClaimTotal != 1 || report.Claims[len(report.Claims)-1].Status != StatusOpen {
		t.Fatalf("unknown claims = %#v", report.Claims)
	}
}

func TestValidateRejectsChangedSafeResolution(t *testing.T) {
	report, err := Evaluate(fstest.MapFS{"main.gooo": {Data: []byte(testSource)}}, "main.gooo", strings.Repeat("c", 40))
	if err != nil {
		t.Fatal(err)
	}
	report.Cases[1].ResolvedIdentity = ConsumerBinding
	if err := Validate(report, false, strings.Repeat("c", 40)); err == nil {
		t.Fatal("changed safe resolution was accepted")
	}
}
