package syntax

import (
	"reflect"
	"strings"
	"testing"
)

const sourceRecordObserve = `package records
namespace records
entity Integer id "records://integer"
entity Box id "records://box" fields { field value id "records://box/value" type integer required one }
activity Select(Integer) -> Box computes "return Box{value: input}" assembling {
 choice "value" field_value at "0" alternative "0 - input" intent "Negate."
 value_case "[0]" -> "{\"value\":0}"
 holdout_value_case "[5]" -> "{\"value\":-5}"
 observe { value_input "[9007199254740993]" oracle "Oracle" max_candidates "2" max_rounds "1" reuse_outputs "true" resolve_unique "true" }
 attempts "2"
}
activity Oracle(Integer) -> Box computes "return Box{value: 0 - input}"
`

func TestRecordObserveSourceFormatCloneAndBounds(t *testing.T) {
	parse := func(source string) *File {
		file, diagnostics := ParseFileWithEntityFieldsSupport("record.gooo", source, EntityFieldsV1Support())
		if diagnostics.HasErrors() {
			t.Fatal(diagnostics)
		}
		return file
	}
	file := parse(sourceRecordObserve)
	activity := file.Declarations[2].(*ActivityDecl)
	formatted, err := FormatWithEntityFieldsSupport(file, EntityFieldsV1Support())
	if err != nil {
		t.Fatal(err)
	}
	again := parse(formatted).Declarations[2].(*ActivityDecl)
	if !reflect.DeepEqual(activity.Assembly.Spec, again.Assembly.Spec) {
		t.Fatal("record observation changed during formatting")
	}
	clone := file.Clone().Declarations[2].(*ActivityDecl).Assembly.Spec.Observation
	clone.ValueInputs[0] = "[7]"
	if activity.Assembly.Spec.Observation.ValueInputs[0] != "[9007199254740993]" {
		t.Fatal("record input lost precision or shared clone storage")
	}
	for _, to := range []string{`value_input "[5]"`, `input "7"`, `value_input "[7]" value_input "[7]"`} {
		source := strings.Replace(sourceRecordObserve, `value_input "[9007199254740993]"`, to, 1)
		if _, diagnostics := ParseFileWithEntityFieldsSupport("record.gooo", source, EntityFieldsV1Support()); !diagnostics.HasErrors() {
			t.Fatal("invalid record observation accepted", to)
		}
	}
}
