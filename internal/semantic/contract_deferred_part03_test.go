package semantic

import (
	"testing"
)

func TestIdenticalDuplicateIDsAreIdempotent(t *testing.T) {
	ns := Namespace("idempotence")
	activity := mustActivity(t, MustIdentity("idempotence://activity/run"), ns, "Run")
	entity := mustEntity(t, MustIdentity("idempotence://entity/input"), ns, "Input")
	ir := NewIR("idempotence", ns)
	for _, node := range []Node{activity, entity} {
		if err := ir.AddNode(node); err != nil {
			t.Fatal(err)
		}
	}
	fact := NewUsedFact(activity.ID, entity.ID)
	if err := ir.AddFact(fact); err != nil {
		t.Fatal(err)
	}
	evidence, err := NewEvidence(
		MustIdentity("idempotence://evidence/run"), GoVerifierID,
		VerificationEvidence, fact.Key(), StableHashString("idempotent"),
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := ir.AddEvidence(evidence); err != nil {
		t.Fatal(err)
	}
	beforeCanonical := ir.Canonical()
	beforeSemantic := ir.StableHash()
	beforeEvidence := ir.EvidenceHash()
	if err := ir.AddNode(activity); err != nil {
		t.Fatal(err)
	}
	if err := ir.AddFact(fact); err != nil {
		t.Fatal(err)
	}
	if err := ir.AddEvidence(evidence); err != nil {
		t.Fatal(err)
	}
	if ir.Canonical() != beforeCanonical || ir.StableHash() != beforeSemantic || ir.EvidenceHash() != beforeEvidence {
		t.Fatal("identical duplicate IDs changed the IR")
	}
	if len(ir.Graph.Nodes()) != 2 || len(ir.Graph.Facts()) != 1 || len(ir.Evidence()) != 1 {
		t.Fatal("identical duplicate IDs were not idempotent")
	}
}
