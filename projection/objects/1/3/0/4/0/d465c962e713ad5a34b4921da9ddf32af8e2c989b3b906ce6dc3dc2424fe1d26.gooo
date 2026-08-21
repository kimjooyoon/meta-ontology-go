package main

import (
	"bytes"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestRunAnalyzeSemanticInvalidIsReadOnly(t *testing.T) {
	directory := t.TempDir()
	filename := filepath.Join(directory, "input.gooo")
	if err := os.WriteFile(filename, []byte(sourceOrderA), 0o640); err != nil {
		t.Fatal(err)
	}
	beforeBytes, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	beforeInfo, err := os.Stat(filename)
	if err != nil {
		t.Fatal(err)
	}
	beforeEntries := directoryEntries(t, directory)
	var stdout, stderr bytes.Buffer
	code := runAnalyzeWithLowerer(
		[]string{filename}, OSFileReader{}, SyntaxSourceParser{}, &stdout, &stderr,
		func(*syntax.File) (semantic.IR, error) { return semantic.IR{}, semantic.ErrUnknownRelation },
	)
	if code != exitFailure || stderr.Len() != 0 || stdout.Len() == 0 {
		t.Fatalf("semantic-invalid analyze = code %d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
	}
	afterBytes, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	afterInfo, err := os.Stat(filename)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(beforeBytes, afterBytes) || !os.SameFile(beforeInfo, afterInfo) ||
		beforeInfo.Mode() != afterInfo.Mode() || !beforeInfo.ModTime().Equal(afterInfo.ModTime()) {
		t.Fatal("semantic-invalid analyze changed its input file")
	}
	if afterEntries := directoryEntries(t, directory); !reflect.DeepEqual(beforeEntries, afterEntries) {
		t.Fatalf("semantic-invalid analyze changed directory entries: before=%v after=%v", beforeEntries, afterEntries)
	}
}
