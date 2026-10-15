package bodycodegen

import (
	"context"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

const recordCompoundSource = `package fieldupdates
namespace fieldupdates
entity Integer id "urn:gooo:type:integer"
entity Ledger id "urn:field-update:ledger" fields {
field total id "urn:field-update:total" type integer required one
field label id "urn:field-update:label" type string required one
field untouched id "urn:field-update:untouched" type integer required one
}
activity Update(기록 Ledger, 증가 Integer) -> Ledger id "urn:field-update:update" computes ` + "`" + `
let result = 기록
result.total += 증가
result.total *= 3
(result).total -= 2
result.total /= 2
result.label += "!"
return result
` + "`\n"

func TestRecordCompoundGenerationAndEvaluation(t *testing.T) {
	generated, err := Generate("field-updates.gooo", []byte(recordCompoundSource), "Update")
	if err != nil || !generated.Report.TypecheckPassed || !generated.Report.DeterministicReplay {
		t.Fatal(err, generated.Report)
	}
	cases := []assemblyspec.ValueCase{
		{Inputs: `[{"total":5,"label":"A","untouched":41},7]`, Expected: `{"total":17,"label":"A!","untouched":41}`},
		{Inputs: `[{"total":-9,"label":"한글","untouched":43},2]`, Expected: `{"total":-11,"label":"한글!","untouched":43}`},
		{Inputs: `[{"total":0,"label":"","untouched":47},0]`, Expected: `{"total":-1,"label":"!","untouched":47}`},
	}
	observed, err := evaluateRecordAssembly(context.Background(), []byte(generated.Source), "Update", generated.Report.RecordTypes, cases)
	if err != nil || len(observed) != 3 {
		t.Fatal(err, observed)
	}
	for _, row := range observed {
		if !row.Passed {
			t.Fatal(row)
		}
	}
	for _, operation := range []string{"+=", "-=", "*=", "/="} {
		if !strings.Contains(generated.Source, operation) {
			t.Fatal("authored operation lost", operation)
		}
	}
	t.Log("three new deterministic evaluator cases; four arithmetic operators, string concatenation, copied input and untouched field")
}

func TestRecordCompoundSourceOrigins(t *testing.T) {
	generated, err := Generate("field-update-origins.gooo", []byte(recordCompoundSource), "Update")
	if err != nil {
		t.Fatal(err)
	}
	writes, reads := 0, 0
	for _, field := range generated.Report.BodyFields {
		if field.Kind == "field_assignment" {
			writes++
			if field.Operation == "" || field.RecordID != "urn:field-update:ledger" {
				t.Fatal(field)
			}
		}
		if field.Kind == "field_read" {
			reads++
		}
	}
	if writes != 5 || reads != 5 {
		t.Fatal("implicit old-value syntax population", writes, reads)
	}
	if _, err := BindBodyFieldSources([]byte(recordCompoundSource), generated.Report.BodyFields); err != nil {
		t.Fatal(err)
	}
	assertRecordCompoundFlow(t)
}
