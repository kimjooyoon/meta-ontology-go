package bodycodegen

import (
	"context"
	"strings"
	"testing"
)

func TestUnusedLocalKeepsInitializationAndAssignment(t *testing.T) {
	body := "let future = input0 / input1\nfuture = 3\nreturn input0"
	generated, err := Generate("unused.gooo", divisionSource("Keep", body), "Keep")
	if err != nil {
		t.Fatal(err)
	}
	for _, fragment := range []string{"var future = input0 / input1", "_ = future", "future = 3"} {
		if !strings.Contains(generated.Source, fragment) {
			t.Fatal("projection lost local evaluation", fragment, generated.Source)
		}
	}
	if !generated.Report.TypecheckPassed || !generated.Report.RouteEquivalence.Equivalent ||
		generated.Report.SourceSemanticUnits != generated.Report.LoweredSemanticUnits {
		t.Fatal("target-local read changed source accounting", generated.Report)
	}
	cases := []IRBodyFillTestCase{{Inputs: []int64{8, 2}, Expected: 8}}
	if _, passed, err := evaluateIntegerCases([]byte(generated.Source), "Keep", cases); err != nil || passed != 1 {
		t.Fatal("unused local changed value", passed, err)
	}
	if _, _, err := evaluateIntegerCases([]byte(generated.Source), "Keep", []IRBodyFillTestCase{{Inputs: []int64{8, 0}}}); err == nil || !strings.Contains(err.Error(), "zero divisor") {
		t.Fatal("unused initializer was dropped", err)
	}
	projection := strings.TrimPrefix(generated.Source, "package division\n")
	oracle := "package main\nimport \"fmt\"\n" + projection + `
func main() {
    fmt.Println(Keep(8, 2))
    defer func() { fmt.Println(recover() != nil) }()
    Keep(8, 0)
}
`
	if got := strings.TrimSpace(runBodyFillGoOracle(t, oracle)); got != "8\ntrue" {
		t.Fatal("native initialization differs", got)
	}
}

func TestUnusedLocalRecordPreflightAndReplay(t *testing.T) {
	source := string(recordUnusedCombinationFixture(t))
	source = strings.ReplaceAll(source, "saved.title", "copy.title")
	source = strings.ReplaceAll(source, "saved.state", "copy.state")
	source = strings.Replace(source, "let saved = copy", "let saved = copy\nlet label = \"pending\"\nlet flag = input1", 1)
	ctx := context.Background()
	plan, err := prepareRecordAssembly(ctx, "unused-record.gooo", []byte(source), "Select")
	if err != nil {
		t.Fatal("unused candidate locals failed preflight", err)
	}
	if flow := recordValueFlow(plan); flow.Status != "RESOLVED" {
		t.Fatal("unused locals prevented source flow", flow.Reason)
	}
	g, err := NewTypedPathGenerator("")
	if err != nil {
		t.Fatal(err)
	}
	result, err := g.GenerateSourceAssembly(ctx, "unused-record.gooo", []byte(source), "Select")
	if err != nil || result.Report.RecordAssembly.FieldsPassed != 15 {
		t.Fatal("unused locals prevented construction", err)
	}
	if strings.Contains(result.GoooSource, "_ = saved") || !strings.Contains(result.Source, "_ = saved") {
		t.Fatal("Go-specific local read leaked into Gooo source")
	}
	if replay, err := RealizeSourceAssembly(ctx, "unused-record.gooo", []byte(source), result); err != nil || replay.ModelCalls != 0 {
		t.Fatal("unused-local record did not replay", err)
	}
}

func TestUnusedLocalStillChecksInitializerTypes(t *testing.T) {
	for _, body := range []string{"let future = missing\nreturn input0", "let future = input0 / 0\nreturn input0",
		"let future = true + input0\nreturn input0"} {
		if _, err := Generate("invalid.gooo", divisionSource("Invalid", body), "Invalid"); err == nil {
			t.Fatal("unused initializer bypassed validation", body)
		}
	}
}

