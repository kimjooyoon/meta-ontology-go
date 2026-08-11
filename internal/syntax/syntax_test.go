package syntax

import "testing"

func TestParseBillingExampleShape(t *testing.T) {
	source := `package billing
namespace billing
entity Order id "billing://entity/order"
activity PayOrder(Order, PaymentMethod) -> Payment`
	file, diagnostics := Parse(source)
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diagnostics)
	}
	if file.Package == nil || file.Package.Name != "billing" || file.Namespace == nil || file.Namespace.Name != "billing" || len(file.Declarations) != 2 {
		t.Fatalf("unexpected file: %#v", file)
	}
	activity, ok := file.Declarations[1].(*ActivityDecl)
	if !ok || activity.Name != "PayOrder" || len(activity.Parameters) != 2 || activity.Result.Name != "Payment" {
		t.Fatalf("unexpected activity: %#v", file.Declarations[1])
	}
}

func TestLexSkipsCommentsAndReportsBadCharacter(t *testing.T) {
	tokens, diagnostics := Lex("// comment\nentity A id \"x\" @")
	if len(tokens) < 5 || tokens[0].Text != "entity" {
		t.Fatalf("unexpected tokens: %#v", tokens)
	}
	if len(diagnostics) != 1 || diagnostics[0].Code != DiagUnexpectedCharacter {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
}

func TestParseUnterminatedString(t *testing.T) {
	_, diagnostics := Parse(`entity A id "missing`)
	if len(diagnostics) == 0 {
		t.Fatal("expected diagnostic")
	}
}
