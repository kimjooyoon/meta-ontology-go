package syntax

import (
	"reflect"
	"testing"
)

func TestQuotedStringASCIIEscapeDiagnosticsRemainDistinct(t *testing.T) {
	sources := []string{
		quotedIDSource("A", `\q`),
		quotedIDSource("A", `\u12xz`),
	}
	for _, source := range sources {
		_, diagnostics := Parse(source)
		if len(diagnostics) != 1 || diagnostics[0].Code != DiagInvalidEscape {
			t.Fatalf("source %q diagnostics = %#v", source, diagnostics)
		}
	}
}
func TestQuotedStringInvalidUTF8DiagnosticsAreDeterministic(t *testing.T) {
	invalid := string([]byte{0xff, 0xfe})
	source := quotedIDSource("注文", "앞"+invalid+"뒤")
	firstFile, firstDiagnostics := ParseFile("repeat.gooo", source)
	secondFile, secondDiagnostics := ParseFile("repeat.gooo", source)
	if !reflect.DeepEqual(firstFile, secondFile) || !reflect.DeepEqual(firstDiagnostics, secondDiagnostics) {
		t.Fatal("repeated malformed UTF-8 parses differed")
	}
	if len(firstDiagnostics) != 2 || firstDiagnostics[0].Span.Start.Offset >= firstDiagnostics[1].Span.Start.Offset {
		t.Fatalf("diagnostics are not source ordered: %#v", firstDiagnostics)
	}
}
func quotedIDSource(name, value string) string {
	return "package p\nnamespace n\nentity " + name + " id \"" + value + "\""
}
