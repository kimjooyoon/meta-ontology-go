package policycompilation

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestCounterexampleRevisionExampleRunsWithoutRepositoryWrites(t *testing.T) {
	source, input, _ := counterexampleRevisionFixture(t)
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root = filepath.Clean(filepath.Join(root, "..", "..", ".."))
	temp := t.TempDir()
	binary := filepath.Join(temp, "counterexample-proposal")
	sourcePath, inputPath := filepath.Join(temp, "policy.gooo"), filepath.Join(temp, "counterexample.json")
	raw, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sourcePath, source, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(inputPath, raw, 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	build := exec.CommandContext(ctx, "go", "build", "-o", binary, "./examples/meta-policy-compilation/counterexample-proposal")
	build.Dir = root
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build public example: %v\n%s", err, output)
	}
	run := func() []byte {
		t.Helper()
		command := exec.CommandContext(ctx, binary, "-policy", sourcePath, "-counterexample", inputPath)
		command.Dir = temp
		output, err := command.Output()
		if err != nil {
			t.Fatalf("execute public example: %v\n%s", err, output)
		}
		return output
	}
	first, replay := run(), run()
	var report PolicyCounterexampleProposal
	if err := json.Unmarshal(first, &report); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, replay) || report.State != "PROPOSED" ||
		report.CounterexampleArtifactDigest != DigestBytes(raw) || report.Revision == nil ||
		report.InputProvenance != "CALLER_DECLARED_NOT_VERIFIED" || report.Admission.State != "UNKNOWN" {
		t.Fatal("public example did not preserve its exact source/request boundary")
	}
	afterSource, sourceErr := os.ReadFile(sourcePath)
	afterInput, inputErr := os.ReadFile(inputPath)
	if sourceErr != nil || inputErr != nil || !bytes.Equal(source, afterSource) || !bytes.Equal(raw, afterInput) {
		t.Fatal("public example modified its caller-owned input files")
	}
	t.Logf("counterexample proposal CLI: state=%s derived_fields=%v admission=%s improvement=%s", report.State, report.DerivedFields, report.Admission.State, report.Improvement)
}
