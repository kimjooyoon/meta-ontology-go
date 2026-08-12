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

func TestDuplicateCandidateFactsArePermutationStable(t *testing.T) {
	ns := Namespace("candidate-duplicates")
	activity := mustActivity(t, MustIdentity("candidate-duplicates://activity/run"), ns, "Run")
	entity := mustEntity(t, MustIdentity("candidate-duplicates://entity/output"), ns, "Output")
	first := NewCandidateFact(activity.ID, Used, entity.ID, "z observation").WithSpan(Span{
		File: "z.gooo", Start: Position{Offset: 20}, End: Position{Offset: 24},
	})
	second := NewCandidateFact(activity.ID, Used, entity.ID, "a observation").WithSpan(Span{
		File: "a.gooo", Start: Position{Offset: 2}, End: Position{Offset: 6},
	})

	build := func(facts ...Fact) Graph {
		graph := NewGraph()
		for _, node := range []Node{activity, entity} {
			if err := graph.AddNode(node); err != nil {
				t.Fatal(err)
			}
		}
		for _, fact := range facts {
			if err := graph.AddCandidate(fact); err != nil {
				t.Fatal(err)
			}
		}
		return graph
	}

	left := build(first, second, first)
	right := build(second, first, second)
	if len(left.Candidates()) != 1 || len(right.Candidates()) != 1 {
		t.Fatalf("duplicate candidates were not collapsed: left=%d right=%d", len(left.Candidates()), len(right.Candidates()))
	}
	if left.Canonical() != right.Canonical() || left.StableHash() != right.StableHash() {
		t.Fatal("duplicate candidate canonical/hash changed with insertion permutation")
	}
	if left.Candidates()[0].Reason != "a observation" || left.Candidates()[0].Span.File != "a.gooo" {
		t.Fatalf("duplicate candidate merge was not deterministic: %#v", left.Candidates()[0])
	}
}

func TestCandidateEvidenceHashIsPermutationStableAndTamperSafe(t *testing.T) {
	ns := Namespace("candidate-evidence-hash")
	activity := mustActivity(t, MustIdentity("candidate-evidence-hash://activity/run"), ns, "Run")
	entity := mustEntity(t, MustIdentity("candidate-evidence-hash://entity/output"), ns, "Output")
	fact := NewCandidateFact(activity.ID, Used, entity.ID, "observed dependency")

	build := func(order []string) IR {
		ir := NewIR("candidate-evidence-hash", ns)
		for _, node := range []Node{activity, entity} {
			if err := ir.AddNode(node); err != nil {
				t.Fatal(err)
			}
		}
		if err := ir.AddCandidate(fact); err != nil {
			t.Fatal(err)
		}
		for _, id := range order {
			evidence, err := NewEvidence(
				MustIdentity("candidate-evidence-hash://evidence/"+id), GoHostedCompilerID,
				CompilerRunEvidence, fact.Key(), StableHashString("candidate evidence "+id),
			)
			if err != nil {
				t.Fatal(err)
			}
			evidence.Status = FactCandidate
			if err := ir.AddEvidence(evidence); err != nil {
				t.Fatal(err)
			}
		}
		return ir
	}

	left := build([]string{"b", "a"})
	right := build([]string{"a", "b"})
	if left.EvidenceCanonical() != right.EvidenceCanonical() || left.EvidenceHash() != right.EvidenceHash() {
		t.Fatal("candidate evidence canonical/hash changed with insertion permutation")
	}
	tampered := left.Evidence()[0]
	tampered.Kind = VerificationEvidence
	beforeCanonical, beforeHash := left.Canonical(), left.EvidenceHash()
	if err := left.AddEvidence(tampered); !errors.Is(err, ErrInvalidEvidence) {
		t.Fatalf("tampered candidate evidence error = %v, want ErrInvalidEvidence", err)
	}
	if left.Canonical() != beforeCanonical || left.EvidenceHash() != beforeHash {
		t.Fatal("tampered candidate evidence changed canonical or hash")
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
		{"alias collision", func(g *Graph) error {
			return g.AddNode(Node{ID: MustIdentity("candidate-hash://entity/alias"), Kind: Entity, Namespace: ns, Name: "Other", Aliases: []string{"Source"}})
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
	contract := readDeferredContract(t, "authorized-rekey")
	if contract.Status != "deferred" || contract.Authoritative {
		t.Fatalf("authorized rekey contract = %#v, want non-authoritative deferred", contract)
	}
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
	beforeEvidenceHash := ir.EvidenceHash()
	if err := ir.Validate(); err != nil {
		t.Fatalf("retained candidate evidence invalidated promoted graph: %v", err)
	}
	if ir.EvidenceHash() != beforeEvidenceHash {
		t.Fatal("promotion changed the retained candidate evidence digest")
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
