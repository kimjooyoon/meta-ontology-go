package main

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

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
