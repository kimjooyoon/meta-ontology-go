package main

import (
	"encoding/json"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"testing"
)

func TestGraphDumpCandidateIsExplicitAndNotInGraphHash(t *testing.T) {
	file, diagnostics := syntax.Parse(sourceOrderA)
	if diagnostics.Error() != nil {
		t.Fatal(diagnostics.Error())
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		t.Fatal(err)
	}
	beforeHash := authoritativeGraphHash(ir.Graph)
	beforeIRHash := authoritativeIRHash(ir)
	if beforeIRHash != ir.StableHash() {
		t.Fatalf("authoritative IR digest disagrees without candidates: %s != %s", beforeIRHash, ir.StableHash())
	}
	order := semantic.MustIdentity("billing://entity/order")
	payment := semantic.MustIdentity("billing://entity/payment")
	if err := ir.AddCandidate(semantic.NewCandidateFact(order, semantic.WasDerivedFrom, payment, "needs review")); err != nil {
		t.Fatal(err)
	}
	dump := newGraphDump([]byte(sourceOrderA), ir)
	if dump.GraphHash != beforeHash || dump.IR.SemanticDigest != beforeIRHash {
		t.Fatalf("candidate changed authoritative digest: graph %s/%s, IR %s/%s", beforeHash, dump.GraphHash, beforeIRHash, dump.IR.SemanticDigest)
	}
	if !hasGraphRelation(dump.Relations, "candidate", string(order), string(semantic.WasDerivedFrom), string(payment)) {
		t.Fatalf("candidate relation was not explicit: %#v", dump.Relations)
	}
}

func TestAuthoritativeIRHashTracksRuntimeBindingEndpointChanges(t *testing.T) {
	file, diagnostics := syntax.Parse(sourceWithRuntimeBinding)
	if diagnostics.Error() != nil {
		t.Fatal(diagnostics.Error())
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		t.Fatal(err)
	}
	baseHash := authoritativeIRHash(ir)
	changed := ir
	changed.RuntimeBindings = append([]semantic.RuntimeBinding(nil), ir.RuntimeBindings...)
	changed.RuntimeBindings[0].ConsumerActivity = semantic.MustIdentity("billing://activity/produce")
	if changed.SemanticCanonical() == ir.SemanticCanonical() {
		t.Fatal("binding endpoint mutation did not change semantic fingerprint")
	}
	if authoritativeIRHash(changed) == baseHash {
		t.Fatal("binding endpoint mutation did not change authoritative inspect hash")
	}
}

func TestInspectJSONCarrierRoundTripPreservesRuntimeBindings(t *testing.T) {
	dump := decodeGraphDump(t, inspectFixtureOutput(t, sourceWithRuntimeBinding))
	payload, err := json.Marshal(dump)
	if err != nil {
		t.Fatal(err)
	}
	var decoded graphDump
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatal(err)
	}
	if len(decoded.RuntimeBindings) != 1 || decoded.RuntimeBindings[0].Entity != "billing://entity/payment" ||
		decoded.RuntimeBindings[0].ProducerPort != "result" || decoded.RuntimeBindings[0].ConsumerPort != "input" {
		t.Fatalf("inspect JSON carrier lost runtime binding: %#v", decoded.RuntimeBindings)
	}
}
