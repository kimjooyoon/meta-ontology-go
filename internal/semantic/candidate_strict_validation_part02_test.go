package semantic

import (
	"errors"
	"testing"
)

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