func TestUnusedLocalTypedRecipe(t *testing.T) {
	ctx := context.Background()
	source := recipeSource("let future = input + 1\nreturn input - 2")
	choice := strings.Replace(recipeOperand, `"occurrence":0`, `"occurrence":1`, 1)
	document, err := DecodeSourcePathDocument(ctx, "unused-recipe.gooo", source, "Assemble", recipeBytes(choice))
	if err != nil {
		t.Fatal("unused local prevented source binding", err)
	}
	result, err := GenerateWithTypedPaths(ctx, "unused-recipe.gooo", source, "Assemble", document, "")
	if err != nil {
		t.Fatal("unused local prevented typed assembly", err)
	}
	if result.Report.BodyPaths.FunctionalCompleteness != 100 ||
		!result.Report.BodyPaths.SourceBinding.Equivalent || !strings.Contains(result.Source, "_ = future") ||
		strings.Contains(result.GoooSource, "_ = future") {
		t.Fatal("unused local changed typed assembly evidence", result)
	}
}

func TestUnusedLocalRetainsPureCallEvaluation(t *testing.T) {
	source := string(divisionSource("Root", "let ignored = Divide(input0, input1)\nreturn input0")) +
		"activity Divide(Integer, Integer) -> Integer computes `let prepared = input0 + input1\nreturn input0 / input1`\n"
	result, err := Generate("unused-call.gooo", []byte(source), "Root")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result.Source, "_ = ignored") || !strings.Contains(result.Source, "_ = prepared") ||
		!result.Report.RouteEquivalence.Equivalent {
		t.Fatal("called-body projection lost unused-local semantics", result)
	}
	if _, _, err := evaluateIntegerCases([]byte(result.Source), "Root", []IRBodyFillTestCase{{Inputs: []int64{8, 0}}}); err == nil {
		t.Fatal("unused pure call was skipped")
	}
	oracle := "package main\nimport \"fmt\"\n" + strings.TrimPrefix(result.Source, "package division\n") + `
func main() {
    fmt.Println(Root(8, 2))
    defer func() { fmt.Println(recover() != nil) }()
    Root(8, 0)
}
`
	if got := strings.TrimSpace(runBodyFillGoOracle(t, oracle)); got != "8\ntrue" {
		t.Fatal("native unused pure call was skipped", got)
	}
}

func TestUnusedLocalInitializationStaysInsideBranch(t *testing.T) {
	for _, condition := range []string{"input1 != 0", "input1 == 0"} {
		body := "if " + condition + " { let future = input0 / input1 }; return input0"
		result, err := Generate("branch.gooo", divisionSource("Branch", body), "Branch")
		if err != nil {
			t.Fatal(err)
		}
		_, passed, err := evaluateIntegerCases([]byte(result.Source), "Branch", []IRBodyFillTestCase{{Inputs: []int64{8, 0}, Expected: 8}})
		if condition == "input1 != 0" && (err != nil || passed != 1) {
			t.Fatal("unentered branch evaluated an unused initializer", passed, err)
		}
		if condition == "input1 == 0" && (err == nil || !strings.Contains(err.Error(), "zero divisor")) {
			t.Fatal("entered branch dropped an unused initializer", err)
		}
	}
}

func TestLocalReadMarkerCanonicalization(t *testing.T) {
	for _, tc := range []struct {
		name, original, projected string
		equal                     bool
	}{
		{"adjacent", "var future = input\nreturn input", "var future = input\n_ = future\nreturn input", true},
		{"multiple", "var x = input\nvar y = input + 1\nreturn input", "var x = input\n_ = x\nvar y = input + 1\n_ = y\nreturn input", true},
		{"nested", "if input > 0 {\nvar x = input\n}\nreturn input", "if input > 0 {\nvar x = input\n_ = x\n}\nreturn input", true},
		{"raw text", "var text = `first\n\nlast`\nreturn input", "var text = `first\n\nlast`\n_ = text\nreturn input", true},
		{"changed text", "var text = `first\n\nlast`\nreturn input", "var text = `first\nlast`\n_ = text\nreturn input", false},
		{"changed initializer", "var x = input\nreturn input", "var x = input + 1\n_ = x\nreturn input", false},
		{"discarded computation", "var x = input\nreturn input", "var x = input\n_ = input / x\nreturn input", false},
		{"other name", "var x = input\nreturn input", "var x = input\n_ = input\nreturn input", false},
		{"nonadjacent", "var x = input\nx = 3\nreturn input", "var x = input\nx = 3\n_ = x\nreturn input", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			generated := []byte("package sample\nfunc Assemble(input int64) int64 {\n" + tc.projected + "\n}")
			receipt, err := routeEquivalence("sample", "Assemble", "int64", "int64", tc.original, generated, "unused_local")
			if err != nil || receipt.Equivalent != tc.equal {
				t.Fatal("local read normalization changed semantics", receipt, err)
			}
		})
	}
}
