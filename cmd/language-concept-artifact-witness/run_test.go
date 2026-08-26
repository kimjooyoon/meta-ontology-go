package main

import (
	"bytes"
	"io"
	"path/filepath"
	"testing"
)

func TestRunProducesAndConsumesArtifact(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(t.TempDir(), "language-concept-artifact.json")
	produced := bytes.Buffer{}
	if err := run(config{root: root, output: output}, &produced); err != nil {
		t.Fatal(err)
	}
	consumed := bytes.Buffer{}
	if err := run(config{root: root, check: output}, &consumed); err != nil {
		t.Fatal(err)
	}
	if produced.String() != consumed.String() {
		t.Fatalf("producer and consumer summaries diverged: %q != %q", produced.String(), consumed.String())
	}
}

func TestRunRejectsRepositoryOutput(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	err = run(config{root: root, output: filepath.Join(root, "artifact.json")}, io.Discard)
	if err == nil {
		t.Fatal("repository output accepted")
	}
}
