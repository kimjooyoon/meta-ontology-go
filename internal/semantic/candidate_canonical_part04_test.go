package semantic

import (
	"testing"
)

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
