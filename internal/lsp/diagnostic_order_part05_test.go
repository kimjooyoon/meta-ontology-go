package lsp

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"testing"
)

func rotateDiagnostics(values syntax.Diagnostics, shift int) syntax.Diagnostics {
	result := append(syntax.Diagnostics(nil), values...)
	if len(result) == 0 {
		return result
	}
	shift %= len(result)
	return append(result[shift:], result[:shift]...)
}
func assertDiagnosticCodes(t *testing.T, diagnostics []Diagnostic, want []string) {
	t.Helper()
	if len(diagnostics) != len(want) {
		t.Fatalf("diagnostics = %#v, want %v", diagnostics, want)
	}
	for index, code := range want {
		if diagnostics[index].Code != code {
			t.Fatalf("diagnostics = %#v, want %v", diagnostics, want)
		}
	}
}
