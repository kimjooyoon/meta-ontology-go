package valueexecution

import "testing"

func TestSelfImprovementExecutionHistoryPreservesUnknownEvidence(t *testing.T) {
	history := NewSelfImprovementExecutionHistory()
	observation := history.Observe()
	if observation.Decision != SelfImprovementExecutionHistoryDecisionInsufficientEvidence {
		t.Fatalf("decision=%q, want insufficient evidence", observation.Decision)
	}
	if observation.Observed != 0 || observation.Digest == "" {
		t.Fatalf("observation=%#v, want empty auditable history", observation)
	}
	if !observation.NonAuthorizing {
		t.Fatal("history observation must remain non-authorizing")
	}
}
