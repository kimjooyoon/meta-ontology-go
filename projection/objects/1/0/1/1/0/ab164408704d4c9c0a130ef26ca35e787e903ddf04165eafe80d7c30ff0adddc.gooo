package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestRunInspectG1SourcePermutationPreservesSemanticProjection(t *testing.T) {
	first := decodeGraphDump(t, inspectFixtureOutput(t, sourceOrderA))
	second := decodeGraphDump(t, inspectFixtureOutput(t, sourceOrderB))
	if first.SourceDigest == second.SourceDigest {
		t.Fatal("source digest ignored a source-order change")
	}
	if first.GraphHash != second.GraphHash || first.IR.SemanticDigest != second.IR.SemanticDigest {
		t.Fatalf("semantic digests changed with declaration order: first=%#v second=%#v", first, second)
	}
	first.SourceDigest, second.SourceDigest = "", ""
	firstJSON, err := json.Marshal(first)
	if err != nil {
		t.Fatal(err)
	}
	secondJSON, err := json.Marshal(second)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(firstJSON, secondJSON) {
		t.Fatalf("canonical graph projection differs by source order:\nfirst=%s\nsecond=%s", firstJSON, secondJSON)
	}
}
func TestRunInspectIsReadOnly(t *testing.T) {
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
	if code := runInspect([]string{filename}, OSFileReader{}, SyntaxSourceParser{}, &stdout, &stderr); code != exitOK {
		t.Fatalf("inspect code = %d, stderr=%q", code, stderr.String())
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
		t.Fatal("inspect changed its input file")
	}
	if afterEntries := directoryEntries(t, directory); !reflect.DeepEqual(beforeEntries, afterEntries) {
		t.Fatalf("inspect changed directory entries: before=%v after=%v", beforeEntries, afterEntries)
	}
}
