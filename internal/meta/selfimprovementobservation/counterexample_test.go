package selfimprovementobservation

import "testing"

func TestBuildFailsClosedForSemanticCounterexamples(t *testing.T) {
	tests := []struct {
		name       string
		mutate     func(*Inputs)
		resolution string
		reason     string
	}{
		{"unknown decision", func(in *Inputs) { in.Report.Value.Decision = "MAYBE"; reseal(&in.Report.Value) }, "LOWER_RESOLUTION", "SOURCE_DECISION_UNKNOWN"},
		{"explicit failure", func(in *Inputs) { in.Report.Value.Decision = "FAIL_CLOSED"; reseal(&in.Report.Value) }, "EXACT", "SOURCE_EXPLICITLY_REJECTED"},
		{"copied digest", func(in *Inputs) { in.Report.Value.Summary.Resources.MaxRSSKiB++ }, "EXACT", "SOURCE_REPORT_DIGEST_INVALID"},
		{"source write", func(in *Inputs) { in.Report.Value.Summary.Effects.RepositoryWrites = 1; reseal(&in.Report.Value) }, "EXACT", "SOURCE_EFFECTS_REJECTED"},
		{"counterexample gap", func(in *Inputs) { in.Counterexamples.Value.Satisfied = 5 }, "EXACT", "COUNTEREXAMPLE_COVERAGE_REJECTED"},
		{"contract gap", func(in *Inputs) { in.Contract.Value.Indicators[6].Verdict = "FAIL" }, "EXACT", "GOOO_OBSERVATION_CONTRACT_REJECTED"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			in, opts := validFixture()
			test.mutate(&in)
			observation := Build(in, opts)
			if observation.Decision != "FAIL_CLOSED" || observation.Resolution != test.resolution || observation.Reason != test.reason {
				t.Fatalf("decision/resolution/reason = %s/%s/%s", observation.Decision, observation.Resolution, observation.Reason)
			}
			if observation.Summary.CandidateCount != 0 || observation.Authority != (Authority{}) {
				t.Fatalf("failure gained authority: %#v", observation)
			}
		})
	}
}
