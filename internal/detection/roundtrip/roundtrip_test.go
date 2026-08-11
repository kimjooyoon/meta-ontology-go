package roundtrip

import (
	"bytes"
	"go/parser"
	"go/token"
	"math/rand"
	"strings"
	"testing"
	"testing/quick"
)

func TestMinimalFixturePassesAllPaths(t *testing.T) {
	fixture := MinimalFixture()
	if len(MinimalDSL()) == 0 || len(MinimalGo()) == 0 {
		t.Fatal("minimal fixture source is empty")
	}
	if _, err := parser.ParseFile(token.NewFileSet(), "minimal.go", MinimalGo(), parser.ParseComments); err != nil {
		t.Fatalf("minimal Go fixture is not executable Go: %v", err)
	}
	fixture.BeforeGo = MinimalGo()
	report := Verify(fixture)
	if !report.OK() {
		t.Fatal(report.Error())
	}
}

func TestDetectorReportsSemanticDriftByStableIdentity(t *testing.T) {
	fixture := MinimalFixture()
	fixture.GoIR.Facts = append([]Fact(nil), fixture.GoIR.Facts...)
	fixture.GoIR.Facts[0].Object = "billing://entity/payment"
	report := CheckGoToIR(fixture.IR, fixture.GoIR)
	if report.OK() {
		t.Fatal("semantic drift was accepted")
	}
	if report.Violations[0].Rule != RuleGoToIR || !strings.Contains(report.Error(), "billing://activity/pay-order") {
		t.Fatalf("unexpected drift report: %s", report.Error())
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
		AllowedIDs: []string{"billing://activity/pay-order"},
	})
	if report.OK() || report.Violations[0].Identity != "billing://entity/payment" {
		t.Fatalf("unrelated region was not reported: %s", report.Error())
	}
}

func TestVerifyInfersLocalityFromSemanticDelta(t *testing.T) {
	fixture := MinimalFixture()
	unrelated := []byte("\n//gooo:generated:start id=\"billing://entity/unrelated\" kind=\"entity\"\ntype Unrelated struct{}\n//gooo:generated:end id=\"billing://entity/unrelated\"\n")
	fixture.BeforeGo = append(MinimalGo(), unrelated...)
	fixture.BeforeIR = fixture.IR
	fixture.AfterIR = fixture.IR
	fixture.AfterIR.Facts = append([]Fact(nil), fixture.IR.Facts...)
	fixture.AfterIR.Facts = append(fixture.AfterIR.Facts, Fact{
		Subject: "billing://activity/pay-order", Predicate: "gooo:invokes", Object: "billing://activity/audit",
	})
	fixture.AfterIR.Nodes = append([]Node(nil), fixture.IR.Nodes...)
	fixture.AfterIR.Nodes = append(fixture.AfterIR.Nodes, Node{ID: "billing://activity/audit", Kind: Activity, Name: "Audit"})
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
	source := []byte("package billinggen\n//gooo:generated:start id=broken\n")
	report := CheckLocality(LocalityInput{Before: source, After: source})
	if report.OK() || report.Violations[0].Rule != RuleMarker {
		t.Fatalf("malformed marker was not reported: %s", report.Error())
	}
}

func TestEquivalentPresentationProperty(t *testing.T) {
	base := MinimalFixture().IR
	property := func(seed uint64) bool {
		candidate := cloneIR(base)
		for index := range candidate.Nodes {
			candidate.Nodes[index].Name = strings.Repeat("renamed", int(seed%3)+1)
			candidate.Nodes[index].Namespace = "display-only"
		}
		return Equivalent(base, candidate) && Fingerprint(base) == Fingerprint(candidate)
	}
	config := &quick.Config{MaxCount: 128, Rand: rand.New(rand.NewSource(41))}
	if err := quick.Check(property, config); err != nil {
		t.Fatal(err)
	}
}

func TestFingerprintOrderProperty(t *testing.T) {
	base := MinimalFixture().IR
	property := func(seed uint64) bool {
		candidate := cloneIR(base)
		random := rand.New(rand.NewSource(int64(seed)))
		random.Shuffle(len(candidate.Nodes), func(i, j int) { candidate.Nodes[i], candidate.Nodes[j] = candidate.Nodes[j], candidate.Nodes[i] })
		random.Shuffle(len(candidate.Facts), func(i, j int) { candidate.Facts[i], candidate.Facts[j] = candidate.Facts[j], candidate.Facts[i] })
		return Equivalent(base, candidate) && Fingerprint(base) == Fingerprint(candidate)
	}
	if err := quick.Check(property, &quick.Config{MaxCount: 128}); err != nil {
		t.Fatal(err)
	}
}

func cloneIR(ir IR) IR {
	result := ir
	result.Nodes = append([]Node(nil), ir.Nodes...)
	result.Facts = append([]Fact(nil), ir.Facts...)
	return result
}
