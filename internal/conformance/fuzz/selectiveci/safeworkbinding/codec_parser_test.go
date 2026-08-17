package safeworkbinding

import "testing"

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

func TestParseDocumentSyntax(t *testing.T) {
	cases := []struct {
		name  string
		input string
	}{
		{
			name:  "empty",
			input: "",
		},
		{
			name:  "open_object",
			input: `{`,
		},
		{
			name:  "object_trailing_comma",
			input: `{"a":1,}`,
		},
		{
			name:  "array_trailing_comma",
			input: `[1,]`,
		},
		{
			name:  "missing_colon",
			input: `{"a" 1}`,
		},
		{
			name:  "missing_value",
			input: `{"a":}`,
		},
		{
			name:  "invalid_escape",
			input: `{"x":"\q"}`,
		},
		{
			name:  "lone_high",
			input: `{"x":"\uD800"}`,
		},
		{
			name:  "lone_low",
			input: `{"x":"\uDC00"}`,
		},
		{
			name:  "invalid_suffix",
			input: `{}!`,
		},
		{
			name:  "invalid_later_suffix",
			input: `{}null!`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			requireDocumentReason(t, []byte(tc.input), ReasonInvalidJSON)
		})
	}
}

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
