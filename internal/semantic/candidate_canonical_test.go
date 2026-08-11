package semantic

import (
	"errors"
	"testing"
)

func TestCandidateFactsStayOutOfAuthoritativeCanonical(t *testing.T) {
	for repetition := 0; repetition < 10; repetition++ {
		reverse := repetition%2 == 1
		authoritative := candidateHashIR(t, false, reverse)
		observed := candidateHashIR(t, true, reverse)
		if err := authoritative.Validate(); err != nil {
			t.Fatal(err)
		}
		if err := observed.Validate(); err != nil {
			t.Fatal(err)
		}
		if authoritative.SemanticCanonical() != observed.SemanticCanonical() {
			t.Fatal("candidate fact entered authoritative semantic canonical form")
		}
		if authoritative.StableHash() != observed.StableHash() {
			t.Fatal("candidate fact changed authoritative stable hash")
		}
		comparison := CompareIR(authoritative, observed)
		if !comparison.Equivalent() {
			t.Fatalf("candidate-only observation changed IR comparison: %#v", comparison)
		}
		if len(observed.Graph.Candidates()) != 1 || observed.Graph.Canonical() == authoritative.Graph.Canonical() {
			t.Fatal("candidate observation was not retained in the audit projection")
		}
	}
}

func TestCandidatePromotionChangesAuthoritativeCanonicalExplicitly(t *testing.T) {
	ir := candidateHashIR(t, true, false)
	candidate := candidateHashFact()
	before := ir.StableHash()
	if _, err := ir.Graph.PromoteCandidate(candidate.Key()); err != nil {
		t.Fatal(err)
	}
	if len(ir.Graph.Candidates()) != 0 || len(ir.Graph.Facts()) != 2 {
		t.Fatalf("promotion state = facts:%d candidates:%d", len(ir.Graph.Facts()), len(ir.Graph.Candidates()))
	}
	if ir.StableHash() == before {
		t.Fatal("explicit candidate promotion did not change authoritative hash")
	}
	expected := candidateHashIR(t, false, false)
	candidate.Status = FactDeterministic
	if err := expected.AddFact(candidate); err != nil {
		t.Fatal(err)
	}
	if ir.StableHash() != expected.StableHash() {
		t.Fatal("promoted authoritative hash differs from deterministic graph")
	}
}

func TestSameEdgeEvidenceRetainsDistinctSpans(t *testing.T) {
	ir := candidateHashIR(t, false, false)
	fact := NewUsedFact(
		MustIdentity("candidate-hash://activity/compile"),
		MustIdentity("candidate-hash://entity/source"),
	)
	before := ir.StableHash()
	beforeEvidenceHash := ir.EvidenceHash()
	first, err := NewEvidence(
		MustIdentity("candidate-hash://evidence/first"), GoVerifierID,
		VerificationEvidence, fact.Key(), StableHashString("edge evidence"),
	)
	if err != nil {
		t.Fatal(err)
	}
	first = first.WithSpan(Span{File: "first.gooo", Start: Position{Offset: 1}, End: Position{Offset: 5}})
	second, err := NewEvidence(
		MustIdentity("candidate-hash://evidence/second"), GoVerifierID,
		VerificationEvidence, fact.Key(), StableHashString("edge evidence"),
	)
	if err != nil {
		t.Fatal(err)
	}
	second = second.WithSpan(Span{File: "second.gooo", Start: Position{Offset: 8}, End: Position{Offset: 13}})
	if err := ir.AddEvidence(first); err != nil {
		t.Fatal(err)
	}
	if err := ir.AddEvidence(second); err != nil {
		t.Fatal(err)
	}
	if err := ir.Validate(); err != nil {
		t.Fatal(err)
	}
	if ir.StableHash() != before || ir.EvidenceHash() == beforeEvidenceHash || len(ir.Evidence()) != 2 {
		t.Fatal("evidence records changed the semantic edge or were not retained")
	}
	if ir.Evidence()[0].Span == ir.Evidence()[1].Span {
		t.Fatal("distinct evidence spans were collapsed")
	}
	conflict := first.WithSpan(Span{File: "changed.gooo", Start: Position{Offset: 20}, End: Position{Offset: 25}})
	if err := ir.AddEvidence(conflict); !errors.Is(err, ErrEvidenceConflict) {
		t.Fatalf("same evidence ID with changed span error = %v, want ErrEvidenceConflict", err)
	}
}

