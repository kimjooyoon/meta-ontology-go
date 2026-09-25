package semantic

import (
	"errors"
	"testing"
)

func TestSelfHostingEvidenceSeparatesMeaningFromProducer(t *testing.T) {
	goIR := selfHostingIR(t, GoHostedCompilerID)
	goooIR := selfHostingIR(t, GoooHostedCompilerID)

	if err := goIR.Validate(); err != nil {
		t.Fatal(err)
	}
	if err := goooIR.Validate(); err != nil {
		t.Fatal(err)
	}
	if !goIR.SemanticallyEquivalent(goooIR) {
		t.Fatal("Go-hosted and gooo-hosted compiler meaning diverged")
	}
	if !goIR.ProvenanceEquivalent(goooIR) {
		t.Fatal("equivalent compiler claims did not share provenance")
	}
	if goIR.EvidenceHash() == goooIR.EvidenceHash() {
		t.Fatal("exact evidence hash discarded the producing host")
	}
	if goIR.ProvenanceHash() != goooIR.ProvenanceHash() {
		t.Fatal("cross-host provenance hash changed for the same claims")
	}
}
func TestEvidenceIdentityIsAppendOnly(t *testing.T) {
	ir := selfHostingIR(t, GoHostedCompilerID)
	record := ir.Evidence()[0]
	before := len(ir.Evidence())
	if err := ir.AddEvidence(record); err != nil {
		t.Fatal(err)
	}
	if got := len(ir.Evidence()); got != before {
		t.Fatalf("idempotent append changed evidence count to %d", got)
	}
	conflict := record
	conflict.Producer = GoooHostedCompilerID
	if err := ir.AddEvidence(conflict); !errors.Is(err, ErrEvidenceConflict) {
		t.Fatalf("conflicting evidence error = %v, want ErrEvidenceConflict", err)
	}
}
func TestEvidenceRequiresStableDigest(t *testing.T) {
	fact := FactKey{
		Subject:   MustIdentity("gooo://activity/compile"),
		Predicate: Used,
		Object:    MustIdentity("gooo://entity/source"),
	}
	_, err := NewEvidence(
		MustIdentity("gooo://evidence/invalid"), GoHostedCompilerID,
		CompilerRunEvidence, fact, "not-a-digest",
	)
	if !errors.Is(err, ErrInvalidEvidence) {
		t.Fatalf("invalid digest error = %v, want ErrInvalidEvidence", err)
	}
}
