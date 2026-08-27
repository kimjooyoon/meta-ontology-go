package nonmonotonicrefutation

import (
	"os"
	"testing"
)

func TestSourceBackedProducerKeepsThreeClaimsAndSixObservations(t *testing.T) {
	source, err := os.ReadFile("../../../examples/nonmonotonic-refutation/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	report, err := Produce("examples/nonmonotonic-refutation/main.gooo", source, 0)
	if err != nil {
		t.Fatal(err)
	}
	if report.Contract.FixedCaseTotal != 3 || report.Contract.FixedClaimTotal != 3 || report.Contract.FixedObservationTotal != 6 || report.Contract.FixedLedgerRowTotal != 6 {
		t.Fatalf("contract = %#v", report.Contract)
	}
	if len(report.Contract.Claims) != 3 || len(report.Contract.Observations) != 6 || report.Effects.RepositoryWrites != 0 || report.Effects.MutationAuthority {
		t.Fatalf("producer report = %#v", report)
	}
	if report.Contract.Claims[0].Proposition != "equals:alpha:input-alpha:1" || report.Contract.Claims[0].Subject != "alpha" || report.Contract.Claims[0].Input != "input-alpha" || report.Contract.Observations[2].ObservedValue != "0" || report.SourceBindingDigest == "" {
		t.Fatalf("source observations = %#v", report.Contract.Observations)
	}
}