func TestRejectedFactAndNodeOperationsDoNotMutateGraph(t *testing.T) {
	ns := Namespace("candidate-hash")
	activity := mustActivity(t, MustIdentity("candidate-hash://activity/compile"), ns, "Compile")
	entity := mustEntity(t, MustIdentity("candidate-hash://entity/source"), ns, "Source")
	output := mustEntity(t, MustIdentity("candidate-hash://entity/output"), ns, "Output")
	verify := mustActivity(t, MustIdentity("candidate-hash://activity/verify"), ns, "Verify")
	agent := mustAgent(t, MustIdentity("candidate-hash://agent/verifier"), ns, "Verifier")
	cases := []struct {
		name string
		add  func(*Graph) error
		want error
	}{
		{"used direction", func(g *Graph) error { return g.AddFact(NewUsedFact(entity.ID, activity.ID)) }, ErrInvalidFact},
		{"generated direction", func(g *Graph) error { return g.AddFact(NewWasGeneratedByFact(activity.ID, output.ID)) }, ErrInvalidFact},
		{"derived direction", func(g *Graph) error { return g.AddFact(NewWasDerivedFromFact(output.ID, activity.ID)) }, ErrInvalidFact},
		{"associated direction", func(g *Graph) error { return g.AddFact(NewWasAssociatedWithFact(verify.ID, output.ID)) }, ErrInvalidFact},
		{"kind rekey", func(g *Graph) error {
			return g.AddNode(Node{ID: activity.ID, Kind: Entity, Namespace: ns, Name: "Compile"})
		}, ErrIdentityConflict},
		{"namespace rekey", func(g *Graph) error {
			return g.AddNode(Node{ID: activity.ID, Kind: Activity, Namespace: Namespace("other"), Name: "Compile"})
		}, ErrIdentityConflict},
		{"name collision", func(g *Graph) error {
			return g.AddNode(Node{ID: MustIdentity("candidate-hash://entity/other"), Kind: Entity, Namespace: ns, Name: "Source"})
		}, ErrNameCollision},
	}
	for _, test := range cases {
		graph := NewGraph()
		for _, node := range []Node{activity, entity, output, verify, agent} {
			if err := graph.AddNode(node); err != nil {
				t.Fatal(err)
			}
		}
		before := graph.Canonical()
		if err := test.add(&graph); !errors.Is(err, test.want) {
			t.Errorf("%s error = %v, want %v", test.name, err, test.want)
		}
		if graph.Canonical() != before {
			t.Errorf("%s mutated graph after rejection", test.name)
		}
	}
}

func TestNamespaceQualifiedNamesAllowDistinctNamespaces(t *testing.T) {
	firstNS := Namespace("first")
	secondNS := Namespace("second")
	graph := NewGraph()
	for _, node := range []Node{
		mustEntity(t, MustIdentity("namespace://entity/first"), firstNS, "Record"),
		mustEntity(t, MustIdentity("namespace://entity/second"), secondNS, "Record"),
	} {
		if err := graph.AddNode(node); err != nil {
			t.Fatal(err)
		}
	}
	if err := graph.Validate(); err != nil {
		t.Fatal(err)
	}
	if _, ok := graph.NodeByName(firstNS, "Record"); !ok {
		t.Fatal("first namespace-qualified name did not resolve")
	}
	if _, ok := graph.NodeByName(secondNS, "Record"); !ok {
		t.Fatal("second namespace-qualified name did not resolve")
	}
}

