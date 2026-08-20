package semantic

import (
	"errors"
	"testing"
)

func TestCandidateEvidenceCannotStandInForAuthoritativeFact(t *testing.T) {
	ns := Namespace("bootstrap")
	activity := mustActivity(t, MustIdentity("bootstrap://activity/compile"), ns, "Compile")
	entity := mustEntity(t, MustIdentity("bootstrap://entity/source"), ns, "Source")
	graph := NewGraph()
	for _, node := range []Node{activity, entity} {
		if err := graph.AddNode(node); err != nil {
			t.Fatal(err)
		}
	}
	fact := NewCandidateFact(activity.ID, Used, entity.ID, "observed but not independently verified")
	if err := graph.AddCandidate(fact); err != nil {
		t.Fatal(err)
	}
	digest := StableHashString("candidate evidence")
	candidateEvidence, err := NewEvidence(MustIdentity("bootstrap://evidence/candidate"), GoHostedCompilerID, CompilerRunEvidence, fact.Key(), digest)
	if err != nil {
		t.Fatal(err)
	}
	candidateEvidence.Status = FactCandidate
	if err := candidateEvidence.ValidateAgainst(graph); err != nil {
		t.Fatalf("candidate evidence did not match candidate fact: %v", err)
	}
	authoritativeEvidence, err := NewEvidence(MustIdentity("bootstrap://evidence/authoritative"), GoVerifierID, VerificationEvidence, fact.Key(), digest)
	if err != nil {
		t.Fatal(err)
	}
	if err := authoritativeEvidence.ValidateAgainst(graph); !errors.Is(err, ErrInvalidEvidence) {
		t.Fatalf("authoritative evidence crossed candidate boundary: %v", err)
	}
	if _, err := graph.PromoteCandidate(fact.Key()); err != nil {
		t.Fatal(err)
	}
	if err := candidateEvidence.ValidateAgainst(graph); err != nil {
		t.Fatalf("retained candidate evidence failed after promotion: %v", err)
	}
	if err := authoritativeEvidence.ValidateAgainst(graph); err != nil {
		t.Fatalf("authoritative evidence did not match promoted fact: %v", err)
	}
}
