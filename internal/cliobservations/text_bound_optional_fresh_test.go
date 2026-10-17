package cliobservations

import (
	"encoding/json"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func TestZeroOptionalRepeatedTextNativeFresh1905(t *testing.T) {
	const source = `package zero1905
namespace zero1905
entity Text id "urn:gooo:type:string"
entity Marker id "bound1905://marker" fields {
 field empty id "bound1905://empty" type string required one max_utf8_bytes "0"
 field omitted id "bound1905://omitted" type string optional one max_utf8_bytes "0"
 field zeros id "bound1905://zeros" type string required many max_utf8_bytes "0"
}
activity Empty(Text) -> Marker id "bound1905://empty-result" computes ` + "`" + `return Marker{empty: "", zeros: []string{"", ""}}` + "`" + `
`
	suite := bodyexecution.CompositionCases{Schema: "gooo/body-composition-cases/v1"}
	for _, input := range []string{`""`, `"present-root"`} {
		suite.Cases = append(suite.Cases, bodyexecution.CompositionCase{
			Inputs:   map[string]json.RawMessage{"app/bound1905:Empty": json.RawMessage(input)},
			Expected: map[string]json.RawMessage{"app/bound1905:Empty": json.RawMessage(`{"empty":"","zeros":["",""]}`)},
		})
	}
	receipt, code := executeTextBound1905(t, "zero-optional-many", []byte(source), "Empty", suite, false)
	assertTextBoundNative1905(t, receipt, code, 2)
	for _, trace := range receipt.Result.Runtime.Traces {
		fields := trace.Deliveries[0].ActualFields
		if len(fields) != 3 || fields[1].ID != "bound1905://omitted" || fields[1].Present == nil || *fields[1].Present {
			t.Fatal("optional absence lost its independent meaning", fields)
		}
	}
}
