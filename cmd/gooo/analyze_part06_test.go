package main

import (
	"bytes"
	"encoding/json"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"testing"
)

func analyzeFixtureOutput(t *testing.T, source string) ([]byte, int, []byte) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := runAnalyze([]string{"fixture.gooo"}, fixtureReader{source: source}, SyntaxSourceParser{}, &stdout, &stderr)
	return stdout.Bytes(), code, stderr.Bytes()
}
func analyzeFixtureBytes(t *testing.T, source string) []byte {
	t.Helper()
	output, code, stderr := analyzeFixtureOutput(t, source)
	if code != exitOK || len(stderr) != 0 {
		t.Fatalf("analyze fixture = code %d, stderr=%q", code, stderr)
	}
	return output
}
func decodeFixPlan(t *testing.T, output []byte) fixPlan {
	t.Helper()
	var plan fixPlan
	if err := json.Unmarshal(output, &plan); err != nil {
		t.Fatalf("decode fix plan: %v; output=%s", err, output)
	}
	return plan
}
func runAnalyzeFile(filename string) ([]byte, int, []byte) {
	var stdout, stderr bytes.Buffer
	code := runAnalyze([]string{filename}, OSFileReader{}, SyntaxSourceParser{}, &stdout, &stderr)
	return stdout.Bytes(), code, stderr.Bytes()
}
func mustAnalyzeEvidence(t *testing.T, id string, fact semantic.FactKey) semantic.Evidence {
	t.Helper()
	evidence, err := semantic.NewEvidence(
		semantic.MustIdentity(id), semantic.GoHostedCompilerID, semantic.CompilerRunEvidence,
		fact, semantic.StableHash([]byte(id)),
	)
	if err != nil {
		t.Fatal(err)
	}
	return evidence
}
