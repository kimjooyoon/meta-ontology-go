package bodycodegen

import (
	"context"
	"strings"
	"testing"
)

func TestNestedCollectionReturnBudget14017(t *testing.T) {
	const prefix = `package nested14017
namespace nested14017
entity Integer id "urn:gooo:type:integer"
entity Batch id "urn:nested14017:batch" fields {field values id "urn:nested14017:values" type integer required many}
activity Finish(Integer)->Integer id "urn:nested14017:finish" computes ` + "`return input+3023`" + `
activity Choose(Batch)->Integer id "urn:nested14017:choose" computes `
	body := `for left in input.values {for right in input.values {if left==right {return Finish(left+right)}}};return 0`
	result, err := Generate("nested14017.gooo", []byte(prefix+"`"+body+"`\n"), "Choose")
	if err != nil || result.Report.CallClosure == nil || result.Report.CallClosure.MaxCallsPerInvocation != 1 {
		t.Fatal("single terminating call within bounded nested collections", err)
	}
	evaluator, err := newCriteriaTupleEvaluator(context.Background(), []byte(result.Source), "Choose", result.Report.RecordTypes, nil)
	if err != nil {
		t.Fatal(err)
	}
	for i, tc := range [][2]string{{`[{"values":[]}]`, `0`}, {`[{"values":[3001,3007]}]`, `9025`}, {`[{"values":[-3011,3019]}]`, `-2999`}} {
		actual, err := evaluator.evaluate(tc[0], i)
		if err != nil || string(actual) != tc[1] {
			t.Fatal("new finite nested collection result", i, string(actual), err)
		}
	}
	for _, rejected := range []string{
		`let result=0;for left in input.values {for right in input.values {result=result+Finish(left+right)}};return result`,
		`for first in input.values {for second in input.values {for third in input.values {return first+second+third}}};return 0`,
	} {
		_, err := Generate("rejected14017.gooo", []byte(prefix+"`"+rejected+"`\n"), "Choose")
		if err == nil || !(strings.Contains(err.Error(), "4096") || strings.Contains(err.Error(), "65536")) {
			t.Fatal("named call or iteration-product bound disappeared", err)
		}
	}
	t.Log("fresh nested source3/3 interpreted tuples;terminating helper budget1;nonterminating nested calls still exceed4096;third collection level exceeds65536 iteration product;models0")
}
