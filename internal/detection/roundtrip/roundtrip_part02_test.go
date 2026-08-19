package roundtrip

import (
	"bytes"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"testing"
)

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
