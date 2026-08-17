package safeworkbinding

import "testing"

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

func TestJSONLex_NumberTokens(t *testing.T) {
	cases := []struct {
		input  string
		offset int
		valid  bool
	}{
		{input: "0x", offset: 1, valid: true},
		{input: "-0,", offset: 2, valid: true},
		{input: "12]", offset: 2, valid: true},
		{input: "-12]", offset: 3, valid: true},
		{input: "1.0,", offset: 3, valid: true},
		{input: "1e2]", offset: 3, valid: true},
		{input: "1E-2}", offset: 4, valid: true},
		{input: "1e+2}", offset: 4, valid: true},
		{input: "01", valid: false},
		{input: "-01", valid: false},
		{input: ".1", valid: false},
		{input: "1.", valid: false},
		{input: "1e", valid: false},
		{input: "+1", valid: false},
		{input: "--1", valid: false},
	}
	for _, tc := range cases {
		parser := jsonParser{data: []byte(tc.input)}
		valid := parser.parseNumber()
		checkToken(t, tc.input, valid, tc.valid, parser.offset, tc.offset)
	}
}

func TestJSONLex_Literals(t *testing.T) {
	cases := []struct {
		input   string
		literal string
		offset  int
		valid   bool
	}{
		{input: "null,", literal: "null", offset: 4, valid: true},
		{input: "true]", literal: "true", offset: 4, valid: true},
		{input: "false}", literal: "false", offset: 5, valid: true},
		{input: "nul", literal: "null", valid: false},
		{input: "truth", literal: "true", valid: false},
		{input: "nil", literal: "nil", valid: false},
		{input: "true", literal: "false", valid: false},
	}
	for _, tc := range cases {
		parser := jsonParser{data: []byte(tc.input)}
		valid := parser.parseLiteral(tc.literal)
		checkToken(t, tc.input, valid, tc.valid, parser.offset, tc.offset)
	}
}

func TestJSONLex_WhitespaceAndHex(t *testing.T) {
	for _, value := range []byte{' ', '\t', '\n', '\r'} {
		parser := jsonParser{data: []byte{value, 'x'}}
		parser.skipSpace()
		if parser.offset != 1 {
			t.Errorf("whitespace %q stopped at %d", value, parser.offset)
		}
	}
	for _, value := range []byte{0, '\v', '\f', 0xA0} {
		parser := jsonParser{data: []byte{value}}
		parser.skipSpace()
		if parser.offset != 0 {
			t.Errorf("non-whitespace %x was consumed", value)
		}
	}
	for index, value := range []byte("0123456789abcdefABCDEF") {
		want := byte(index)
		if index >= 16 {
			want -= 6
		}
		got, ok := hexValue(value)
		if !ok || got != want {
			t.Errorf("hex %q failed", value)
		}
	}
	for _, value := range []byte{'g', 'G', '/', ' '} {
		if _, ok := hexValue(value); ok {
			t.Errorf("non-hex %q accepted", value)
		}
	}
}
