package bodylibrary

import (
	"context"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestLibraryPackageIntroductionKeepsOriginsOutsideGoDocumentation(t *testing.T) {
	const source = `package docs
namespace docs
entity Integer id "urn:gooo:type:integer"
activity Keep(Integer) -> Integer id "urn:docs:keep" computes "return input"
`
	generated, err := GenerateWithStageExports(context.Background(), "docs.gooo", []byte(source), "Keep", "views",
		[]StageExport{{Name: "Keep", ActivityID: "urn:docs:keep"}})
	if err != nil {
		t.Fatal(err)
	}
	positions := token.NewFileSet()
	file, err := parser.ParseFile(positions, "library.go", generated, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	metadata, err := doc.NewFromFiles(positions, []*ast.File{file}, "example.test/views", doc.PreserveAST)
	if err != nil {
		t.Fatal(err)
	}
	if got := doc.Synopsis(metadata.Doc); got != "Package views exposes a source-defined Gooo computation through Evaluate." {
		t.Fatal("generated package introduction describes metadata", got)
	}
	for _, marker := range []string{"Exported stage", "Lowered Gooo source:", "Typed plan:"} {
		if strings.Contains(metadata.Doc, marker) || !strings.Contains(generated, marker) {
			t.Fatal("source origin was lost or promoted to package documentation", marker)
		}
	}
	if !ast.IsGenerated(file) {
		t.Fatal("standard generated-source recognition was lost")
	}
}
