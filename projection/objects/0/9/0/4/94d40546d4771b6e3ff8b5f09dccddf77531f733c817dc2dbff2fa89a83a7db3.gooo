package main

import (
	"bytes"
	"encoding/json"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"reflect"
	"testing"
)

func TestGraphDumpEvidenceDoesNotBecomeProvenance(t *testing.T) {
	firstIR := lowerInspectFixtureIR(t)
	secondIR := lowerInspectFixtureIR(t)
	facts := firstIR.Graph.DeterministicFacts()
	if len(facts) == 0 {
		t.Fatal("fixture has no deterministic fact for evidence")
	}
	evidenceA, err := semantic.NewEvidence(
		semantic.MustIdentity("billing://evidence/a"), semantic.GoHostedCompilerID,
		semantic.CompilerRunEvidence, facts[0].Key(), semantic.StableHash([]byte("fixture")),
	)
	if err != nil {
		t.Fatal(err)
	}
	evidenceB, err := semantic.NewEvidence(
		semantic.MustIdentity("billing://evidence/b"), semantic.GoHostedCompilerID,
		semantic.VerificationEvidence, facts[0].Key(), semantic.StableHash([]byte("verification")),
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := firstIR.AddEvidence(evidenceB); err != nil {
		t.Fatal(err)
	}
	if err := firstIR.AddEvidence(evidenceA); err != nil {
		t.Fatal(err)
	}
	if err := secondIR.AddEvidence(evidenceA); err != nil {
		t.Fatal(err)
	}
	if err := secondIR.AddEvidence(evidenceB); err != nil {
		t.Fatal(err)
	}
	firstDump := newGraphDump([]byte(sourceOrderA), firstIR)
	secondDump := newGraphDump([]byte(sourceOrderA), secondIR)
	wantRefs := []string{evidenceA.ID.String(), evidenceB.ID.String()}
	if firstDump.Evidence.Status != "available" || !reflect.DeepEqual(firstDump.Evidence.Refs, wantRefs) {
		t.Fatalf("evidence status = %#v", firstDump.Evidence)
	}
	if firstDump.Provenance.Status != "missing" || firstDump.Provenance.Refs != nil || firstDump.Provenance.Reason == "" {
		t.Fatalf("evidence was falsely reported as provenance: %#v", firstDump.Provenance)
	}
	firstJSON, err := json.Marshal(firstDump)
	if err != nil {
		t.Fatal(err)
	}
	secondJSON, err := json.Marshal(secondDump)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(firstJSON, secondJSON) {
		t.Fatalf("evidence insertion order changed graph dump:\nfirst=%s\nsecond=%s", firstJSON, secondJSON)
	}
}
