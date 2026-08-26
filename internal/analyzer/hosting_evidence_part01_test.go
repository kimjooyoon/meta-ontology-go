package analyzer

import (
	"reflect"
	"testing"
)

func TestHostingViewsSeparateFactsCandidatesAndImplementation(t *testing.T) {
	report := hostingPairReport(t)
	facts := report.AuthoritativeFacts()
	candidates := report.CandidateEvidence()
	details := report.ImplementationEvidence()
	if len(facts) != 3 || len(candidates) != 1 || len(details) != 3 {
		t.Fatalf("evidence views = facts %d, candidates %d, implementation %d; want 3, 1, 3",
			len(facts), len(candidates), len(details))
	}
	for _, fact := range facts {
		if fact.Kind != EvidenceKindFact || fact.Status != EvidenceStatusDeterministic {
			t.Errorf("authoritative fact = %#v, want deterministic fact", fact)
		}
	}
	candidate := candidates[0]
	if candidate.Kind != EvidenceKindCandidate || candidate.Status != EvidenceStatusCandidate {
		t.Fatalf("candidate view = %#v, want candidate evidence", candidate)
	}
	wantOptions := []Identity{
		NewIdentity("fraud", "fraud://activity/check"),
		NewIdentity("security", "security://activity/check"),
	}
	if !reflect.DeepEqual(candidate.Options, wantOptions) {
		t.Fatalf("candidate options = %#v, want %#v", candidate.Options, wantOptions)
	}
	if candidate.Subject.ID == facts[0].Subject.ID && candidate.Kind == facts[0].Kind {
		t.Fatal("candidate was mixed into authoritative facts")
	}

	candidates[0].Options[0].ID = "mutated"
	if report.CandidateEvidence()[0].Options[0].ID == "mutated" {
		t.Fatal("candidate view exposed report-owned options")
	}
}
