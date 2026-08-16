package main

import (
	"bytes"
	"encoding/json"
	"go/ast"
	goparser "go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestRunGenerateWritesDeterministicProjection(t *testing.T) {
	outputDir := t.TempDir()
	var stdout, stderr bytes.Buffer
	args := []string{"billing.gooo", "--out", outputDir}
	code := runGenerate(args, fixtureReader{source: validSource}, SyntaxSourceParser{}, &stdout, &stderr)
	if code != exitOK || stderr.Len() != 0 {
		t.Fatalf("generate result = code %d, stdout %q, stderr %q", code, stdout.String(), stderr.String())
	}
	path := filepath.Join(outputDir, "semantic.gooo.go")
	first, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(first), `//gooo:generated:start id="billing://activity/pay-order"`) {
		t.Fatalf("generated source lacks activity marker:\n%s", first)
	}
	if code := runGenerate(args, fixtureReader{source: validSource}, SyntaxSourceParser{}, &bytes.Buffer{}, &bytes.Buffer{}); code != exitOK {
		t.Fatalf("second generate code = %d", code)
	}
	second, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("identical source did not generate identical output")
	}
	if _, err := os.Stat(filepath.Join(outputDir, generatedManifestFileName)); err != nil {
		t.Fatalf("generate did not publish manifest: %v", err)
	}
}

func TestRunGeneratePreservesPreviousGoAndPublishesManifest(t *testing.T) {
	root := t.TempDir()
	sourcePath := filepath.Join(root, "main.gooo")
	if err := os.WriteFile(sourcePath, []byte(validSource), 0o640); err != nil {
		t.Fatal(err)
	}
	firstDir := filepath.Join(root, "first")
	var stdout, stderr bytes.Buffer
	if code := runGenerate([]string{sourcePath, "--out", firstDir}, OSFileReader{}, SyntaxSourceParser{}, &stdout, &stderr); code != exitOK {
		t.Fatalf("initial generate = %d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
	}
	firstPath := filepath.Join(firstDir, generatedFileName)
	first, err := os.ReadFile(firstPath)
	if err != nil {
		t.Fatal(err)
	}
	userBody := "return Order{\n\t\t// user-owned\n\t}"
	previous := bytes.Replace(first, []byte("return Order{}"), []byte(userBody), 1)
	if bytes.Equal(first, previous) {
		t.Fatal("fixture did not create a user-owned previous body")
	}
	previousPath := filepath.Join(root, "previous.go")
	if err := os.WriteFile(previousPath, previous, 0o640); err != nil {
		t.Fatal(err)
	}
	secondDir := filepath.Join(root, "second")
	manifestPath := filepath.Join(root, "evidence", "projection.jsonl")
	stdout.Reset()
	stderr.Reset()
	if code := runGenerate([]string{sourcePath, "--out", secondDir, "--previous-go", previousPath, "--manifest", manifestPath}, OSFileReader{}, SyntaxSourceParser{}, &stdout, &stderr); code != exitOK {
		t.Fatalf("previous-Go generate = %d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
	}
	second, err := os.ReadFile(filepath.Join(secondDir, generatedFileName))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(second, []byte("// user-owned")) || !bytes.Contains(second, []byte(userBody)) {
		t.Fatalf("previous Go slot was not preserved:\n%s", second)
	}
	manifestBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	var manifest projectionManifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		t.Fatalf("manifest is not JSONL JSON: %v", err)
	}
	if manifest.Schema != projectionManifestSchema || manifest.Status != "pass" || !manifest.PreviousGoProvided || manifest.PreviousGoDigest == "" || !manifest.ProtectedBytesEqual || manifest.ResponseDigest == "" || manifest.EvidenceManifest.PayloadSHA256 == "" {
		t.Fatalf("incomplete previous-Go manifest: %#v", manifest)
	}
	if got, err := os.ReadFile(sourcePath); err != nil || !bytes.Equal(got, []byte(validSource)) {
		t.Fatalf("source was modified: %v", err)
	}
}

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

func TestRunGenerateRejectsInvalidInputWithoutMutation(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{name: "syntax", source: "package billing\nentity Broken id \"x\" @", want: "error"},
		{name: "semantic", source: `package billing
namespace billing
entity Order id "billing://entity/order"
activity PayOrder(PayOrder) -> Order
`, want: "generation failed"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			outputDir := t.TempDir()
			outputPath := filepath.Join(outputDir, generatedFileName)
			original := []byte("existing generated bytes\n")
			if err := os.WriteFile(outputPath, original, 0o640); err != nil {
				t.Fatal(err)
			}
			before, err := os.Stat(outputPath)
			if err != nil {
				t.Fatal(err)
			}
			beforeEntries := directoryEntries(t, outputDir)
			var stdout, stderr bytes.Buffer
			code := runGenerate([]string{"fixture.gooo", "--out", outputDir}, fixtureReader{source: test.source}, SyntaxSourceParser{}, &stdout, &stderr)
			if code != exitFailure || stdout.Len() != 0 || !strings.Contains(stderr.String(), test.want) {
				t.Fatalf("%s generate = code %d, stdout=%q, stderr=%q", test.name, code, stdout.String(), stderr.String())
			}
			after, err := os.Stat(outputPath)
			if err != nil {
				t.Fatal(err)
			}
			got, err := os.ReadFile(outputPath)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, original) || !os.SameFile(before, after) || before.Mode() != after.Mode() || !before.ModTime().Equal(after.ModTime()) || !reflect.DeepEqual(beforeEntries, directoryEntries(t, outputDir)) {
				t.Fatalf("%s rejection mutated output state", test.name)
			}
		})
	}
}

