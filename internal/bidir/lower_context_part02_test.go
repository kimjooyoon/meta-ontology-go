package bidir

import (
	"context"
	"errors"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"reflect"
	"testing"
)

func TestLowerContextReplayIsDeterministicAndNonMutating(t *testing.T) {
	file := lowerContextFixture(t)
	before, err := DocumentFromSyntax(file)
	if err != nil {
		t.Fatal(err)
	}
	first, err := LowerContext(context.Background(), file)
	if err != nil {
		t.Fatal(err)
	}
	second, err := LowerContext(context.Background(), file)
	if err != nil {
		t.Fatal(err)
	}
	if first.SemanticCanonical() != second.SemanticCanonical() {
		t.Fatalf("replayed lowering is not deterministic:\n%s\n%s", first.SemanticCanonical(), second.SemanticCanonical())
	}
	after, err := DocumentFromSyntax(file)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, after) {
		t.Fatal("successful lowering mutated syntax input")
	}
}
func assertLowerCanceled(t *testing.T, got semantic.IR, err error) {
	t.Helper()
	if !errors.Is(err, ErrLowerCanceled) || err.Error() != ErrLowerCanceled.Error() {
		t.Fatalf("cancellation error = %v", err)
	}
	if !reflect.DeepEqual(got, semantic.IR{}) {
		t.Fatalf("canceled lowerer exposed partial IR: %#v", got)
	}
}
func lowerContextFixture(t *testing.T) *syntax.File {
	t.Helper()
	file, diagnostics := syntax.ParseFile("cancel.gooo", `package billing
namespace billing
entity Order id "billing://entity/order"
entity Payment id "billing://entity/payment"
activity PayOrder(Order) -> Payment`)
	if diagnostics.Error() != nil {
		t.Fatal(diagnostics.Error())
	}
	return file
}
func repeatedLowerDocument(count int) Document {
	document := Document{Package: "billing", Namespace: "billing"}
	for index := 0; index < count; index++ {
		document.Declarations = append(document.Declarations, Declaration{
			Kind: EntityKind,
			ID:   ID("billing://entity/item-" + zeroPad(index)),
			Name: "Item" + zeroPad(index),
		})
	}
	return document
}
