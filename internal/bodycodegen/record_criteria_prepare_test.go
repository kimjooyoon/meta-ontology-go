package bodycodegen

import (
	"context"
	"errors"
	"strings"
	"testing"
)

const typedCriteriaSource = `package criteria
namespace criteria
entity Integer id "criteria://integer"
entity Boolean id "criteria://boolean"
entity Box id "criteria://box" fields { field value id "criteria://value" type integer required one }
activity Select(Integer) -> Box computes "return Box{value:input}" assembling {
 choice "value" field_value at "0" alternative "0-input" intent "Negate."
 criteria { predicate "negative" activity "Negative" input "[7]" holdout_input "[9]" }
 attempts "2"
}
activity Negative(Integer,Box) -> Boolean computes "return input1.value == 0-input0"`

func TestCriteriaPreflightBindsUnaryInputAndFixedPredicate(t *testing.T) {
	ctx := context.Background()
	source := []byte(typedCriteriaSource)
	owner, err := prepareActivityBody("criteria.gooo", source, "Select")
	if err != nil {
		t.Fatal(err)
	}
	spec, err := SourceAssembly(ctx, "criteria.gooo", source, "Select")
	if err != nil {
		t.Fatal(err)
	}
	prepared, err := prepareRecordCriteria(ctx, "criteria.gooo", source, owner, spec.Criteria)
	if err != nil || len(prepared) != 1 || prepared[0].activityID != "criteria://activity/negative" ||
		prepared[0].goSHA256 != digest([]byte(prepared[0].program.Source)) || prepared[0].declaration.ID != "negative" {
		t.Fatal("typed criterion was not bound", prepared, err)
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := prepareRecordCriteria(canceled, "criteria.gooo", source, owner, spec.Criteria); !errors.Is(err, context.Canceled) {
		t.Fatal("criteria preflight lost cancellation", err)
	}
}

func TestCriteriaPreflightRejectsWrongSignatureAndUnfixedClosure(t *testing.T) {
	variants := []string{
		strings.Replace(typedCriteriaSource, `activity "Negative"`, `activity "Select"`, 1),
		strings.Replace(typedCriteriaSource, `activity "Negative"`, `activity "Missing"`, 1),
		strings.Replace(typedCriteriaSource, "Negative(Integer,Box) -> Boolean", "Negative(Boolean,Box) -> Boolean", 1),
		strings.Replace(typedCriteriaSource, "Negative(Integer,Box) -> Boolean", "Negative(Integer,Integer) -> Boolean", 1),
		strings.Replace(typedCriteriaSource, "Negative(Integer,Box) -> Boolean", "Negative(Integer,Box) -> Integer", 1),
		typedCriteriaSource + ` assembling { choice "direction" operand_order at "0" intent "Different." case "0" -> "0" attempts "2" }`,
		strings.Replace(typedCriteriaSource, "input1.value == 0-input0", "Probe(input0) > 0", 1) + `
activity Probe(Integer) -> Integer computes "return input - 1" assembling {
 choice "direction" operand_order at "0" intent "Different." case "1" -> "0" attempts "2"
}`,
	}
	for i, text := range variants {
		source := []byte(text)
		owner, err := prepareActivityBodySource("criteria.gooo", source, "Select")
		if err != nil {
			t.Fatalf("fixture %d cannot prepare source: %v", i, err)
		}
		spec, err := SourceAssembly(context.Background(), "criteria.gooo", source, "Select")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := prepareRecordCriteria(context.Background(), "criteria.gooo", source, owner, spec.Criteria); err == nil {
			t.Fatalf("invalid criterion %d accepted", i)
		}
	}
}
