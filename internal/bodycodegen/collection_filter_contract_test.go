package bodycodegen

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

func filterSource(kind, predicate string) string {
	valueName := "Value"
	valueDeclaration := fmt.Sprintf("entity Value id %q\n", "urn:gooo:type:"+kind)
	if kind == "boolean" {
		valueName, valueDeclaration = "Boolean", ""
	}
	return fmt.Sprintf("package filtering\nnamespace filtering\n%s"+
		"entity Boolean id \"urn:gooo:type:boolean\"\n"+
		"entity Batch id \"urn:filter:batch\" fields {field values id \"urn:filter:values\" type %s required many}\n"+
		"activity Keep(%s) -> Boolean id \"urn:filter:keep\" computes %q\n"+
		"activity Select(Batch) -> Batch id \"urn:filter:select\" computes `return Batch{values:filter_values(input.values,Keep)}`\n",
		valueDeclaration, kind, valueName, predicate)
}

func TestCollectionFilterPreservesOrderAndScalarType(t *testing.T) {
	for _, row := range []struct{ kind, predicate, input, expected string }{
		{"integer", "return input > 0", "[-1,0,2,2,-3]", "[2,2]"},
		{"string", `return len(input) > 0`, `["","a","한",""]`, `["a","한"]`},
		{"boolean", "return input", "[false,true,true,false]", "[true,true]"},
		{"integer", "return input > 0", "[]", "[]"},
	} {
		t.Run(row.kind+row.input, func(t *testing.T) {
			generated, err := Generate("filter.gooo", []byte(filterSource(row.kind, row.predicate)), "Select")
			if err != nil {
				t.Fatal(err)
			}
			cases := []assemblyspec.ValueCase{{Inputs: `[{"values":` + row.input + `}]`,
				Expected: `{"values":` + row.expected + `}`}}
			observed, err := evaluateRecordAssembly(context.Background(), []byte(generated.Source),
				"Select", generated.Report.RecordTypes, cases)
			if err != nil || len(observed) != 1 || !observed[0].Passed {
				t.Fatal("filter differs from authored expectation", err, observed)
			}
		})
	}
}

func TestCollectionFilterInverseRejectsChangedSelection(t *testing.T) {
	source := filterSource("integer", "return input > 0")
	prepared, err := prepareActivityBodySource("filter.gooo", []byte(source), "Select")
	if err != nil {
		t.Fatal(err)
	}
	calls, _, _, err := prepared.resolvePureCalls(prepared.body)
	if err != nil {
		t.Fatal(err)
	}
	generated, err := Generate("filter.gooo", []byte(source), "Select")
	if err != nil {
		t.Fatal(err)
	}
	for _, change := range [][2]string{
		{"[256]int64", "[257]int64"},
		{"goooFilterApply(goooFilterInput[goooFilterIndex])", "goooFilterApply(goooFilterInput[0])"},
		{"goooFilterResult[goooFilterLength] = goooFilterInput[goooFilterIndex]", "goooFilterResult[goooFilterLength] = goooFilterInput[0]"},
		{"goooFilterResult[:goooFilterLength:goooFilterLength]", "goooFilterResult[:goooFilterLength]"},
	} {
		changed := strings.Replace(generated.Source, change[0], change[1], 1)
		if changed == generated.Source {
			t.Fatal("counterexample did not alter the generated projection", change)
		}
		proof, err := routeEquivalenceWithCalls(prepared.packageName, prepared.activity.Name,
			prepared.parameters, prepared.outputType, prepared.body, []byte(changed), preserveRoute,
			calls, prepared.allRecords)
		if err == nil && proof.Equivalent {
			t.Fatal("changed filter allocation, predicate or order passed inverse rule", change)
		}
	}
}

func TestCollectionFilterRejectsUnownedOrWrongPredicate(t *testing.T) {
	base := filterSource("string", "return len(input) > 0")
	for _, change := range []struct{ name, before, after string }{
		{"missing", "input.values,Keep", "input.values,Unknown"},
		{"wrong output", "Keep(Value) -> Boolean", "Keep(Value) -> Value"},
		{"wrong input", "Keep(Value) -> Boolean", "Keep(Boolean) -> Boolean"},
		{"anonymous", "input.values,Keep", "input.values,func(value string) bool {return true}"},
	} {
		t.Run(change.name, func(t *testing.T) {
			source := strings.Replace(base, change.before, change.after, 1)
			if source == base {
				t.Fatal("counterexample did not change source")
			}
			if _, err := Generate("filter.gooo", []byte(source), "Select"); err == nil {
				t.Fatal("invalid source predicate was accepted")
			}
		})
	}
}
