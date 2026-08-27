package partialknowledgecomposition

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	independent "github.com/kimjooyoon/meta-ontology-go/internal/meta/partialknowledgecomposition/verify"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const testHeadSHA = "0123456789abcdef0123456789abcdef01234567"

func testInput(t *testing.T, source []byte, mode InterventionMode) Input {
	t.Helper()
	return Input{Repository: "kimjooyoon/meta-ontology-go", HeadSHA: testHeadSHA, SourcePath: SourcePath, Source: source, Intervention: mode}
}

func testSource(t *testing.T) []byte {
	t.Helper()
	root := filepath.Join("..", "..", "..")
	source, err := os.ReadFile(filepath.Join(root, SourcePath))
	if err != nil {
		t.Fatal(err)
	}
	return source
}

func TestGoooSourceProvidesFiveComputedObservationReceipts(t *testing.T) {
	source := testSource(t)
	file, diagnostics := syntax.ParseFile(SourcePath, string(source))
	if file == nil || diagnostics.HasErrors() {
		t.Fatalf("source diagnostics=%v file=%#v", diagnostics, file)
	}
	model, err := parseSource(SourcePath, source)
	if err != nil {
		t.Fatal(err)
	}
	if len(model.Cases) != FixedDenominator || model.SemanticIRDigest == "" {
		t.Fatalf("source model = %#v", model)
	}
	if len(file.Declarations) != 23 {
		t.Fatalf("declaration count = %d, want 23", len(file.Declarations))
	}
}

func TestCompositionDerivesKnowledgeCausesFromEvidence(t *testing.T) {
	cases := []struct {
		name        string
		left, right Evidence
		state       State
		claim       string
	}{
		{"exact", Evidence{Operation: "left", Required: "a", Observed: "a", ObservedAvailable: true, PriorState: "OPEN"}, Evidence{Operation: "right", Required: "b", Observed: "b", ObservedAvailable: true, PriorState: "OPEN"}, StateExact, "DISCHARGED"},
		{"direct unknown", Evidence{Operation: "left", Required: "a", PriorState: "OPEN"}, Evidence{Operation: "right", Required: "b", Observed: "b", ObservedAvailable: true, PriorState: "OPEN"}, StateDirectUnknown, "OPEN"},
		{"dependency blocked", Evidence{Operation: "left", Required: "a", Observed: "a", ObservedAvailable: true, PriorState: "OPEN"}, Evidence{Operation: "right", Required: "b", DependencyClaimID: "receipt", PriorState: "OPEN"}, StateDependencyBlocked, "OPEN"},
		{"invariant", Evidence{Operation: "left", Required: "a", Observed: "a", ObservedAvailable: true, PriorState: "OPEN"}, Evidence{Operation: "right", Required: "b", Observed: "b", ObservedAvailable: true, InvariantEvidence: "read-only", PriorState: "OPEN"}, StateInvariantOnly, "OPEN"},
		{"mixed", Evidence{Operation: "left", Required: "a", PriorState: "OPEN"}, Evidence{Operation: "right", Required: "b", DependencyClaimID: "receipt", PriorState: "OPEN"}, StateMixedUnresolved, "OPEN"},
	}
	for _, current := range cases {
		t.Run(current.name, func(t *testing.T) {
			value := Compose(current.left, current.right)
			if value.State != current.state {
				t.Fatalf("composition = %#v", value)
			}
			_, resolution, _, topSuccess := classify(value)
			if (resolution == "EXACT") != (current.state == StateExact) || topSuccess != (current.state == StateExact) {
				t.Fatalf("classification = %q/%v", resolution, topSuccess)
			}
			if transitionState(value.State) != current.claim {
				t.Fatalf("claim transition = %q, want %q", transitionState(value.State), current.claim)
			}
		})
	}
}

func TestCompositionIsOrderIndependent(t *testing.T) {
	left := Evidence{Operation: "source", Required: "source-shape", PriorState: "OPEN"}
	right := Evidence{Operation: "binding", Required: "receipt-shape", DependencyClaimID: "receipt", PriorState: "OPEN"}
	if forward, reverse := Compose(left, right), Compose(right, left); !reflect.DeepEqual(forward, reverse) {
		t.Fatalf("forward=%#v reverse=%#v", forward, reverse)
	}
}

