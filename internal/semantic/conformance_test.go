package semantic

import (
	"errors"
	"os"
	"testing"
)

func TestPROVConformanceCoversKindsAndCoreRelations(t *testing.T) {
	ns := Namespace("bootstrap")
	source := mustEntity(t, MustIdentity("bootstrap://entity/source"), ns, "Source")
	output := mustEntity(t, MustIdentity("bootstrap://entity/output"), ns, "Output")
	compile := mustActivity(t, MustIdentity("bootstrap://activity/compile"), ns, "Compile")
	verify := mustActivity(t, MustIdentity("bootstrap://activity/verify"), ns, "Verify")
	ci := mustAgent(t, GoVerifierID, ns, "Go verifier")
	graph := NewGraph()
	for _, node := range []Node{source, output, compile, verify, ci} {
		if err := graph.AddNode(node); err != nil {
			t.Fatal(err)
		}
	}
	facts := []Fact{
		NewUsedFact(compile.ID, source.ID),
		NewWasGeneratedByFact(output.ID, compile.ID),
		NewWasDerivedFromFact(output.ID, source.ID),
		NewWasAssociatedWithFact(verify.ID, ci.ID),
	}
	for _, fact := range facts {
		if err := graph.AddFact(fact); err != nil {
			t.Fatal(err)
		}
	}
	if err := graph.Validate(); err != nil {
		t.Fatal(err)
	}
	if got := len(graph.Facts()); got != len(facts) {
		t.Fatalf("deterministic PROV fact count = %d, want %d", got, len(facts))
	}
}

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

func TestEvidenceFreshnessFixtureRejectsChangedPayload(t *testing.T) {
	payload, err := os.ReadFile("testdata/evidence-freshness.txt")
	if err != nil {
		t.Fatal(err)
	}
	const wantDigest = "9ea86bfd939cc1218a0e2352a4623282e0d44739c88fe3f97f8f762efee68de0"
	if got := StableHash(payload); got != wantDigest {
		t.Fatalf("freshness fixture digest = %s, want %s", got, wantDigest)
	}
	fact := FactKey{
		Subject: MustIdentity("bootstrap://activity/verify"), Predicate: Used,
		Object: MustIdentity("bootstrap://entity/output"),
	}
	evidence, err := NewEvidence(
		MustIdentity("bootstrap://evidence/freshness"), GoVerifierID,
		VerificationEvidence, fact, StableHash(payload),
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := evidence.ValidateFresh(payload); err != nil {
		t.Fatalf("fresh fixture was rejected: %v", err)
	}
	stalePayload := append(append([]byte(nil), payload...), '\n')
	if err := evidence.ValidateFresh(stalePayload); !errors.Is(err, ErrStaleEvidence) {
		t.Fatalf("changed fixture was accepted: %v", err)
	}
}

func TestPROVCanonicalDigestIsPinnedAndOrderIndependent(t *testing.T) {
	left := conformanceGraph(t, false)
	right := conformanceGraph(t, true)
	if left.SemanticCanonical() != right.SemanticCanonical() {
		t.Fatal("PROV canonical form changed with insertion order")
	}
	const wantDigest = "aa0e5232c314cf31b63caf6268578c3dc7b7763f97b9600d911ba776c568b156"
	if got := left.StableHash(); got != wantDigest {
		t.Fatalf("PROV canonical digest = %s, want %s", got, wantDigest)
	}
}

func conformanceGraph(t *testing.T, reverse bool) Graph {
	t.Helper()
	ns := Namespace("bootstrap")
	nodes := []Node{
		mustEntity(t, MustIdentity("bootstrap://entity/source"), ns, "Renamed source"),
		mustEntity(t, MustIdentity("bootstrap://entity/output"), ns, "Renamed output"),
		mustActivity(t, MustIdentity("bootstrap://activity/compile"), ns, "Compiler run"),
		mustActivity(t, MustIdentity("bootstrap://activity/verify"), ns, "Verifier run"),
		mustAgent(t, GoVerifierID, ns, "Protected verifier"),
	}
	facts := []Fact{
		NewUsedFact(nodes[2].ID, nodes[0].ID),
		NewWasGeneratedByFact(nodes[1].ID, nodes[2].ID),
		NewWasDerivedFromFact(nodes[1].ID, nodes[0].ID),
		NewWasAssociatedWithFact(nodes[3].ID, nodes[4].ID),
	}
	if reverse {
		reverseNodes(nodes)
		reverseFacts(facts)
	}
	graph := NewGraph()
	for _, node := range nodes {
		if err := graph.AddNode(node); err != nil {
			t.Fatal(err)
		}
	}
	for _, fact := range facts {
		if err := graph.AddFact(fact); err != nil {
			t.Fatal(err)
		}
	}
	return graph
}

func reverseNodes(nodes []Node) {
	for left, right := 0, len(nodes)-1; left < right; left, right = left+1, right-1 {
		nodes[left], nodes[right] = nodes[right], nodes[left]
	}
}

func reverseFacts(facts []Fact) {
	for left, right := 0, len(facts)-1; left < right; left, right = left+1, right-1 {
		facts[left], facts[right] = facts[right], facts[left]
	}
}
