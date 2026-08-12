package analyzer

import (
	"reflect"
	"testing"
)

func TestHostingContractsKeepFutureStageDeferred(t *testing.T) {
	goContract := ContractFor(StageGoHosted)
	if !goContract.Valid() || !goContract.PromotionReady() {
		t.Fatalf("Go-hosted contract = %#v, want implemented", goContract)
	}
	future := ContractFor(StageGoooHosted)
	if !future.Valid() || future.Status != ContractDeferred || future.PromotionReady() {
		t.Fatalf("gooo-hosted contract = %#v, want deferred", future)
	}
	if !containsRequirement(future.Requirements, RequirementHostComparison) || !containsRequirement(future.Requirements, RequirementIndependentGate) {
		t.Fatalf("gooo-hosted requirements = %#v, want comparison and independent gate", future.Requirements)
	}
	unknown := ContractFor(HostStage("future-host"))
	if unknown.Valid() || unknown.PromotionReady() {
		t.Fatalf("unknown contract = %#v, want invalid and not ready", unknown)
	}
}

func TestGoHostedEvidenceMirrorsEveryDeltaView(t *testing.T) {
	source := []byte(`package billing

import fraud "example.com/fraud"

//gooo:semantic activity id="billing://activity/run"
func Run(order Order) {
	fraud.Check(order)
	helper()
}

//gooo:semantic entity id="billing://entity/order"
type Order struct{}
`)
	registry := NewRegistry()
	for _, id := range []string{"fraud://activity/check", "security://activity/check"} {
		registry.MustRegister(Registration{
			Ref:      SymbolRef{PackagePath: "example.com/fraud", PackageName: "fraud", Name: "Check"},
			Kind:     KindActivity,
			Identity: NewIdentity("fraud", id),
		})
	}
	result, err := AnalyzeSource("run.go", source, registry)
	if err != nil {
		t.Fatal(err)
	}
	report := result.GoHostedEvidence()
	if !report.Complete() || report.Reason != "" {
		t.Fatalf("Go-hosted report = %#v, want complete without deferral", report)
	}
	counts := make(map[EvidenceKind]int)
	for _, record := range report.Records {
		if !record.Valid() {
			t.Errorf("invalid evidence record = %#v", record)
		}
		counts[record.Kind]++
	}
	if !reflect.DeepEqual(counts, map[EvidenceKind]int{
		EvidenceKindFact: 1, EvidenceKindCandidate: 1, EvidenceKindImplementation: 2,
	}) {
		t.Fatalf("evidence counts = %#v", counts)
	}
	if got := report.Contract.Producer; got != "go" {
		t.Fatalf("producer = %q, want go", got)
	}
}

func TestGoooHostedEvidenceIsDeferredWithoutAFalseDigest(t *testing.T) {
	report := (Result{}).GoooHostedEvidence()
	if report.Complete() {
		t.Fatal("deferred gooo-hosted report was complete")
	}
	if report.ComparisonDigest() != "" {
		t.Fatalf("deferred digest = %q, want empty", report.ComparisonDigest())
	}
	if report.Reason == "" || len(report.Records) != 0 {
		t.Fatalf("deferred report = %#v, want reason and no records", report)
	}
}

func TestEvidenceComparisonExcludesHostMetadata(t *testing.T) {
	source := []byte(`package billing

//gooo:semantic activity id="billing://activity/run"
func Run(order Order) {}

//gooo:semantic entity id="billing://entity/order"
type Order struct{}
`)
	result, err := AnalyzeSource("run.go", source, NewRegistry())
	if err != nil {
		t.Fatal(err)
	}
	goReport := result.GoHostedEvidence()
	futureContract := ContractFor(StageGoooHosted)
	futureContract.Status = ContractImplemented
	future := EvidenceReport{Contract: futureContract, Records: append([]EvidenceRecord(nil), goReport.Records...)}
	if goReport.ComparisonDigest() == "" || goReport.ComparisonDigest() != future.ComparisonDigest() {
		t.Fatalf("host-specific comparison digests differ: go=%q future=%q", goReport.ComparisonDigest(), future.ComparisonDigest())
	}
}

func containsRequirement(requirements []ContractRequirement, want ContractRequirement) bool {
	for _, requirement := range requirements {
		if requirement == want {
			return true
		}
	}
	return false
}
