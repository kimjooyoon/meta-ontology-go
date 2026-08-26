package analyzer

import (
	"slices"
	"testing"
)

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
	return slices.Contains(requirements, want)
}
