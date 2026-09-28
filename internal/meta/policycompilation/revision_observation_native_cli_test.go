package policycompilation

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
	"time"
)

// This is a synthetic native-CI user path, not external utility or an
// independent revision verifier. The real CLI stdout is retained as bytes.
func TestPolicyRevisionWitnessCLIEmitsBoundExecutionEvidence(t *testing.T) {
	source, request := revisionObservationFixture(t)
	work := t.TempDir()
	sourcePath := filepath.Join(work, "policy.gooo")
	requestPath := filepath.Join(work, "request.json")
	requestBytes, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	// Preserve noncanonical input framing to distinguish the raw artifact
	// identity from the canonical typed request identity.
	requestBytes = append([]byte(" \n\t"), requestBytes...)
	requestBytes = append(requestBytes, '\n')
	if err := os.WriteFile(sourcePath, source, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(requestPath, requestBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	repositoryRoot, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	binaryPath := filepath.Join(work, "witness")
	if runtime.GOOS == "windows" {
		binaryPath += ".exe"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	build := exec.CommandContext(ctx, "go", "build", "-trimpath", "-o", binaryPath, "./cmd/meta-policy-compilation-witness")
	build.Dir = repositoryRoot
	build.Env = append(os.Environ(), "GOTOOLCHAIN=go1.27.0")
	buildStart := time.Now()
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build native witness: %v\n%s", err, output)
	}
	buildWallMS := time.Since(buildStart).Milliseconds()
	binaryBytes, err := os.ReadFile(binaryPath)
	if err != nil {
		t.Fatal(err)
	}

	command := exec.CommandContext(ctx, binaryPath, "-policy", "policy.gooo", "-observe-revision", "request.json")
	command.Dir = work
	command.Env = append(os.Environ(), "GOTOOLCHAIN=go1.27.0")
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	commandStart := time.Now()
	if err := command.Run(); err != nil {
		t.Fatalf("execute native revision witness: %v\nstdout=%s\nstderr=%s", err, stdout.Bytes(), stderr.Bytes())
	}
	commandWallMS := time.Since(commandStart).Milliseconds()
	var report PolicyRevisionObservation
	if err := decodeStrictJSON(stdout.Bytes(), &report); err != nil {
		t.Fatalf("decode actual native CLI stdout: %v\n%s", err, stdout.Bytes())
	}
	sourceAfter, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	requestAfter, err := os.ReadFile(requestPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(sourceAfter, source) || !bytes.Equal(requestAfter, requestBytes) {
		t.Fatal("native CLI changed the caller's input bytes")
	}
	canonicalRequest, err := canonicalJSON(request)
	if err != nil {
		t.Fatal(err)
	}
	if report.Schema != PolicyRevisionObservationSchema || report.SourceFile != "policy.gooo" ||
		!reflect.DeepEqual(report.Request, request) || report.RequestDigest != DigestBytes(canonicalRequest) ||
		report.RequestArtifactDigest != DigestBytes(requestBytes) || report.OriginalPolicy.SourceDigest != DigestBytes(source) ||
		report.CandidatePolicy.SourceDigest != DigestBytes([]byte(report.CandidateSource)) ||
		report.CandidateSource == string(source) || report.OriginalPolicy.SemanticDigest == report.CandidatePolicy.SemanticDigest {
		t.Fatal("native CLI request, source, or candidate binding differs from the caller")
	}
	want := PolicyRevisionObservationCounts{
		RequestedCasePairs:           3,
		ObservedCasePairs:            3,
		SameInputObservedPairs:       2,
		RequestedTransitionsObserved: 1,
		SourceComparisons:            12,
		ReplayComparisons:            6}
	if report.Counts != want || report.ExecutionStatus != "COMPLETED" || report.ExecutionConformance != "PASS" {
		t.Fatalf("native CLI execution accounting differs: %+v status=%s conformance=%s", report.Counts, report.ExecutionStatus, report.ExecutionConformance)
	}
	for side, phase := range []PolicyRevisionExecution{report.Baseline, report.Candidate} {
		if !phase.Complete || phase.GeneratedJudgeSource == "" ||
			DigestBytes([]byte(phase.GeneratedJudgeSource)) != phase.GeneratedJudgeDigest ||
			len(phase.DeclaredInputs) != 3 || len(phase.SourceResults) != 3 ||
			len(phase.FirstResults) != 3 || len(phase.ReplayResults) != 3 {
			t.Fatalf("native CLI source or result cohort is incomplete for side %d", side)
		}
		for index, pair := range request.Cases {
			input := pair.Baseline
			if side == 1 {
				input = pair.Candidate
			}
			if phase.DeclaredInputs[index] != input || phase.FirstResults[index].CaseID != input.ID ||
				!sameResult(phase.FirstResults[index], phase.SourceResults[index]) ||
				!sameResult(phase.FirstResults[index], phase.ReplayResults[index]) {
				t.Fatalf("native CLI altered input or source/replay evidence at side %d case %d", side, index)
			}
		}
	}
	if len(report.Transitions) != 3 || !report.Transitions[0].RequestedTransitionObserved ||
		report.Transitions[0].InputsIdentical || report.Transitions[0].CausalAttribution != "UNASSESSED" ||
		report.Baseline.FirstResults[0].Decision != DecisionPass ||
		report.Candidate.FirstResults[0].Decision != DecisionFailClosed ||
		report.Candidate.FirstResults[1].MatchedCondition != ConditionSourceMismatch ||
		report.Candidate.FirstResults[2].Decision != DecisionUnknown {
		t.Fatal("native CLI did not preserve the requested transition, stale evidence, or missing evidence")
	}
	if report.Admission.State != "UNKNOWN" || report.Admission.Stage == "" || report.Admission.Step == "" ||
		report.Admission.Reason != "INDEPENDENT_REVISION_EVIDENCE_MISSING" || report.Admission.UnknownClass != "DIRECT_MISSING" ||
		report.Admission.NextOperation == "" || report.Admission.BlockedBy == nil || len(report.Admission.BlockedBy) != 0 ||
		report.MutationAuthority != 0 || report.PromotionAuthority != 0 || report.Improvement != "UNKNOWN" ||
		report.InputProvenance != "CALLER_DECLARED_NOT_VERIFIED" || report.RepositoryObservation != "NOT_PERFORMED" {
		t.Fatal("native CLI acquired unobserved admission, provenance, effects, or authority")
	}

	receipt := map[string]any{}
	receipt["schema"] = "gooo/meta-policy-revision-cli-evidence/v1"
	receipt["evidence_class"] = EvidenceSyntheticFixture
	receipt["command"] = command.Args
	receipt["working_directory"] = command.Dir
	receipt["process_exit_code"] = command.ProcessState.ExitCode()
	receipt["witness_sha256"] = DigestBytes(binaryBytes)
	receipt["witness_build_wall_ms"] = buildWallMS
	receipt["cli_wall_ms"] = commandWallMS
	receipt["source_before_sha256"] = DigestBytes(source)
	receipt["source_after_sha256"] = DigestBytes(sourceAfter)
	receipt["request_before_sha256"] = DigestBytes(requestBytes)
	receipt["request_after_sha256"] = DigestBytes(requestAfter)
	receipt["canonical_request_sha256"] = report.RequestDigest
	receipt["raw_stdout"] = stdout.String()
	receipt["stdout_sha256"] = DigestBytes(stdout.Bytes())
	receipt["raw_stderr"] = stderr.String()
	receipt["stderr_sha256"] = DigestBytes(stderr.Bytes())
	receipt["counts"] = report.Counts
	receipt["admission"] = report.Admission
	receipt["execution_conformance"] = report.ExecutionConformance
	receipt["repository_observation"] = report.RepositoryObservation
	receipt["improvement"] = report.Improvement
	receiptBytes, err := json.Marshal(receipt)
	if err != nil {
		t.Fatal(err)
	}
	// Existing native test.json artifacts retain this single-line receipt.
	// raw_stdout preserves the actual CLI bytes, not a replacement rendering.
	t.Logf("POLICY_REVISION_CLI_OBSERVATION=%s", receiptBytes)
}
