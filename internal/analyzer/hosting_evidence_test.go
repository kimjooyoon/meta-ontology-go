package analyzer

import (
	"os"
	"reflect"
	"testing"
)

func TestHostingViewsSeparateFactsCandidatesAndImplementation(t *testing.T) {
	report := hostingPairReport(t)
	facts := report.AuthoritativeFacts()
	candidates := report.CandidateEvidence()
	details := report.ImplementationEvidence()
	if len(facts) != 3 || len(candidates) != 1 || len(details) != 0 {
		t.Fatalf("evidence views = facts %d, candidates %d, implementation %d; want 3, 1, 0", len(facts), len(candidates), len(details))
	}
	for _, fact := range facts {
		if fact.Kind != EvidenceKindFact || fact.Status != EvidenceStatusDeterministic {
			t.Errorf("authoritative fact = %#v, want deterministic fact", fact)
		}
	}
	candidate := candidates[0]
	if candidate.Kind != EvidenceKindCandidate || candidate.Status != EvidenceStatusCandidate {
		t.Fatalf("candidate view = %#v, want candidate evidence", candidate)
	}
	wantOptions := []Identity{
		NewIdentity("fraud", "fraud://activity/check"),
		NewIdentity("security", "security://activity/check"),
	}
	if !reflect.DeepEqual(candidate.Options, wantOptions) {
		t.Fatalf("candidate options = %#v, want %#v", candidate.Options, wantOptions)
	}
	if candidate.Subject.ID == facts[0].Subject.ID && candidate.Kind == facts[0].Kind {
		t.Fatal("candidate was mixed into authoritative facts")
	}

	candidates[0].Options[0].ID = "mutated"
	if report.CandidateEvidence()[0].Options[0].ID == "mutated" {
		t.Fatal("candidate view exposed report-owned options")
	}
}

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

func TestHostingPairAnalysisIsRepeatable(t *testing.T) {
	source, err := os.ReadFile("testdata/hosting_pair.go")
	if err != nil {
		t.Fatal(err)
	}
	registry := hostingPairRegistry()
	sources := []SourceFile{{Filename: "testdata/hosting_pair.go", PackagePath: "example.com/billing", Source: source}}
	first, err := AnalyzePackage(sources, registry)
	if err != nil {
		t.Fatal(err)
	}
	second, err := AnalyzePackage(sources, registry)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first.Delta, second.Delta) || first.GoHostedEvidence().ComparisonDigest() != second.GoHostedEvidence().ComparisonDigest() {
		t.Fatalf("repeated analysis changed evidence:\nfirst=%#v\nsecond=%#v", first.Delta, second.Delta)
	}
}

func hostingPairReport(t *testing.T) EvidenceReport {
	t.Helper()
	source, err := os.ReadFile("testdata/hosting_pair.go")
	if err != nil {
		t.Fatal(err)
	}
	result, err := AnalyzePackage([]SourceFile{{
		Filename: "testdata/hosting_pair.go", PackagePath: "example.com/billing", Source: source,
	}}, hostingPairRegistry())
	if err != nil {
		t.Fatal(err)
	}
	return result.GoHostedEvidence()
}

func hostingPairRegistry() *Registry {
	registry := NewRegistry()
	for namespace, id := range map[string]string{
		"fraud": "fraud://activity/check", "security": "security://activity/check",
	} {
		registry.MustRegister(Registration{
			Ref:  SymbolRef{PackagePath: "example.com/fraud", PackageName: "fraud", Name: "Check"},
			Kind: KindActivity, Identity: NewIdentity(namespace, id),
		})
	}
	return registry
}

func hostingSpan(filename string, startLine, startColumn, endLine, endColumn, startOffset, endOffset int) Span {
	return Span{
		Filename: filename,
		Start:    Position{Offset: startOffset, Line: startLine, Column: startColumn},
		End:      Position{Offset: endOffset, Line: endLine, Column: endColumn},
	}
}
