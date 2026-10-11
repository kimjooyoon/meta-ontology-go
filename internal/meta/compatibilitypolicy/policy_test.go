package compatibilitypolicy

import (
	"bytes"
	"os"
	"testing"
)

func TestCommittedEvaluatorMatchesCanonicalGoooSource(t *testing.T) {
	const sourcePath = "examples/self-improvement-observation/observation.gooo"
	source, err := os.ReadFile("../../../" + sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	_, generated, err := GenerateNamed(sourcePath, source)
	if err != nil {
		t.Fatal(err)
	}
	committed, err := os.ReadFile("generated/evaluator.go")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(generated, committed) {
		t.Fatal("stale compatibility evaluator: from repository root run go run ./scripts/self-improvement-compiler-compatibility -mode generate -contract " + sourcePath + " -output internal/meta/compatibilitypolicy/generated/evaluator.go")
	}
	if _, err := Load(sourcePath, source); err != nil {
		t.Fatal("canonical source must load through its generated evaluator:", err)
	}
}
