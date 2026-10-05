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
	search := `package assembly
namespace assembly
entity Integer id "assembly://integer"
activity Clamp(Integer) -> Integer computes "return __GOOO_BODY_HOLE_floor__" assembling {
    search hole "floor" grammar "integer-offset-constant/v1" intent "Map the training domain." max_candidates "16"
    case "-1" -> "0"
    case "1" -> "1"
    holdout_case "-2" -> "0"
    attempts "4"
}`
	testAssemblyParseFormatCloneAndSpans(t, search)
	fill := `package assembly
namespace assembly
entity Integer id "assembly://integer"
activity Lift(Integer) -> Integer computes ` + "`" + `let base = __GOOO_BODY_HOLE_seed__
let increment = __GOOO_BODY_HOLE_step__
return base + increment` + "`" + ` assembling {
    source_fill intent "Compose a base and a small increment." {
        hole "seed"
        hole "step"
        candidate "add_one" {
            fill "seed" "input + 0"
            fill "step" "1"
        }
        candidate "double" {
            fill "seed" "input * 2"
            fill "step" "0"
        }
    }
    case "0" -> "1"
    case "2" -> "3"
}`
	testAssemblyParseFormatCloneAndSpans(t, fill)
	derived := strings.Replace(fill, `candidate "add_one" {`+"\n"+
		`            fill "seed" "input + 0"`+"\n"+
		`            fill "step" "1"`+"\n"+
		`        }`+"\n"+
		`        candidate "double" {`+"\n"+
		`            fill "seed" "input * 2"`+"\n"+
		`            fill "step" "0"`+"\n"+
		`        }`+"\n", `derive grammar "integer-offset-constant/v1" max_expressions "8" max_candidates "16"`+"\n", 1)
	testAssemblyParseFormatCloneAndSpans(t, derived)
	mixed := strings.Replace(derived,
		`derive grammar "integer-offset-constant/v1" max_expressions "8" max_candidates "16"`,
		"derive assignments max_candidates \"16\" {\n"+
			"            hole \"seed\" grammar \"integer-offset-constant/v1\" max_expressions \"8\"\n"+
			"            hole \"step\" grammar \"integer-predicate/v1\" max_expressions \"8\"\n"+
			"        }", 1)
	// The grammar declarations are type-checked by body generation; syntax parsing
	// still owns their canonical source representation and clone behavior.
	testAssemblyParseFormatCloneAndSpans(t, mixed)
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
	if clone.Assembly.Spec.FillPlan != nil {
		if clone.Assembly.Spec.FillPlan.Generation != nil {
			clone.Assembly.Spec.FillPlan.Generation.MaxCandidates++
			if activity.Assembly.Spec.FillPlan.Generation.MaxCandidates == clone.Assembly.Spec.FillPlan.Generation.MaxCandidates {
				t.Fatal("syntax clone shares source fill generation storage")
			}
			if len(clone.Assembly.Spec.FillPlan.Generation.HoleGrammars) > 0 {
				clone.Assembly.Spec.FillPlan.Generation.HoleGrammars[0].Grammar = "changed"
				if activity.Assembly.Spec.FillPlan.Generation.HoleGrammars[0].Grammar == "changed" {
					t.Fatal("syntax clone shares per-hole grammar storage")
				}
			}
		} else {
			clone.Assembly.Spec.FillPlan.Candidates[0].Fills[0].Expression = "changed"
			if activity.Assembly.Spec.FillPlan.Candidates[0].Fills[0].Expression == "changed" {
				t.Fatal("syntax clone shares source fill plan storage")
			}
		}
	} else if clone.Assembly.Spec.Search != nil {
		originalIntent := activity.Assembly.Spec.Search.Intent
		originalHoldout := activity.Assembly.Spec.HoldoutCases[0].Expected
		clone.Assembly.Spec.Search.Intent = "changed"
		clone.Assembly.Spec.HoldoutCases[0].Expected = 5
		if activity.Assembly.Spec.Search.Intent != originalIntent || activity.Assembly.Spec.HoldoutCases[0].Expected != originalHoldout {
			t.Fatal("syntax clone shares search assembly storage")
		}
	} else {
		clone.Assembly.Spec.Choices[0].Intent = "changed"
		clone.Assembly.Spec.Cases[0].Expected = 0
		if activity.Assembly.Spec.Choices[0].Intent == "changed" || activity.Assembly.Spec.Cases[0].Expected == 0 {
			t.Fatal("syntax clone shares assembly storage")
		}
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

func TestAssemblyIRSearchRejectsMixedAndLeakingContracts(t *testing.T) {
	source := `package assembly
namespace assembly
entity Integer id "assembly://integer"
activity Clamp(Integer) -> Integer computes "return __GOOO_BODY_HOLE_floor__" assembling {
    search hole "floor" grammar "integer-offset-constant/v1" intent "Map values." max_candidates "8"
    case "-1" -> "0"
    holdout_case "1" -> "1"
    attempts "4"
}`
	for _, mutation := range []struct{ from, to string }{
		{`grammar "integer-offset-constant/v1"`, `grammar "arbitrary-go/v1"`},
		{`max_candidates "8"`, `max_candidates "1"`},
		{`intent "Map values."`, `intent ""`},
		{`hole "floor"`, `hole "bad-id!"`},
		{`holdout_case "1" -> "1"`, `holdout_case "-1" -> "1"`},
		{`attempts "4"`, `attempts "9"`},
		{`attempts "4"`, `attempts "4" choice "x" operand_order at "0" intent "mixed"`},
	} {
		changed := strings.Replace(source, mutation.from, mutation.to, 1)
		if _, diagnostics := Parse(changed); !diagnostics.HasErrors() {
			t.Fatalf("invalid IR search contract accepted: %q -> %q", mutation.from, mutation.to)
		}
	}
}

func TestSourceFillDerivationRejectsOpenOrUnboundedContracts(t *testing.T) {
	source := `package sample
namespace sample
entity Integer id "sample://integer"
activity Lift(Integer) -> Integer computes ` + "`" + `let base = __GOOO_BODY_HOLE_seed__
let increment = __GOOO_BODY_HOLE_step__
return base + increment` + "`" + ` assembling {
    source_fill intent "Compose a finite assignment." {
        hole "seed"
        hole "step"
        derive grammar "integer-offset-constant/v1" max_expressions "8" max_candidates "8"
    }
    case "0" -> "1"
}`
	for _, mutation := range []struct{ from, to string }{
		{`integer-offset-constant/v1`, `arbitrary-go/v1`},
		{`max_expressions "8"`, `max_expressions "1"`},
		{`max_candidates "8"`, `max_candidates "1"`},
		{`max_candidates "8"`, `max_candidates "17"`},
		{`derive grammar "integer-offset-constant/v1" max_expressions "8" max_candidates "8"`,
			`derive grammar "integer-offset-constant/v1" max_expressions "8" max_candidates "8" candidate "manual" { fill "seed" "input" fill "step" "1" }`},
	} {
		changed := strings.Replace(source, mutation.from, mutation.to, 1)
		if _, diagnostics := Parse(changed); !diagnostics.HasErrors() {
			t.Fatalf("invalid source-fill generation contract accepted: %q -> %q", mutation.from, mutation.to)
		}
	}
}
