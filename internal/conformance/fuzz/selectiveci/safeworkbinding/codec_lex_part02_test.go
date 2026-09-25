package safeworkbinding

import (
	"testing"
)

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
