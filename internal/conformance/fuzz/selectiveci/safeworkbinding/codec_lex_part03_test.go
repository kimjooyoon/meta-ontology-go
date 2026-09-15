package safeworkbinding

import (
	"testing"
)

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
