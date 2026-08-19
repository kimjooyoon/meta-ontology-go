package semantic

import (
	"errors"
	"testing"
)

func TestCandidateEvidenceKindClosureRejectsTampering(t *testing.T) {
	ns := Namespace("candidate-evidence")
	activity := mustActivity(t, MustIdentity("candidate-evidence://activity/run"), ns, "Run")
	entity := mustEntity(t, MustIdentity("candidate-evidence://entity/output"), ns, "Output")
	ir := NewIR("candidate-evidence", ns)
	for _, node := range []Node{activity, entity} {
		if err := ir.AddNode(node); err != nil {
			t.Fatal(err)
		}
	}
	fact := NewCandidateFact(activity.ID, Used, entity.ID, "observed dependency")
	if err := ir.AddCandidate(fact); err != nil {
		t.Fatal(err)
	}
	for _, kind := range []EvidenceKind{VerificationEvidence, ComparisonEvidence} {
		evidence, err := NewEvidence(
			MustIdentity("candidate-evidence://evidence/"+kind.String()), GoVerifierID,
			kind, fact.Key(), StableHashString("candidate evidence"),
		)
		if err != nil {
			t.Fatal(err)
		}
		evidence.Status = FactCandidate
		beforeCanonical, beforeHash := ir.Canonical(), ir.EvidenceHash()
		if err := ir.AddEvidence(evidence); !errors.Is(err, ErrInvalidEvidence) {
			t.Fatalf("candidate %s evidence error = %v, want ErrInvalidEvidence", kind, err)
		}
		if ir.Canonical() != beforeCanonical || ir.EvidenceHash() != beforeHash || len(ir.Evidence()) != 0 {
			t.Fatalf("candidate %s evidence tamper mutated IR", kind)
		}
	}
}
