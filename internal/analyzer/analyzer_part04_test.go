package analyzer

import (
	"reflect"
	"testing"
)

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
	if len(result.Delta.ImplementationDetails) != 2 ||
		result.Delta.ImplementationDetails[1].Reference != "order.Validate" {
		t.Fatalf("implementation details = %#v", result.Delta.ImplementationDetails)
	}
}
