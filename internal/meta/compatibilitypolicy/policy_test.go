package compatibilitypolicy

import (
	"bytes"
	"os"
	"regexp"
	"strings"
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

func TestNewGoooToolchainContractCanRegenerateEvaluator(t *testing.T) {
	const sourcePath = "examples/self-improvement-observation/observation.gooo"
	source, err := os.ReadFile("../../../" + sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	const replacement = "go1.99.1|go test ./...|go test -race ./..."
	changed := strings.Replace(string(source), TestContract, replacement, 1)
	if changed == string(source) {
		t.Fatal("Gooo test contract was not found")
	}
	policy, generated, err := GenerateNamed(sourcePath, []byte(changed))
	if err != nil || policy.TestContract != replacement || !regexp.MustCompile(`(?m)^\s*TestContract\s*=\s*"`+regexp.QuoteMeta(replacement)+`"$`).Match(generated) {
		t.Fatalf("new source contract did not project into Go: %v", err)
	}
	if _, err := Load(sourcePath, []byte(changed)); err == nil {
		t.Fatal("new source bypassed committed evaluator regeneration")
	}
}
