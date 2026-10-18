package bodycodegen

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

func TestCollectionMapScalarPairs(t *testing.T) {
	for _, row := range []struct{ in, out, body, input, expected string }{
		{"integer", "integer", "return input + 1", "[-1,0,2]", "[0,1,3]"},
		{"integer", "string", `if input > 0 {return "yes"}; return "no"`, "[-1,0,2]", `["no","no","yes"]`},
		{"integer", "boolean", "return input > 0", "[-1,0,2]", "[false,false,true]"},
		{"string", "integer", "return int64(len(input))", `["","a","한"]`, "[0,1,3]"},
		{"string", "string", `return "[" + input + "]"`, `["","a","한"]`, `["[]","[a]","[한]"]`},
		{"string", "boolean", "return len(input) > 0", `["","a","한"]`, "[false,true,true]"},
		{"boolean", "integer", "if input {return 1}; return 0", "[true,false,true]", "[1,0,1]"},
		{"boolean", "string", `if input {return "yes"}; return "no"`, "[true,false,true]", `["yes","no","yes"]`},
		{"boolean", "boolean", "return !input", "[true,false,true]", "[false,true,false]"},
	} {
		t.Run(row.in+"-"+row.out, func(t *testing.T) {
			source := fmt.Sprintf("package mapping\nnamespace mapping\nentity A id %q\nentity B id %q\n"+
				"entity Input id \"urn:map:in\" fields {field items id \"urn:map:items\" type %s required many}\n"+
				"entity Output id \"urn:map:out\" fields {field items id \"urn:map:outputs\" type %s required many}\n"+
				"activity Convert(A) -> B id \"urn:map:convert\" computes %q\n"+
				"activity Run(Input) -> Output id \"urn:map:run\" computes `return Output{items:map_values(input.items,Convert)}`\n",
				"urn:gooo:type:"+row.in, "urn:gooo:type:"+row.out, row.in, row.out, row.body)
			if row.in == row.out {
				source = strings.Replace(source, fmt.Sprintf("entity B id %q\n", "urn:gooo:type:"+row.out), "", 1)
				source = strings.Replace(source, "-> B", "-> A", 1)
			}
			generated, err := Generate("map.gooo", []byte(source), "Run")
			if err != nil {
				t.Fatal(err)
			}
			cases := []assemblyspec.ValueCase{{Inputs: `[{"items":` + row.input + `}]`, Expected: `{"items":` + row.expected + `}`}}
			observed, err := evaluateRecordAssembly(context.Background(), []byte(generated.Source), "Run", generated.Report.RecordTypes, cases)
			if err != nil || len(observed) != 1 || !observed[0].Passed {
				t.Fatal("scalar map differs from authored expectation", err, observed)
			}
		})
	}
}

func TestCollectionMapEquivalenceRejectsChangedTraversal(t *testing.T) {
	p, err := prepareActivityBodySource("mapping.gooo", []byte(collectionMapFixture), "Project")
	if err != nil {
		t.Fatal(err)
	}
	calls, _, _, err := p.resolvePureCalls(p.body)
	if err != nil {
		t.Fatal(err)
	}
	generated, err := Generate("mapping.gooo", []byte(collectionMapFixture), "Project")
	if err != nil {
		t.Fatal(err)
	}
	for _, change := range [][2]string{
		{"[256]string", "[257]string"},
		{"goooMapApply(goooMapInput[goooMapIndex])", "goooMapApply(goooMapInput[0])"},
		{"goooMapResult[goooMapIndex] =", "goooMapResult[0] ="},
		{"goooMapResult[:goooMapLength:goooMapLength]", "goooMapResult[:goooMapLength]"},
		{`"[" + input + "]"`, `"(" + input + ")"`},
	} {
		changed := strings.Replace(generated.Source, change[0], change[1], 1)
		if changed == generated.Source {
			t.Fatal("counterexample did not change source", change)
		}
		proof, err := routeEquivalenceWithCalls(p.packageName, p.activity.Name, p.parameters, p.outputType, p.body, []byte(changed), preserveRoute, calls, p.allRecords)
		if err == nil && proof.Equivalent {
			t.Fatal("changed allocation, traversal or mapper passed inverse rule", change)
		}
	}
}

func TestCollectionMapRetainsCallBound(t *testing.T) {
	for _, count := range []int{15, 16} {
		body := strings.TrimSuffix(strings.Repeat("Echo(input)+", count), "+")
		source := strings.Replace(collectionMapFixture, `"[" + input + "]"`, body, 1) +
			"activity Echo(Text) -> Text id \"urn:map:echo\" computes `return input`\n"
		generated, err := Generate("bound.gooo", []byte(source), "Project")
		if count == 15 {
			if err != nil || generated.Report.CallClosure.MaxCallsPerInvocation != 4096 {
				t.Fatal("exact existing call bound", err, generated.Report.CallClosure)
			}
		} else {
			failure, ok := errors.AsType[*PureCallBudgetError](err)
			if !ok || failure.CallerID != "urn:mapping:project" || failure.CalleeID != "urn:mapping:quote" ||
				failure.Multiplicity != 256 || failure.MultiplicityCapped || failure.PrefixBefore != 0 ||
				failure.CalleeBound != 16 || failure.PrefixAfter != 4352 || failure.Limit != 4096 {
				t.Fatal("map lost its actual mapper identity or call bound", err)
			}
		}
	}
}

func TestCollectionMapLoopBudgetReportsCappedMultiplicity(t *testing.T) {
	for _, count := range []int{16, 17} {
		body := fmt.Sprintf("let result = input; for let i = 0; i < %d; i++ { result = Paths{values:map_values(result.values,Quote)} }; return result", count)
		source := strings.Replace(collectionMapFixture, "return Paths{values: map_values(input.values, Quote)}", body, 1)
		generated, err := Generate("loop-map.gooo", []byte(source), "Project")
		if count == 16 {
			if err != nil || generated.Report.CallClosure.MaxCallsPerInvocation != 4096 {
				t.Fatal("inclusive map/loop bound changed", err)
			}
			continue
		}
		failure, ok := errors.AsType[*PureCallBudgetError](err)
		if !ok || failure.Callee != "Quote" || failure.CalleeID != "urn:mapping:quote" ||
			failure.Multiplicity != 4097 || !failure.MultiplicityCapped || failure.PrefixAfter != 4097 ||
			failure.CalleeBound != 0 || failure.PrefixBefore != 0 || generated.Source != "" {
			t.Fatal("map/loop refusal lost the saturated bound", err)
		}
	}
}

func TestCollectionMapRejectsUnsupportedMappers(t *testing.T) {
	for _, expression := range []string{
		"map_values(input.values)", "map_values(input.values,Quote,Quote)",
		"map_values(input.values,Missing)", "map_values(input.values,input)",
		"map_values(input.values,Quote(input.values[0]))", "map_values(input.values,Project)",
		"map_values(input.values[0],Quote)", "map_values([]int64{1},Quote)",
	} {
		source := strings.Replace(collectionMapFixture, "map_values(input.values, Quote)", expression, 1)
		if source == collectionMapFixture {
			t.Fatal("refusal fixture did not change")
		}
		if _, err := Generate("invalid-map.gooo", []byte(source), "Project"); err == nil {
			t.Fatal("unsupported map accepted", expression)
		}
	}
}
