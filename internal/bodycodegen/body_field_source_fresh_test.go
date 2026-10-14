package bodycodegen

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestFieldSourceCapture16121(t *testing.T) {
	const prefix = `package fieldsource
namespace fieldsource16121
entity Input id "urn:fieldsource16121:input" fields {
 field amount id "urn:fieldsource16121:amount" type integer required one
 field label id "urn:fieldsource16121:label" type string required one
}
entity State id "urn:fieldsource16121:state" fields {
 field score id "urn:fieldsource16121:score" type integer required one
 field tag id "urn:fieldsource16121:tag" type string required one
}
`
	const body = `let state=State{score:input.amount+16121,tag:input.label}
state.score=state.score+17
state.tag="기원:"+state.tag
return state`
	var reports []Report
	for _, literal := range []string{"`" + body + "`", strconv.Quote(body)} {
		source := []byte(prefix + `activity Build(Input)->State id "urn:fieldsource16121:build" computes ` + literal + "\n")
		result, err := Generate("field-source16121.gooo", source, "Build")
		if err != nil {
			t.Fatal(err)
		}
		origins := result.Report.BodyFields
		if len(origins) != 8 {
			t.Fatalf("typed field population got %d: %+v", len(origins), origins)
		}
		bound, err := BindBodyFieldSources(source, origins)
		if err != nil {
			t.Fatal(err)
		}
		counts := map[string]int{}
		for i, origin := range origins {
			counts[origin.Kind]++
			if origin.RecordID == "" || origin.FieldID == "" || origin.NativePopulation <= origin.NativeOrdinal {
				t.Fatal("identity", origin)
			}
			for _, location := range []*BodyExpressionSource{bound[i].Value, bound[i].Target} {
				if location.Status != "MAPPED" || location.Start < 0 || location.End > len(source) || location.SourceSHA256 != digest(source) {
					t.Fatal("source", location)
				}
				if len(source[location.Start:location.End]) == 0 {
					t.Fatal("empty literal range")
				}
			}
		}
		if counts["record_constructor"] != 2 || counts["field_assignment"] != 2 || counts["field_read"] != 4 {
			t.Fatal(counts)
		}
		mutated := append([]BodyFieldOrigin(nil), origins...)
		mutated[0].FieldID = "other"
		if EqualRetainedFieldOrigins(mutated, origins) || !EqualRetainedFieldOrigins(nil, origins) {
			t.Fatal("retained field replay claims")
		}
		reports = append(reports, result.Report)
	}
	if root := os.Getenv("GOOO_FRESH_SOURCE_OUTPUT"); root != "" {
		raw, err := json.MarshalIndent(reports, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, "field-capture16121.json"), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Log("raw/escaped Korean source;16 typed field slots mapped;no producer execution/model;original literal byte spans retained")
}
