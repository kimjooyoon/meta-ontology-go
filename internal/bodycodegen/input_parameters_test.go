package bodycodegen

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestMultipleBodyParametersPreserveTypesAndAllLoweringRoutes(t *testing.T) {
	parameters := []InputParameter{{"input0", "int64"}, {"input1", "bool"}, {"input2", "string"}}
	body := `if input1 && input0 > 0 { return input2 + "!" } else { return input2 }`
	for _, route := range []string{preserveRoute, guardReturnRoute, mergeResultRoute} {
		result, err := generateRouteParameters("multi", "Label", "multi://label", parameters, "string", body, route)
		if err != nil {
			t.Fatalf("%s: %v", route, err)
		}
		if !result.report.RouteEquivalence.Equivalent || !reflect.DeepEqual(result.report.InputParameters, parameters) || !strings.Contains(string(result.source), "input0 int64, input1 bool, input2 string") {
			t.Fatalf("signature lost on %s: %+v", route, result.report)
		}
	}
	generated, err := generateRouteParameters("multi", "Label", "multi://label", parameters, "string", body, preserveRoute)
	if err != nil {
		t.Fatal(err)
	}
	changed := strings.Replace(string(generated.source), "input0 int64, input1 bool", "input1 bool, input0 int64", 1)
	receipt, err := routeEquivalenceParameters("multi", "Label", parameters, "string", body, []byte(changed), "test")
	if err != nil || receipt.Equivalent {
		t.Fatalf("reordered signature accepted: %v %+v", err, receipt)
	}
}

func TestMultipleBodyParametersAreImmutableAndBounded(t *testing.T) {
	for _, body := range []string{"input1 = 0; return input0", "let input0 = 1; return input1", "if input0 > 0 { input1 = 1; return input0 } else { return input1 }", "return input", "return input2", "return input0 + true"} {
		source := fmt.Sprintf("package multi\nnamespace multi\nentity Integer id \"multi://integer\"\nactivity Pair(Integer, Integer) -> Integer computes %q\n", body)
		if _, err := Generate("multi.gooo", []byte(source), "Pair"); err == nil {
			t.Fatalf("invalid body accepted: %s", body)
		}
	}
	inputs := strings.TrimSuffix(strings.Repeat("Integer, ", 16), ", ")
	source := fmt.Sprintf("package multi\nnamespace multi\nentity Integer id \"multi://integer\"\nactivity Full(%s) -> Integer computes \"return input0 + input15\"\n", inputs)
	result, err := GenerateWithPlanner(context.Background(), "multi.gooo", []byte(source), "Full", "", "")
	if err != nil || len(result.Report.InputParameters) != 16 {
		t.Fatalf("16 input bound: %v", err)
	}
	source = strings.Replace(source, "Full("+inputs, "Full("+inputs+", Integer", 1)
	if _, err := Generate("multi.gooo", []byte(source), "Full"); err == nil {
		t.Fatal("17 inputs accepted")
	}
}
