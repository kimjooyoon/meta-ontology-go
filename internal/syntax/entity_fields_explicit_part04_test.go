package syntax

import (
	"reflect"
	"testing"
)

func TestSupportedEntityFieldsParserReuseIsDeterministic(t *testing.T) {
	source := `package p namespace n entity E id "urn:e" fields { field F id "urn:f" type "gooo:string" optional many }`
	parser := NewParserWithEntityFieldsSupport(source, supportedEntityFields())
	first, firstDiagnostics := parser.Parse()
	second, secondDiagnostics := parser.Parse()
	if first != second || !reflect.DeepEqual(first, second) || !reflect.DeepEqual(firstDiagnostics, secondDiagnostics) {
		t.Fatal("explicit parser reuse was not immutable and deterministic")
	}
}
func FuzzParseEntityFieldsSupported(f *testing.F) {
	for _, seed := range []string{
		`package p namespace n entity E id "urn:e" fields {}`,
		`package p namespace n entity E id "urn:e" fields { field F id "urn:f" type string required one }`,
		`package p namespace n entity E id "urn:e" fields { field F id "urn:f" type "gooo:string" optional many }`,
		"package p\nnamespace n\nentity E id \"urn:e\" fields { field 名 id \"urn:f\" type \"urn:t\" required one }",
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, source string) {
		first, firstDiagnostics := ParseWithEntityFieldsSupport(source, supportedEntityFields())
		second, secondDiagnostics := ParseWithEntityFieldsSupport(source, supportedEntityFields())
		if !reflect.DeepEqual(first, second) || !reflect.DeepEqual(firstDiagnostics, secondDiagnostics) {
			t.Fatal("explicit parser replay changed")
		}
		if first != nil && !firstDiagnostics.HasErrors() {
			if _, err := FormatWithEntityFieldsSupport(first, supportedEntityFields()); err != nil {
				t.Fatalf("valid explicit parse cannot format: %v", err)
			}
		}
	})
}
