package syntax

import (
	"reflect"
	"strings"
	"testing"
)

func TestDeferredEntityFieldsRejectsWithoutPartialASTOrWrite(t *testing.T) {
	source := `package billing
namespace billing
entity Order id "billing://entity/order" fields {
    field Number id "billing://field/number" type string required one
}

activity PayOrder(Order) -> Order`
	filename := "entity-fields.gooo"
	original := source
	want := Diagnostic{
		Severity: SeverityError,
		Code:     DiagEntityFieldsDeferred,
		Message:  "entity fields are deferred and unsupported by the public syntax",
		Span: Span{
			Filename: filename,
			Start:    Position{Offset: 75, Line: 3, Column: 42},
			End:      Position{Offset: 81, Line: 3, Column: 48},
		},
	}

	file, diagnostics := ParseFile(filename, source)
	if file != nil || !reflect.DeepEqual(diagnostics, Diagnostics{want}) {
		t.Fatalf("deferred parse result = %#v, %#v; want no AST and %#v", file, diagnostics, want)
	}
	formatted, formatDiagnostics, err := FormatSource(filename, source)
	if formatted != "" || !reflect.DeepEqual(formatDiagnostics, diagnostics) ||
		err == nil || err.Error() != want.String() {
		t.Fatalf("deferred format result = %q, %#v, %v", formatted, formatDiagnostics, err)
	}
	secondFile, secondDiagnostics := ParseFile(filename, source)
	if secondFile != nil || !reflect.DeepEqual(secondDiagnostics, diagnostics) || source != original {
		t.Fatal("deferred rejection was not deterministic or source was changed")
	}
	if !strings.Contains(source, "fields {") {
		t.Fatal("test source lost its EntityFields marker")
	}
}
func TestFieldlessBillingSyntaxRemainsSupportedWhileFieldsAreDeferred(t *testing.T) {
	source := `package billing
namespace billing

entity Order id "billing://entity/order"
entity PaymentMethod id "billing://entity/payment-method"
entity Payment id "billing://entity/payment"

activity PayOrder(Order, PaymentMethod) -> Payment
`
	file, diagnostics := ParseFile("billing.gooo", source)
	if len(diagnostics) != 0 || file == nil || len(file.Declarations) != 4 {
		t.Fatalf("billing parse = %#v, %#v", file, diagnostics)
	}
	formatted, err := Format(file)
	want := "package billing\nnamespace billing\n\n" +
		"entity Order id \"billing://entity/order\"\n" +
		"entity PaymentMethod id \"billing://entity/payment-method\"\n" +
		"entity Payment id \"billing://entity/payment\"\n" +
		"activity PayOrder(Order, PaymentMethod) -> Payment\n"
	if err != nil || formatted != want {
		t.Fatalf("billing format = %q, %v", formatted, err)
	}
}
