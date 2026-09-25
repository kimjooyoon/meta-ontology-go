package syntax

import "testing"

func TestParseValidTable(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		packageName string
		namespace   string
		entities    int
		activities  int
	}{
		{
			name: "billing example",
			source: `package billing
namespace billing
entity Order id "billing://entity/order"
entity PaymentMethod id "billing://entity/payment-method"
entity Payment id "billing://entity/payment"
activity PayOrder(Order, PaymentMethod) -> Payment`,
			packageName: "billing", namespace: "billing", entities: 3, activities: 1,
		},
		{
			name: "empty parameter list",
			source: `package p
namespace n
activity Tick() -> Result`,
			packageName: "p", namespace: "n", activities: 1,
		},
		{
			name: "unicode identifiers",
			source: `package 도메인
namespace 도메인
entity 주문 id "urn:order"
activity 결제(주문) -> 주문`,
			packageName: "도메인", namespace: "도메인", entities: 1, activities: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			file, diagnostics := Parse(test.source)
			if len(diagnostics) != 0 {
				t.Fatalf("unexpected diagnostics: %v", diagnostics)
			}
			if file.Package == nil || file.Package.Name != test.packageName || file.Namespace == nil || file.Namespace.Name != test.namespace {
				t.Fatalf("headers = %#v", file)
			}
			if len(file.Declarations) != test.entities+test.activities || len(file.Decls) != len(file.Declarations) {
				t.Fatalf("declarations = %#v", file.Declarations)
			}
			for _, declaration := range file.Declarations {
				switch declaration.(type) {
				case *EntityDecl:
					if test.entities == 0 {
						t.Fatalf("unexpected entity declaration")
					}
					test.entities--
				case *ActivityDecl:
					if test.activities == 0 {
						t.Fatalf("unexpected activity declaration")
					}
					test.activities--
				default:
					t.Fatalf("unexpected declaration type %T", declaration)
				}
			}
			if test.entities != 0 || test.activities != 0 {
				t.Fatalf("declaration counts were not preserved")
			}
		})
	}
}
