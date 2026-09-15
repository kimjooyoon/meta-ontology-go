package roundtrip

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestMinimalFixturePassesAllPaths(t *testing.T) {
	fixture := MinimalFixture()
	if len(MinimalDSL()) == 0 || len(MinimalGo()) == 0 {
		t.Fatal("minimal fixture source is empty")
	}
	if _, err := parser.ParseFile(token.NewFileSet(), "minimal.go", MinimalGo(), parser.ParseComments); err != nil {
		t.Fatalf("minimal Go fixture is not executable Go: %v", err)
	}
	if err := fixture.IR.Validate(); err != nil {
		t.Fatalf("minimal semantic fixture is invalid: %v", err)
	}
	if report := Verify(fixture); !report.OK() {
		t.Fatal(report.Error())
	}
}
func TestDetectorReportsSemanticDriftByStableIdentity(t *testing.T) {
	fixture := MinimalFixture()
	mutated := rebuildIR(t, fixture.IR, nil, func(fact semantic.Fact) (semantic.Fact, bool) {
		if fact.Predicate == semantic.Used {
			fact.Object = semantic.MustIdentity("billing://entity/payment")
		}
		return fact, fact.Predicate == semantic.Used
	})
	report := CheckGoToIR(fixture.IR, mutated)
	if report.OK() {
		t.Fatal("semantic drift was accepted")
	}
	if report.Violations[0].Rule != RuleGoToIR || !strings.Contains(report.Error(), "billing://activity/pay-order") {
		t.Fatalf("unexpected drift report: %s", report.Error())
	}
}
func TestEquivalentIgnoresPresentationNamesAndAliases(t *testing.T) {
	fixture := MinimalFixture()
	renamed := rebuildIR(t, fixture.IR, func(node semantic.Node) semantic.Node {
		node.Name = "Display " + node.Name
		node.Aliases = []string{"presentation-" + node.ID.String()}
		return node
	}, nil)
	if !Equivalent(fixture.IR, renamed) {
		t.Fatal("presentation-only rename changed semantic equivalence")
	}
	if Fingerprint(fixture.IR) != Fingerprint(renamed) {
		t.Fatal("presentation-only rename changed semantic fingerprint")
	}
}