func TestEvaluateBuildsOpenClaimsAndSeparateConformance(t *testing.T) {
	receipt, err := Evaluate(testInput(t, testSource(t), InterventionNone))
	if err != nil {
		t.Fatal(err)
	}
	if receipt.Decision != DecisionCalculusProven || receipt.Resolution != ResolutionCalculus || receipt.SubjectResolution != SubjectResolution {
		t.Fatalf("receipt boundary = %#v", receipt)
	}
	if receipt.Summary.ExactCases != 1 || receipt.Summary.DirectUnknownCases != 1 || receipt.Summary.DependencyBlockedCases != 1 || receipt.Summary.InvariantOnlyCases != 1 || receipt.Summary.MixedUnresolvedCases != 1 || receipt.Summary.OpenClaims != 4 || receipt.Summary.DischargedClaims != 1 {
		t.Fatalf("summary = %#v", receipt.Summary)
	}
	if receipt.Claims[0].From != "OPEN" || receipt.Claims[0].To != "DISCHARGED" || receipt.Claims[1].To != "OPEN" || receipt.Claims[2].To != "OPEN" || receipt.Claims[3].To != "OPEN" || receipt.Claims[4].To != "OPEN" {
		t.Fatalf("claims = %#v", receipt.Claims)
	}
	if receipt.RepositoryWrites != 0 || receipt.PromotionAuthorized {
		t.Fatalf("authority = %#v", receipt)
	}
}

func TestIndependentVerifierReconstructsSourceReceipt(t *testing.T) {
	input := testInput(t, testSource(t), InterventionNone)
	receipt, err := Evaluate(input)
	if err != nil {
		t.Fatal(err)
	}
	receiptJSON, err := json.Marshal(receipt)
	if err != nil {
		t.Fatal(err)
	}
	report, err := independent.Verify(independent.Input{Repository: input.Repository, HeadSHA: input.HeadSHA, SourcePath: input.SourcePath, Source: input.Source, InterventionMode: string(input.Intervention), Receipt: receiptJSON})
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != "VERIFIED" || !report.IndependentEvaluator || report.OpenClaims != 4 {
		t.Fatalf("report = %#v", report)
	}
}

func TestSemanticInterventionChangesObservationAndClaim(t *testing.T) {
	source := testSource(t)
	base, err := Evaluate(testInput(t, source, InterventionNone))
	if err != nil {
		t.Fatal(err)
	}
	changed, err := Evaluate(testInput(t, source, InterventionSemantic))
	if err != nil {
		t.Fatal(err)
	}
	if base.Cases[1].Result.State != StateDirectUnknown || changed.Cases[1].Result.State != StateExact || base.Claims[1].To != "OPEN" || changed.Claims[1].To != "DISCHARGED" {
		t.Fatalf("causality base=%#v changed=%#v", base.Cases[1], changed.Cases[1])
	}
}

func TestCommentOnlyInterventionPreservesSemanticProjection(t *testing.T) {
	source := testSource(t)
	base, err := Evaluate(testInput(t, source, InterventionNone))
	if err != nil {
		t.Fatal(err)
	}
	commented := append([]byte("// nonsemantic comment intervention\n"), source...)
	comment, err := Evaluate(testInput(t, commented, InterventionCommentOnly))
	if err != nil {
		t.Fatal(err)
	}
	if base.SemanticProjectionDigest != comment.SemanticProjectionDigest || base.SemanticIRDigest != comment.SemanticIRDigest {
		t.Fatalf("comment changed semantic projection")
	}
	if base.SourceDigest == comment.SourceDigest {
		t.Fatalf("comment did not change source digest")
	}
}

func TestConclusionLabelsAreNotAcceptedAsSourceEvidence(t *testing.T) {
	source := string(testSource(t))
	source = strings.Replace(source, "|case=exact-pair|", "|state=EXACT|case=exact-pair|", 1)
	if _, err := Evaluate(testInput(t, []byte(source), InterventionNone)); err == nil {
		t.Fatal("conclusion label was accepted as source evidence")
	}
}
