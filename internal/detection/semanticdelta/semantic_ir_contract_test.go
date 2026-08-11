package semanticdelta

import (
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func TestSemanticIRAdapterExcludesCandidateFacts(t *testing.T) {
	base := selfHostingFixtureIR(t, semantic.GoHostedCompilerID)
	after, err := base.Normalized()
	if err != nil {
		t.Fatal(err)
	}
	candidate := semantic.NewCandidateFact(
		semantic.MustIdentity("gooo://activity/compile"), semantic.Used,
		semantic.MustIdentity("gooo://entity/uncertain"), "candidate implementation reference",
	)
	if err := after.AddCandidate(candidate); err != nil {
		t.Fatal(err)
	}
	adapter := semanticIRAdapter()
	snapshot, err := adapter.Snapshot(after)
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Facts) != 1 {
		t.Fatalf("candidate fact crossed adapter boundary: %#v", snapshot.Facts)
	}
	delta, err := adapter.Diff(base, after)
	if err != nil {
		t.Fatal(err)
	}
	if !delta.IsEmpty() {
		t.Fatalf("candidate-only change became semantic delta: %#v", delta)
	}
}

func TestSemanticIRAdapterComparesHostClaimsWithoutClaimingGoooExecution(t *testing.T) {
	goIR := selfHostingFixtureIR(t, semantic.GoHostedCompilerID)
	goooIR := selfHostingFixtureIR(t, semantic.GoooHostedCompilerID)
	adapter := semanticIRAdapter()
	goSnapshot, err := adapter.Snapshot(goIR)
	if err != nil {
		t.Fatal(err)
	}
	goooSnapshot, err := adapter.Snapshot(goooIR)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(goSnapshot, goooSnapshot) {
		t.Fatalf("host-specific evidence changed semantic snapshot:\nGo=%#v\nGooo=%#v", goSnapshot, goooSnapshot)
	}
	if !goIR.SemanticallyEquivalent(goooIR) || !goIR.ProvenanceEquivalent(goooIR) {
		t.Fatal("host claims were not comparable after normalization")
	}
	if goIR.EvidenceHash() == goooIR.EvidenceHash() {
		t.Fatal("exact evidence hash erased producer identity")
	}
	if goIR.ProvenanceHash() != goooIR.ProvenanceHash() {
		t.Fatal("comparison evidence changed with producer identity")
	}
	// The fixture models a future comparison contract only. It does not assert
	// that a gooo-hosted compiler has executed or passed any promotion stage.
}

func semanticIRAdapter() Adapter[semantic.IR] {
	return Adapter[semantic.IR]{
		Nodes: func(ir semantic.IR) ([]Node, error) {
			normalized, err := ir.Normalized()
			if err != nil {
				return nil, err
			}
			nodes := make([]Node, 0, len(normalized.Graph.Nodes()))
			for _, node := range normalized.Graph.Nodes() {
				nodes = append(nodes, Node{ID: node.ID.String(), Kind: node.Kind.String()})
			}
			return nodes, nil
		},
		Facts: func(ir semantic.IR) ([]Fact, error) {
			normalized, err := ir.Normalized()
			if err != nil {
				return nil, err
			}
			facts := make([]Fact, 0, len(normalized.Graph.DeterministicFacts()))
			for _, fact := range normalized.Graph.DeterministicFacts() {
				facts = append(facts, Fact{
					Subject: fact.Subject.String(), Predicate: fact.Predicate.String(), Object: fact.Object.String(),
				})
			}
			return facts, nil
		},
	}
}

func selfHostingFixtureIR(t *testing.T, producer semantic.ID) semantic.IR {
	t.Helper()
	ir := semantic.NewIR("self-hosting-fixture", semantic.Namespace("self-host"))
	activity := semantic.MustIdentity("gooo://activity/compile")
	entity := semantic.MustIdentity("gooo://entity/source")
	uncertain := semantic.MustIdentity("gooo://entity/uncertain")
	activityNode, err := semantic.NewActivity(activity, ir.Namespace, "Compile")
	if err != nil {
		t.Fatal(err)
	}
	entityNode, err := semantic.NewEntity(entity, ir.Namespace, "Source")
	if err != nil {
		t.Fatal(err)
	}
	uncertainNode, err := semantic.NewEntity(uncertain, ir.Namespace, "Uncertain")
	if err != nil {
		t.Fatal(err)
	}
	for _, node := range []semantic.Node{activityNode, entityNode, uncertainNode} {
		if err := ir.AddNode(node); err != nil {
			t.Fatal(err)
		}
	}
	fact := semantic.NewUsedFact(activity, entity)
	if err := ir.AddFact(fact); err != nil {
		t.Fatal(err)
	}
	evidence, err := semantic.NewEvidence(
		semantic.MustIdentity("gooo://evidence/compile"), producer,
		semantic.CompilerRunEvidence, fact.Key(), semantic.StableHashString("fixture"),
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := ir.AddEvidence(evidence); err != nil {
		t.Fatal(err)
	}
	return ir
}
