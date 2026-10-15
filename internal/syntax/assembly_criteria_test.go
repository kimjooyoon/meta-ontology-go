package syntax

import (
	"reflect"
	"strings"
	"testing"
)

const criteriaSource = `package requirements
namespace requirements
entity Integer id "requirements://integer"
entity Boolean id "requirements://boolean"
entity Box id "requirements://box" fields { field value id "requirements://box/value" type integer required one }
activity Select(Integer) -> Box computes "return Box{value:input}" assembling {
 choice "value" field_value at "0" alternative "0-input" intent "Negate."
 criteria {
  predicate "negative" activity "Negative"
  input "[9007199254740993]"
  holdout_input "[9007199254740995]"
 }
 attempts "2"
}
activity Negative(Integer,Box) -> Boolean computes "return input1.value == 0-input0"
`

func TestCriteriaSourceRoundTripCloneAndInvalidDeclarations(t *testing.T) {
	parse := func(source string) *File {
		file, diagnostics := ParseFileWithEntityFieldsSupport("criteria.gooo", source, EntityFieldsV1Support())
		if diagnostics.HasErrors() {
			t.Fatal(diagnostics)
		}
		return file
	}
	file := parse(criteriaSource)
	formatted, err := FormatWithEntityFieldsSupport(file, EntityFieldsV1Support())
	if err != nil {
		t.Fatal(err)
	}
	before := file.Declarations[3].(*ActivityDecl).Assembly.Spec
	after := parse(formatted).Declarations[3].(*ActivityDecl).Assembly.Spec
	if !reflect.DeepEqual(before, after) {
		t.Fatal("criteria changed during formatting")
	}
	cloned := file.Clone().Declarations[3].(*ActivityDecl).Assembly.Spec.Criteria
	cloned.Predicates[0].Activity, cloned.Inputs[0], cloned.HoldoutInputs[0] = "Changed", "[1]", "[2]"
	if before.Criteria.Predicates[0].Activity != "Negative" || before.Criteria.Inputs[0] != "[9007199254740993]" ||
		before.Criteria.HoldoutInputs[0] != "[9007199254740995]" {
		t.Fatal("criteria clone aliases source")
	}
	for _, source := range []string{
		strings.Replace(criteriaSource, `predicate "negative" activity "Negative"`, `predicate "negative" activity "Negative" predicate "negative" activity "Negative"`, 1),
		strings.Replace(criteriaSource, `holdout_input "[9007199254740995]"`, `holdout_input "[9007199254740993]"`, 1),
		strings.Replace(criteriaSource, " attempts", ` criteria { predicate "other" activity "Negative" input "[3]" } attempts`, 1),
	} {
		if _, diagnostics := ParseFileWithEntityFieldsSupport("criteria.gooo", source, EntityFieldsV1Support()); !diagnostics.HasErrors() {
			t.Fatal("invalid or duplicate criteria block accepted")
		}
	}
}
