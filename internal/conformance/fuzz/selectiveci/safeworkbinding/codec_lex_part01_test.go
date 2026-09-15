package safeworkbinding

import (
	"testing"
)

type stringCase struct {
	name  string
	input string
	want  string
	valid bool
}

func checkStringCases(t *testing.T, cases []stringCase) {
	for _, tc := range cases {
		parser := jsonParser{data: []byte(tc.input)}
		got, valid := parser.parseString()
		if valid != tc.valid {
			t.Errorf("%s: valid=%v, want %v", tc.name, valid, tc.valid)
		}
		if valid && got != tc.want {
			t.Errorf("%s: got %q, want %q", tc.name, got, tc.want)
		}
		if valid && parser.offset != len(tc.input) {
			t.Errorf("%s: offset=%d, want %d", tc.name, parser.offset, len(tc.input))
		}
	}
}
func checkToken(t *testing.T, name string, valid, wantValid bool, offset, wantOffset int) {
	if valid != wantValid {
		t.Errorf("%s: valid=%v, want %v", name, valid, wantValid)
	}
	if valid && offset != wantOffset {
		t.Errorf("%s: offset=%d, want %d", name, offset, wantOffset)
	}
}
func TestJSONLex_StringEscapes(t *testing.T) {
	checkStringCases(t, []stringCase{
		{name: "ordinary", input: `"hello"`, want: "hello", valid: true},
		{name: "quote", input: `"\""`, want: `"`, valid: true},
		{name: "backslash", input: `"\\"`, want: `\`, valid: true},
		{name: "slash", input: `"\/"`, want: "/", valid: true},
		{name: "backspace", input: `"\b"`, want: "\b", valid: true},
		{name: "formfeed", input: `"\f"`, want: "\f", valid: true},
		{name: "newline", input: `"\n"`, want: "\n", valid: true},
		{name: "carriage_return", input: `"\r"`, want: "\r", valid: true},
		{name: "tab", input: `"\t"`, want: "\t", valid: true},
		{name: "unicode", input: `"\u0061"`, want: "a", valid: true},
		{name: "raw_replacement", input: `"�"`, want: "�", valid: true},
		{name: "escaped_replacement", input: `"\uFFFD"`, want: "�", valid: true},
		{name: "raw_control", input: string([]byte{'"', 1, '"'}), valid: false},
		{name: "unknown_escape", input: `"\q"`, valid: false},
		{name: "truncated_escape", input: `"\u12"`, valid: false},
	})
}
func TestJSONLex_Surrogates(t *testing.T) {
	checkStringCases(t, []stringCase{
		{name: "pair", input: `"\uD83D\uDE00"`, want: "😀", valid: true},
		{name: "lone_high", input: `"\uD800"`, valid: false},
		{name: "lone_low", input: `"\uDC00"`, valid: false},
		{name: "non_low_follower", input: `"\uD800\u0061"`, valid: false},
		{name: "truncated_pair", input: `"\uD800\u"`, valid: false},
		{name: "invalid_hex", input: `"\uD800\u12ZZ"`, valid: false},
	})
}
