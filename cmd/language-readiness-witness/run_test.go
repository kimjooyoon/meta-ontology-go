package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	readinessartifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/improvement"
)

const testSHA = "0000000000000000000000000000000000000000"

func TestRunPublishesExactEightOfTwentyFour(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	input := conceptInput(t, root)
	output := filepath.Join(t.TempDir(), "language-readiness-artifact.json")
	produced := bytes.Buffer{}
	if err := run(config{root: root, input: input, output: output, expectedSHA: testSHA}, &produced); err != nil {
		t.Fatal(err)
	}
	consumed := bytes.Buffer{}
	if err := run(config{root: root, input: input, check: output, expectedSHA: testSHA}, &consumed); err != nil {
		t.Fatal(err)
	}
	if produced.String() != consumed.String() {
		t.Fatalf("producer and consumer summaries diverged: %q != %q", produced.String(), consumed.String())
	}
	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	receipt := readinessartifact.Receipt{}
	if err := json.Unmarshal(data, &receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.Snapshot.Summary.Completed != 8 || receipt.Snapshot.Summary.Total != 24 ||
		receipt.Snapshot.Summary.ReadinessBPS != 3333 || receipt.FixedPoint.Decision != improvement.NoChange {
		t.Fatalf("receipt = %+v", receipt)
	}
}

func TestRunRejectsRepositoryOutput(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	err = run(config{
		root: root, input: conceptInput(t, root),
		output: filepath.Join(root, "readiness.json"), expectedSHA: testSHA,
	}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("repository output accepted")
	}
}
