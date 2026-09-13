package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/policycompilation"
)

func TestIndependentRevisionCompositionFlagBoundary(t *testing.T) {
	base := []string{"-policy", "policy.gooo", "-observe-revision", "request.json", "-revision-operation", "operation.gooo"}
	digest := policycompilation.DigestBytes([]byte("consumer"))
	for name, args := range map[string][]string{
		"consumer-only": {"-revision-consumer", "consumer"},
		"missing-digest": append(append([]string{}, base...), "-revision-consumer", "consumer"),
		"digest-only": append(append([]string{}, base...), "-revision-consumer-digest", digest),
		"bad-digest": append(append([]string{}, base...), "-revision-consumer", "consumer", "-revision-consumer-digest", "bad"),
		"no-gooo": {"-policy", "policy.gooo", "-observe-revision", "request.json", "-revision-consumer", "consumer", "-revision-consumer-digest", digest},
		"valid": append(append([]string{}, base...), "-revision-consumer", "consumer", "-revision-consumer-digest", digest),
	} {
		t.Run(name, func(t *testing.T) {
			flags := flag.NewFlagSet(name, flag.ContinueOnError)
			flags.SetOutput(io.Discard)
			for _, key := range []string{"policy", "observe-revision", "revision-operation", "revision-consumer", "revision-consumer-digest"} {
				flags.String(key, "", "")
			}
			if err := flags.Parse(args); err != nil {
				t.Fatal(err)
			}
			requested, err := revisionObservationMode(flags)
			if !requested || (err == nil) != (name == "valid") {
				t.Fatalf("mode=%v error=%v", requested, err)
			}
		})
	}
}

