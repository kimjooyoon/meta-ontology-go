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

func TestActivityContractRejectsInvalidFactsAtomically(t *testing.T) {
	ns := Namespace("contract-boundary")
	activity := mustActivity(t, MustIdentity("contract-boundary://activity/run"), ns, "Run")
	input := mustEntity(t, MustIdentity("contract-boundary://entity/input"), ns, "Input")
	output := mustEntity(t, MustIdentity("contract-boundary://entity/output"), ns, "Output")
	otherActivity := mustActivity(t, MustIdentity("contract-boundary://activity/other"), ns, "Other")
	agent := mustAgent(t, MustIdentity("contract-boundary://agent/verifier"), ns, "Verifier")
	missing := MustIdentity("contract-boundary://entity/missing")
	for _, test := range []struct {
		name     string
		contract ActivityContract
		wantErr  error
	}{
		{
			name: "missing input endpoint", contract: ActivityContract{
				Activity: activity.ID, Inputs: []ID{input.ID, missing},
			}, wantErr: ErrNodeNotFound,
		},
		{
			name: "missing output endpoint", contract: ActivityContract{
				Activity: activity.ID, Outputs: []ID{output.ID, missing},
			}, wantErr: ErrNodeNotFound,
		},
		{
			name: "wrong input kind", contract: ActivityContract{
				Activity: activity.ID, Inputs: []ID{input.ID, otherActivity.ID},
			}, wantErr: ErrInvalidFact,
		},
		{
			name: "wrong output kind", contract: ActivityContract{
				Activity: activity.ID, Outputs: []ID{output.ID, otherActivity.ID},
			}, wantErr: ErrInvalidFact,
		},
		{
			name: "missing agent endpoint", contract: ActivityContract{
				Activity: activity.ID, Agents: []ID{agent.ID, missing},
			}, wantErr: ErrNodeNotFound,
		},
		{
			name: "wrong agent kind", contract: ActivityContract{
				Activity: activity.ID, Agents: []ID{agent.ID, input.ID},
			}, wantErr: ErrInvalidFact,
		},
	} {
		ir := NewIR("contract-boundary", ns)
		for _, node := range []Node{activity, input, output, otherActivity, agent} {
			if err := ir.AddNode(node); err != nil {
				t.Fatal(err)
			}
		}
		before := ir.Canonical()
		if err := ir.AddActivityContract(test.contract); !errors.Is(err, test.wantErr) {
			t.Fatalf("%s error = %v, want %v", test.name, err, test.wantErr)
		}
		if ir.Canonical() != before || len(ir.Graph.Facts()) != 0 || len(ir.Evidence()) != 0 {
			t.Fatalf("%s partially mutated graph, IR, or evidence", test.name)
		}
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

func TestCandidateDirectionMismatchIsRejectedAtomically(t *testing.T) {
	ns := Namespace("deferred")
	activity := mustActivity(t, MustIdentity("deferred://activity/run"), ns, "Run")
	entity := mustEntity(t, MustIdentity("deferred://entity/input"), ns, "Input")
	output := mustEntity(t, MustIdentity("deferred://entity/output"), ns, "Output")
	agent := mustAgent(t, MustIdentity("deferred://agent/verifier"), ns, "Verifier")
	cases := []struct {
		name      string
		subject   ID
		predicate Relation
		object    ID
	}{
		{name: "used reversed", subject: entity.ID, predicate: Used, object: activity.ID},
		{name: "generated reversed", subject: activity.ID, predicate: WasGeneratedBy, object: output.ID},
		{name: "derived reversed", subject: activity.ID, predicate: WasDerivedFrom, object: output.ID},
		{name: "associated reversed", subject: agent.ID, predicate: WasAssociatedWith, object: activity.ID},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			graph := NewGraph()
			for _, node := range []Node{activity, entity, output, agent} {
				if err := graph.AddNode(node); err != nil {
					t.Fatal(err)
				}
			}
			fact := NewCandidateFact(test.subject, test.predicate, test.object, "untyped observation")
			beforeCanonical, beforeHash := graph.Canonical(), graph.StableHash()
			if err := graph.AddCandidate(fact); !errors.Is(err, ErrInvalidFact) {
				t.Fatalf("candidate direction error = %v, want ErrInvalidFact", err)
			}
			if graph.HasCandidate(fact.Key()) || graph.HasFact(fact.Key()) {
				t.Fatal("rejected candidate direction crossed the graph boundary")
			}
			if graph.Canonical() != beforeCanonical || graph.StableHash() != beforeHash {
				t.Fatal("rejected candidate direction mutated the graph")
			}
		})
	}
}

func TestMissingCandidatePromotionDoesNotInitializeOrMutateGraph(t *testing.T) {
	graph := Graph{}
	key := FactKey{
		Subject: MustIdentity("deferred://activity/missing"), Predicate: Used,
		Object: MustIdentity("deferred://entity/missing"),
	}
	beforeCanonical, beforeHash := graph.Canonical(), graph.StableHash()
	if _, err := graph.PromoteCandidate(key); !errors.Is(err, ErrCandidateNotFound) {
		t.Fatalf("missing candidate error = %v, want ErrCandidateNotFound", err)
	}
	if graph.Canonical() != beforeCanonical || graph.StableHash() != beforeHash {
		t.Fatal("missing candidate promotion changed canonical/hash")
	}
	if graph.nodes != nil || graph.names != nil || graph.facts != nil || graph.candidates != nil {
		t.Fatal("missing candidate promotion initialized graph storage")
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
