package bodycodegen

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

const pureCallsFixture = "package calls\nnamespace calls\nentity Integer id \"gooo://calls/integer\"\n" +
	"activity Twice(Integer) -> Integer computes `let result = input * 2; return result`\n" +
	"activity Add(Integer, Integer) -> Integer computes `return input0 + input1`\n" +
	"activity Score(Integer) -> Integer computes `if input < 0 { return Twice(-input) } else { return Add(Twice(input), 3) }`\n"

func TestPureActivityCallsGenerateInterpretAndBindDependencies(t *testing.T) {
	source := []byte(pureCallsFixture)
	result, err := Generate("calls.gooo", source, "Score")
	if err != nil {
		t.Fatal(err)
	}
	r := result.Report
	if r.CallClosure == nil || len(r.CallClosure.Activities) != 2 || len(r.CallClosure.Edges) != 3 || r.CallClosure.MaxCallsPerInvocation != 3 ||
		!r.RouteEquivalence.Equivalent || !r.DeterministicReplay {
		t.Fatal("missing typed call closure", r)
	}
	cases := []IRBodyFillTestCase{{Input: -5, Expected: 10}, {Input: 4, Expected: 11}, {Input: 9007199254740993, Expected: 18014398509481989}}
	_, passed, err := evaluateIntegerCasesContext(context.Background(), []byte(result.Source), "Score", cases)
	if err != nil || passed != len(cases) {
		t.Fatal("closed interpreter differs", passed, err)
	}
	changed, err := Generate("calls.gooo", []byte(strings.Replace(pureCallsFixture, "input * 2", "input * 3", 1)), "Score")
	if err != nil || changed.Report.CallClosure.Activities[0].ProgramSHA256 == r.CallClosure.Activities[0].ProgramSHA256 ||
		changed.Report.RouteEquivalence.SourceSemanticDigest == r.RouteEquivalence.SourceSemanticDigest {
		t.Fatal("callee body was not included in semantic identity", err)
	}
}

func TestPureActivityCallsRejectUnknownEffectsAndCycles(t *testing.T) {
	for _, replacement := range []struct{ old, next string }{
		{"Twice(-input)", "Missing(input)"}, {"Twice(-input)", "float64(input)"},
		{"Twice(-input)", "Twice(\"text\")"}, {"Twice(-input)", "Twice(input, input)"},
		{"Twice(-input)", "Twice()"}, {"Twice(-input)", "external.Twice(input)"},
		{"input * 2", "Score(input)"}, {"let result = input * 2; return result", "input = 1; return input"},
		{"let result = input * 2; return result", "for { }; return input"},
	} {
		t.Run(replacement.next, func(t *testing.T) {
			source := strings.Replace(pureCallsFixture, replacement.old, replacement.next, 1)
			if _, err := Generate("calls.gooo", []byte(source), "Score"); err == nil {
				t.Fatal("unsupported pure call generated", replacement.next)
			}
		})
	}
}

func TestPureActivityCallsBoundRepeatedCallExpansion(t *testing.T) {
	var source strings.Builder
	source.WriteString("package calls\nnamespace calls\nentity Integer id \"gooo://calls/integer\"\nactivity F0(Integer) -> Integer computes `return input`\n")
	for i := 1; i <= 13; i++ {
		fmt.Fprintf(&source, "activity F%d(Integer) -> Integer computes `return F%d(input) + F%d(input)`\n", i, i-1, i-1)
	}
	if _, err := Generate("calls.gooo", []byte(source.String()), "F13"); err == nil || !strings.Contains(err.Error(), "4096") {
		t.Fatal("branching call closure escaped its static budget", err)
	}
}

func TestPureActivityCallRecordArgumentsAreValues(t *testing.T) {
	source := []byte("package calls\nnamespace calls\n" +
		"entity Item id \"gooo://calls/item\" fields { field value id \"gooo://calls/item/value\" type integer required one }\n" +
		"activity Copy(Item) -> Item computes `let result = input; result.value = input.value + 1; return result`\n" +
		"activity Run(Item) -> Item computes `let changed = Copy(input); return Item{value: changed.value + input.value}`\n")
	result, err := Generate("calls.gooo", source, "Run")
	if err != nil {
		t.Fatal(err)
	}
	cases := []assemblyspec.ValueCase{{Inputs: `[{"value":2}]`, Expected: `{"value":5}`},
		{Inputs: `[{"value":-3}]`, Expected: `{"value":-5}`}}
	observed, err := evaluateRecordAssembly(context.Background(), []byte(result.Source), "Run", result.Report.RecordTypes, cases)
	if err != nil || len(observed) != 2 || !observed[0].Passed || !observed[1].Passed {
		t.Fatal("record call changed a caller value", err, observed)
	}
}

func TestPureActivityCallsReusedClosureKeepsDepthBound(t *testing.T) {
	var source strings.Builder
	source.WriteString("package calls\nnamespace calls\nentity Integer id \"gooo://calls/integer\"\nactivity F0(Integer) -> Integer computes `return input`\n")
	for i := 1; i <= 16; i++ {
		fmt.Fprintf(&source, "activity F%d(Integer) -> Integer computes `return F%d(input)`\n", i, i-1)
	}
	for _, primed := range []int{0, 8, 16} {
		for _, last := range []int{15, 16} {
			if primed > last {
				continue
			}
			root := fmt.Sprintf("activity Run(Integer) -> Integer computes `return F%d(input) + F%d(input)`\n", primed, last)
			_, err := Generate("calls.gooo", []byte(source.String()+root), "Run")
			if last == 15 && err != nil {
				t.Fatal("valid reused closure rejected", primed, err)
			}
			if last == 16 && (err == nil || !strings.Contains(err.Error(), "16 call levels")) {
				t.Fatal("cached callee escaped the depth bound", primed, err)
			}
		}
	}
}
