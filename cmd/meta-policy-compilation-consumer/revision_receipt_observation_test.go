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

func revisionReceiptFixture(t *testing.T) ([]byte, []byte, policycompilation.PolicyRevisionObservation) {
	t.Helper()
	source := sourceObservationFixture(t)
	change := policycompilation.PolicyDecisionRevision{
		ExpectedSourceDigest: digestBytes(source), Condition: "SEMANTIC_EQUIVALENCE",
		FromDecision: "PASS", ToDecision: "FAIL_CLOSED",
	}
	proposal, err := policycompilation.ProposePolicyDecisionRevision("policy.gooo", source,
		"metapolicycompilation", "metapolicycompilation", change)
	if err != nil {
		t.Fatal(err)
	}
	input := func(id string, policy policycompilation.CompiledPolicy) policycompilation.Case {
		return policycompilation.Case{
			ID: id, ValidatorExpectation: "PASS", EvidenceClass: "SYNTHETIC_FIXTURE", Provenance: "native receipt observer fixture",
			ProducerAvailable: true, ConsumerAvailable: true, ObservedSourceDigest: policy.SourceDigest,
			ObservedArtifactSourceDigest: policy.SourceDigest, ObservedIndependentDigest: policy.SemanticDigest,
			ObservedGeneratedJudgeDigest: digestBytes(policycompilation.GenerateJudge(policy)),
		}
	}
	freshBefore, freshAfter := input("fresh", proposal.Original), input("fresh", proposal.Candidate)
	freshAfter.ValidatorExpectation = "FAIL_CLOSED"
	stale := input("stale", proposal.Original)
	missing := policycompilation.Case{ID: "missing", EvidenceClass: "SYNTHETIC_FIXTURE", Provenance: "explicit absent input"}
	request := policycompilation.PolicyRevisionObservationRequest{
		ExpectedSourceDigest: change.ExpectedSourceDigest, Condition: change.Condition,
		FromDecision: change.FromDecision, ToDecision: change.ToDecision,
		Cases: []policycompilation.PolicyRevisionCasePair{
			{Baseline: freshBefore, Candidate: freshAfter},
			{Baseline: stale, Candidate: stale},
			{Baseline: missing, Candidate: missing},
		},
	}
	requestBytes, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	requestBytes = append([]byte(" \n\t"), requestBytes...)
	requestBytes = append(requestBytes, '\n')
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	report, err := policycompilation.ObservePolicyDecisionRevision(ctx, "policy.gooo", source,
		"metapolicycompilation", "metapolicycompilation", request)
	if err != nil {
		t.Fatal(err)
	}
	report.RequestArtifactDigest = digestBytes(requestBytes)
	return source, requestBytes, report
}

