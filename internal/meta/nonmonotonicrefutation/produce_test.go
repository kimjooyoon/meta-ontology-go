package nonmonotonicrefutation

import (
	"os"
	"testing"
)

func TestSourceBackedProducerKeepsThreeClaimsAndEightObservations(t *testing.T) {
	source, err := os.ReadFile("../../../examples/nonmonotonic-refutation/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	report, err := Produce("examples/nonmonotonic-refutation/main.gooo", source, true)
	if err != nil {
		t.Fatal(err)
	}
	if report.Contract.FixedCaseTotal != 3 || report.Contract.FixedClaimTotal != 3 || report.Contract.FixedObservationTotal != 8 || report.Contract.FixedLedgerRowTotal != 8 {
		t.Fatalf("contract = %#v", report.Contract)
	}
	if len(report.Contract.Claims) != 3 || len(report.Contract.Observations) != 8 || !report.Effects.NetRepositoryStatusUnchanged || report.Effects.RepositoryWriteObservation != "NONE_OBSERVED_IN_NET_STATUS" || report.Effects.MutationAuthorityResolution != "UNKNOWN" {
		t.Fatalf("producer report = %#v", report)
	}
	if report.Contract.Policy.ID == "" || report.Contract.Policy.PolicyDigest == "" || report.Contract.Claims[0].Proposition != "equals:alpha:input-alpha:1" || report.Contract.Claims[0].Subject != "alpha" || report.Contract.Claims[0].Input != "input-alpha" || report.Contract.Observations[2].ObservedValue != "0" || report.Contract.Observations[2].EvidenceDigest == "" || report.Contract.Observations[5].ObservedValue != "0" || report.Contract.Observations[6].RevisionRelation != RevisionSupersedes || report.Contract.Observations[7].RevisionRelation != RevisionSupersedes || report.SourceBindingDigest == "" {
		t.Fatalf("source observations = %#v", report.Contract.Observations)
	}
}
