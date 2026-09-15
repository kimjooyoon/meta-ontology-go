package safeworkbinding

import (
	"testing"
)

func TestParseDocumentDecodedDuplicateKeys(t *testing.T) {
	cases := []struct {
		name  string
		input string
	}{
		{
			name:  "decoded_key",
			input: `{"a":1,"\u0061":2}`,
		},
		{
			name:  "nested_array",
			input: `{"task_id":[{"x":1,"x":2}]}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			requireDocumentReason(t, []byte(tc.input), ReasonDuplicateKey)
		})
	}
	requireDocumentReason(t, []byte(`[{"a":1},{"a":2}]`), ReasonInvalidSchema)
}
func TestParseDocumentPrecedence(t *testing.T) {
	requireDocumentReason(t, []byte(`{"a":1,"a":2`), ReasonInvalidJSON)
	requireDocumentReason(t, []byte(`{"a":1,"a":2}null`), ReasonDuplicateKey)
	requireDocumentReason(t, []byte(`{}{"a":1,"\u0061":2}`), ReasonDuplicateKey)
	requireDocumentReason(t, []byte(`null{}`), ReasonTrailingValue)
}
func TestParseDocumentTrailingValues(t *testing.T) {
	requireDocumentReason(t, []byte(`{}null`), ReasonTrailingValue)
	requireDocumentReason(t, []byte(`{}null[]`), ReasonTrailingValue)
	requireDocumentReason(t, []byte(`{}`), ReasonNone)
}
func TestParseDocumentRootKinds(t *testing.T) {
	requireDocumentReason(t, []byte(`null`), ReasonInvalidSchema)
	requireDocumentReason(t, []byte(`[]`), ReasonInvalidSchema)
	requireDocumentReason(t, []byte(`0`), ReasonInvalidSchema)
	requireDocumentReason(t, []byte(`"x"`), ReasonInvalidSchema)
}
func TestParseDocumentReplacementCharacter(t *testing.T) {
	value := requireDocumentReason(t, []byte(`{"x":"\uFFFD"}`), ReasonNone)
	member, ok := value.object["x"]
	if !ok {
		t.Fatal("missing member x")
	}
	if member.kind != jsonStringValue {
		t.Fatalf("member kind=%v, want string", member.kind)
	}
	if member.text != "�" {
		t.Fatalf("member text=%q, want replacement character", member.text)
	}
}
