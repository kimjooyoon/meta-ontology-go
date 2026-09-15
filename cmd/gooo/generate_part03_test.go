package main

import (
	"bytes"
	"go/ast"
	goparser "go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"testing"
)

func TestRunGenerateConformanceRoundTripIsReadOnly(t *testing.T) {
	fixture := filepath.Join("..", "..", "examples", "conformance", "main.gooo")
	beforeBytes, err := os.ReadFile(fixture)
	if err != nil {
		t.Fatal(err)
	}
	beforeInfo, err := os.Stat(fixture)
	if err != nil {
		t.Fatal(err)
	}
	outputDir := t.TempDir()
	code := runGenerate([]string{fixture, "--out", outputDir}, OSFileReader{}, SyntaxSourceParser{}, &bytes.Buffer{}, &bytes.Buffer{})
	if code != exitOK {
		t.Fatalf("conformance generate code = %d", code)
	}
	generatedPath := filepath.Join(outputDir, generatedFileName)
	generated, err := os.ReadFile(generatedPath)
	if err != nil {
		t.Fatal(err)
	}
	fileSet := token.NewFileSet()
	file, err := goparser.ParseFile(fileSet, generatedPath, generated, goparser.ParseComments)
	if err != nil {
		t.Fatalf("generated Go does not parse: %v", err)
	}
	config := types.Config{}
	if _, err := config.Check(file.Name.Name, fileSet, []*ast.File{file}, nil); err != nil {
		t.Fatalf("generated Go does not type-check: %v", err)
	}
	afterBytes, err := os.ReadFile(fixture)
	if err != nil {
		t.Fatal(err)
	}
	afterInfo, err := os.Stat(fixture)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(beforeBytes, afterBytes) || !os.SameFile(beforeInfo, afterInfo) || beforeInfo.Mode() != afterInfo.Mode() || !beforeInfo.ModTime().Equal(afterInfo.ModTime()) {
		t.Fatal("generate mutated the source fixture")
	}
}
func TestRunGenerateRejectsInvalidUsage(t *testing.T) {
	var stderr bytes.Buffer
	code := runGenerate([]string{"billing.gooo"}, fixtureReader{}, SyntaxSourceParser{}, &bytes.Buffer{}, &stderr)
	if code != exitUsage || stderr.String() != "usage: gooo generate <file.gooo> --out <directory>\n" {
		t.Fatalf("usage = code %d, stderr %q", code, stderr.String())
	}
}
