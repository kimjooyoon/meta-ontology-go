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

func selfHostingIR(t *testing.T, producer ID) IR {
	t.Helper()
	ir := NewIR("self-hosted-compiler", Namespace("self-host"))
	source := mustEntity(t, MustIdentity("gooo://entity/source"), ir.Namespace, "Source")
	output := mustEntity(t, MustIdentity("gooo://entity/ir"), ir.Namespace, "SemanticIR")
	report := mustEntity(t, MustIdentity("gooo://entity/report"), ir.Namespace, "VerificationReport")
	compile := mustActivity(t, MustIdentity("gooo://activity/compile"), ir.Namespace, "Compile")
	verify := mustActivity(t, MustIdentity("gooo://activity/verify"), ir.Namespace, "Verify")
	ci := mustAgent(t, CIVerifierID, ir.Namespace, "CI verifier")
	for _, node := range []Node{source, output, report, compile, verify, ci} {
		if err := ir.AddNode(node); err != nil {
			t.Fatal(err)
		}
	}
	facts := []Fact{
		NewUsedFact(compile.ID, source.ID),
		NewWasGeneratedByFact(output.ID, compile.ID),
		NewUsedFact(verify.ID, output.ID),
		NewWasGeneratedByFact(report.ID, verify.ID),
		NewWasAssociatedWithFact(verify.ID, ci.ID),
	}
	for _, fact := range facts {
		if err := ir.AddFact(fact); err != nil {
			t.Fatal(err)
		}
	}
	digest := StableHashString("self-hosting source and verification artifact")
	for index, fact := range facts[:4] {
		evidenceProducer := producer
		evidenceKind := CompilerRunEvidence
		if index == 3 {
			evidenceProducer = GoVerifierID
			evidenceKind = VerificationEvidence
		}
		evidence, err := NewEvidence(
			MustIdentity("gooo://evidence/fact/"+string(rune('a'+index))),
			evidenceProducer, evidenceKind, fact.Key(), digest,
		)
		if err != nil {
			t.Fatal(err)
		}
		if err := ir.AddEvidence(evidence); err != nil {
			t.Fatal(err)
		}
	}
	return ir
}
