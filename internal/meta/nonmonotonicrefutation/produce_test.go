package nonmonotonicrefutation

import "testing"

func TestCanonicalProducerKeepsThreeRevisionCases(t *testing.T) {
	report := Produce("examples/nonmonotonic-refutation/main.gooo", []byte("package nonmonotonicrefutation\nactivity Observe\n"))
	if report.Contract.FixedClaimTotal != 3 || report.Contract.FixedTransitionTotal != 6 {
		t.Fatalf("contract = %#v", report.Contract)
	}
	if len(report.Contract.Cases) != 3 || report.Effects.RepositoryWrites != 0 || report.Effects.MutationAuthority {
		t.Fatalf("producer report = %#v", report)
	}
	if len(report.Contract.Cases[1].Evidence) != 2 || report.Contract.Cases[1].ExpectedFinalStatus != StatusRefuted {
		t.Fatalf("counterevidence case = %#v", report.Contract.Cases[1])
	}
	if len(report.Contract.Cases[2].Evidence) != 3 || report.Contract.Cases[2].ExpectedFinalStatus != StatusDischarged {
		t.Fatalf("re-evaluation case = %#v", report.Contract.Cases[2])
	}
}
