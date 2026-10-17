package valueexecution

import (
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/textbound"
)

func TestCompiledRecordPlanKeepsTextDomainFresh1905(t *testing.T) {
	const source = `package bound_transport
namespace bound_transport
entity Box id "bound1905://transport/box" fields {
 field value id "bound1905://transport/value" type string required one max_utf8_bytes "1030"
}
activity Echo(Box) -> Box id "bound1905://transport/echo" computes "record.forward:v1"
`
	plan, err := CompileRecordPlan("transport.gooo", []byte(source))
	if err != nil || plan.programs["Echo"].fields[0].TextBound != textbound.New(1030) {
		t.Fatal("source domain was lost before record transport", err)
	}
	for _, size := range []int{1030, 1031} {
		execution, err := plan.Execute(map[string]RecordFields{"Echo": {"value": strings.Repeat("x", size)}})
		if size == 1031 {
			if err == nil || execution.ApplyCalls != 0 || len(execution.Results) != 0 {
				t.Fatal("out-of-domain input received transport credit", execution, err)
			}
			continue
		}
		if err != nil || execution.Results["Echo"].Fields["value"] != strings.Repeat("x", size) {
			t.Fatal("admitted source value changed", execution, err)
		}
	}
	wider, err := CompileRecordPlan("transport.gooo", []byte(strings.Replace(source, `"1030"`, `"1031"`, 1)))
	if err != nil || wider.SemanticFingerprint == plan.SemanticFingerprint || wider.SourceDigest == plan.SourceDigest {
		t.Fatal("a different field domain retained old transport identity", err)
	}
}

func TestRecordTransportUsesDeclaredTextDomainFresh1905(t *testing.T) {
	field := RecordField{TypeID: string(semantic.BuiltinStringTypeID), TextBound: textbound.New(2)}
	if !recordFieldValueMatches(field, "ab") || recordFieldValueMatches(field, "한") {
		t.Fatal("transport did not count decoded UTF-8 bytes")
	}
	field.Cardinality = semantic.Many
	if !recordFieldValueMatches(field, []any{"", "ab"}) || recordFieldValueMatches(field, []any{"ab", "abc"}) {
		t.Fatal("per-element domain was lost")
	}
	field.TextBound = textbound.New(0)
	if !recordFieldValueMatches(field, []any{}) || recordFieldValueMatches(field, []any{"x"}) {
		t.Fatal("zero bound and empty collection changed meaning")
	}
}

func TestRecordTransportPreservesRawEscapeRefusalFresh1905(t *testing.T) {
	for _, raw := range []string{`{"\ud800":1}`, `{"value":"\udfff"}`, `{"value":"a","value":"b"}`} {
		if fields, err := DecodeRecordInput([]byte(raw)); err == nil || fields != nil {
			t.Fatal("raw record ambiguity normalized", fields, err)
		}
	}
	if _, err := DecodeRecordInput([]byte(`{"value":"\ufffd"}`)); err != nil {
		t.Fatal("an explicit replacement scalar is a valid Text value", err)
	}
}
