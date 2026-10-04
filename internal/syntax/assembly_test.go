package syntax

import (
	"reflect"
	"strings"
	"testing"
)

const assemblySource = `package assembly
namespace assembly
entity Integer id "assembly://integer"
activity Adjust(Integer) -> Integer computes "return input - 2" assembling {
    choice "offset" operand_order at "0" intent "상수에서 입력을 뺀다. Subtract input from the constant."
    case "-9223372036854775808" -> "-9223372036854775806"
    case "4" -> "-2"
    attempts "2"
}`

func TestAssemblyParseFormatCloneAndSpans(t *testing.T) {
	testAssemblyParseFormatCloneAndSpans(t, assemblySource)
	checkpoint := strings.Replace(assemblySource, "assembling {", `assembling {
    baseline "return input - 2"
    picked "offset" -> "layout_reverse"`, 1)
	testAssemblyParseFormatCloneAndSpans(t, checkpoint)
}

func testAssemblyParseFormatCloneAndSpans(t *testing.T, assemblySource string) {
	t.Helper()
	file, diagnostics := ParseFile("inline.gooo", assemblySource)
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	activity := file.Declarations[1].(*ActivityDecl)
	if activity.Assembly == nil || activity.Assembly.Span.Filename != "inline.gooo" ||
		assemblySource[activity.Assembly.Span.Start.Offset:activity.Assembly.Span.End.Offset] !=
			assemblySource[strings.Index(assemblySource, "assembling"):] {
		t.Fatal("assembly lost its source span", activity)
	}
	formatted, err := Format(file)
	if err != nil {
		t.Fatal(err)
	}
	again, diagnostics := Parse(formatted)
	if diagnostics.HasErrors() || !reflect.DeepEqual(activity.Assembly.Spec, again.Declarations[1].(*ActivityDecl).Assembly.Spec) {
		t.Fatal("assembly changed during formatting", diagnostics)
	}
	fixed, err := Format(again)
	if err != nil || formatted != fixed {
		t.Fatal("assembly formatting is not a fixed point", err)
	}
	clone := file.Clone().Declarations[1].(*ActivityDecl)
	clone.Assembly.Spec.Choices[0].Intent = "changed"
	clone.Assembly.Spec.Cases[0].Expected = 0
	if activity.Assembly.Spec.Choices[0].Intent == "changed" || activity.Assembly.Spec.Cases[0].Expected == 0 {
		t.Fatal("syntax clone shares assembly storage")
	}
	if len(clone.Assembly.Spec.Picked) != 0 {
		clone.Assembly.Spec.Picked[0].Label = "layout_forward"
		if activity.Assembly.Spec.Picked[0].Label != "layout_reverse" {
			t.Fatal("syntax clone shares checkpoint storage")
		}
	}
}

func TestAssemblyRejectsIncompleteOrAmbiguousDeclarations(t *testing.T) {
	for _, mutation := range []struct{ from, to string }{
		{` computes "return input - 2"`, ""},
		{`operand_order`, `unknown_kind`},
		{`at "0"`, `at "128"`},
		{`at "0"`, `at "-1"`},
		{`at "0"`, `at "9223372036854775808"`},
		{`at "0"`, `within "0"`},
		{`intent "상수`, `alternative "other" intent "상수`},
		{`operand_order`, `local_reference`},
		{`attempts "2"`, `attempts "0"`},
		{`attempts "2"`, `attempts "65"`},
		{`attempts "2"`, `attempts "2" attempts "2"`},
		{`attempts "2"`, ""},
		{`case "4" -> "-2"`, `case "4" -> "bad"`},
		{`attempts "2"`, `attempts "2" seed "a" seed "b"`},
		{`attempts "2"`, `attempts "2" baseline ""`},
		{`attempts "2"`, `attempts "2" baseline "return input - 2"`},
		{`attempts "2"`, `attempts "2" picked "offset" -> "layout_reverse"`},
		{`attempts "2"`, `attempts "2" baseline "return input - 2" baseline "return input - 2"`},
	} {
		source := strings.Replace(assemblySource, mutation.from, mutation.to, 1)
		if _, diagnostics := Parse(source); !diagnostics.HasErrors() {
			t.Fatalf("invalid assembly accepted: %s -> %s", mutation.from, mutation.to)
		}
	}
	for _, block := range []string{
		strings.Repeat(`choice "same" operand_order at "0" intent "subtract" `, 17),
		`choice "one" operand_order at "0" intent "subtract" ` + strings.Repeat(`case "1" -> "1" `, 129),
		`case "1" -> "1" attempts "1"`,
		`choice "one" operand_order at "0" intent "subtract" attempts "1"`,
	} {
		prefix := assemblySource[:strings.Index(assemblySource, "assembling")]
		if _, diagnostics := Parse(prefix + "assembling { " + block + " }"); !diagnostics.HasErrors() {
			t.Fatal("missing or over-budget assembly accepted")
		}
	}
}
