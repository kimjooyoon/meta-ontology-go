package partialknowledgecomposition

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	independent "github.com/kimjooyoon/meta-ontology-go/internal/meta/partialknowledgecomposition/verify"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const testHeadSHA = "0123456789abcdef0123456789abcdef01234567"

func testInput(t *testing.T) Input {
	t.Helper()
	root := filepath.Join("..", "..", "..")
	source, err := os.ReadFile(filepath.Join(root, SourcePath))
	if err != nil {
		t.Fatal(err)
	}
	fixtureBytes, err := os.ReadFile(filepath.Join(root, "examples", "partial-knowledge-composition", "cases.json"))
	if err != nil {
		t.Fatal(err)
	}
	var fixture Fixture
	if err := json.Unmarshal(fixtureBytes, &fixture); err != nil {
		t.Fatal(err)
	}
	return Input{Repository: "kimjooyoon/meta-ontology-go", HeadSHA: testHeadSHA,
		SourcePath: SourcePath, Source: source, Fixture: fixture}
}

func TestGoooSourceUsesTheDeclaredKnowledgeValues(t *testing.T) {
	input := testInput(t)
	file, diagnostics := syntax.Parse(string(input.Source))
	if len(diagnostics) != 0 || file == nil {
		t.Fatalf("Gooo source diagnostics=%v file=%#v", diagnostics, file)
	}
	if len(file.Declarations) != 18 {
		t.Fatalf("declaration count = %d, want 18", len(file.Declarations))
	}
}

func TestCompositionCalculusKeepsKnowledgeCauses(t *testing.T) {
	cases := []struct {
		name       string
		left       Operand
		right      Operand
		state      State
		direct     []string
		blocked    []string
		invariants []string
	}{
		{name: "exact", left: Operand{Operation: "left", State: StateExact}, right: Operand{Operation: "right", State: StateExact}, state: StateExact},
		{name: "direct unknown", left: Operand{Operation: "left", State: StateDirectUnknown}, right: Operand{Operation: "right", State: StateExact}, state: StateDirectUnknown, direct: []string{"left"}},
		{name: "dependency blocked", left: Operand{Operation: "left", State: StateExact}, right: Operand{Operation: "right", State: StateDependencyBlocked, BlockedDependency: "missing"}, state: StateDependencyBlocked, blocked: []string{"missing"}},
		{name: "invariant", left: Operand{Operation: "left", State: StateExact}, right: Operand{Operation: "right", State: StateInvariantOnly, Invariants: []string{"read-only"}}, state: StateInvariantOnly, invariants: []string{"read-only"}},
		{name: "mixed", left: Operand{Operation: "left", State: StateDirectUnknown}, right: Operand{Operation: "right", State: StateDependencyBlocked, BlockedDependency: "missing"}, state: StateMixedUnresolved, direct: []string{"left"}, blocked: []string{"missing"}},
	}
	for _, current := range cases {
		t.Run(current.name, func(t *testing.T) {
			value := Compose(current.left, current.right)
			if value.State != current.state || !sameStrings(value.DirectUnknowns, current.direct) ||
				!sameStrings(value.BlockedDependencies, current.blocked) || !sameStrings(value.PreservedInvariants, current.invariants) {
				t.Fatalf("composition = %#v", value)
			}
			_, _, topSuccess := classify(value)
			if topSuccess != (current.state == StateExact) {
				t.Fatalf("top success for %q = %v", current.state, topSuccess)
			}
		})
	}
}

func TestCompositionIsOrderIndependent(t *testing.T) {
	left := Operand{Operation: "source", State: StateDirectUnknown}
	right := Operand{Operation: "binding", State: StateDependencyBlocked, BlockedDependency: "receipt"}
	forward := Compose(left, right)
	reverse := Compose(right, left)
	if !reflect.DeepEqual(forward, reverse) {
		t.Fatalf("forward composition=%#v, reverse composition=%#v", forward, reverse)
	}
}

func TestEvaluateBuildsClosedReceiptAndClaims(t *testing.T) {
	input := testInput(t)
	receipt, err := Evaluate(input)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateReceipt(receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.Summary.ExactCases != 1 || receipt.Summary.DirectUnknownCases != 1 ||
		receipt.Summary.DependencyBlockedCases != 1 || receipt.Summary.InvariantOnlyCases != 1 ||
		receipt.Summary.MixedUnresolvedCases != 1 || receipt.Summary.TopSuccessCases != 1 {
		t.Fatalf("summary = %#v", receipt.Summary)
	}
	if len(receipt.Claims) != FixedDenominator || receipt.Claims[0].From != "OPEN" || receipt.Claims[0].To != "DISCHARGED" || receipt.Claims[1].To != "UNKNOWN" || receipt.Claims[2].To != "BLOCKED" || receipt.Claims[3].To != "INVARIANT_PRESERVED" || receipt.Claims[4].To != "UNRESOLVED" {
		t.Fatalf("claim transitions = %#v", receipt.Claims)
	}
}

func TestIndependentVerifierReconstructsReceipt(t *testing.T) {
	input := testInput(t)
	receipt, err := Evaluate(input)
	if err != nil {
		t.Fatal(err)
	}
	receiptJSON, err := json.Marshal(receipt)
	if err != nil {
		t.Fatal(err)
	}
	fixtureJSON, err := json.Marshal(input.Fixture)
	if err != nil {
		t.Fatal(err)
	}
	report, err := independent.Verify(independent.Input{
		Repository: input.Repository, HeadSHA: input.HeadSHA, SourcePath: input.SourcePath,
		Source: input.Source, Fixture: fixtureJSON, Receipt: receiptJSON,
	})
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != "VERIFIED" || !report.IndependentEvaluator || report.PromotionAuthorized || report.RepositoryWrites != 0 {
		t.Fatalf("independent report = %#v", report)
	}
}

func TestUnknownExpectationCannotChangeComposition(t *testing.T) {
	input := testInput(t)
	input.Fixture.Cases[1].ExpectedState = StateExact
	if _, err := Evaluate(input); err == nil {
		t.Fatal("tampered direct unknown expectation was accepted")
	}
}

func sameStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
