package query

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"testing"
)

func evidenceFixture(t *testing.T, reverse bool) semantic.IR {
	t.Helper()
	ir := semantic.NewIR("fixture", "fixture")
	activity, err := semantic.NewActivity("fixture://activity/compile", "fixture", "Compile")
	if err != nil {
		t.Fatal(err)
	}
	source, err := semantic.NewEntity("fixture://entity/source", "fixture", "Source")
	if err != nil {
		t.Fatal(err)
	}
	output, err := semantic.NewEntity("fixture://entity/output", "fixture", "Output")
	if err != nil {
		t.Fatal(err)
	}
	nodes := []semantic.Node{activity, source, output}
	facts := []semantic.Fact{
		semantic.NewUsedFact(activity.ID, source.ID),
		semantic.NewWasGeneratedByFact(output.ID, activity.ID),
	}
	if reverse {
		nodes[0], nodes[2] = nodes[2], nodes[0]
		facts[0], facts[1] = facts[1], facts[0]
	}
	for _, node := range nodes {
		if err := ir.AddNode(node); err != nil {
			t.Fatal(err)
		}
	}
	digest := semantic.StableHashString("fixture payload/v1")
	for _, fact := range facts {
		if err := ir.AddFact(fact); err != nil {
			t.Fatal(err)
		}
		evidence, err := semantic.NewEvidence(
			semantic.MustIdentity("fixture://evidence/"+string(fact.Predicate)),
			semantic.GoHostedCompilerID, semantic.CompilerRunEvidence, fact.Key(), digest,
		)
		if err != nil {
			t.Fatal(err)
		}
		if err := ir.AddEvidence(evidence); err != nil {
			t.Fatal(err)
		}
	}
	return ir
}
func projectionLabel(metadata ProjectionMetadata, view string) AuthorityLabel {
	for _, label := range metadata.AuthorityLabels {
		if label.View == view {
			return label
		}
	}
	return AuthorityLabel{}
}
