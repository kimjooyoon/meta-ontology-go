package syntax

import (
	"reflect"
	"strings"
	"testing"
)

func TestCriteriaQuestionsRoundTripAndSemanticCarrier(t *testing.T) {
	source := strings.Replace(sourceRecordObserve, `value_case "[0]" -> "{\"value\":0}"`,
		`criteria { predicate "negative" activity "Negative" input "[0]" }`, 1)
	source = strings.Replace(source, `value_input "[9007199254740993]" oracle "Oracle"`,
		`criteria_input "[9007199254740993]" criteria_input "[-7]"`, 1)
	file, diagnostics := ParseFileWithEntityFieldsSupport("questions.gooo", source, EntityFieldsV1Support())
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	formatted, err := FormatWithEntityFieldsSupport(file, EntityFieldsV1Support())
	if err != nil {
		t.Fatal(err)
	}
	other, diagnostics := ParseFileWithEntityFieldsSupport("questions.gooo", formatted, EntityFieldsV1Support())
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	a, b := file.Declarations[2].(*ActivityDecl).Assembly.Spec, other.Declarations[2].(*ActivityDecl).Assembly.Spec
	if !reflect.DeepEqual(a, b) || len(a.Observation.CriteriaInputs) != 2 {
		t.Fatal("questions changed in source roundtrip")
	}
	clone := file.Clone().Declarations[2].(*ActivityDecl).Assembly.Spec
	clone.Observation.CriteriaInputs[0] = "[3]"
	if a.Observation.CriteriaInputs[0] != "[9007199254740993]" {
		t.Fatal("question clone lost exact integer")
	}
}
