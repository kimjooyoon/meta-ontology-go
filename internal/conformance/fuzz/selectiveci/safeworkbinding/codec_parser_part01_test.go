package safeworkbinding

import (
	"testing"
)

func requireDocumentReason(t *testing.T, input []byte, want Reason) jsonValue {
	t.Helper()
	value, reason := parseDocument(input)
	if reason != want {
		t.Fatalf("reason=%v, want %v", reason, want)
	}
	if reason != ReasonNone {
		if value.kind != jsonNullValue {
			t.Errorf("kind=%v, want zero", value.kind)
		}
		if value.text != "" {
			t.Errorf("text=%q, want empty", value.text)
		}
		if value.object != nil {
			t.Errorf("object=%v, want nil", value.object)
		}
	}
	return value
}
func TestParseDocumentBOMPrecedesUTF8(t *testing.T) {
	input := []byte{0xEF, 0xBB, 0xBF, 0x7B, 0x7D, 0xFF}
	requireDocumentReason(t, input, ReasonBOMForbidden)
}
func TestParseDocumentUTF8BeforeJSON(t *testing.T) {
	input := []byte{0x7B, 0x7D, 0xFF}
	requireDocumentReason(t, input, ReasonInvalidUTF8)
}
func TestParseDocumentRecursiveValues(t *testing.T) {
	empty := requireDocumentReason(t, []byte(`{}`), ReasonNone)
	if empty.kind != jsonObjectValue {
		t.Fatalf("empty kind=%v, want object", empty.kind)
	}
	if empty.object == nil {
		t.Fatal("empty object map is nil")
	}
	if len(empty.object) != 0 {
		t.Fatalf("empty object has %d members", len(empty.object))
	}
	input := []byte(" {\"a\":[null,true,false,0,\"x\",{}]}\r\n")
	value := requireDocumentReason(t, input, ReasonNone)
	member, ok := value.object["a"]
	if !ok {
		t.Fatal("missing member a")
	}
	if member.kind != jsonArrayValue {
		t.Fatalf("member kind=%v, want array", member.kind)
	}
}