func revisionReceiptJSON(t *testing.T, value any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestRevisionReceiptReconstructsGeneratedEvidenceAndRejectsCounterexamples(t *testing.T) {
	source, requestBytes, report := revisionReceiptFixture(t)
	reportBytes := revisionReceiptJSON(t, report)
	t.Run("source-bound-complete-receipt", func(t *testing.T) {
		got, err := reconstructRevisionReceipt("policy.gooo", source, requestBytes, reportBytes,
			"metapolicycompilation", "metapolicycompilation")
		if err != nil || got.Decision != "RECEIPT_CONSISTENT_ONLY" || len(got.Checks) != 6 ||
			got.IndependentComparisons != 18 || got.Counts != report.Counts ||
			got.RequestArtifactDigest == got.CanonicalRequestDigest ||
			got.ReportArtifactDigest != digestBytes(reportBytes) || got.SourceDigest != digestBytes(source) {
			t.Fatalf("receipt reconstruction differs from exact observations: %+v error=%v", got, err)
		}
		for _, check := range got.Checks {
			if check.State != "CLOSED" {
				t.Fatalf("complete supplied receipt did not close bounded check: %+v", check)
			}
		}
		unknown := got.ExecutionUnknown
		if got.ExecutionObserved || got.MutationAuthority != 0 || got.PromotionAuthority != 0 ||
			got.Improvement != "UNKNOWN" || unknown.State != "UNKNOWN" || unknown.Stage == "" ||
			unknown.Step == "" || unknown.Reason != "PROCESS_EXECUTION_NOT_ATTESTED" ||
			unknown.UnknownClass != "DIRECT_MISSING" || unknown.NextOperation == "" ||
			unknown.BlockedBy == nil || len(unknown.BlockedBy) != 0 ||
			len(got.BaselineSource.RuleBindings) != 8 || len(got.CandidateSource.RuleBindings) != 8 {
			t.Fatal("receipt acquired execution authority or lost source/meta/UNKNOWN bindings")
		}
		replay, replayError := reconstructRevisionReceipt("policy.gooo", source, requestBytes, reportBytes,
			"metapolicycompilation", "metapolicycompilation")
		if replayError != nil || !bytes.Equal(revisionReceiptJSON(t, got), revisionReceiptJSON(t, replay)) {
			t.Fatal("read-only receipt reconstruction is not deterministic")
		}
	})
	for _, name := range []string{
		"raw-request-framing", "duplicate-json-key", "unknown-json-field", "source-digest-mismatch",
		"unrequested-semantic-change", "silently-rebound-snapshot", "changed-case-id", "tampered-unknown-frontier",
		"missing-replay-result", "counter-overclaim", "forged-completion", "authority-escalation",
	} {
		t.Run(name, func(t *testing.T) {
			var changed policycompilation.PolicyRevisionObservation
			if err := json.Unmarshal(reportBytes, &changed); err != nil {
				t.Fatal(err)
			}
			rawSource, rawRequest := bytes.Clone(source), bytes.Clone(requestBytes)
			switch name {
			case "raw-request-framing":
				rawRequest = append(rawRequest, '\n')
			case "source-digest-mismatch":
				rawSource = append(rawSource, []byte("\n// independently changed input\n")...)
			case "unrequested-semantic-change":
				changed.CandidateSource = strings.Replace(changed.CandidateSource,
					"namespace metapolicycompilation", "namespace unrequested", 1)
			case "silently-rebound-snapshot":
				changed.Candidate.DeclaredInputs[1] = changed.Request.Cases[0].Candidate
			case "changed-case-id":
				changed.Candidate.FirstResults[0].CaseID = "unrequested-case"
			case "tampered-unknown-frontier":
				changed.Candidate.SourceResults[2].BlockedBy = []string{"fabricated-frontier"}
			case "missing-replay-result":
				changed.Candidate.ReplayResults = changed.Candidate.ReplayResults[:2]
			case "counter-overclaim":
				changed.Counts.RequestedTransitionsObserved++
			case "forged-completion":
				changed.Candidate.ExecutionError, changed.Candidate.Complete = "interrupted batch", false
			case "authority-escalation":
				changed.PromotionAuthority = 1
			}
			rawReport := revisionReceiptJSON(t, changed)
			switch name {
			case "duplicate-json-key":
				rawReport = []byte(strings.Replace(string(rawReport), "\"schema\":", "\"schema\":\"forged\",\"schema\":", 1))
			case "unknown-json-field":
				rawReport = append(rawReport[:len(rawReport)-1], []byte(",\"undeclared_authority\":true}")...)
			}
			got, err := reconstructRevisionReceipt("policy.gooo", rawSource, rawRequest, rawReport,
				"metapolicycompilation", "metapolicycompilation")
			if err == nil || (got.Schema != "" && got.Decision != "REFUTED") {
				t.Fatalf("counterexample was accepted: decision=%s error=%v", got.Decision, err)
			}
		})
	}
	t.Run("honest-incomplete-attempt-remains-unknown", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		failed, err := policycompilation.ObservePolicyDecisionRevision(ctx, "policy.gooo", source,
			"metapolicycompilation", "metapolicycompilation", report.Request)
		if err == nil {
			t.Fatal("cancelled producer unexpectedly succeeded")
		}
		failed.RequestArtifactDigest = digestBytes(requestBytes)
		got, err := reconstructRevisionReceipt("policy.gooo", source, requestBytes, revisionReceiptJSON(t, failed),
			"metapolicycompilation", "metapolicycompilation")
		if err != nil || got.Decision != "UNKNOWN" || got.Counts.FailedBatches != 1 ||
			got.Counts.ObservedCasePairs != 0 || got.ExecutionObserved {
			t.Fatalf("honest incomplete evidence was rewritten: %+v error=%v", got, err)
		}
	})
}

