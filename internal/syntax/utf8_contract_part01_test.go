package syntax

import (
	"strings"
	"testing"
)

func TestQuotedStringInvalidUTF8PolicyAcrossRecoveryPaths(t *testing.T) {
	invalid := string([]byte{0xff, 0xfe})
	tests := []struct {
		name       string
		source     string
		codes      []DiagnosticCode
		byteOffset int
	}{
		{
			name:       "ordinary content",
			source:     quotedIDSource("A", invalid),
			codes:      []DiagnosticCode{DiagInvalidUTF8, DiagInvalidUTF8},
			byteOffset: strings.IndexByte(quotedIDSource("A", invalid), 0xff),
		},
		{
			name:       "after backslash",
			source:     quotedIDSource("A", "\\"+invalid),
			codes:      []DiagnosticCode{DiagInvalidUTF8, DiagInvalidUTF8},
			byteOffset: strings.IndexByte(quotedIDSource("A", "\\"+invalid), 0xff),
		},
		{
			name:       "inside unicode escape",
			source:     quotedIDSource("A", "\\u"+invalid+"00"),
			codes:      []DiagnosticCode{DiagInvalidUTF8, DiagInvalidUTF8, DiagInvalidEscape},
			byteOffset: strings.IndexByte(quotedIDSource("A", "\\u"+invalid+"00"), 0xff),
		},
		{
			name:       "unicode and CRLF",
			source:     "package p\r\nnamespace 도메인\r\nentity 注文 id \"앞" + invalid + "뒤\"\r\n",
			codes:      []DiagnosticCode{DiagInvalidUTF8, DiagInvalidUTF8},
			byteOffset: strings.IndexByte("package p\r\nnamespace 도메인\r\nentity 注文 id \"앞"+invalid+"뒤\"\r\n", 0xff),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, diagnostics := ParseFile("utf8.gooo", test.source)
			if len(diagnostics) != len(test.codes) {
				t.Fatalf("diagnostics = %#v, want %v", diagnostics, test.codes)
			}
			for index, diagnostic := range diagnostics {
				if diagnostic.Code != test.codes[index] {
					t.Fatalf("diagnostic %d = %q, want %q", index, diagnostic.Code, test.codes[index])
				}
				if index < 2 {
					wantOffset := test.byteOffset + index
					if diagnostic.Span.Start.Offset != wantOffset || diagnostic.Span.End.Offset != wantOffset+1 {
						t.Fatalf("diagnostic %d span = %#v, want byte [%d,%d)", index, diagnostic.Span, wantOffset, wantOffset+1)
					}
					if test.name == "unicode and CRLF" && (diagnostic.Span.Start.Line != 3 || diagnostic.Span.Start.Column != 16+index) {
						t.Fatalf("diagnostic %d position = %#v, want 3:%d", index, diagnostic.Span.Start, 16+index)
					}
				}
			}
		})
	}
}
func TestQuotedStringInvalidUTF8BlocksFormatting(t *testing.T) {
	source := quotedIDSource("A", string([]byte{'x', 0xff, 'y'}))
	formatted, diagnostics, err := FormatSource("invalid.gooo", source)
	if err == nil || len(diagnostics) != 1 || diagnostics[0].Code != DiagInvalidUTF8 || formatted != "" {
		t.Fatalf("invalid UTF-8 format result = %q, %#v, %v", formatted, diagnostics, err)
	}
}
