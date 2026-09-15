package semantic

import (
	"errors"
	"testing"
)

func TestPROVFixtureCoversEvidenceForEveryCoreRelation(t *testing.T) {
	ir := provFixture(t, GoHostedCompilerID)
	if err := ir.Validate(); err != nil {
		t.Fatal(err)
	}
	if got := len(ir.Graph.Facts()); got != 4 {
		t.Fatalf("PROV fixture facts = %d, want 4", got)
	}
	if got := len(ir.Evidence()); got != 4 {
		t.Fatalf("PROV fixture evidence = %d, want 4", got)
	}
	counts := make(map[Relation]int)
	for _, fact := range ir.Graph.Facts() {
		counts[fact.Predicate]++
	}
	for _, relation := range []Relation{Used, WasGeneratedBy, WasDerivedFrom, WasAssociatedWith} {
		if counts[relation] != 1 {
			t.Fatalf("PROV relation %s count = %d, want 1", relation, counts[relation])
		}
	}
	const wantDigest = "a4955dfa90475116d85286f95884b718409d0f3afe51e567d77cac09cab4033d"
	if got := ir.StableHash(); got != wantDigest {
		t.Fatalf("PROV fixture digest = %s, want %s", got, wantDigest)
	}
}
func TestPROVFixtureComparesHostsWithoutErasingProducerEvidence(t *testing.T) {
	goIR := provFixture(t, GoHostedCompilerID)
	goooIR := provFixture(t, GoooHostedCompilerID)
	if !goIR.SemanticallyEquivalent(goooIR) {
		t.Fatal("host-specific PROV fixtures changed semantic meaning")
	}
	if !goIR.ProvenanceEquivalent(goooIR) {
		t.Fatal("host-specific PROV fixtures changed comparable provenance")
	}
	if goIR.EvidenceHash() == goooIR.EvidenceHash() {
		t.Fatal("exact evidence hash erased producer identity")
	}
}
func TestPROVRelationKindMatrixRejectsInvalidEdges(t *testing.T) {
	cases := []struct {
		relation        Relation
		subject, object Kind
	}{
		{Used, Entity, Entity},
		{WasGeneratedBy, Activity, Entity},
		{WasDerivedFrom, Entity, Activity},
		{WasAssociatedWith, Activity, Entity},
	}
	for _, test := range cases {
		err := test.relation.ValidateKinds(test.subject, test.object)
		if !errors.Is(err, ErrInvalidFact) {
			t.Errorf("%s(%s, %s) error = %v, want ErrInvalidFact", test.relation, test.subject, test.object, err)
		}
	}
}