func TestRevisionReceiptModeRejectsMixedOrIncompleteInputs(t *testing.T) {
	for _, args := range [][]string{
		{"-policy", "policy.gooo", "-revision-request", "request.json"},
		{"-policy", "policy.gooo", "-observe-revision-receipt", "report.json"},
		{"-policy", "policy.gooo", "-revision-request", "request.json", "-observe-revision-receipt", "report.json", "-output", "forbidden"},
		{"-policy", "policy.gooo", "-revision-request", "request.json", "-observe-revision-receipt", "report.json", "-observe-source"},
		{"-policy", "policy.gooo", "-revision-request", "request.json", "-observe-revision-receipt", "report.json", "positional"},
	} {
		flags := flag.NewFlagSet("receipt", flag.ContinueOnError)
		flags.SetOutput(io.Discard)
		for _, name := range []string{"policy", "revision-request", "observe-revision-receipt", "output"} {
			flags.String(name, "", "")
		}
		flags.String("profile-package", "metapolicycompilation", "")
		flags.String("profile-namespace", "metapolicycompilation", "")
		flags.Bool("observe-source", false, "")
		if err := flags.Parse(args); err != nil {
			t.Fatal(err)
		}
		if requested, err := revisionReceiptMode(flags); !requested || err == nil {
			t.Fatalf("mixed or incomplete receipt mode accepted: %v", args)
		}
	}
}

func TestRevisionReceiptObserverCLIEmitsReadOnlyBoundReport(t *testing.T) {
	source, requestBytes, report := revisionReceiptFixture(t)
	reportBytes := revisionReceiptJSON(t, report)
	work := t.TempDir()
	files := []struct {
		name string
		data []byte
	}{{"policy.gooo", source}, {"request.json", requestBytes}, {"report.json", reportBytes}}
	for _, file := range files {
		if err := os.WriteFile(filepath.Join(work, file.name), file.data, 0400); err != nil {
			t.Fatal(err)
		}
	}
	binary := filepath.Join(work, "receipt-observer")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	build := exec.CommandContext(ctx, "go", "build", "-trimpath", "-o", binary, ".")
	build.Env = append(os.Environ(), "GOTOOLCHAIN=go1.27.0")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build actual receipt observer: %v\n%s", err, output)
	}
	command := exec.CommandContext(ctx, binary, "-policy", filepath.Join(work, "policy.gooo"),
		"-revision-request", filepath.Join(work, "request.json"), "-observe-revision-receipt", filepath.Join(work, "report.json"))
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := command.Run(); err != nil {
		t.Fatalf("execute actual receipt observer: %v\n%s\n%s", err, stdout.Bytes(), stderr.Bytes())
	}
	var got revisionReceiptObservation
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil || got.Decision != "RECEIPT_CONSISTENT_ONLY" ||
		got.ExecutionObserved || len(got.Checks) != 6 || got.IndependentComparisons != 18 {
		t.Fatalf("native observer output is incomplete: %+v error=%v", got, err)
	}
	for _, file := range files {
		after, err := os.ReadFile(filepath.Join(work, file.name))
		if err != nil || !bytes.Equal(after, file.data) {
			t.Fatalf("observer changed original input %s: %v", file.name, err)
		}
	}
	binaryBytes, err := os.ReadFile(binary)
	if err != nil {
		t.Fatal(err)
	}
	receipt := map[string]any{
		"schema": "gooo/meta-policy-revision-receipt-cli-evidence/v1",
		"evidence_class": "SYNTHETIC_FIXTURE", "command": command.Args,
		"process_exit_code": command.ProcessState.ExitCode(), "observer_digest": digestBytes(binaryBytes),
		"source_digest": digestBytes(source), "request_artifact_digest": digestBytes(requestBytes),
		"input_report_digest": digestBytes(reportBytes), "raw_stdout": stdout.String(),
		"stdout_digest": digestBytes(stdout.Bytes()), "raw_stderr": stderr.String(),
		"stderr_digest": digestBytes(stderr.Bytes()), "input_bytes_unchanged_at_end": true,
		"execution_provenance": got.ExecutionUnknown, "improvement": got.Improvement,
	}
	t.Logf("POLICY_REVISION_RECEIPT_OBSERVATION=%s", revisionReceiptJSON(t, receipt))
}
