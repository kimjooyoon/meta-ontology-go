package bidir

import (
	"strconv"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func whitespaceProgramSyntax(t *testing.T, program string) *syntax.File {
	t.Helper()
	source := "package whitespace\nnamespace whitespace\n" +
		"entity Text id \"gooo://std/text\"\n" +
		"activity Main(Text) -> Text id \"gooo://whitespace/main\" computes " + strconv.Quote(program)
	file, diagnostics := syntax.Parse(source)
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	return file
}

func TestSourceComputesWhitespaceKeepsBodyContentAndOriginalSpan(t *testing.T) {
	canonical := `return " left " + input + " right "`
	raw := "\n\t" + canonical + "\n  "
	file := whitespaceProgramSyntax(t, raw)
	activity := file.Declarations[1].(*syntax.ActivityDecl)
	span := activity.ValueProgramSpan
	document, err := DocumentFromSyntax(file)
	if err != nil {
		t.Fatal(err)
	}
	if activity.ValueProgram != raw || activity.ValueProgramSpan != span || span.IsEmpty() {
		t.Fatal("lowering altered the original body or its source position")
	}
	if document.Declarations[1].Attributes[ActivityValueProgramAttribute] != canonical {
		t.Fatal("surrounding whitespace was not normalized, or literal content changed")
	}
	core, err := LowerDocument(document)
	if err != nil {
		t.Fatal(err)
	}
	other, err := Lower(whitespaceProgramSyntax(t, canonical))
	if err != nil || core.StableHash() != other.StableHash() {
		t.Fatal("presentation whitespace changed semantic meaning", err)
	}
}

func TestWhitespaceOnlyComputesStillHasNoExecutableBody(t *testing.T) {
	if _, err := Lower(whitespaceProgramSyntax(t, "\n\t  ")); err == nil {
		t.Fatal("empty body was silently accepted")
	}
	file := whitespaceProgramSyntax(t, "return input")
	document, err := DocumentFromSyntax(file)
	if err != nil {
		t.Fatal(err)
	}
	// Parser-neutral callers still provide canonical attributes. Normalization is
	// owned by the syntax adapter, rather than by arbitrary semantic mutations.
	document.Declarations[1].Attributes[ActivityValueProgramAttribute] = "\nreturn input\n"
	if _, err := LowerDocument(document); err == nil || !strings.Contains(err.Error(), "non-canonical") {
		t.Fatal("non-canonical semantic mutation was accepted", err)
	}
}
