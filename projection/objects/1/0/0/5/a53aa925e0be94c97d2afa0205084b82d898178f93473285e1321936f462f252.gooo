package analyzer

import (
	"os"
	"reflect"
	"testing"
)

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
