package bodycodegen

import (
	"fmt"
	"strconv"
	"testing"
)

func TestConditionOriginRetainedAcrossRoutes2303(t *testing.T) {
	for i, route := range []string{preserveRoute, guardReturnRoute, mergeResultRoute} {
		threshold, offset := 47+i*2, 9+i*2
		body := fmt.Sprintf("if input < %d { return input + %d } else { return input - %d }", threshold, offset, offset)
		generated, err := generateRoute("routes2303", "Calculate", "routes2303://calculate", "int64", "int64", body, route)
		if err != nil {
			t.Fatal(route, err)
		}
		origins := generated.report.BodyConditions
		if len(origins) != 1 || origins[0].Expression != fmt.Sprintf("input < %d", threshold) || origins[0].Relation != "retained_condition_node" {
			t.Fatal("route dropped authored condition", route, origins)
		}
		source := []byte("package routes2303\nnamespace routes2303\nentity Integer id \"routes2303://integer\"\nactivity Calculate(Integer) -> Integer id \"routes2303://calculate\" computes " + strconv.Quote(body) + "\n")
		bound, err := BindBodyConditionSources(source, origins)
		if err != nil || len(bound) != 1 || bound[0].Status != "MAPPED" || bound[0].Declaration == nil || bound[0].Declaration.StableID != "routes2303://calculate" {
			t.Fatal("route lost source declaration", route, err, bound)
		}
	}
	t.Log("three distinct route bodies retain source conditions and activity declarations; static projection only")
}
