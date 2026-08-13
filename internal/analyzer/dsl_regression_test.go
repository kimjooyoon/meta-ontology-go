package analyzer

import (
	"os"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestBillingDSLStableIDsDriveGoSignatureAnalysis(t *testing.T) {
	source, err := os.ReadFile("../../examples/billing/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	file, diagnostics := syntax.ParseFile("examples/billing/main.gooo", string(source))
	if diagnostics.HasErrors() {
		t.Fatalf("billing DSL diagnostics = %#v", diagnostics)
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		t.Fatal(err)
	}

	registry := NewRegistry()
	for _, node := range ir.Graph.Nodes() {
		kind, ok := analyzerKind(node.Kind)
		if !ok {
			continue
		}
		if err := registry.Register(Registration{
			Ref:      SymbolRef{PackagePath: "billing", PackageName: "billing", Name: node.Name},
			Kind:     kind,
			Identity: NewIdentity(ir.Namespace.String(), string(node.ID)),
		}); err != nil {
			t.Fatal(err)
		}
	}

	result, err := AnalyzePackage([]SourceFile{{
		Filename:    "generated.go",
		PackagePath: "billing",
		Source: []byte(`package billing

func PayOrder(order Order, method PaymentMethod) Payment { return }

type Order struct{}
type PaymentMethod struct{}
type Payment struct{}
`),
	}}, registry)
	if err != nil {
		t.Fatal(err)
	}
	if result.Diagnostics.HasErrors() {
		t.Fatalf("generated Go diagnostics = %#v", result.Diagnostics)
	}
	if len(result.Delta.Added) != 3 {
		t.Fatalf("facts = %#v, want two uses and one generation", result.Delta.Added)
	}
	uses := 0
	generated := 0
	for _, fact := range result.Delta.Added {
		switch fact.Relation {
		case RelationUses:
			uses++
			if fact.Object.ID != "billing://entity/order" && fact.Object.ID != "billing://entity/payment-method" {
				t.Fatalf("unexpected use fact = %#v", fact)
			}
		case RelationGenerates:
			generated++
			if fact.Object.ID != "billing://entity/payment" {
				t.Fatalf("unexpected generated fact = %#v", fact)
			}
		default:
			t.Fatalf("unexpected relation fact = %#v", fact)
		}
	}
	if uses != 2 || generated != 1 {
		t.Fatalf("fact shape = uses:%d generated:%d", uses, generated)
	}
}

func analyzerKind(kind semantic.Kind) (SymbolKind, bool) {
	switch kind {
	case semantic.Entity:
		return KindEntity, true
	case semantic.Activity:
		return KindActivity, true
	default:
		return "", false
	}
}
