package roundtrip

import (
	"bytes"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
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

func TestSemanticDeltaAndReportAreDeterministicAcrossInputOrder(t *testing.T) {
	fixture := MinimalFixture()
	permuted := rebuildIR(t, fixture.IR, nil, nil)
	if !Equivalent(fixture.IR, permuted) || CheckRoundTrip(fixture.IR, permuted).Error() != "roundtrip verification passed" {
		t.Fatal("permuted semantic input was not equivalent")
	}
	first, err := SemanticDelta(fixture.IR, permuted)
	if err != nil {
		t.Fatal(err)
	}
	second, err := SemanticDelta(permuted, fixture.IR)
	if err != nil {
		t.Fatal(err)
	}
	if len(first.TouchedIDs) != 0 || len(first.AffectedIDs) != 0 || len(second.TouchedIDs) != 0 {
		t.Fatalf("permutation produced a semantic delta: %#v %#v", first, second)
	}

	mutated := rebuildIR(t, fixture.IR, func(node semantic.Node) semantic.Node {
		if node.ID == semantic.MustIdentity("billing://entity/order") {
			node.Namespace = semantic.Namespace("other")
		}
		return node
	}, nil)
	repeated := CheckGoToIR(fixture.IR, mutated)
	for i := 0; i < 20; i++ {
		if got := CheckGoToIR(fixture.IR, mutated).Error(); got != repeated.Error() {
			t.Fatalf("detector output changed on repeat %d: %q != %q", i, got, repeated.Error())
		}
	}
}

func TestLocalityAllowsImplementationEditOutsideGeneratedRegions(t *testing.T) {
	before := MinimalGo()
	after := bytes.Replace(before, []byte("package billinggen\n"), []byte("package billinggen\n\nvar Keep = 7\n"), 1)
	report := CheckLocality(LocalityInput{Before: before, After: after})
	if !report.OK() {
		t.Fatal(report.Error())
	}
}

func TestLocalityIgnoresHandwrittenSlotBody(t *testing.T) {
	before := bytes.Replace(MinimalGo(), []byte("func PayOrder(order Order) Payment { return Payment{} }"), []byte("//gooo:slot:start id=\"billing://activity/pay-order/implementation\"\nfunc PayOrder(order Order) Payment { return Payment{} }\n//gooo:slot:end id=\"billing://activity/pay-order/implementation\""), 1)
	after := bytes.Replace(before, []byte("func PayOrder(order Order) Payment { return Payment{} }"), []byte("func PayOrder(order Order) Payment { return Payment{ } }"), 1)
	report := CheckLocality(LocalityInput{Before: before, After: after})
	if !report.OK() {
		t.Fatal(report.Error())
	}
}

func TestLocalityReportsUnrelatedGeneratedRegion(t *testing.T) {
	before := MinimalGo()
	after := bytes.Replace(before, []byte("type Payment struct{}"), []byte("type Payment struct{ Amount int }"), 1)
	report := CheckLocality(LocalityInput{
		Before:     before,
		After:      after,
		AllowedIDs: []semantic.ID{semantic.MustIdentity("billing://activity/pay-order")},
	})
	if report.OK() || report.Violations[0].Identity != "billing://entity/payment" {
		t.Fatalf("unrelated region was not reported: %s", report.Error())
	}
}

func TestVerifyInfersLocalityFromSemanticDelta(t *testing.T) {
	fixture := MinimalFixture()
	unrelated := []byte("\n//gooo:generated:start id=\"billing://entity/unrelated\" kind=\"entity\"\ntype Unrelated struct{}\n//gooo:generated:end id=\"billing://entity/unrelated\" kind=\"entity\"\n")
	fixture.BeforeGo = append(MinimalGo(), unrelated...)
	fixture.BeforeIR = fixture.IR
	fixture.AfterIR = rebuildIR(t, fixture.IR, func(node semantic.Node) semantic.Node {
		return node
	}, func(fact semantic.Fact) (semantic.Fact, bool) {
		return fact, false
	})
	newActivity := mustFixtureNode(t, semantic.Activity, "billing://activity/audit", "Audit")
	newEntity := mustFixtureNode(t, semantic.Entity, "billing://entity/audit-record", "AuditRecord")
	fixture.AfterIR = addNodesAndFacts(t, fixture.AfterIR, []semantic.Node{newActivity, newEntity}, []semantic.Fact{
		semantic.NewUsedFact(semantic.MustIdentity("billing://activity/audit"), semantic.MustIdentity("billing://entity/audit-record")),
	})
	fixture.AfterGo = bytes.Replace(fixture.BeforeGo, []byte("type Unrelated struct{}"), []byte("type Unrelated struct{ Value int }"), 1)
	report := Verify(fixture)
	if report.OK() {
		t.Fatal("changed unrelated generated region was accepted")
	}
	if !strings.Contains(report.Error(), "billing://entity/unrelated") {
		t.Fatalf("locality finding missing: %s", report.Error())
	}
}

func TestMalformedGeneratedMarkersAreReported(t *testing.T) {
	source := []byte("package billinggen\n//gooo:generated:start id=\"broken\" kind=\"entity\"\n")
	report := CheckLocality(LocalityInput{Before: source, After: source})
	if report.OK() || report.Violations[0].Rule != RuleMarker {
		t.Fatalf("malformed marker was not reported: %s", report.Error())
	}
}

func TestMarkerRejectsSlotAndRegionIdentityCollisions(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "duplicate slot across regions",
			source: `package billinggen
//gooo:generated:start id="billing://entity/first" kind="entity"
//gooo:slot:start id="billing://slot/shared"
func First() {}
//gooo:slot:end id="billing://slot/shared"
//gooo:generated:end id="billing://entity/first" kind="entity"
//gooo:generated:start id="billing://entity/second" kind="entity"
//gooo:slot:start id="billing://slot/shared"
func Second() {}
//gooo:slot:end id="billing://slot/shared"
//gooo:generated:end id="billing://entity/second" kind="entity"
`,
		},
		{
			name: "slot nested in region",
			source: `package billinggen
//gooo:generated:start id="billing://shared" kind="entity"
//gooo:slot:start id="billing://shared"
func Shared() {}
//gooo:slot:end id="billing://shared"
//gooo:generated:end id="billing://shared" kind="entity"
`,
		},
		{
			name: "region after slot",
			source: `package billinggen
//gooo:generated:start id="billing://entity/first" kind="entity"
//gooo:slot:start id="billing://shared"
func First() {}
//gooo:slot:end id="billing://shared"
//gooo:generated:end id="billing://entity/first" kind="entity"
//gooo:generated:start id="billing://shared" kind="entity"
type Shared struct{}
//gooo:generated:end id="billing://shared" kind="entity"
`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			report := CheckLocality(LocalityInput{Before: []byte(test.source), After: []byte(test.source)})
			if report.OK() || report.Violations[0].Rule != RuleMarker {
				t.Fatalf("identity collision was not reported: %s", report.Error())
			}
		})
	}
}

