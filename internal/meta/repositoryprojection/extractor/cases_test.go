package extractor

import (
	"bytes"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractionCounterexamples(t *testing.T) {
	cases := []struct{ name, logical, source, reason string }{
		{"unsafe path", "../escape.go", "package p\nfunc F() {}\n", "WRITE_SET_ESCAPE"},
		{"unresolved import", "x.go", "package p\nimport \"example.invalid/missing\"\nfunc F() {}\n", "UNRESOLVED_IMPORT"},
		{"build conflict", "x.go", "//go:build linux\n// +build windows\n\npackage p\nfunc F() {}\n", "BUILD_TAG_CONFLICT"},
		{"identity collision", "x.go", "package p\nfunc F() {}\nfunc F() {}\n", "DECLARATION_IDENTITY_COLLISION"},
		{"embed missing import", "x.go", "package p\n\n//go:embed fixture.txt\nvar fixture string\n", "EMBED_IMPORT_MISSING"},
		{"linkname missing unsafe", "x.go", "package p\n\n//go:linkname linked runtime.linked\nfunc linked() {}\n", "LINKNAME_UNSAFE_IMPORT_MISSING"},
		{"cgo import relocation", "x.go", "package p\n\n/* #include <stdio.h> */\nimport \"C\"\n\nfunc F() {}\n", "CGO_IMPORT_RELOCATION_UNSAFE"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.test\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			if tc.logical == "x.go" {
				if err := os.WriteFile(filepath.Join(root, tc.logical), []byte(tc.source), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			_, _, err := Extract(root, tc.logical)
			var got Failure
			if !errors.As(err, &got) || got.Reason != tc.reason {
				t.Fatalf("reason=%v want %s", err, tc.reason)
			}
		})
	}
}

func TestExtractionPreservesDirectiveImports(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	source := []byte("package p\n\nimport (\n\t_ \"embed\"\n\t_ \"unsafe\"\n)\n\n//go:embed fixture.txt\nvar fixture string\n\n//go:linkname linked runtime.linked\nfunc linked() {}\n")
	if err := os.WriteFile(filepath.Join(root, "x.go"), source, 0o644); err != nil {
		t.Fatal(err)
	}
	generated, _, err := Extract(root, "x.go")
	if err != nil {
		t.Fatal(err)
	}
	embedFound, linknameFound := false, false
	for path, data := range generated {
		if path == "x.go" {
			continue
		}
		file, parseErr := parser.ParseFile(token.NewFileSet(), path, data, parser.ParseComments)
		if parseErr != nil {
			t.Fatalf("generated helper is not parseable: %s: %v", path, parseErr)
		}
		if hasDeclarationDirective(file, "embed") {
			embedFound = true
			if !hasImportAlias(file, "embed", "_") {
				t.Fatalf("embed directive helper lost embed import: %s", path)
			}
		}
		if hasDeclarationDirective(file, "linkname") {
			linknameFound = true
			if !hasImportAlias(file, "unsafe", "_") {
				t.Fatalf("linkname directive helper lost unsafe import: %s", path)
			}
		}
	}
	if !embedFound {
		t.Fatal("generated helper lost //go:embed declaration directive")
	}
	if !linknameFound {
		t.Fatal("generated helper lost //go:linkname declaration directive")
	}
}

func hasImportAlias(file *ast.File, path, alias string) bool {
	for _, declaration := range file.Decls {
		group, ok := declaration.(*ast.GenDecl)
		if !ok || group.Tok.String() != "import" {
			continue
		}
		for _, raw := range group.Specs {
			spec, ok := raw.(*ast.ImportSpec)
			if !ok || spec.Path.Value != `"`+path+`"` || spec.Name == nil || spec.Name.Name != alias {
				continue
			}
			return true
		}
	}
	return false
}

func hasDeclarationDirective(file *ast.File, directive string) bool {
	prefix := "//go:" + directive
	for _, declaration := range file.Decls {
		var docs *ast.CommentGroup
		switch node := declaration.(type) {
		case *ast.FuncDecl:
			docs = node.Doc
		case *ast.GenDecl:
			docs = node.Doc
		}
		if docs == nil {
			continue
		}
		for _, comment := range docs.List {
			text := strings.TrimSpace(comment.Text)
			if text == prefix || strings.HasPrefix(text, prefix+" ") || strings.HasPrefix(text, prefix+"\t") {
				return true
			}
		}
	}
	return false
}

func TestExtractionReplayIsByteIdentical(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	source := []byte("package p\n\nfunc First() {}\n\nfunc Second() {}\n")
	if err := os.WriteFile(filepath.Join(root, "x.go"), source, 0o644); err != nil {
		t.Fatal(err)
	}
	first, paths, err := Extract(root, "x.go")
	if err != nil {
		t.Fatal(err)
	}
	second, secondPaths, err := Extract(root, "x.go")
	if err != nil || len(paths) != len(secondPaths) {
		t.Fatal(err)
	}
	for _, path := range paths {
		if !bytes.Equal(first[path], second[path]) {
			t.Fatalf("nondeterministic output for %s", path)
		}
	}
}
