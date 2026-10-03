package bodycodegen

import (
	"os"
	"strconv"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestGenerateRawComputesMatchesQuotedBody(t *testing.T) {
	raw, err := os.ReadFile("../../examples/body-codegen/raw-computes.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	quoted, diagnostics, err := syntax.FormatSource("main.gooo", string(raw))
	if err != nil || diagnostics.HasErrors() {
		t.Fatal(err)
	}
	a, err := Generate("main.gooo", raw, "ClampBelowZero")
	if err != nil {
		t.Fatal(err)
	}
	b, err := Generate("main.gooo", []byte(quoted), "ClampBelowZero")
	if err != nil {
		t.Fatal(err)
	}
	if a.Source != b.Source || a.Report.Decision != "PASS" || !a.Report.TypecheckPassed ||
		!a.Report.DeterministicReplay || a.Report.SourceSemanticUnits != b.Report.SourceSemanticUnits {
		t.Fatalf("raw/quoted generated source differs:\nraw:\n%s\nquoted:\n%s", a.Source, b.Source)
	}
	assertCompletenessReceipt(t, a.Report.CompletenessReceipt)
}

func TestGenerateRawComputesPreservesTextEscapes(t *testing.T) {
	prefix := "package raw\nnamespace raw\nentity Text id \"raw://entity/text\"\nactivity Message(Text) -> Text computes "
	body := `return "한글 \"a\"\\n"`
	a, err := Generate("text.gooo", []byte(prefix+"`"+body+"`\n"), "Message")
	if err != nil {
		t.Fatal(err)
	}
	b, err := Generate("text.gooo", []byte(prefix+strconv.Quote(body)+"\n"), "Message")
	if err != nil || a.Source != b.Source || a.Report.Decision != "PASS" {
		t.Fatal("raw outer literal changed inner text escapes", err, a.Source, b.Source)
	}
}
