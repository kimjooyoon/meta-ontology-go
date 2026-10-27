package main

import (
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestPackageGenerateIntroductionFollowsCompilerMetadata(t *testing.T) {
	root := t.TempDir()
	writeWorkspaceSource(t, root, "main.gooo", libraryHygieneSource)
	manifest := writeWorkspaceManifest(t, root, `{"schema":"gooo/package-workspace-manifest/v1","entry":{"package_path":"hygiene","activity":"Keep"},"packages":[{"path":"hygiene","sources":["main.gooo"]}]}`)
	generated := generateLibraryCommand(t, "views", manifest)
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
		t.Fatal("compiler metadata became the package synopsis", got)
	}
	if !strings.HasPrefix(string(generated), "// Compiler version=") || strings.Contains(metadata.Doc, "Compiler version=") || !ast.IsGenerated(file) {
		t.Fatal("compiler origins or generated status changed")
	}
}
