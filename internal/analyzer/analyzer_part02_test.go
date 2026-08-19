package analyzer

import (
	"testing"
)

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
