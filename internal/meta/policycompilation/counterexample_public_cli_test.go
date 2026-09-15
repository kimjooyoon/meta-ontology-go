package policycompilation

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
)

func buildCounterexamplePublicCLI(t *testing.T, ctx context.Context, work string) string {
	t.Helper()
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(work, "gooo")
	command := exec.CommandContext(ctx, "go", "build", "-o", binary, "./cmd/gooo")
	command.Dir = root
	command.Env = append(os.Environ(), "GOTOOLCHAIN=go1.27.0")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build public Gooo CLI once: %v: %s", err, output)
	}
	return binary
}

func runCounterexamplePublicCLI(t *testing.T, ctx context.Context, binary, work string, args ...string) []byte {
	t.Helper()
	command := exec.CommandContext(ctx, binary, args...)
	command.Dir = work
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := command.Run(); err != nil {
		t.Fatalf("public Gooo command %v: %v: %s: %s", args, err, stdout.Bytes(), stderr.Bytes())
	}
	event := make(map[string]any)
	event["args"] = command.Args
	event["process_exit_code"] = command.ProcessState.ExitCode()
	event["stdout_digest"] = DigestBytes(stdout.Bytes())
	event["stderr_digest"] = DigestBytes(stderr.Bytes())
	t.Logf("COUNTEREXAMPLE_PUBLIC_COMMAND=%s", counterexampleExecutionJSON(t, event))
	return append([]byte(nil), stdout.Bytes()...)
}

func generateCounterexamplePublicProposal(t *testing.T, ctx context.Context, binary, work, project, policyPath string, request PolicyRevisionObservationRequest) (PublicPolicyRevisionReport, []byte) {
	t.Helper()
	arguments := func(outputRoot string) []string {
		return []string{
			"generate", policyPath, "--out", outputRoot, "--profile", PublicPolicyRevisionProfileID,
			"--profile-package", "metapolicycompilation", "--profile-namespace", "metapolicycompilation",
			"--profile-project-root", project, "--profile-source-digest", request.ExpectedSourceDigest,
			"--profile-condition", request.Condition, "--profile-from-decision", request.FromDecision,
			"--profile-to-decision", request.ToDecision, "--json",
		}
	}
	first := runCounterexamplePublicCLI(t, ctx, binary, work, arguments(filepath.Join(work, "first"))...)
	replay := runCounterexamplePublicCLI(t, ctx, binary, work, arguments(filepath.Join(work, "replay"))...)
	if !bytes.Equal(first, replay) {
		t.Fatal("public counterexample proposal JSON did not replay")
	}
	var report PublicPolicyRevisionReport
	if err := decodeStrictJSON(first, &report); err != nil {
		t.Fatal(err)
	}
	if report.Profile != PublicPolicyRevisionProfileID || report.Schema != PublicPolicyRevisionReportSchema ||
		!reflect.DeepEqual(report.GeneratedFiles, []string{"candidate.gooo", "proposal.json"}) ||
		report.ExecutionObserved || report.CurrentConformance != PublicGenerationConformanceUnknown ||
		report.RepositoryWrites != 0 || report.MutationAuthority != 0 || report.PromotionAuthority != 0 {
		t.Fatal("public proposal acquired a different artifact or authority contract")
	}
	for _, name := range report.GeneratedFiles {
		original := readCounterexamplePublicArtifact(t, filepath.Join(work, "first", name))
		replayed := readCounterexamplePublicArtifact(t, filepath.Join(work, "replay", name))
		if !bytes.Equal(original, replayed) || (name == "proposal.json" && !bytes.Equal(first, original)) {
			t.Fatalf("public artifact %s lost byte identity", name)
		}
	}
	for _, directory := range []string{"first", "replay"} {
		entries, err := os.ReadDir(filepath.Join(work, directory))
		if err != nil || len(entries) != 2 {
			t.Fatalf("public revision did not emit exactly two artifacts: %v", err)
		}
	}
	return report, first
}

func readCounterexamplePublicArtifact(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
