package syntax

import (
	"strings"
	"testing"
)

func TestActivityValueProgramParsesAndFormatsCanonically(t *testing.T) {
	source := `package valuewitness
namespace valuewitness
entity Integer id "gooo://value-witness/entity/integer"
activity Increment(Integer) -> Integer computes "int.add:1"`
	file, diagnostics := Parse(source)
	if diagnostics.HasErrors() {
		t.Fatalf("parse diagnostics: %v", diagnostics)
	}
	activity := file.Declarations[1].(*ActivityDecl)
	if !activity.ValueProgramPresent || activity.ValueProgram != "int.add:1" || activity.ValueProgramSpan.IsEmpty() {
		t.Fatalf("value program was not retained: %#v", activity)
	}
	var output strings.Builder
	if err := formatActivity(&output, activity); err != nil {
		t.Fatal(err)
	}
	if got, want := output.String(), `activity Increment(Integer) -> Integer computes "int.add:1"`; got != want {
		t.Fatalf("formatted activity = %q, want %q", got, want)
	}
}

func TestActivityValueProgramRequiresQuotedProgram(t *testing.T) {
	_, diagnostics := Parse(`activity Increment(Integer) -> Integer computes int.add:1`)
	if !diagnostics.HasErrors() {
		t.Fatal("unquoted value program was accepted")
	}
}
