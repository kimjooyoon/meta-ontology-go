package main

import (
	"errors"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func TestCompilerEvidenceComparesHostsWithoutClaimingStageSuccess(t *testing.T) {
	goIR := evidenceFixture(t)
	goooIR := evidenceFixture(t)
	emitter := SemanticEvidenceEmitter{}
	goEvidence, err := emitter.EmitCompilerEvidence(goIR, semantic.GoHostedCompilerID)
	if err != nil {
		t.Fatal(err)
	}
	goooEvidence, err := emitter.EmitCompilerEvidence(goooIR, semantic.GoooHostedCompilerID)
	if err != nil {
		t.Fatal(err)
	}
	if len(goEvidence) != len(goooEvidence) || len(goEvidence) == 0 {
		t.Fatalf("evidence counts = %d and %d", len(goEvidence), len(goooEvidence))
	}
	for _, record := range goEvidence {
		if record.Kind != semantic.CompilerRunEvidence {
			t.Fatalf("evidence kind = %s", record.Kind)
		}
		if err := goIR.AddEvidence(record); err != nil {
			t.Fatal(err)
		}
	}
	for _, record := range goooEvidence {
		if err := goooIR.AddEvidence(record); err != nil {
			t.Fatal(err)
		}
	}
	if !goIR.ProvenanceEquivalent(goooIR) {
		t.Fatal("host evidence was not comparable")
	}
	if goIR.EvidenceHash() == goooIR.EvidenceHash() {
		t.Fatal("audit hashes should retain producer identity")
	}
	if goIR.ProvenanceHash() != goooIR.ProvenanceHash() {
		t.Fatal("comparison hashes should ignore producer identity")
	}
}

func TestCompilerEvidenceRejectsUnknownProducer(t *testing.T) {
	_, err := (SemanticEvidenceEmitter{}).EmitCompilerEvidence(evidenceFixture(t), semantic.MustIdentity("gooo://host/compiler/unknown"))
	if err == nil {
		t.Fatal("unknown producer was accepted")
	}
}

func TestCompilerEvidenceRemainsAppendOnly(t *testing.T) {
	ir := evidenceFixture(t)
	records, err := (SemanticEvidenceEmitter{}).EmitCompilerEvidence(ir, semantic.GoHostedCompilerID)
	if err != nil {
		t.Fatal(err)
	}
	if err := ir.AddEvidence(records[0]); err != nil {
		t.Fatal(err)
	}
	if err := ir.AddEvidence(records[0]); err != nil {
		t.Fatalf("idempotent append failed: %v", err)
	}
	conflict := records[0]
	conflict.Producer = semantic.GoooHostedCompilerID
	if err := ir.AddEvidence(conflict); !errors.Is(err, semantic.ErrEvidenceConflict) {
		t.Fatalf("conflicting append error = %v", err)
	}
}

func evidenceFixture(t *testing.T) semantic.IR {
	t.Helper()
	namespace := semantic.Namespace("billing")
	ir := semantic.NewIR("billing", namespace)
	activity, err := semantic.NewActivity(semantic.MustIdentity("billing://activity/pay"), namespace, "Pay")
	if err != nil {
		t.Fatal(err)
	}
	entity, err := semantic.NewEntity(semantic.MustIdentity("billing://entity/order"), namespace, "Order")
	if err != nil {
		t.Fatal(err)
	}
	if err := ir.AddNode(activity); err != nil {
		t.Fatal(err)
	}
	if err := ir.AddNode(entity); err != nil {
		t.Fatal(err)
	}
	if err := ir.AddActivityContract(semantic.ActivityContract{Activity: activity.ID, Inputs: []semantic.ID{entity.ID}}); err != nil {
		t.Fatal(err)
	}
	return ir
}
