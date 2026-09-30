package bodycodegen

import (
	"context"
	"fmt"
	"go/format"
	"testing"
)

func TestRouteEquivalenceNormalizesOnlyOuterExpressionParentheses(t *testing.T) {
	for _, condition := range []string{
		"input >= 3 && input <= 8",
		"input < 0 || input > 0",
		"(input + 1) * 2 < 10",
	} {
		for _, route := range []string{preserveRoute, guardReturnRoute, mergeResultRoute} {
			t.Run(condition+"/"+route, func(t *testing.T) {
				body := "if (((" + condition + "))) { return 2 } else { return -1 }"
				generated, err := generateRoute("bodycodegen", "Choose", "parenthesis-test", "int64", "int64", body, route)
				if err != nil {
					t.Fatal(err)
				}
				if generated.report.RouteEquivalence.Decision != "PASS" || !generated.report.RouteEquivalence.Equivalent {
					t.Fatalf("formatting changed equivalence: %#v", generated.report.RouteEquivalence)
				}
			})
		}
	}
}

func TestRouteEquivalenceRetainsInnerGroupingAndOperandOrder(t *testing.T) {
	for _, changed := range []string{
		"input + 1 * 2 < 10",
		"(input + 2) * 1 < 10",
	} {
		t.Run(changed, func(t *testing.T) {
			body := "if (((input + 1) * 2 < 10)) { return 2 } else { return -1 }"
			generated, err := format.Source([]byte(fmt.Sprintf("package bodycodegen\nfunc Choose(input int64) int64 { if %s { return 2 }; return -1 }", changed)))
			if err != nil {
				t.Fatal(err)
			}
			receipt, err := routeEquivalence("bodycodegen", "Choose", "int64", "int64", body, generated, "test-changed-condition")
			if err != nil {
				t.Fatal(err)
			}
			if receipt.Equivalent || receipt.Decision != "FAIL_CLOSED" {
				t.Fatalf("changed grouping received a pass: %#v", receipt)
			}
		})
	}
}

func TestIRBodyFillSupportsBooleanConditionHole(t *testing.T) {
	source := []byte(`package bodycodegen
namespace bodycodegen
entity Integer id "bodycodegen://entity/integer"
activity Choose(Integer) -> Integer computes "if __GOOO_BODY_HOLE_choice__ { return 2 } else { return -1 }"
`)
	plan := IRBodyFillPlan{Schema: bodyFillPlanSchema, Intent: "Choose the inclusive band from 3 to 8", HoleID: "choice",
		Candidates: []IRBodyFillCandidate{
			{ID: "inclusive", Expression: "input >= 3 && input <= 8"},
			{ID: "exclusive", Expression: "input > 3 && input < 8"},
		},
		TestCases: []IRBodyFillTestCase{
			{Input: 2, Expected: -1}, {Input: 3, Expected: 2},
			{Input: 8, Expected: 2}, {Input: 9, Expected: -1},
		},
	}
	result, err := GenerateWithIRBodyFill(context.Background(), "condition.gooo", source, "Choose", plan, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if result.Report.BodyFill.TestCasesPassed != 4 || result.Report.RouteEquivalence.Decision != "PASS" || result.Report.RouteEquivalence.Method != routeEquivalenceMethod {
		t.Fatalf("Boolean hole lacks matching finite evidence: %#v", result.Report)
	}
}
