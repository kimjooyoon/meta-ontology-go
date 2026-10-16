package bodyexecution

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/textbound"
)

func TestTextBoundAndAggregateBudgetStaySeparateFresh1905(t *testing.T) {
	text := strings.Repeat("\x00", CompositionInputMaxBytes/6+1)
	value, _ := json.Marshal(map[string]string{"value": text})
	record := bodycodegen.RecordType{Name: "Box", Fields: []bodycodegen.RecordField{{
		Name: "value", TypeID: "urn:gooo:type:string", TextBound: textbound.New(CompositionInputMaxBytes + 1),
	}}}
	if _, err := canonicalRecord(value, record); err != nil {
		t.Fatal("field domain rejected an in-range decoded value", err)
	}
	raw, _ := json.Marshal(map[string]any{"schema": CompositionInputsSchema,
		"inputs": []any{map[string]json.RawMessage{"Entry": value}}})
	if len(raw) <= CompositionInputMaxBytes {
		t.Fatal("fixture did not cross the independent wire budget")
	}
	if _, err := DecodeCompositionInputs(raw); err == nil {
		t.Fatal("large declared field widened the aggregate input budget")
	}
}

func TestCapacityHintCannotChooseSourceDomainFresh1905(t *testing.T) {
	hint := CompositionTransportCapacity()
	hint.Schema, hint.TextUTF8Bytes, hint.InputJSONBytes, hint.Rows = "malformed-advisory", 1000000, -1, -1
	if _, err := canonicalRecordFieldWithBound([]byte(`"abc"`), "Text", "one", textbound.New(2)); err == nil {
		t.Fatal("advisory mutation widened the source field")
	}
	if _, err := canonicalRecordFieldWithBound([]byte(`"ab"`), "Text", "one", textbound.New(2)); err != nil {
		t.Fatal("malformed advisory blocked a source-valid value", err, hint)
	}
	if CompositionTransportCapacity().TextUTF8Bytes != 1024 || CompositionTransportCapacity().InputJSONBytes != 1<<20 {
		t.Fatal("inspection value became mutable admission state")
	}
}
