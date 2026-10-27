package bodycodegen

import (
	"os"
	"reflect"
	"testing"
)

func TestRecordScopeRetainsDeclaredIdentitiesAndCalledBodyTypes(t *testing.T) {
	raw, err := os.ReadFile("../../examples/body-codegen/selected-record-scope.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	file, diagnostics := ParseBodyFile("scope.gooo", raw)
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	_, all, err := resolveBodyModel(file)
	if err != nil || len(all) != 20 {
		t.Fatal("declaration view", len(all), err)
	}
	selected, err := SelectActivityRecords(file, all, []string{"Evaluate"})
	if err != nil {
		t.Fatal(err)
	}
	names := []string{}
	for _, record := range selected {
		names = append(names, record.Name)
		found := false
		for _, original := range all {
			if original.ID == record.ID {
				found = reflect.DeepEqual(original, record)
			}
		}
		if !found {
			t.Fatal("selected record lost its original layout", record)
		}
	}
	if !reflect.DeepEqual(names, []string{"Request", "Result", "Scratch"}) {
		t.Fatal("helper constructor or local shadow was misclassified", names)
	}
	t.Log(`GOOO_RECORD_SCOPE_OBSERVATION={"declared":20,"selected":3,"helper_only_record":1,"shadowed_record_name_counted":false,"model_calls":0}`)
}
