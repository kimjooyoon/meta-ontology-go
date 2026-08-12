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

func TestCandidateDirectionMismatchRemainsNonAuthoritative(t *testing.T) {
	ns := Namespace("deferred")
	activity := mustActivity(t, MustIdentity("deferred://activity/run"), ns, "Run")
	entity := mustEntity(t, MustIdentity("deferred://entity/input"), ns, "Input")
	graph := NewGraph()
	for _, node := range []Node{activity, entity} {
		if err := graph.AddNode(node); err != nil {
			t.Fatal(err)
		}
	}
	fact := NewCandidateFact(entity.ID, Used, activity.ID, "untyped observation")
	if err := graph.AddCandidate(fact); err != nil {
		t.Fatal(err)
	}
	if !graph.HasCandidate(fact.Key()) || graph.HasFact(fact.Key()) {
		t.Fatal("candidate direction mismatch crossed the authoritative boundary")
	}
	if err := graph.Validate(); !errors.Is(err, ErrGraphInvalid) {
		t.Fatalf("invalid candidate direction was not fail-closed: %v", err)
	}
	if _, err := graph.PromoteCandidate(fact.Key()); !errors.Is(err, ErrInvalidFact) {
		t.Fatalf("invalid candidate promotion error = %v, want ErrInvalidFact", err)
	}
}

func TestQualifiedPROVCounterpartIsDeferredAndRejected(t *testing.T) {
	relation := Relation("wasGeneratedBy#qualified")
	if relation.Valid() {
		t.Fatal("qualified relation became a bare relation implicitly")
	}
	graph := NewGraph()
	fact := NewFact(
		MustIdentity("deferred://entity/output"), relation,
		MustIdentity("deferred://activity/run"),
	)
	if err := graph.AddFact(fact); !errors.Is(err, ErrUnknownRelation) {
		t.Fatalf("qualified relation error = %v, want ErrUnknownRelation", err)
	}
	if len(graph.AllFacts()) != 0 {
		t.Fatal("deferred qualified relation mutated graph")
	}
	t.Log("DEFERRED: semantic-ir/v1 has no qualified relation/event identity schema")
}

func TestIdentityReplacementIsNotImplicitlyEquivalent(t *testing.T) {
	ns := Namespace("deferred")
	oldID := MustIdentity("deferred://entity/old")
	newID := MustIdentity("deferred://entity/new")
	graph := NewGraph()
	if err := graph.AddNode(mustEntity(t, oldID, ns, "Record")); err != nil {
		t.Fatal(err)
	}
	before := graph.StableHash()
	if err := graph.AddNode(mustEntity(t, newID, ns, "RecordV2")); err != nil {
		t.Fatal(err)
	}
	if graph.StableHash() == before {
		t.Fatal("new ID was silently treated as an authorized rekey")
	}
	if _, ok := graph.Node(oldID); !ok {
		t.Fatal("old ID disappeared without a continuity contract")
	}
	if _, ok := graph.Node(newID); !ok {
		t.Fatal("new ID was not retained as a distinct declaration")
	}
	t.Log("DEFERRED: ID continuity/rekey authorization requires a future delta contract")
}
