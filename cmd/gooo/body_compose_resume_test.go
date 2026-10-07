package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"
)

func runComposeResumeFixture(t *testing.T, options ...string) bodyCompositionOutput {
	t.Helper()
	args := []string{"body-compose", "--source", "../../examples/package-diagnostic-replay/diagnostics.gooo.fixture",
		"--cases", "../../examples/assembly-policy/cases.json"}
	var out, diagnostics bytes.Buffer
	if code := run(append(args, options...), &out, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String())
	}
	var result bodyCompositionOutput
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	return result
}

func TestBodyComposeResumeSavesRunsAndReplays(t *testing.T) {
	first := filepath.Join(t.TempDir(), "first")
	prior := runComposeResumeFixture(t, "--assembly-policy", "../../examples/assembly-policy/checkpoint.gooo.fixture",
		"--policy-activity", "Checkpoint", "--out", first)
	if r := prior.Composition.Steps[0].Generation.Report.RecordAssembly; len(r.Attempts) != 1 || r.Passed == r.Total {
		t.Fatal("checkpoint did not retain a partial first attempt", r)
	}
	continued := filepath.Join(t.TempDir(), "continued")
	result := runComposeResumeFixture(t, "--resume-composition", filepath.Join(first, "composition.json"),
		"--assembly-policy", "../../examples/assembly-explainer/main.gooo.fixture", "--policy-activity", "Explain", "--out", continued)
	r := result.Composition.Steps[0].Generation.Report.RecordAssembly
	if !result.GeneratedNow || result.Composition.Model != nil || len(r.Attempts) != 4 ||
		r.Continuation.RetainedAttempts != 1 || r.Continuation.AddedAttempts != 3 || r.Continuation.NewModelCalls != 0 ||
		result.Runtime.FinitePassed != 4 || result.Runtime.ModelCalls != 0 {
		t.Fatal("continuation did not reach the expected native observations", r, result.Runtime)
	}
	replayed := runComposeResumeFixture(t, "--composition", filepath.Join(continued, "composition.json"))
	if replayed.GeneratedNow || replayed.Runtime.FinitePassed != 4 || replayed.Runtime.ModelCalls != 0 ||
		replayed.Composition.GeneratedSHA256 != result.Composition.GeneratedSHA256 {
		t.Fatal("continued program did not replay", replayed.Runtime)
	}
}

func TestBodyComposeResumeRequiresPolicyAndNoModel(t *testing.T) {
	base := []string{"body-compose", "--source", "source", "--cases", "cases", "--resume-composition", "saved"}
	policy := []string{"--assembly-policy", "policy", "--policy-activity", "Explain"}
	for _, extra := range [][]string{nil, append(policy[:4:4], "--model", "model"),
		append(policy[:4:4], "--fill-model", "model"), append(policy[:4:4], "--composition", "saved")} {
		var out, diagnostics bytes.Buffer
		if code := run(append(append([]string(nil), base...), extra...), &out, &diagnostics); code != exitUsage {
			t.Fatal("invalid continuation arguments", extra, code, diagnostics.String())
		}
	}
}
