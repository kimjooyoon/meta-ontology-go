package syntax

import (
	"reflect"
	"testing"
)

func FuzzParseMalformedSeeds(f *testing.F) {
	for _, seed := range malformedSyntaxSeeds() {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, source string) {
		firstFile, firstDiagnostics := ParseFile("fuzz.gooo", source)
		secondFile, secondDiagnostics := ParseFile("fuzz.gooo", source)
		if !reflect.DeepEqual(firstFile, secondFile) || !reflect.DeepEqual(firstDiagnostics, secondDiagnostics) {
			t.Fatal("parser result is not deterministic")
		}
		if firstFile == nil {
			t.Fatal("parser returned nil file")
		}
		if _, err := Format(firstFile); err != nil && !firstDiagnostics.HasErrors() {
			t.Fatalf("valid parse could not be formatted: %v", err)
		}
	})
}

func malformedSyntaxSeeds() []string {
	return []string{
		"",
		"@ package p namespace n",
		"package",
		"package p namespace",
		"package p namespace n entity",
		"package p namespace n entity A id",
		"package p namespace n entity A id \"unterminated",
		"package p namespace n activity Run(One Two) -> Result",
		"package p namespace n activity Run(, ) -> Result",
		"package p namespace n activity Run(One) Result",
		"/* unterminated",
		string([]byte{'p', 'a', 'c', 'k', 'a', 'g', 'e', ' ', 0xff}),
		string(append([]byte("package p\nnamespace n\nentity A id \""), 0xff, 0xfe, '"')),
		quotedIDSource("A", `\ud800`),
		"package 도메인\nnamespace 도메인\nentity 注文 id \"urn:注文\"\n",
		"package p\rnamespace n\rentity A id \"urn:a\"\ractivity Run() -> A\r",
	}
}
