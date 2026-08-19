package lsp

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"reflect"
	"testing"
)

func TestAdapterReplayUsesAllEqualStartTieBreaks(t *testing.T) {
	source := "0123456789012345678901234567890123456789"
	raw := syntax.Diagnostics{
		syntaxDiagnosticAt("cross-view.gooo", 20, 23, syntax.SeverityWarning, "a.code", "warning"),
		syntaxDiagnosticAt("cross-view.gooo", 10, 12, syntax.SeverityError, "early", "early"),
		syntaxDiagnosticAt("cross-view.gooo", 20, 22, syntax.SeverityError, "z.code", "short-end"),
		syntaxDiagnosticAt("cross-view.gooo", 20, 23, syntax.SeverityError, "z.code", "z-message"),
		syntaxDiagnosticAt("cross-view.gooo", 20, 23, syntax.SeverityError, "a.code", "z-message"),
		syntaxDiagnosticAt("cross-view.gooo", 20, 23, syntax.SeverityError, "a.code", "a-message"),
	}
	want := expectedAdapterReplay(t, source, raw)
	for replay := 0; replay < 64; replay++ {
		permuted := rotateDiagnostics(raw, replay)
		result, err := adaptSyntaxResult("cross-view.gooo", source, &syntax.File{}, permuted)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(want, result.Diagnostics) {
			t.Fatalf("replay %d changed adapter order: want=%#v got=%#v", replay, want, result.Diagnostics)
		}
	}
	assertDiagnosticPairs(t, want, []diagnosticPair{
		{Code: "early", Message: "early"},
		{Code: "z.code", Message: "short-end"},
		{Code: "a.code", Message: "a-message"},
		{Code: "a.code", Message: "z-message"},
		{Code: "z.code", Message: "z-message"},
		{Code: "a.code", Message: "warning"},
	})
}
