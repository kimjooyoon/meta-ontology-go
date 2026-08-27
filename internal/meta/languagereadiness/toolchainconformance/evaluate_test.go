package toolchainconformance

import "testing"

func TestEvaluateClosesAllToolchainSurfaces(t *testing.T) {
	report := Evaluate(fixtureInput(t))
	if err := Validate(report, fixtureHead); err != nil {
		t.Fatal(err)
	}
	summary := report.Summary
	if summary.SurfacesSatisfied != 9 || summary.CasesSatisfied != 179 ||
		summary.IndicatorsSatisfied != 151 || summary.ProofsPassed != 27 ||
		summary.TamperRejections != 13 || len(report.Indicators) != 28 {
		t.Fatalf("summary = %#v", summary)
	}
	if len(report.Proofs) != 3 || !report.Proofs[0].Passed ||
		!report.Proofs[1].Passed || !report.Proofs[2].Passed {
		t.Fatalf("proofs = %#v", report.Proofs)
	}
}

func TestEvaluateReplaysWithoutRepositoryWrites(t *testing.T) {
	first := Evaluate(fixtureInput(t))
	replay := Evaluate(fixtureInput(t))
	if first.ReportDigest != replay.ReportDigest ||
		first.RepositoryWrites != 0 || first.MutationAuthorized {
		t.Fatalf("first = %#v replay = %#v", first, replay)
	}
}