func TestMalformedSemanticSnapshotsAreReported(t *testing.T) {
	report := CheckDSLToIR(semantic.IR{}, MinimalFixture().IR)
	if report.OK() || report.Violations[0].Rule != RuleSnapshot {
		t.Fatalf("malformed semantic snapshot was not reported: %s", report.Error())
	}
}

func TestVerifyDoesNotMutateInputs(t *testing.T) {
	fixture := MinimalFixture()
	beforeGo := append([]byte(nil), fixture.BeforeGo...)
	afterGo := append([]byte(nil), fixture.AfterGo...)
	beforeHash := fixture.IR.StableHash()
	allowed := append([]semantic.ID(nil), fixture.AllowedIDs...)
	if report := Verify(fixture); !report.OK() {
		t.Fatal(report.Error())
	}
	if !bytes.Equal(fixture.BeforeGo, beforeGo) || !bytes.Equal(fixture.AfterGo, afterGo) {
		t.Fatal("verification mutated source bytes")
	}
	if fixture.IR.StableHash() != beforeHash || len(fixture.AllowedIDs) != len(allowed) {
		t.Fatal("verification mutated semantic input")
	}
}

func rebuildIR(t *testing.T, base semantic.IR, transformNode func(semantic.Node) semantic.Node, transformFact func(semantic.Fact) (semantic.Fact, bool)) semantic.IR {
	t.Helper()
	result := semantic.NewIR(base.Package, base.Namespace)
	for _, node := range base.Graph.Nodes() {
		if transformNode != nil {
			node = transformNode(node)
		}
		if err := result.AddNode(node); err != nil {
			t.Fatalf("rebuild node %s: %v", node.ID, err)
		}
	}
	for _, fact := range base.Graph.DeterministicFacts() {
		keep := true
		if transformFact != nil {
			fact, keep = transformFact(fact)
		}
		if keep {
			if err := result.AddFact(fact); err != nil {
				t.Fatalf("rebuild fact %s: %v", fact.Key(), err)
			}
		}
	}
	return result
}

func addNodesAndFacts(t *testing.T, base semantic.IR, nodes []semantic.Node, facts []semantic.Fact) semantic.IR {
	t.Helper()
	result := rebuildIR(t, base, nil, nil)
	for _, node := range nodes {
		if err := result.AddNode(node); err != nil {
			t.Fatalf("add node %s: %v", node.ID, err)
		}
	}
	for _, fact := range facts {
		if err := result.AddFact(fact); err != nil {
			t.Fatalf("add fact %s: %v", fact.Key(), err)
		}
	}
	return result
}

func mustFixtureNode(t *testing.T, kind semantic.Kind, id, name string) semantic.Node {
	t.Helper()
	node, err := semantic.NewNodeFromStrings(kind, id, "billing", name)
	if err != nil {
		t.Fatal(err)
	}
	return node
}
