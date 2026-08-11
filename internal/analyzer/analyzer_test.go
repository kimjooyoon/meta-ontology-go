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
	if !reflect.DeepEqual(details, []string{"json.Marshal", "strings.TrimSpace"}) {
		t.Fatalf("implementation details = %#v", details)
	}
}

func TestAnalyzePackageLiftsAnnotationsBeforeVisitingReferences(t *testing.T) {
	files := []SourceFile{
		{
			Filename:    "billing/activity.go",
			PackagePath: "example.com/billing",
			Source: []byte(`package billing

//gooo:semantic activity id="billing://activity/pay" namespace=billing
func Pay(order Order) (result Payment) { return }
`),
		},
		{
			Filename:    "billing/semantic.go",
			PackagePath: "example.com/billing",
			Source: []byte(`package billing

//gooo:semantic entity id="billing://entity/order" namespace=billing
type Order struct{}

//gooo:semantic entity id="billing://entity/payment" namespace=billing
type Payment struct{}
`),
		},
	}

	result, err := AnalyzePackage(files, NewRegistry())
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Registrations) != 3 {
		t.Fatalf("registrations = %d, want 3", len(result.Registrations))
	}
	if len(result.Delta.Added) != 2 {
		t.Fatalf("facts = %#v, want parameter use and result generation", result.Delta.Added)
	}
	if result.Delta.Added[0].Object.ID != "billing://entity/order" || result.Delta.Added[1].Object.ID != "billing://entity/payment" {
		t.Fatalf("facts = %#v", result.Delta.Added)
	}
}

func TestAmbiguousRegisteredReferenceRemainsCandidate(t *testing.T) {
	source, err := os.ReadFile("testdata/ambiguous.go")
	if err != nil {
		t.Fatal(err)
	}

	registry := NewRegistry()
	for _, id := range []string{"fraud://activity/check", "security://activity/check"} {
		registry.MustRegister(Registration{
			Ref:      SymbolRef{PackagePath: "example.com/fraud", PackageName: "fraud", Name: "Check"},
			Kind:     KindActivity,
			Identity: NewIdentity("registered", id),
		})
	}

	result, err := AnalyzeSource("testdata/ambiguous.go", source, registry)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Delta.Added) != 1 {
		t.Fatalf("facts = %#v, want only the deterministic parameter use", result.Delta.Added)
	}
	if len(result.Delta.Candidates) != 1 {
		t.Fatalf("candidates = %#v, want one", result.Delta.Candidates)
	}
	candidate := result.Delta.Candidates[0]
	if candidate.Relation != RelationInvokes || candidate.Reference != "fraud.Check" || candidate.Subject.ID != "billing://activity/settle" {
		t.Fatalf("candidate = %#v", candidate)
	}
	wantOptions := []Identity{
		NewIdentity("registered", "fraud://activity/check"),
		NewIdentity("registered", "security://activity/check"),
	}
	if !reflect.DeepEqual(candidate.Options, wantOptions) {
		t.Fatalf("candidate options = %#v, want %#v", candidate.Options, wantOptions)
	}
}

func TestUnregisteredCallIsImplementationDetail(t *testing.T) {
	source := []byte(`package billing

//gooo:semantic activity id="billing://activity/run"
func Run() { helper() }

func helper() {}
`)
	result, err := AnalyzeSource("run.go", source, NewRegistry())
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Delta.Added) != 0 || len(result.Delta.Candidates) != 0 {
		t.Fatalf("delta = %#v, want no semantic facts", result.Delta)
	}
	if len(result.Delta.ImplementationDetails) != 1 || result.Delta.ImplementationDetails[0].Reference != "helper" {
		t.Fatalf("implementation details = %#v", result.Delta.ImplementationDetails)
	}
}

func TestRegistryResolutionIsDeterministicRegardlessOfRegistrationOrder(t *testing.T) {
	source := []byte(`package app

import domain "example.com/domain"

//gooo:semantic activity id="app://activity/run"
func Run(input domain.Order) domain.Payment {
	return domain.Payment{}
}
`)

	makeRegistry := func(reverse bool) *Registry {
		registry := NewRegistry()
		entries := []Registration{
			{Ref: SymbolRef{PackagePath: "example.com/domain", PackageName: "domain", Name: "Payment"}, Kind: KindEntity, Identity: NewIdentity("domain", "domain://entity/payment")},
			{Ref: SymbolRef{PackagePath: "example.com/domain", PackageName: "domain", Name: "Order"}, Kind: KindEntity, Identity: NewIdentity("domain", "domain://entity/order")},
		}
		if reverse {
			entries[0], entries[1] = entries[1], entries[0]
		}
		for _, entry := range entries {
			registry.MustRegister(entry)
		}
		return registry
	}

	first, err := AnalyzeSource("run.go", source, makeRegistry(false))
	if err != nil {
		t.Fatal(err)
	}
	second, err := AnalyzeSource("run.go", source, makeRegistry(true))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first.Delta, second.Delta) {
		t.Fatalf("delta depends on registration order:\nfirst=%#v\nsecond=%#v", first.Delta, second.Delta)
	}
}

