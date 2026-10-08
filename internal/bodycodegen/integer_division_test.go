package bodycodegen

import (
	"context"
	"fmt"
	"go/token"
	"math"
	"strings"
	"testing"
)

func divisionSource(name, body string) []byte {
	return []byte("package division\nnamespace division\nentity Integer id \"division://integer\"\nactivity " + name + "(Integer, Integer) -> Integer computes `" + body + "`\n")
}

func TestIntegerDivisionAndRemainderMatchNativeGo(t *testing.T) {
	rows := [][4]int64{
		{7, 3, 2, 1}, {-7, 3, -2, -1}, {7, -3, -2, 1}, {-7, -3, 2, -1},
		{math.MinInt64, -1, math.MinInt64, 0}, {math.MinInt64, 1, math.MinInt64, 0},
		{math.MaxInt64, 2, math.MaxInt64 / 2, 1}, {1, math.MinInt64, 0, 1}, {0, 7, 0, 0},
	}
	oracle := "package main\nimport \"fmt\"\n"
	want, calls := []string{}, "func main() {\n"
	for index, operator := range []string{"/", "%"} {
		name := []string{"Quotient", "Remainder"}[index]
		generated, err := Generate("division.gooo", divisionSource(name, "return input0 "+operator+" input1"), name)
		if err != nil || !generated.Report.TypecheckPassed || !generated.Report.DeterministicReplay {
			t.Fatal("division body was not lowered", name, err)
		}
		var cases []IRBodyFillTestCase
		for _, row := range rows {
			cases = append(cases, IRBodyFillTestCase{Inputs: []int64{row[0], row[1]}, Expected: row[index+2]})
			calls += fmt.Sprintf("fmt.Println(%s(%d, %d))\n", name, row[0], row[1])
			want = append(want, fmt.Sprint(row[index+2]))
		}
		if _, passed, err := evaluateIntegerCases([]byte(generated.Source), name, cases); err != nil || passed != len(cases) {
			t.Fatal("candidate evaluator arithmetic differs", name, passed, err)
		}
		body, ok := strings.CutPrefix(generated.Source, "package division\n")
		if !ok {
			t.Fatal("unexpected generated package", generated.Source)
		}
		oracle += body + "\n"
	}
	if got := strings.TrimSpace(runBodyFillGoOracle(t, oracle+calls+"}\n")); got != strings.Join(want, "\n") {
		t.Fatal("compiled quotient/remainder differs from signed finite expectations", got)
	}
}

func TestIntegerDivisionZeroAndShortCircuit(t *testing.T) {
	for _, operator := range []token.Token{token.QUO, token.REM} {
		if _, err := evaluateIntegerBinary(operator, int64(4), int64(0)); err == nil || !strings.Contains(err.Error(), "zero") {
			t.Fatal("zero divisor must return an evaluation error", operator, err)
		}
		if _, err := Generate("zero.gooo", divisionSource("ConstantZero", "return input0 "+operator.String()+" 0"), "ConstantZero"); err == nil || !strings.Contains(err.Error(), "zero") {
			t.Fatal("constant zero divisor passed type checking", operator, err)
		}
		for _, body := range []string{
			"if input1 == 0 || input0 " + operator.String() + " input1 == 0 { return 1 }; return 0",
			"if input1 != 0 && input0 " + operator.String() + " input1 != 0 { return 0 }; return 1",
		} {
			generated, err := Generate("guard.gooo", divisionSource("Guard", body), "Guard")
			if err != nil {
				t.Fatal(err)
			}
			if _, passed, err := evaluateIntegerCases([]byte(generated.Source), "Guard", []IRBodyFillTestCase{{Inputs: []int64{4, 0}, Expected: 1}}); err != nil || passed != 1 {
				t.Fatal("short circuit evaluated a zero divisor", passed, err)
			}
		}
	}
}

func TestIntegerDivisionInsideDeclaredIRFill(t *testing.T) {
	source := []byte("package division\nnamespace division\nentity Integer id \"division://integer\"\nactivity Half(Integer) -> Integer computes `return __GOOO_BODY_HOLE_value__`\n")
	plan := IRBodyFillPlan{Schema: bodyFillPlanSchema, HoleID: "value", Intent: "integer half, truncated toward zero", Candidates: []IRBodyFillCandidate{
		{ID: "half", Expression: "input / 2"}, {ID: "parity", Expression: "input % 2"},
	}, TestCases: []IRBodyFillTestCase{{Input: 2, Expected: 1}, {Input: 3, Expected: 1}, {Input: -3, Expected: -1}}}
	result, err := GenerateWithIRBodyFill(context.Background(), "fill.gooo", source, "Half", plan, "", "")
	if err != nil || result.Report.BodyFill == nil || result.Report.BodyFill.SelectedCandidateID != "half" {
		t.Fatal("declared division candidate was not selected", err)
	}
}
