package bodycodegen

import "testing"

func TestRecordCollectionGraphContextDoesNotInventScalarTypes(t *testing.T) {
	for _, item := range []struct{ typeID, cardinality, want string }{
		{"urn:gooo:type:integer", "", "int64"},
		{"urn:gooo:type:integer", "many", ""},
		{"urn:source:child", "", ""},
	} {
		body := preparedBody{allRecords: []RecordType{{Name: "Parent", ID: "urn:source:parent", Fields: []RecordField{{ID: "urn:source:field", TypeID: item.typeID, Cardinality: item.cardinality}}}}}
		if got := recordGraphInputType(body, RecordFlowNode{FieldID: "urn:source:field"}); got != item.want {
			t.Fatalf("context fabricated a scalar for %s/%s: %q", item.typeID, item.cardinality, got)
		}
	}
}
