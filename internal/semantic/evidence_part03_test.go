package semantic

import (
	"testing"
)

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
