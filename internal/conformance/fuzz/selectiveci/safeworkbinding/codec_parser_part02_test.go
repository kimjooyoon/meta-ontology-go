package safeworkbinding

import (
	"testing"
)

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
