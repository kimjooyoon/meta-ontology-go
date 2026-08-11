package main

import (
	"bytes"
	"os"
	"path/filepath"
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
}

func TestRunGenerateRejectsInvalidUsage(t *testing.T) {
	var stderr bytes.Buffer
	code := runGenerate([]string{"billing.gooo"}, fixtureReader{}, SyntaxSourceParser{}, &bytes.Buffer{}, &stderr)
	if code != exitUsage || stderr.String() != "usage: gooo generate <file.gooo> --out <directory>\n" {
		t.Fatalf("usage = code %d, stderr %q", code, stderr.String())
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
