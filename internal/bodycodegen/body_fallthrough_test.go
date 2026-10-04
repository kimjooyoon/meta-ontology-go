package bodycodegen

import (
	"context"
	"strconv"
	"strings"
	"testing"
)

func TestPreservedBodiesSupportGuardReturnAndAssignmentFallthrough(t *testing.T) {
	for _, body := range []string{
		"if input < 0 { return 0 }; return input",
		"let value = input; if value < 0 { value = 0 }; return value",
		"let value = input; if value < 2 { if value < 0 { value = 0 } }; return value",
		"if input < 0 { return 0 }; if input < 10 { return input }; return 10",
	} {
		source := []byte("package fieldflow\nnamespace fieldflow\nentity Integer id \"fallthrough://integer\"\nactivity Choose(Integer) -> Integer computes " + strconv.Quote(body) + "\n")
		result, err := Generate("fallthrough.gooo", source, "Choose")
		if err != nil || !result.Report.TypecheckPassed || !result.Report.DeterministicReplay || !result.Report.RouteEquivalence.Equivalent ||
			result.Report.Route != preserveRoute || strings.Contains(result.Source, "else {") {
			t.Fatalf("guard/fallthrough projection failed: %v %+v", err, result.Report)
		}
	}
}

func TestRecordAssemblyGuardReturnAndNestedUpdatesReplay(t *testing.T) {
	source := string(recordAssemblyFixture(t))
	old := `let copy = input0
if input1 && copy.state != "ready" {
    copy = Candidate{title: "draft", state: "wait", reason: "deferred"}
} else {
    copy = input0
}
return copy`
	constructor := `Candidate{title: "draft", state: "wait", reason: "deferred"}`
	for _, body := range []string{
		`let copy = input0; if input1 && copy.state != "ready" { copy = ` + constructor + ` }; return copy`,
		`if !input1 || input0.state == "ready" { return input0 }; return ` + constructor,
		`let copy = input0; if input1 { if copy.state != "ready" { copy = ` + constructor + ` } }; return copy`,
	} {
		changed := []byte(strings.Replace(source, old, body, 1))
		g, _ := NewTypedPathGenerator("")
		result, err := g.GenerateSourceAssembly(context.Background(), "record-fallthrough.gooo", changed, "Select")
		if err != nil || result.Report.RecordAssembly == nil || result.Report.RecordAssembly.FieldsPassed != 15 {
			t.Fatal("current record guard/fallthrough selection", err)
		}
		if _, err := RealizeSourceAssembly(context.Background(), "record-fallthrough.gooo", changed, result); err != nil {
			t.Fatal("current record guard/fallthrough replay", err)
		}
	}
}

func TestFallthroughStillRequiresReturnAndValidScope(t *testing.T) {
	for _, body := range []string{
		"if input < 0 { return 0 }",
		"if input < 0 { let branch = 0 }; return branch",
		"let chosen = input; if input < 0 { input = 0 }; return chosen",
		"let chosen = input; if input < 0 { chosen = true }; return chosen",
	} {
		source := []byte("package fieldflow\nnamespace fieldflow\nentity Integer id \"fallthrough://integer\"\nactivity Choose(Integer) -> Integer computes " + strconv.Quote(body) + "\n")
		if _, err := Generate("invalid-fallthrough.gooo", source, "Choose"); err == nil {
			t.Fatal("missing return, escaped branch scope, parameter write or wrong type accepted", body)
		}
	}
}
