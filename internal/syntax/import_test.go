package syntax

import "testing"

func TestImportsParseCloneAndFormatInStableOrder(t *testing.T) {
	source := `package app
namespace app
import "z/package"
import "a/package"
entity Text id "urn:gooo:text"
`
	file, diagnostics := Parse(source)
	if diagnostics.HasErrors() || len(file.Imports) != 2 {
		t.Fatalf("imports did not parse: diagnostics=%v file=%#v", diagnostics, file)
	}
	clone := file.Clone()
	clone.Imports[0].Path = "changed"
	if file.Imports[0].Path != "z/package" {
		t.Fatalf("clone shares import storage: original=%#v clone=%#v", file.Imports, clone.Imports)
	}
	formatted, err := Format(file)
	if err != nil {
		t.Fatalf("format imports: %v", err)
	}
	want := "package app\nnamespace app\nimport \"a/package\"\nimport \"z/package\"\n\nentity Text id \"urn:gooo:text\"\n"
	if formatted != want {
		t.Fatalf("unexpected canonical imports:\nwant=%q\n got=%q", want, formatted)
	}
	again, diagnostics := Parse(formatted)
	if diagnostics.HasErrors() {
		t.Fatalf("formatted imports did not parse: %v", diagnostics)
	}
	second, err := Format(again)
	if err != nil || second != formatted {
		t.Fatalf("import formatting did not reach a fixed point: err=%v\nfirst=%q\nsecond=%q", err, formatted, second)
	}
}

func TestDuplicateSourceImportsAreRejected(t *testing.T) {
	file, diagnostics := Parse(`package app
namespace app
import "core"
import "core"
`)
	if file == nil || !diagnostics.HasErrors() {
		t.Fatalf("duplicate source imports were accepted: diagnostics=%v file=%#v", diagnostics, file)
	}
}