func TestIndependentRevisionCompositionPublicCommand(t *testing.T) {
	work := t.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()
	binaries := []string{}
	for index, pkg := range []string{".", "../meta-policy-compilation-consumer"} {
		name := filepath.Join(work, []string{"witness", "consumer"}[index])
		if runtime.GOOS == "windows" {
			name += ".exe"
		}
		command := exec.CommandContext(ctx, "go", "build", "-trimpath", "-o", name, pkg)
		command.Env = append(os.Environ(), "GOTOOLCHAIN=go1.27.0")
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("build %s: %v\n%s", pkg, err, output)
		}
		binaries = append(binaries, name)
	}
	source, err := os.ReadFile("../../examples/meta-policy-compilation/policy.gooo")
	if err != nil {
		t.Fatal(err)
	}
	request := independentRevisionRequest(t, source)
	requestBytes, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	requestBytes = append([]byte(" \n"), requestBytes...)
	inputs := map[string][]byte{
		"policy.gooo": source, "request.json": requestBytes,
		"operation.gooo": policycompilation.PolicyRevisionOperationContract(),
	}
	for name, data := range inputs {
		if err := os.WriteFile(filepath.Join(work, name), data, 0400); err != nil {
			t.Fatal(err)
		}
	}
	consumerBytes, err := os.ReadFile(binaries[1])
	if err != nil {
		t.Fatal(err)
	}
	digest := policycompilation.DigestBytes(consumerBytes)
	invoke := func(path, expected string) (revisionIndependentObservation, error) {
		command := exec.CommandContext(ctx, binaries[0], "-policy", filepath.Join(work, "policy.gooo"),
			"-observe-revision", filepath.Join(work, "request.json"), "-revision-operation", filepath.Join(work, "operation.gooo"),
			"-revision-consumer", path, "-revision-consumer-digest", expected)
		var stdout, stderr bytes.Buffer
		command.Stdout, command.Stderr = &stdout, &stderr
		runError := command.Run()
		var report revisionIndependentObservation
		if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
			t.Fatalf("composition omitted its bounded report: %v\n%s\n%s", err, stdout.Bytes(), stderr.Bytes())
		}
		return report, runError
	}
	report, err := invoke(binaries[1], digest)
	if err != nil || report.Decision != "INDEPENDENT_RECONSTRUCTION_OBSERVED" || report.Pending != nil ||
		report.Operation == nil || report.Operation.Observation == nil || !report.Consumer.Started || report.Consumer.ExitCode != 0 {
		t.Fatalf("independent composition did not close: %+v error=%v", report, err)
	}
	operation := report.Operation
	if operation.Binding.Program != policycompilation.PolicyRevisionOperationProgram ||
		operation.Observation.Admission.State != "UNKNOWN" || report.Improvement != "UNKNOWN" ||
		report.MutationAuthority != 0 || report.PromotionAuthority != 0 ||
		report.Consumer.ExpectedExecutableDigest != digest || report.Consumer.ObservedExecutableDigest != digest ||
		report.Consumer.StdoutDigest != policycompilation.DigestBytes([]byte(report.Consumer.Stdout)) {
		t.Fatal("composition lost its source/consumer binding or silently granted authority")
	}
	for name, before := range inputs {
		after, err := os.ReadFile(filepath.Join(work, name))
		if err != nil || !bytes.Equal(before, after) {
			t.Fatal("composition changed a caller input")
		}
	}
	for name, path := range map[string]string{"wrong-pin": binaries[1], "missing": filepath.Join(work, "absent")} {
		t.Run(name, func(t *testing.T) {
			failed, err := invoke(path, policycompilation.DigestBytes([]byte("different executable")))
			if err == nil || failed.Decision != "UNKNOWN" || failed.Pending == nil ||
				failed.Pending.Stage != "EXECUTABLE_BINDING" || failed.Consumer.Started || failed.Operation != nil {
				t.Fatal("unbound executable acquired execution evidence")
			}
		})
	}
	payload, err := json.Marshal(operation.Observation)
	if err != nil {
		t.Fatal(err)
	}
	testIndependentConsumerCounterexamples(t, []byte(report.Consumer.Stdout), *operation, payload)
	witness, err := json.Marshal(map[string]any{
		"decision": report.Decision, "gooo_binding": operation.Binding, "counts": operation.Observation.Counts,
		"consumer_executable_digest": digest, "consumer_stdout_digest": report.Consumer.StdoutDigest,
		"consumer_process_started": report.Consumer.Started, "consumer_exit_code": report.Consumer.ExitCode,
		"consumer_wall_ms": report.Consumer.WallMilliseconds, "admission": operation.Observation.Admission,
		"input_provenance": operation.Observation.InputProvenance, "improvement": report.Improvement,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("INDEPENDENT_REVISION_COMPOSITION_WITNESS=%s", witness)
}

func independentRevisionRequest(t *testing.T, source []byte) policycompilation.PolicyRevisionObservationRequest {
	t.Helper()
	revision := policycompilation.PolicyDecisionRevision{
		ExpectedSourceDigest: policycompilation.DigestBytes(source), Condition: policycompilation.ConditionSemanticEquivalence,
		FromDecision: policycompilation.DecisionPass, ToDecision: policycompilation.DecisionFailClosed,
	}
	proposal, err := policycompilation.ProposePolicyDecisionRevision("policy.gooo", source, "metapolicycompilation", "metapolicycompilation", revision)
	if err != nil {
		t.Fatal(err)
	}
	input := func(id string, policy policycompilation.CompiledPolicy) policycompilation.Case {
		return policycompilation.Case{
			ID: id, ValidatorExpectation: "PASS", EvidenceClass: "SYNTHETIC_FIXTURE", Provenance: "independent composition fixture",
			ProducerAvailable: true, ConsumerAvailable: true, ObservedSourceDigest: policy.SourceDigest,
			ObservedArtifactSourceDigest: policy.SourceDigest, ObservedIndependentDigest: policy.SemanticDigest,
			ObservedGeneratedJudgeDigest: policycompilation.DigestBytes(policycompilation.GenerateJudge(policy)),
		}
	}
	before, after := input("fresh", proposal.Original), input("fresh", proposal.Candidate)
	after.ValidatorExpectation = "FAIL_CLOSED"
	stale := input("stale", proposal.Original)
	missing := policycompilation.Case{ID: "missing", EvidenceClass: "SYNTHETIC_FIXTURE", Provenance: "explicit absent evidence"}
	return policycompilation.PolicyRevisionObservationRequest{
		ExpectedSourceDigest: revision.ExpectedSourceDigest, Condition: revision.Condition,
		FromDecision: revision.FromDecision, ToDecision: revision.ToDecision,
		Cases: []policycompilation.PolicyRevisionCasePair{
			{Baseline: before, Candidate: after}, {Baseline: stale, Candidate: stale}, {Baseline: missing, Candidate: missing},
		},
	}
}

func testIndependentConsumerCounterexamples(t *testing.T, original []byte,
	operation policycompilation.PolicyRevisionOperationObservation, report []byte) {
	t.Helper()
	for _, field := range []string{"schema", "source_digest", "request_artifact_digest", "canonical_request_digest",
		"report_artifact_digest", "decision", "checks", "independent_result_comparisons", "policy_execution_observed",
		"improvement", "mutation_authority", "promotion_authority"} {
		t.Run("missing-"+field, func(t *testing.T) {
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(original, &fields); err != nil {
				t.Fatal(err)
			}
			delete(fields, field)
			data, err := json.Marshal(fields)
			if err != nil {
				t.Fatal(err)
			}
			if decision, err := checkRevisionConsumerReport(data, operation, report); err == nil || decision != "UNKNOWN" {
				t.Fatal("missing consumer field became success")
			}
		})
	}
	for name, data := range map[string][]byte{
		"null": []byte("null"), "trailing": append(bytes.Clone(original), []byte("{}")...),
		"unknown-decision": bytes.Replace(original, []byte("RECEIPT_CONSISTENT_ONLY"), []byte("FIXED_POINT"), 1),
		"wrong-source": bytes.Replace(original, []byte(operation.PolicySourceDigest), []byte(policycompilation.DigestBytes([]byte("wrong"))), 1),
		"duplicate": []byte(strings.Replace(string(original), "\"schema\":", "\"schema\":\"forged\",\"schema\":", 1)),
	} {
		t.Run(name, func(t *testing.T) {
			if decision, err := checkRevisionConsumerReport(data, operation, report); err == nil || decision != "UNKNOWN" {
				t.Fatal("invalid consumer output became success")
			}
		})
	}
}
