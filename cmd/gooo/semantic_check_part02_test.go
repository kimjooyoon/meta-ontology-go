package main

import (
	"bytes"
	"errors"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestSemanticCheckPropagatesInvalidRelationDeterministically(t *testing.T) {
	lower := func(*syntax.File) (semantic.IR, error) { return semantic.IR{}, semantic.ErrUnknownRelation }
	first, second := bytes.Buffer{}, bytes.Buffer{}
	_, firstErr := semanticCheckIRWithLowerer(&syntax.File{}, time.Second, lower)
	_, secondErr := semanticCheckIRWithLowerer(&syntax.File{}, time.Second, lower)
	if !errors.Is(firstErr, semantic.ErrUnknownRelation) || !errors.Is(secondErr, semantic.ErrUnknownRelation) {
		t.Fatalf("relation errors = %v/%v", firstErr, secondErr)
	}
	if !reportSemanticDiagnostic("fixture.gooo", &syntax.File{}, firstErr, &first) ||
		!reportSemanticDiagnostic("fixture.gooo", &syntax.File{}, secondErr, &second) || first.String() != second.String() {
		t.Fatalf("relation diagnostics are not deterministic: %q/%q", first.String(), second.String())
	}
	if !bytes.Contains(first.Bytes(), []byte("semantic.invalid-relation")) {
		t.Fatalf("relation diagnostic = %q", first.String())
	}
}
func TestSemanticCheckUsesVersionedReadOnlyBoundary(t *testing.T) {
	if semanticCheckSchemaVersion != "gooo-semantic-check/v1" {
		t.Fatalf("semantic check schema = %q", semanticCheckSchemaVersion)
	}
	directory := t.TempDir()
	filename := filepath.Join(directory, "valid.gooo")
	if err := os.WriteFile(filename, []byte(validSource), 0o640); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	beforeInfo, err := os.Stat(filename)
	if err != nil {
		t.Fatal(err)
	}
	beforeEntries := directoryEntries(t, directory)
	var stdout, stderr bytes.Buffer
	code := runCheck([]string{"--semantic", filename}, OSFileReader{}, SyntaxSourceParser{}, &stdout, &stderr)
	if code != exitOK || stdout.String() != "ok: "+filename+"\n" || stderr.String() != deferredCheckProvenance+"\n" {
		t.Fatalf("semantic check = code %d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
	}
	after, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	afterInfo, err := os.Stat(filename)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) || !os.SameFile(beforeInfo, afterInfo) || beforeInfo.Mode() != afterInfo.Mode() || !beforeInfo.ModTime().Equal(afterInfo.ModTime()) {
		t.Fatal("semantic check changed valid input")
	}
	if afterEntries := directoryEntries(t, directory); !reflect.DeepEqual(beforeEntries, afterEntries) {
		t.Fatalf("semantic check changed directory entries: before=%v after=%v", beforeEntries, afterEntries)
	}
}