func TestIdentityLabelRenamePreservesMeaningAndEvidence(t *testing.T) {
	ir := provFixture(t, GoHostedCompilerID)
	beforeSemantic := ir.StableHash()
	beforeEvidence := ir.EvidenceHash()
	ns := ir.Namespace
	sourceID := MustIdentity("prov-fixture://entity/source")
	if err := ir.AddNode(Node{
		ID: sourceID, Kind: Entity, Namespace: ns, Name: "Renamed source", Aliases: []string{"Input"},
	}); err != nil {
		t.Fatal(err)
	}
	if ir.StableHash() != beforeSemantic || ir.EvidenceHash() != beforeEvidence {
		t.Fatal("identity label rename changed semantic meaning or evidence set")
	}
	if _, ok := ir.Graph.NodeByName(ns, "Source"); ok {
		t.Fatal("old label remained after rename")
	}
	if node, ok := ir.Graph.NodeByName(ns, "Input"); !ok || node.ID != sourceID {
		t.Fatal("new alias did not resolve to stable identity")
	}
}

func TestIdentityRekeyAuthorizationIsDeferred(t *testing.T) {
	t.Log("DEFERRED: semantic-ir/v1 has no ID continuity or rekey authorization contract")
}

func TestCandidateEvidenceRemainsCandidateAfterGraphPromotion(t *testing.T) {
	ir := candidateHashIR(t, true, false)
	fact := candidateHashFact()
	evidence, err := NewEvidence(
		MustIdentity("candidate-hash://evidence/observation"),
		GoHostedCompilerID, CompilerRunEvidence, fact.Key(),
		StableHashString("candidate observation"),
	)
	if err != nil {
		t.Fatal(err)
	}
	evidence.Status = FactCandidate
	if err := ir.AddEvidence(evidence); err != nil {
		t.Fatal(err)
	}
	if err := ir.Validate(); err != nil {
		t.Fatal(err)
	}
	if _, err := ir.Graph.PromoteCandidate(fact.Key()); err != nil {
		t.Fatal(err)
	}
	if len(ir.Evidence()) != 1 || ir.Evidence()[0].Status != FactCandidate {
		t.Fatal("candidate evidence was erased or silently reclassified")
	}
	if err := ir.Validate(); !errors.Is(err, ErrGraphInvalid) {
		t.Fatalf("promoted graph accepted stale candidate evidence: %v", err)
	}
}

func TestPROVAttributionIsExplicitlyUnsupported(t *testing.T) {
	relation := Relation("wasAttributedTo")
	if relation.Valid() {
		t.Fatal("unsupported attribution relation became valid implicitly")
	}
	graph := NewGraph()
	fact := NewFact(
		MustIdentity("prov-guard://entity/output"), relation,
		MustIdentity("prov-guard://agent/owner"),
	)
	if err := graph.AddFact(fact); !errors.Is(err, ErrUnknownRelation) {
		t.Fatalf("unsupported attribution error = %v, want ErrUnknownRelation", err)
	}
	if len(graph.AllFacts()) != 0 {
		t.Fatal("rejected attribution mutated graph")
	}
}

func candidateHashIR(t *testing.T, includeCandidate, reverse bool) IR {
	t.Helper()
	ns := Namespace("candidate-hash")
	ir := NewIR("candidate-hash", ns)
	nodes := []Node{
		mustEntity(t, MustIdentity("candidate-hash://entity/source"), ns, "Source"),
		mustEntity(t, MustIdentity("candidate-hash://entity/output"), ns, "Output"),
		mustActivity(t, MustIdentity("candidate-hash://activity/compile"), ns, "Compile"),
	}
	facts := []Fact{NewUsedFact(nodes[2].ID, nodes[0].ID)}
	if includeCandidate {
		facts = append(facts, candidateHashFact())
	}
	if reverse {
		reverseNodes(nodes)
		reverseFacts(facts)
	}
	for _, node := range nodes {
		if err := ir.AddNode(node); err != nil {
			t.Fatal(err)
		}
	}
	for _, fact := range facts {
		if includeCandidate && fact.Status == FactCandidate {
			if err := ir.AddCandidate(fact); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if err := ir.AddFact(fact); err != nil {
			t.Fatal(err)
		}
	}
	return ir
}

func candidateHashFact() Fact {
	return NewCandidateFact(
		MustIdentity("candidate-hash://activity/compile"), Used,
		MustIdentity("candidate-hash://entity/output"),
		"observed output dependency",
	)
}
