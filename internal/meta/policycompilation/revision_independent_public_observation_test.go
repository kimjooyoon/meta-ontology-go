package policycompilation

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// Observe the installed command boundary from the already collected library
// test package. The read-only consumer and CI authority remain unchanged.
func TestIndependentRevisionCompositionPublicObservation(t *testing.T) {
	source, request := revisionObservationFixture(t)
	directory := t.TempDir()
	_, location, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate the source checkout")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(location), "../../.."))
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	witness := buildIndependentPublicCommand(t, ctx, root, directory, "meta-policy-compilation-witness")
	consumer := buildIndependentPublicCommand(t, ctx, root, directory, "meta-policy-compilation-consumer")
	requestBytes, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	requestBytes = append([]byte(" \n\t"), requestBytes...)
	inputs := map[string][]byte{
		"policy.gooo": source, "request.json": requestBytes, "operation.gooo": PolicyRevisionOperationContract(),
	}
	for name, data := range inputs {
		if err := os.WriteFile(filepath.Join(directory, name), data, 0400); err != nil {
			t.Fatal(err)
		}
	}
	consumerBytes, err := os.ReadFile(consumer)
	if err != nil {
		t.Fatal(err)
	}
	consumerDigest := DigestBytes(consumerBytes)
	command := exec.CommandContext(ctx, witness,
		"-policy", filepath.Join(directory, "policy.gooo"),
		"-observe-revision", filepath.Join(directory, "request.json"),
		"-revision-operation", filepath.Join(directory, "operation.gooo"),
		"-revision-consumer", consumer, "-revision-consumer-digest", consumerDigest,
		"-profile-package", "metapolicycompilation", "-profile-namespace", "metapolicycompilation")
	command.Env = append(os.Environ(), "GOTOOLCHAIN=go1.27.0")
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	start := time.Now()
	runError := command.Run()
	wallMilliseconds := time.Since(start).Milliseconds()
	if runError != nil {
		t.Fatalf("public independent composition: %v\n%s\n%s", runError, stdout.Bytes(), stderr.Bytes())
	}
	envelope := independentPublicField[map[string]json.RawMessage](t, stdout.Bytes())
	if independentPublicField[string](t, envelope["schema"]) != "gooo/meta-policy-revision-independent-observation/v1" ||
		independentPublicField[string](t, envelope["decision"]) != "INDEPENDENT_RECONSTRUCTION_OBSERVED" {
		t.Fatal("public command did not produce its versioned observation")
	}
	operation := independentPublicField[PolicyRevisionOperationObservation](t, envelope["operation"])
	if operation.Observation == nil || operation.NativeWorkerInvocations != 1 ||
		operation.PolicySourceDigest != DigestBytes(source) || operation.RequestArtifactDigest != DigestBytes(requestBytes) ||
		operation.Binding.Program != PolicyRevisionOperationProgram || !operation.Binding.UsedPolicySource ||
		!operation.Binding.UsedRevisionRequest || !operation.Binding.GeneratedObservation {
		t.Fatal("public observation lost its actual Gooo source/request/native binding")
	}
	process := independentPublicField[map[string]json.RawMessage](t, envelope["consumer_process"])
	consumerStdout := independentPublicField[string](t, process["stdout"])
	if !independentPublicField[bool](t, process["started"]) || independentPublicField[int](t, process["exit_code"]) != 0 ||
		independentPublicField[string](t, process["expected_executable_digest"]) != consumerDigest ||
		independentPublicField[string](t, process["observed_executable_digest"]) != consumerDigest ||
		independentPublicField[string](t, process["stdout_digest"]) != DigestBytes([]byte(consumerStdout)) {
		t.Fatal("consumer process evidence does not bind the actual executable and output")
	}
	independent := independentPublicField[map[string]json.RawMessage](t, []byte(consumerStdout))
	comparisons := independentPublicField[int](t, independent["independent_result_comparisons"])
	if comparisons != 6*len(request.Cases) || independentPublicField[bool](t, independent["policy_execution_observed"]) {
		t.Fatal("reconstruction counts or historical process boundary changed")
	}
	counts := operation.Observation.Counts
	if counts.RequestedCasePairs != len(request.Cases) || counts.ObservedCasePairs != len(request.Cases) ||
		counts.SourceComparisons != 4*len(request.Cases) || counts.ReplayComparisons != 2*len(request.Cases) ||
		counts.SourceMismatches != 0 || counts.ReplayMismatches != 0 || counts.FailedBatches != 0 {
		t.Fatal("public observation did not retain the exact supplied case cohort")
	}
	admission := operation.Observation.Admission
	if admission.State != "UNKNOWN" || admission.Reason != "INDEPENDENT_REVISION_EVIDENCE_MISSING" ||
		admission.Stage == "" || admission.Step == "" || admission.UnknownClass == "" ||
		admission.NextOperation == "" || admission.BlockedBy == nil ||
		independentPublicField[string](t, envelope["improvement"]) != "UNKNOWN" ||
		independentPublicField[int](t, envelope["mutation_authority"]) != 0 ||
		independentPublicField[int](t, envelope["promotion_authority"]) != 0 {
		t.Fatal("command composition silently granted unobserved admission or authority")
	}
	for name, original := range inputs {
		actual, err := os.ReadFile(filepath.Join(directory, name))
		if err != nil || !bytes.Equal(original, actual) {
			t.Fatal("public command modified a supplied input")
		}
	}
	event, err := json.Marshal(map[string]any{
		"decision": "INDEPENDENT_RECONSTRUCTION_OBSERVED", "gooo_binding": operation.Binding,
		"counts": counts, "independent_result_comparisons": comparisons,
		"public_command_exit_code": command.ProcessState.ExitCode(), "public_command_wall_ms": wallMilliseconds,
		"consumer_executable_digest": consumerDigest, "consumer_stdout_digest": DigestBytes([]byte(consumerStdout)),
		"public_stdout_digest": DigestBytes(stdout.Bytes()), "public_stdout": stdout.String(),
		"admission": admission, "evidence_class": "SYNTHETIC_FIXTURE", "improvement": "UNKNOWN",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("INDEPENDENT_REVISION_PUBLIC_OBSERVATION_WITNESS=%s", event)
}

func buildIndependentPublicCommand(t *testing.T, ctx context.Context, root, directory, name string) string {
	t.Helper()
	binary := filepath.Join(directory, name)
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	command := exec.CommandContext(ctx, "go", "build", "-trimpath", "-o", binary, "./cmd/"+name)
	command.Dir = root
	command.Env = append(os.Environ(), "GOTOOLCHAIN=go1.27.0")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build public %s command: %v\n%s", name, err, output)
	}
	return binary
}

func independentPublicField[T any](t *testing.T, data json.RawMessage) T {
	t.Helper()
	var value T
	if len(data) == 0 || bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		t.Fatal("public observation omitted a required value")
	}
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatal(err)
	}
	return value
}
