package bodycodegen

import (
	"context"
	"strconv"
	"strings"
	"testing"
)

func TestCollectionBoundsAndSourceRules7781(t *testing.T) {
	const prefix = `package bounds7781
namespace bounds7781
entity Integer id "urn:gooo:type:integer"
entity Batch id "urn:bounds7781:batch" fields {
 field items id "urn:bounds7781:items" type integer required many
}
activity Sum(Batch) -> Integer id "urn:bounds7781:sum" computes `
	body := `let total = 0; for item in input.items { total = total + item }; return total`
	source := []byte(prefix + strconv.Quote(body) + "\n")
	generated, err := Generate("bounds7781.gooo", source, "Sum")
	if err != nil {
		t.Fatal(err)
	}
	evaluator, err := newCriteriaTupleEvaluator(context.Background(), []byte(generated.Source), "Sum", generated.Report.RecordTypes, nil)
	if err != nil {
		t.Fatal(err)
	}
	boundary := `[{"items":[` + strings.Repeat("64,", 255) + `64]}]`
	result, err := evaluator.evaluate(boundary, 0)
	if err != nil || string(result) != "16384" {
		t.Fatal("256-item exact finite evaluation", string(result), err)
	}
	oversized := `[{"items":[` + strings.Repeat("64,", 256) + `64]}]`
	if _, err := evaluator.evaluate(oversized, 1); err == nil {
		t.Fatal("257-item input accepted by finite evaluator")
	}
	for _, rejected := range []string{
		`for item in input.items { item = 7 }; return 0`,
		`for item in input.items { for second in input.items { for third in input.items { return item + second + third } } }; return 0`,
		`let item = 0; for item in input.items { return item }; return 0`,
		`for letter in "abc" { return 0 }; return 0`,
	} {
		if _, err := Generate("bounds7781-rejected.gooo", []byte(prefix+strconv.Quote(rejected)+"\n"), "Sum"); err == nil {
			t.Fatal("iteration beyond its source contract accepted", rejected)
		}
	}
	t.Log("7781: 256-item finite sum exact; 257-item input and four unsupported source loop forms rejected; model0")
}