func TestUnknownMethodDoesNotResolveToAnUnrelatedRegisteredMethod(t *testing.T) {
	source := []byte(`package billing

//gooo:semantic activity id="billing://activity/run"
func Run(order Order) {
	order.Validate()
}

type Order struct{}
`)
	registry := NewRegistry()
	registry.MustRegister(Registration{
		Ref:      SymbolRef{PackagePath: "example.com/billing", PackageName: "billing", Receiver: "Invoice", Name: "Validate"},
		Kind:     KindActivity,
		Identity: NewIdentity("billing", "billing://activity/validate"),
	})

	result, err := AnalyzePackage([]SourceFile{{Filename: "run.go", PackagePath: "example.com/billing", Source: source}}, registry)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Delta.Added) != 0 || len(result.Delta.Candidates) != 0 {
		t.Fatalf("delta = %#v, want no semantic relation", result.Delta)
	}
	if len(result.Delta.ImplementationDetails) != 1 || result.Delta.ImplementationDetails[0].Reference != "order.Validate" {
		t.Fatalf("implementation details = %#v", result.Delta.ImplementationDetails)
	}
}

func TestNamespacePreservesAmbiguityAndStableOptionOrder(t *testing.T) {
	source := []byte(`package billing

import fraud "example.com/fraud"

//gooo:semantic activity id="billing://activity/run"
func Run() { fraud.Check() }
`)
	makeRegistry := func(reverse bool) *Registry {
		registry := NewRegistry()
		entries := []Registration{
			{Ref: SymbolRef{PackagePath: "example.com/fraud", PackageName: "fraud", Name: "Check"}, Kind: KindActivity, Identity: NewIdentity("zeta", "urn:check")},
			{Ref: SymbolRef{PackagePath: "example.com/fraud", PackageName: "fraud", Name: "Check"}, Kind: KindActivity, Identity: NewIdentity("alpha", "urn:check")},
		}
		if reverse {
			entries[0], entries[1] = entries[1], entries[0]
		}
		for _, entry := range entries {
			registry.MustRegister(entry)
		}
		return registry
	}

	first, err := AnalyzePackage([]SourceFile{{Filename: "run.go", PackagePath: "example.com/billing", Source: source}}, makeRegistry(false))
	if err != nil {
		t.Fatal(err)
	}
	second, err := AnalyzePackage([]SourceFile{{Filename: "run.go", PackagePath: "example.com/billing", Source: source}}, makeRegistry(true))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first.Delta, second.Delta) {
		t.Fatalf("delta depends on registration order:\nfirst=%#v\nsecond=%#v", first.Delta, second.Delta)
	}
	if got := first.Delta.Candidates[0].Options; !reflect.DeepEqual(got, []Identity{NewIdentity("alpha", "urn:check"), NewIdentity("zeta", "urn:check")}) {
		t.Fatalf("candidate options = %#v", got)
	}
}

func TestLocalBindingDoesNotLiftPackageSymbol(t *testing.T) {
	source := []byte(`package billing

//gooo:semantic activity id="billing://activity/run"
func Run() {
	Check := func() {}
	Check()
}
`)
	registry := NewRegistry()
	registry.MustRegister(Registration{
		Ref:      SymbolRef{PackagePath: "billing", PackageName: "billing", Name: "Check"},
		Kind:     KindActivity,
		Identity: NewIdentity("billing", "billing://activity/check"),
	})

	result, err := AnalyzeSource("run.go", source, registry)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Delta.Added) != 0 || len(result.Delta.Candidates) != 0 {
		t.Fatalf("delta = %#v, want no semantic relation", result.Delta)
	}
	if len(result.Delta.ImplementationDetails) != 1 || result.Delta.ImplementationDetails[0].Reference != "Check" {
		t.Fatalf("implementation details = %#v", result.Delta.ImplementationDetails)
	}
}
