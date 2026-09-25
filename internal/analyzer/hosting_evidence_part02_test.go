package analyzer

import (
	"reflect"
	"testing"
)

func TestHostingPairHasStableIDsSpansAndIndependentFutureComparison(t *testing.T) {
	report := hostingPairReport(t)
	if !report.Complete() {
		t.Fatalf("Go-hosted report = %#v, want complete", report)
	}
	want := []EvidenceRecord{
		{
			Kind: EvidenceKindFact, Status: EvidenceStatusDeterministic,
			Subject: NewIdentity("", "billing://activity/pay-order"), Relation: RelationUses,
			Object: NewIdentity("", "billing://entity/order"),
			Span:   hostingSpan("testdata/hosting_pair.go", 6, 21, 6, 26, 130, 135),
		},
		{
			Kind: EvidenceKindFact, Status: EvidenceStatusDeterministic,
			Subject: NewIdentity("", "billing://activity/pay-order"), Relation: RelationGenerates,
			Object: NewIdentity("", "billing://entity/payment"),
			Span:   hostingSpan("testdata/hosting_pair.go", 6, 28, 6, 35, 137, 144),
		},
		{
			Kind: EvidenceKindCandidate, Status: EvidenceStatusCandidate,
			Subject: NewIdentity("", "billing://activity/pay-order"), Relation: RelationInvokes,
			Reference: "fraud.Check", Options: []Identity{
				NewIdentity("fraud", "fraud://activity/check"),
				NewIdentity("security", "security://activity/check"),
			}, Span: hostingSpan("testdata/hosting_pair.go", 7, 2, 7, 13, 148, 159),
			Reason: "multiple registered semantic symbols match",
		},
		{
			Kind: EvidenceKindImplementation, Status: EvidenceStatusImplementation,
			Reference: "order", Span: hostingSpan("testdata/hosting_pair.go", 7, 14, 7, 19, 160, 165),
			Reason: "symbol reference: unregistered semantic symbol", IdentityState: IdentityUnresolved,
		},
		{
			Kind: EvidenceKindImplementation, Status: EvidenceStatusImplementation,
			Reference: "OrderID", Span: hostingSpan("testdata/hosting_pair.go", 8, 17, 8, 24, 183, 190),
			Reason: "symbol reference: unregistered semantic symbol", IdentityState: IdentityUnresolved,
		},
		{
			Kind: EvidenceKindImplementation, Status: EvidenceStatusImplementation,
			Reference: "order.ID", Span: hostingSpan("testdata/hosting_pair.go", 8, 26, 8, 34, 192, 200),
			Reason: "symbol reference: unregistered semantic symbol", IdentityState: IdentityUnresolved,
		},
		{
			Kind: EvidenceKindFact, Status: EvidenceStatusDeterministic,
			Subject: NewIdentity("", "billing://activity/pay-order"), Relation: RelationUses,
			Object: NewIdentity("", "billing://entity/payment"),
			Span:   hostingSpan("testdata/hosting_pair.go", 8, 9, 8, 16, 175, 182),
		},
	}
	sortEvidenceRecords(want)
	if !reflect.DeepEqual(report.Records, want) {
		t.Fatalf("Go-hosted records changed:\ngot=%#v\nwant=%#v", report.Records, want)
	}
	futureContract := ContractFor(StageGoooHosted)
	futureContract.Status = ContractImplemented
	future := EvidenceReport{Contract: futureContract, Records: append([]EvidenceRecord(nil), want...)}
	if !reflect.DeepEqual(report.AuthoritativeFacts(), future.AuthoritativeFacts()) {
		t.Fatalf("authoritative facts differ between hosts:\ngo=%#v\nfuture=%#v", report.AuthoritativeFacts(), future.AuthoritativeFacts())
	}
	if report.ComparisonDigest() != future.ComparisonDigest() {
		t.Fatalf("host comparison digest differs: go=%q future=%q", report.ComparisonDigest(), future.ComparisonDigest())
	}

	deferred := (Result{}).GoooHostedEvidence()
	if deferred.Complete() || deferred.ComparisonDigest() != "" {
		t.Fatalf("future analyzer state = %#v, want explicit deferred state", deferred)
	}
}
