package workspaceexecution

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func TestRecordInputScalarFill8851(t *testing.T) {
	cases := []struct {
		name, typeID, fieldType, train, expected, held, heldExpected, native, bad string
	}{
		{"text", "string", "string", `[{"position":1,"values":["left","서울"]}]`, `"서울"`,
			`[{"position":2,"values":["달","해","별"]}]`, `"별"`,
			`[{"inputs":{"Choose":{"values":["봄","여름"],"position":1}},"expected":{"Choose":"여름"}},{"inputs":{"Choose":{"values":["east","west"],"position":0}},"expected":{"Choose":"east"}}]`, "true"},
		{"boolean", "boolean", "boolean", `[{"position":1,"values":[false,true]}]`, `true`,
			`[{"position":2,"values":[true,true,false]}]`, `false`,
			`[{"inputs":{"Choose":{"values":[true,false],"position":0}},"expected":{"Choose":true}},{"inputs":{"Choose":{"values":[false,true,false],"position":2}},"expected":{"Choose":false}}]`, `"wrong type"`},
	}
	var observations []any
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			source := fmt.Sprintf(`package scalar%s8851
namespace scalar%s8851
entity Answer id "urn:gooo:type:%s"
entity Choices id "urn:scalar%s8851:choices" fields {
 field values id "urn:scalar%s8851:values" type %s required many
 field position id "urn:scalar%s8851:position" type integer required one
}
activity Choose(Choices) -> Answer id "urn:scalar%s8851:choose" computes "return input.values[__GOOO_BODY_HOLE_position__]" assembling {
 source_fill intent "목록에서 요청한 위치의 값을 읽는다" {
  hole "position"
  candidate "wrongtype" { fill "position" %q }
  candidate "first" { fill "position" "0" }
  candidate "requested" { fill "position" "input.position" }
 }
 value_case %q -> %q
 holdout_value_case %q -> %q
}
`, c.name, c.name, c.typeID, c.name, c.name, c.fieldType, c.name, c.name, c.bad, c.train, c.expected, c.held, c.heldExpected)
			suite, err := bodyexecution.DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":` + c.native + `}`))
			if err != nil {
				t.Fatal(err)
			}
			generation, err := bodyexecution.GenerateComposition(context.Background(), "scalar-"+c.name+".gooo", []byte(source), suite, "")
			if err != nil {
				t.Fatal("typed scalar fill", err)
			}
			fill := generation.Steps[0].Generation.Report.BodyFill
			if fill == nil || fill.SelectedCandidateID != "requested" || fill.TestCasesPassed != 1 || fill.TestCasesTotal != 1 ||
				fill.HoldoutCasesPassed != 1 || fill.HoldoutCasesTotal != 1 || len(fill.RejectedCandidates) != 1 ||
				fill.Evaluator != "gooo/bodycodegen-typed-value-interpreter/v1" {
				t.Fatal("scalar selection, holdout and rejected type", fill)
			}
			runtime, err := bodyexecution.ExecuteComposition(context.Background(), "scalar-"+c.name+".gooo", []byte(source), generation, suite, "")
			if err != nil || runtime.FinitePassed != 2 || runtime.FiniteTotal != 2 || runtime.ModelCalls != 0 {
				t.Fatal(err, runtime)
			}
			observations = append(observations, map[string]any{"source": source, "generation": generation, "runtime": runtime, "cases": suite})
		})
	}
	if output := os.Getenv("GOOO_ACCESS_OUTPUT"); output != "" {
		raw, _ := json.MarshalIndent(observations, "", "  ")
		if err := os.WriteFile(filepath.Join(output, "scalar8851.json"), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Log("8851: source type identities resolve Text/Boolean aliases; each selects requested index1/1, holdout1/1, rejects wrong-typed index; native4/4; model0")
}
