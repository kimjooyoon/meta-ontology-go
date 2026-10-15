package bodycodegen

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/jointdecision"
)

const recordHoldoutContextSource = `package contextheld
namespace contextheld
entity Text id "contextheld://text"
entity Box id "contextheld://box" fields {
 field a id "contextheld://box/a" type string required one
 field b id "contextheld://box/b" type string required one
 field c id "contextheld://box/c" type string required one
}
activity Build(Text) -> Box id "contextheld://build" computes "return Box{a: input, b: input, c: input}" assembling {
 choice "a" field_value at "0" alternative "\"backup\"" intent "Use the supplied text."
 choice "b" field_value at "1" alternative "\"backup\"" intent "Use the supplied text."
 choice "c" field_value at "2" alternative "\"backup\"" intent "Use the supplied text."
 value_case "[\"\"]" -> "{\"a\":\"\",\"b\":\"\",\"c\":\"\"}"
 holdout_value_case "[\"fresh\"]" -> "{\"a\":\"fresh\",\"b\":\"fresh\",\"c\":\"fresh\"}"
 attempts "8"
}
`

func TestRecordHoldoutIsAbsentFromEveryModelContext(t *testing.T) {
	ctx := context.Background()
	before, err := prepareRecordAssembly(ctx, "context.gooo", []byte(recordHoldoutContextSource), "Build")
	if err != nil {
		t.Fatal(err)
	}
	after, err := prepareRecordAssembly(ctx, "context.gooo", []byte(strings.ReplaceAll(recordHoldoutContextSource, "fresh", "different")), "Build")
	if err != nil {
		t.Fatal(err)
	}
	for _, version := range []string{"", jointdecision.RecordFieldFeatureVersion, jointdecision.RecordSharedFeatureVersion, jointdecision.RecordOriginSharedFeatureVersion, jointdecision.RecordGraphSharedFeatureVersion} {
		a, b := recordPlanModelContext(before, version), recordPlanModelContext(after, version)
		if a.Status != "ENCODED" || a.Text == "" || !reflect.DeepEqual(a, b) {
			t.Fatalf("held input or answer changed model context %q: %#v %#v", version, a, b)
		}
	}
}

func TestRecordHoldoutTypesAreCheckedBeforeSelection(t *testing.T) {
	source := strings.Replace(recordHoldoutSource, `holdout_value_case "[5]" -> "{\"value\":-5}"`, `holdout_value_case "[5]" -> "{\"value\":\"wrong\"}"`, 1)
	g, _ := NewTypedPathGenerator("")
	r, err := g.GenerateSourceAssembly(context.Background(), "held.gooo", []byte(source), "Choose")
	if err == nil || !strings.Contains(err.Error(), "record holdout") || r.Report.RecordAssembly != nil {
		t.Fatal("held type mismatch reached selection", err, r)
	}
}
