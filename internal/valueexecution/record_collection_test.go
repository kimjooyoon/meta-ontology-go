package valueexecution

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestOrderedCollectionTransport7409(t *testing.T) {
	source := []byte(`package transport7409
namespace transport7409
entity Batch id "urn:transport7409:batch" fields {
 field amounts id "urn:transport7409:amounts" type integer required many
 field labels id "urn:transport7409:labels" type string optional many
 field flags id "urn:transport7409:flags" type boolean optional many
}
activity Start(Batch) -> Batch id "urn:transport7409:start" computes "record.forward:v1"
activity Finish(Batch) -> Batch id "urn:transport7409:finish" computes "record.forward:v1"
bind Start.result -> Finish.input
`)
	plan, err := CompileRecordPlan("전달7409.gooo", source)
	if err != nil {
		t.Fatal(err)
	}
	for _, input := range []string{
		`{"amounts":[9007199254741009,-2,31],"labels":["나","a"],"flags":[true,false]}`,
		`{"amounts":[],"labels":[],"flags":[]}`,
		`{"amounts":[17]}`,
	} {
		fields, err := DecodeRecordInput([]byte(input))
		if err != nil {
			t.Fatal(err)
		}
		result, err := plan.Execute(map[string]RecordFields{"Start": fields})
		if err != nil || result.ApplyCalls != 2 || result.Deliveries != 1 {
			t.Fatal("ordered transport", err, result)
		}
		start, finish := result.Results["Start"], result.Results["Finish"]
		expected, _ := json.Marshal(fields)
		actual, _ := json.Marshal(finish.Fields)
		if string(actual) != string(expected) || finish.ParentResultDigest != start.ResultDigest || finish.RootInputDigest != start.RootInputDigest {
			t.Fatal("array ordering, exact numbers or input origin changed", string(actual), finish)
		}
		if len(fields["amounts"].([]any)) > 0 {
			fields["amounts"].([]any)[0] = int64(-88)
			start.Fields["amounts"].([]any)[0] = int64(-99)
			detached, _ := json.Marshal(finish.Fields)
			if string(detached) != string(expected) {
				t.Fatal("caller or sibling observation aliases transported array")
			}
		}
	}
	for _, input := range []string{`{"amounts":null}`, `{"amounts":[null]}`, `{"amounts":[[]]}`,
		`{"amounts":[1.5]}`, `{"amounts":[true]}`, `{"amounts":[] ,"labels":null}`} {
		fields, err := DecodeRecordInput([]byte(input))
		if err == nil {
			_, err = plan.Execute(map[string]RecordFields{"Start": fields})
		}
		if err == nil {
			t.Fatal("non-profile collection accepted", input)
		}
	}
	oversized := `{"amounts":[` + strings.Repeat("1,", 256) + `1]}`
	if _, err := DecodeRecordInput([]byte(oversized)); err == nil {
		t.Fatal("257-item field exceeded source profile")
	}
	t.Log("7409: three ordered inputs through two activities, exact int64, absent/empty optional arrays, detached observations and seven invalid forms; model0")
}
