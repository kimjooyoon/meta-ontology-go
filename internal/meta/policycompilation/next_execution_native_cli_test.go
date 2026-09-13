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

func TestNextPolicyExecutionPublicHandoff(t *testing.T) {
	source, revision := revisionObservationFixture(t)
	directory := t.TempDir()
	_, location, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate checkout")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(location), "../../.."))
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	witness := buildIndependentPublicCommand(t, ctx, root, directory, "meta-policy-compilation-witness")
	consumer := buildIndependentPublicCommand(t, ctx, root, directory, "meta-policy-compilation-consumer")
	next := filepath.Join(directory, "next-execution")
	if runtime.GOOS == "windows" {
		next += ".exe"
	}
	build := exec.CommandContext(ctx, "go", "build", "-trimpath", "-o", next, "./cmd/meta-policy-compilation-witness/next-execution")
	build.Dir, build.Env = root, append(os.Environ(), "GOTOOLCHAIN=go1.27.0")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build next invocation: %v\n%s", err, output)
	}
	write := func(name string, data []byte) string {
		t.Helper()
		path := filepath.Join(directory, name)
		if err := os.WriteFile(path, data, 0400); err != nil {
			t.Fatal(err)
		}
		return path
	}
	invoke := func(binary string, arguments ...string) ([]byte, int) {
		t.Helper()
		command := exec.CommandContext(ctx, binary, arguments...)
		command.Env = append(os.Environ(), "GOTOOLCHAIN=go1.27.0")
		var stdout, stderr bytes.Buffer
		command.Stdout, command.Stderr = &stdout, &stderr
		err := command.Run()
		if command.ProcessState == nil {
			t.Fatalf("start public command: %v\n%s", err, stderr.Bytes())
		}
		return stdout.Bytes(), command.ProcessState.ExitCode()
	}
	revisionBytes, err := json.Marshal(revision)
	if err != nil {
		t.Fatal(err)
	}
	consumerBytes, err := os.ReadFile(consumer)
	if err != nil {
		t.Fatal(err)
	}
	policyPath := write("original.gooo", source)
	original, originalExit := invoke(witness, "-policy", policyPath,
		"-observe-revision", write("revision.json", revisionBytes),
		"-revision-operation", write("operation.gooo", PolicyRevisionOperationContract()),
		"-revision-consumer", consumer, "-revision-consumer-digest", DigestBytes(consumerBytes))
	if originalExit != 0 {
		t.Fatalf("original public composition failed: exit=%d %s", originalExit, original)
	}
	envelope := independentPublicField[map[string]json.RawMessage](t, original)
	operation := independentPublicField[PolicyRevisionOperationObservation](t, envelope["operation"])
	if operation.Observation == nil {
		t.Fatal("missing original observation")
	}
	candidate := operation.Observation.CandidatePolicy
	candidateSource := []byte(operation.Observation.CandidateSource)
	candidatePath := write("next-policy.gooo", candidateSource)
	predecessorPath := write("predecessor.json", original)
	fresh := Case{
		ID: "next-fresh", ValidatorExpectation: DecisionFailClosed, EvidenceClass: "SYNTHETIC_FIXTURE",
		Provenance: "EXPLICIT_NEXT_INVOCATION_FIXTURE", ProducerAvailable: true, ConsumerAvailable: true,
		ObservedSourceDigest: candidate.SourceDigest, ObservedArtifactSourceDigest: candidate.SourceDigest,
		ObservedGeneratedJudgeDigest: DigestBytes(GenerateJudge(candidate)), ObservedIndependentDigest: candidate.SemanticDigest,
	}
	stale, upper := fresh, fresh
	stale.ID, stale.ObservedSourceDigest = "next-stale", DigestBytes(source)
	upper.ID, upper.UpperDecision = "next-unknown-upper", "FIXED_POINT"
	missing := Case{ID: "next-missing", ValidatorExpectation: DecisionUnknown,
		EvidenceClass: "SYNTHETIC_FIXTURE", Provenance: "EXPLICIT_MISSING_NEXT_EVIDENCE"}
	request := NextPolicyExecutionRequest{
		Schema: NextPolicyExecutionRequestSchema, ExpectedSourceDigest: candidate.SourceDigest,
		ExpectedSemanticDigest: candidate.SemanticDigest, PredecessorDigest: DigestBytes(original),
		Cases: []Case{fresh, stale, missing, upper},
	}
	raw, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	requestPath := write("next-request.json", raw)
	output, exit := invoke(next, "-policy", candidatePath, "-predecessor", predecessorPath, "-request", requestPath)
	report := independentPublicField[NextPolicyExecutionReport](t, output)
	if exit != 0 || report.Decision != "NEXT_EXECUTION_OBSERVED" || report.GeneratedBatchInvocations != 1 ||
		report.RequestedCases != 4 || report.ObservedCases != 4 || report.ExpectationComparisons != 4 ||
		report.ExpectationMismatches != 0 || report.ResultBindingMismatches != 0 ||
		report.SourceDigest != candidate.SourceDigest || report.PredecessorDigest != DigestBytes(original) ||
		report.Admission.State != "UNKNOWN" || report.MutationAuthority != 0 || report.PromotionAuthority != 0 {
		t.Fatalf("next invocation did not use the pinned candidate: exit=%d %s", exit, output)
	}
	negativeCases := []struct {
		name     string
		edit     func(*NextPolicyExecutionRequest)
		decision string
		exit     int
		batches  int
	}{
		{"changed-expectation", func(r *NextPolicyExecutionRequest) { r.Cases[0].ValidatorExpectation = DecisionPass }, "REFUTED", 1, 1},
		{"wrong-source-pin", func(r *NextPolicyExecutionRequest) { r.ExpectedSourceDigest = DigestBytes(source) }, "REFUTED", 1, 0},
		{"wrong-predecessor-pin", func(r *NextPolicyExecutionRequest) { r.PredecessorDigest = DigestBytes([]byte("other")) }, "REFUTED", 1, 0},
		{"null-request", nil, "UNKNOWN", 2, 0},
	}
	negativeResults := make(map[string]NextPolicyExecutionReport)
	for _, negative := range negativeCases {
		value := request
		value.Cases = append([]Case(nil), request.Cases...)
		encoded := []byte("null")
		if negative.edit != nil {
			negative.edit(&value)
			encoded, err = json.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
		}
		result, code := invoke(next, "-policy", candidatePath, "-predecessor", predecessorPath,
			"-request", write(negative.name+".json", encoded))
		observed := independentPublicField[NextPolicyExecutionReport](t, result)
		if code != negative.exit || observed.Decision != negative.decision || observed.GeneratedBatchInvocations != negative.batches {
			t.Fatalf("negative %s: exit=%d %s", negative.name, code, result)
		}
		if observed.Decision == "UNKNOWN" && (observed.Pending == nil || observed.Pending.BlockedBy == nil ||
			observed.Pending.Stage == "" || observed.Pending.Step == "" || observed.Pending.Reason == "" ||
			observed.Pending.UnknownClass == "" || observed.Pending.NextOperation == "") {
			t.Fatal("UNKNOWN lost its causal fields")
		}
		negativeResults[negative.name] = observed
	}
	for path, expected := range map[string][]byte{
		policyPath: source, candidatePath: candidateSource, predecessorPath: original, requestPath: raw,
	} {
		actual, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(expected, actual) {
			t.Fatal("handoff modified a supplied source, predecessor or expectation")
		}
	}
	event, err := json.Marshal(map[string]any{
		"source_operation": operation.Binding, "predecessor_digest": DigestBytes(original),
		"original_policy_digest": DigestBytes(source), "selected_source_digest": DigestBytes(candidateSource),
		"public_stdout_digest": DigestBytes(output), "public_stdout": string(output), "public_exit": exit,
		"negative_cases": negativeResults, "evidence_class": "SYNTHETIC_FIXTURE",
		"admission": "UNKNOWN", "improvement": "UNKNOWN",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("NEXT_POLICY_EXECUTION_PUBLIC_WITNESS=%s", event)
}
