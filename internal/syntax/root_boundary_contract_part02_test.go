package syntax

import (
	"reflect"
	"strings"
	"testing"
)

func TestParserRecoveryAndDiagnosticOrderingDoNotMutateInputs(t *testing.T) {
	invalid := string([]byte{0xff})
	source := "package 도메인\r\nnamespace 도메인\r\nentity 注文 id \"앞" + invalid + "뒤\"\r\n"
	originalSource := source
	parser := NewParserFile("immutable.gooo", source)
	file, diagnostics := parser.Parse()
	repeatedFile, repeatedDiagnostics := parser.Parse()
	originalDiagnostics := append(Diagnostics(nil), diagnostics...)
	firstSorted := diagnostics.SortBySpan()
	secondSorted := diagnostics.SortBySpan()
	invalidOffset := strings.IndexByte(source, 0xff)

	if source != originalSource {
		t.Fatal("parser modified source input")
	}
	if !reflect.DeepEqual(diagnostics, originalDiagnostics) {
		t.Fatal("sorting modified diagnostic input")
	}
	if !reflect.DeepEqual(firstSorted, secondSorted) {
		t.Fatalf("diagnostic ordering was not deterministic: %#v vs %#v", firstSorted, secondSorted)
	}
	if !reflect.DeepEqual(file, repeatedFile) || !reflect.DeepEqual(diagnostics, repeatedDiagnostics) {
		t.Fatal("repeated recovery changed the parsed result")
	}
	if file.Span.End.Offset != len(source) || file.Span.End.Line != 4 || file.Span.End.Column != 1 {
		t.Fatalf("recovered root span = %#v", file.Span)
	}
	if len(diagnostics) != 1 || diagnostics[0].Code != DiagInvalidUTF8 {
		t.Fatalf("malformed UTF-8 diagnostics = %#v", diagnostics)
	}
	invalidSpan := diagnostics[0].Span
	if invalidSpan.Start.Offset != invalidOffset || invalidSpan.End.Offset != invalidOffset+1 || invalidSpan.Start.Line != 3 || invalidSpan.Start.Column != 16 {
		t.Fatalf("invalid UTF-8 span = %#v, want byte %d at 3:16", invalidSpan, invalidOffset)
	}
	for index, diagnostic := range firstSorted {
		if index == 0 || diagnostic.Span.Start.Offset >= firstSorted[index-1].Span.Start.Offset {
			continue
		}
		t.Fatalf("diagnostics were not ordered by byte offset: %#v", firstSorted)
	}
}
