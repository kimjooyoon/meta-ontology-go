package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func revisionReceiptNativeWitness(t *testing.T, work string) string {
	t.Helper()
	binary := filepath.Join(work, "revision-witness")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	build := exec.CommandContext(ctx, "go", "build", "-trimpath", "-o", binary, "../meta-policy-compilation-witness")
	build.Env = append(os.Environ(), "GOTOOLCHAIN=go1.27.0")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build separate native revision witness: %v\n%s", err, output)
	}
	return binary
}

func revisionReceiptRunFixture(t *testing.T, binary, work, phase string, requestBytes []byte, missingToolchain bool) revisionWireObservation {
	t.Helper()
	requestPath := filepath.Join(work, phase+"-request.json")
	if err := os.WriteFile(requestPath, requestBytes, 0400); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, binary, "-policy", filepath.Join(work, "policy.gooo"), "-observe-revision", requestPath)
	command.Env = append(os.Environ(), "GOTOOLCHAIN=go1.27.0")
	if missingToolchain {
		environment := make([]string, 0, len(command.Env)+1)
		for _, entry := range command.Env {
			if !strings.HasPrefix(strings.ToUpper(entry), "PATH=") {
				environment = append(environment, entry)
			}
		}
		command.Env = append(environment, "PATH="+filepath.Join(work, "missing-toolchain"))
	}
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	runError := command.Run()
	if (runError != nil) != missingToolchain {
		t.Fatalf("native fixture %s completion differs: %v\n%s\n%s", phase, runError, stdout.Bytes(), stderr.Bytes())
	}
	var report revisionWireObservation
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("native fixture %s omitted its receipt: %v\n%s", phase, err, stdout.Bytes())
	}
	if report.RequestArtifactDigest != digestBytes(requestBytes) {
		t.Fatal("native fixture did not retain the supplied raw request")
	}
	t.Logf("REVISION_RECEIPT_FIXTURE_PROCESS=%s", revisionReceiptJSON(t, map[string]any{
		"phase": phase, "process_exit_code": command.ProcessState.ExitCode(),
		"request_digest": digestBytes(requestBytes), "stdout_digest": digestBytes(stdout.Bytes()),
		"raw_stdout": stdout.String(), "raw_stderr": stderr.String(),
	}))
	return report
}

func revisionReceiptFixture(t *testing.T) ([]byte, []byte, revisionWireObservation, func() revisionWireObservation) {
	t.Helper()
	source := sourceObservationFixture(t)
	work := t.TempDir()
	if err := os.WriteFile(filepath.Join(work, "policy.gooo"), source, 0400); err != nil {
		t.Fatal(err)
	}
	binary := revisionReceiptNativeWitness(t, work)
	missing := revisionWireCase{ID: "missing", EvidenceClass: "SYNTHETIC_FIXTURE", Provenance: "explicit absent input"}
	request := revisionWireRequest{
		ExpectedSourceDigest: digestBytes(source), Condition: "SEMANTIC_EQUIVALENCE",
		FromDecision: "PASS", ToDecision: "FAIL_CLOSED",
		Cases: []revisionWirePair{{Baseline: missing, Candidate: missing}},
	}
	bootstrap := revisionReceiptRunFixture(t, binary, work, "bootstrap", revisionReceiptJSON(t, request), false)
	input := func(id string, policy revisionWirePolicy, execution revisionWireExecution) revisionWireCase {
		return revisionWireCase{
			ID: id, ValidatorExpectation: "PASS", EvidenceClass: "SYNTHETIC_FIXTURE", Provenance: "native receipt observer fixture",
			ProducerAvailable: true, ConsumerAvailable: true, ObservedSourceDigest: policy.SourceDigest,
			ObservedArtifactSourceDigest: policy.SourceDigest, ObservedIndependentDigest: policy.SemanticDigest,
			ObservedGeneratedJudgeDigest: digestBytes([]byte(execution.GeneratedJudgeSource)),
		}
	}
	freshBefore := input("fresh", bootstrap.OriginalPolicy, bootstrap.Baseline)
	freshAfter := input("fresh", bootstrap.CandidatePolicy, bootstrap.Candidate)
	freshAfter.ValidatorExpectation = "FAIL_CLOSED"
	stale := input("stale", bootstrap.OriginalPolicy, bootstrap.Baseline)
	request.Cases = []revisionWirePair{
		{Baseline: freshBefore, Candidate: freshAfter},
		{Baseline: stale, Candidate: stale},
		{Baseline: missing, Candidate: missing},
	}
	requestBytes, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	requestBytes = append([]byte(" \n\t"), requestBytes...)
	requestBytes = append(requestBytes, '\n')
	report := revisionReceiptRunFixture(t, binary, work, "complete", requestBytes, false)
	failedAttempt := func() revisionWireObservation {
		return revisionReceiptRunFixture(t, binary, work, "missing-toolchain", requestBytes, true)
	}
	return source, requestBytes, report, failedAttempt
}
