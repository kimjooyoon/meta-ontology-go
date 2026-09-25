package analyzer

import (
	"reflect"
	"testing"
)

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
