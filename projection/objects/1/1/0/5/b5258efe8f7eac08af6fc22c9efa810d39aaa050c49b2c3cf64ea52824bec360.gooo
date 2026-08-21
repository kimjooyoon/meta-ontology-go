package semantic

import (
	"encoding/json"
	"errors"
	"os"
	"testing"
)

type deferredContract struct {
	ID            string `json:"id"`
	Status        string `json:"status"`
	Authoritative bool   `json:"authoritative"`
}

func readDeferredContract(t *testing.T, id string) deferredContract {
	t.Helper()
	data, err := os.ReadFile("testdata/deferred-contracts.json")
	if err != nil {
		t.Fatal(err)
	}
	var contracts []deferredContract
	if err := json.Unmarshal(data, &contracts); err != nil {
		t.Fatal(err)
	}
	for _, contract := range contracts {
		if contract.ID == id {
			return contract
		}
	}
	t.Fatalf("deferred contract %q is missing", id)
	return deferredContract{}
}
func TestDeferredSemanticBoundariesAreMachineReadable(t *testing.T) {
	for _, id := range []string{
		"parser-formatter-crlf", "authorized-rekey", "analyzer-adapter",
	} {
		contract := readDeferredContract(t, id)
		if contract.Status != "deferred" || contract.Authoritative {
			t.Fatalf("contract %q = %#v, want non-authoritative deferred", id, contract)
		}
	}
}
func TestUnknownFactInputsFailAtomicallyAtAddBoundaries(t *testing.T) {
	ns := Namespace("boundary")
	activity := mustActivity(t, MustIdentity("boundary://activity/run"), ns, "Run")
	entity := mustEntity(t, MustIdentity("boundary://entity/input"), ns, "Input")
	unknown := MustIdentity("boundary://entity/missing")
	ir := NewIR("boundary", ns)
	for _, node := range []Node{activity, entity} {
		if err := ir.AddNode(node); err != nil {
			t.Fatal(err)
		}
	}
	before := ir.Canonical()
	beforeEvidence := ir.EvidenceHash()
	if err := ir.AddFact(NewUsedFact(activity.ID, unknown)); !errors.Is(err, ErrNodeNotFound) {
		t.Fatalf("unknown deterministic endpoint error = %v, want ErrNodeNotFound", err)
	}
	if err := ir.AddCandidate(NewCandidateFact(activity.ID, Used, unknown, "unresolved")); !errors.Is(err, ErrNodeNotFound) {
		t.Fatalf("unknown candidate endpoint error = %v, want ErrNodeNotFound", err)
	}
	unknownRelation := NewFact(activity.ID, Relation("calls"), entity.ID)
	if err := ir.AddFact(unknownRelation); !errors.Is(err, ErrUnknownRelation) {
		t.Fatalf("unknown deterministic relation error = %v, want ErrUnknownRelation", err)
	}
	if err := ir.AddCandidate(unknownRelation); !errors.Is(err, ErrUnknownRelation) {
		t.Fatalf("unknown candidate relation error = %v, want ErrUnknownRelation", err)
	}
	if ir.Canonical() != before || ir.EvidenceHash() != beforeEvidence || len(ir.Evidence()) != 0 {
		t.Fatal("rejected endpoint/relation operation mutated graph, IR, or evidence")
	}
}