func TestRunGenerateRejectsInvalidInputBeforeCreatingOutputRoot(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{name: "syntax", source: "package billing\nentity Broken id \"x\" @", want: "error"},
		{name: "semantic", source: `package billing
namespace billing
entity Order id "billing://entity/order"
activity PayOrder(PayOrder) -> Order
`, want: "generation failed"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parent := t.TempDir()
			outputDir := filepath.Join(parent, "new-output")
			manifestPath := filepath.Join(parent, "manifest.jsonl")
			beforeEntries := directoryEntries(t, parent)
			var stdout, stderr bytes.Buffer
			code := runGenerate([]string{"fixture.gooo", "--out", outputDir, "--manifest", manifestPath}, fixtureReader{source: test.source}, SyntaxSourceParser{}, &stdout, &stderr)
			if code != exitFailure || stdout.Len() != 0 || !strings.Contains(stderr.String(), test.want) {
				t.Fatalf("%s generate = code %d, stdout=%q, stderr=%q", test.name, code, stdout.String(), stderr.String())
			}
			if _, err := os.Lstat(outputDir); !os.IsNotExist(err) {
				t.Fatalf("%s rejection created output root: %v", test.name, err)
			}
			if _, err := os.Lstat(manifestPath); !os.IsNotExist(err) {
				t.Fatalf("%s rejection created manifest: %v", test.name, err)
			}
			if !reflect.DeepEqual(beforeEntries, directoryEntries(t, parent)) {
				t.Fatalf("%s rejection changed parent directory entries", test.name)
			}
		})
	}
}

func TestGenerateDigestMatchSkipsWrite(t *testing.T) {
	outputDir := t.TempDir()
	args := []string{"billing.gooo", "--out", outputDir}
	if code := runGenerate(args, fixtureReader{source: validSource}, SyntaxSourceParser{}, &bytes.Buffer{}, &bytes.Buffer{}); code != exitOK {
		t.Fatalf("initial generate code = %d", code)
	}
	path := filepath.Join(outputDir, generatedFileName)
	before, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(outputDir, 0o555); err != nil {
		t.Skipf("cannot make output root read-only: %v", err)
	}
	defer os.Chmod(outputDir, 0o755)
	if code := runGenerate(args, fixtureReader{source: validSource}, SyntaxSourceParser{}, &bytes.Buffer{}, &bytes.Buffer{}); code != exitOK {
		t.Fatalf("digest-only generate code = %d", code)
	}
	after, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !os.SameFile(before, after) || before.Mode() != after.Mode() || !before.ModTime().Equal(after.ModTime()) {
		t.Fatal("digest match rewrote the generated file")
	}
}
