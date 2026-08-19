package semantic

import (
	"testing"
)

func provFixture(t *testing.T, compiler ID) IR {
	t.Helper()
	ns := Namespace("prov-fixture")
	ir := NewIR("prov-fixture", ns)
	source := mustEntity(t, MustIdentity("prov-fixture://entity/source"), ns, "Source")
	output := mustEntity(t, MustIdentity("prov-fixture://entity/output"), ns, "Output")
	compile := mustActivity(t, MustIdentity("prov-fixture://activity/compile"), ns, "Compile")
	verify := mustActivity(t, MustIdentity("prov-fixture://activity/verify"), ns, "Verify")
	agent := mustAgent(t, GoVerifierID, ns, "Go verifier")
	for _, node := range []Node{source, output, compile, verify, agent} {
		if err := ir.AddNode(node); err != nil {
			t.Fatal(err)
		}
	}
	facts := []Fact{
		NewUsedFact(compile.ID, source.ID),
		NewWasGeneratedByFact(output.ID, compile.ID),
		NewWasDerivedFromFact(output.ID, source.ID),
		NewWasAssociatedWithFact(verify.ID, agent.ID),
	}
	digest := StableHashString("prov fixture payload/v1")
	for index, fact := range facts {
		if err := ir.AddFact(fact); err != nil {
			t.Fatal(err)
		}
		producer, kind := compiler, CompilerRunEvidence
		if fact.Predicate == WasAssociatedWith {
			producer, kind = GoVerifierID, VerificationEvidence
		}
		evidence, err := NewEvidence(
			MustIdentity("prov-fixture://evidence/"+string(rune('a'+index))),
			producer, kind, fact.Key(), digest,
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
