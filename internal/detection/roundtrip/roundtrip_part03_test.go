package roundtrip

import (
	"bytes"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"strings"
	"testing"
)

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
