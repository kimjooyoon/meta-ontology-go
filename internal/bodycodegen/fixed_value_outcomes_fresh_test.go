package bodycodegen

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestFixedValueOutcomesFresh1759(t *testing.T) {
	for _, mode := range []string{"mismatch", "wrong-output-type", "missing-output-field", "wrong-input-type"} {
		t.Run(mode, func(t *testing.T) {
			inputs, expected := `[true]`, `{"enabled":true}`
			switch mode {
			case "wrong-output-type":
				expected = `{"enabled":"false"}`
			case "missing-output-field":
				expected = `{}`
			case "wrong-input-type":
				inputs = `["true"]`
			}
			source := []byte(fmt.Sprintf(`package fixedoutcome
namespace fixedoutcome
entity Boolean id "urn:gooo:type:boolean"
entity Gate id "fixedoutcome://gate" fields {
    field enabled id "fixedoutcome://gate/enabled" type boolean required one
}
activity Build(Boolean) -> Gate computes "return Gate{enabled: false}" assembling {
    value_case %q -> %q
    attempts "1"
}
`, inputs, expected))
			generator, err := NewTypedPathGenerator("")
			if err != nil {
				t.Fatal(err)
			}
			result, err := generator.GenerateSourceAssembly(context.Background(), "fixed-outcomes-1759.gooo", source, "Build")
			if mode != "mismatch" {
				if err == nil {
					t.Fatal("ill-typed fixed value case was accepted")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			r := result.Report.RecordAssembly
			if r == nil || r.Status != "PARTIAL_FINITE" || r.Total != 1 || r.Passed != 0 || r.FieldsTotal != 1 ||
				r.FieldsPassed != 0 || len(r.Attempts) != 1 || len(r.Ranking) != 1 || r.ModelCalls != 0 || r.ModelRequested ||
				!strings.Contains(r.Scope, "fixed record body") {
				t.Fatalf("mismatch invented completion or search: %+v", r)
			}
		})
	}
}
