package analyzer

import (
	"os"
	"reflect"
	"sort"
	"testing"
)

func TestAnalyzeSourceLiftsOnlyRegisteredSymbols(t *testing.T) {
	source, err := os.ReadFile("testdata/registered.go")
	if err != nil {
		t.Fatal(err)
	}
	registry := NewRegistry()
	registry.MustRegister(Registration{
		Ref:      SymbolRef{PackagePath: "example.com/fraud", PackageName: "fraud", Name: "Check"},
		Kind:     KindActivity,
		Identity: NewIdentity("fraud", "fraud://activity/check"),
	})

	result, err := AnalyzeSource("testdata/registered.go", source, registry)
	if err != nil {
		t.Fatal(err)
	}

	want := []struct {
		relation Relation
		object   string
	}{
		{RelationUses, "billing://entity/order"},
		{RelationGenerates, "billing://entity/payment"},
		{RelationInvokes, "fraud://activity/check"},
		{RelationUses, "billing://entity/payment"},
	}
	if len(result.Delta.Added) != len(want) {
		t.Fatalf("facts = %d, want %d: %#v", len(result.Delta.Added), len(want), result.Delta.Added)
	}
	for index, expected := range want {
		fact := result.Delta.Added[index]
		if fact.Subject.ID != "billing://activity/pay-order" || fact.Relation != expected.relation || fact.Object.ID != expected.object {
			t.Errorf("fact[%d] = %#v, want %s %s", index, fact, expected.relation, expected.object)
		}
		if fact.Span.Filename != "testdata/registered.go" || fact.Span.Start.Line == 0 || fact.Span.End.Offset <= fact.Span.Start.Offset {
			t.Errorf("fact[%d] has invalid source span: %#v", index, fact.Span)
		}
	}
	if len(result.Delta.Candidates) != 0 {
		t.Fatalf("candidates = %#v, want none", result.Delta.Candidates)
	}
	var details []string
	for _, detail := range result.Delta.ImplementationDetails {
		details = append(details, detail.Reference)
	}
	sort.Strings(details)
	if !reflect.DeepEqual(details, []string{
		"OrderID", "json.Marshal", "normalized", "normalized", "order", "order.ID", "order.ID", "strings.TrimSpace",
	}) {
		t.Fatalf("implementation details = %#v", details)
	}
	for _, detail := range result.Delta.ImplementationDetails {
		if detail.IdentityState != IdentityUnresolved || detail.Span.Start.Offset < 0 || detail.Reason == "" {
			t.Fatalf("incomplete unresolved detail = %#v", detail)
		}
	}
}
