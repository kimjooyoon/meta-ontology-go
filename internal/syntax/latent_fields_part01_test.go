package syntax

import (
	"reflect"
	"strings"
	"testing"
)

func TestParsedEntitiesHaveNoReachableFields(t *testing.T) {
	source := `package billing
namespace billing
entity Order id "billing://entity/order"
activity PayOrder(Order) -> Order`
	file, diagnostics := ParseFile("billing.gooo", source)
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diagnostics)
	}
	entity := file.Declarations[0].(*EntityDecl)
	if entity.Fields != nil {
		t.Fatalf("parser populated latent fields: %#v", entity.Fields)
	}
	formatted, err := Format(file)
	if err != nil {
		t.Fatalf("existing source became unformattable: %v", err)
	}
	if !strings.Contains(formatted, `entity Order id "billing://entity/order"`) {
		t.Fatalf("formatted source lost entity declaration: %q", formatted)
	}
}
func TestProposedFieldSourceRemainsRejectedWithoutPartialFieldAST(t *testing.T) {
	source := `package billing
namespace billing
entity Order id "billing://entity/order" field Name id "billing://field/name" type string required one`
	first, firstDiagnostics := ParseFile("latent-fields.gooo", source)
	second, secondDiagnostics := ParseFile("latent-fields.gooo", source)
	if !reflect.DeepEqual(first, second) || !reflect.DeepEqual(firstDiagnostics, secondDiagnostics) {
		t.Fatal("proposed field source was not rejected deterministically")
	}
	if len(firstDiagnostics) == 0 || firstDiagnostics[0].Code != DiagUnexpectedDeclaration {
		t.Fatalf("field source diagnostics = %#v", firstDiagnostics)
	}
	entity := first.Declarations[0].(*EntityDecl)
	if len(entity.Fields) != 0 {
		t.Fatalf("rejected source produced partial field AST: %#v", entity.Fields)
	}
}
