package verify_test

import (
	"encoding/json"
	"testing"

	meta "github.com/kimjooyoon/meta-ontology-go/internal/meta/partialknowledgecomposition"
	independent "github.com/kimjooyoon/meta-ontology-go/internal/meta/partialknowledgecomposition/verify"
)

func TestVerifierRejectsTamperedReceipt(t *testing.T) {
	input := fixtureInput(t)
	receipt, err := meta.Evaluate(input)
	if err != nil {
		t.Fatal(err)
	}
	receipt.Cases[1].Result.State = meta.StateExact
	receiptJSON, err := json.Marshal(receipt)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := independent.Verify(independent.Input{
		Repository: input.Repository, HeadSHA: input.HeadSHA, SourcePath: input.SourcePath,
		Source: input.Source, Fixture: fixtureJSON(input), Receipt: receiptJSON,
	}); err == nil {
		t.Fatal("tampered receipt was accepted")
	}
}

func fixtureInput(t *testing.T) meta.Input {
	t.Helper()
	return meta.Input{
		Repository: "kimjooyoon/meta-ontology-go",
		HeadSHA:    "0123456789abcdef0123456789abcdef01234567",
		SourcePath: meta.SourcePath,
		Source:     []byte("package partialknowledgecomposition\nnamespace partialknowledgecomposition\nentity DirectUnknown id \"direct\"\nentity DependencyBlocked id \"blocked\"\nentity InvariantOnly id \"invariant\"\nactivity Compose(MetaValue, MetaValue) -> MetaValue\n"),
		Fixture: meta.Fixture{
			Schema: meta.FixtureSchema, SourcePath: meta.SourcePath,
			FixedDenominator: meta.FixedDenominator,
			Cases: []meta.Case{
				caseInput("exact-pair", meta.Operand{Operation: "left-exact", State: meta.StateExact}, meta.Operand{Operation: "right-exact", State: meta.StateExact}, meta.StateExact, "PASS", "ALL_OPERATIONS_EXACT", meta.ProofCoherence),
				caseInput("direct-unknown", meta.Operand{Operation: "source", State: meta.StateDirectUnknown}, meta.Operand{Operation: "receipt", State: meta.StateExact}, meta.StateDirectUnknown, "FAIL_CLOSED", "DIRECT_UNKNOWN_NOT_PROMOTED", meta.ProofFoundation),
				caseInput("dependency-blocked", meta.Operand{Operation: "source", State: meta.StateExact}, meta.Operand{Operation: "binding", State: meta.StateDependencyBlocked, BlockedDependency: "receipt"}, meta.StateDependencyBlocked, "FAIL_CLOSED", "DEPENDENCY_BLOCKED_NOT_PROMOTED", meta.ProofCoherence),
				caseInput("invariant-preservation", meta.Operand{Operation: "source", State: meta.StateExact}, meta.Operand{Operation: "writes", State: meta.StateInvariantOnly, Invariants: []string{"repository-writes-zero"}}, meta.StateInvariantOnly, "HOLD", "KNOWN_INVARIANT_PRESERVED", meta.ProofFoundation),
				caseInput("mixed-unknown-and-blocked", meta.Operand{Operation: "source", State: meta.StateDirectUnknown}, meta.Operand{Operation: "binding", State: meta.StateDependencyBlocked, BlockedDependency: "receipt"}, meta.StateMixedUnresolved, "FAIL_CLOSED", "MIXED_UNRESOLVED_KNOWLEDGE", meta.ProofRegression),
			},
		},
	}
}

func caseInput(id string, left, right meta.Operand, state meta.State, decision, reason string, proof meta.ProofChoice) meta.Case {
	return meta.Case{ID: id, Producer: meta.Producer, Consumer: meta.Consumer, MetaOperation: meta.MetaOperation, ProofChoice: proof, Left: left, Right: right, ExpectedState: state, ExpectedDecision: decision, ExpectedReason: reason}
}

func fixtureJSON(input meta.Input) []byte {
	raw, _ := json.Marshal(input.Fixture)
	return raw
}
