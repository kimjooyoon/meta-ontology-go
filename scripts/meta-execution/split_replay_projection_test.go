package main

import (
	"errors"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/operationconformance"
)

func TestSplitReplayIgnoresExecutionPathsAndAuditDigests(t *testing.T) {
	firstEvidence := splitReplayEvidenceFixture()
	secondEvidence := splitReplayEvidenceFixture()
	first := splitReplayProcessFixture()
	second := first
	second.StdoutBytes = 999
	second.StderrBytes = 777
	second.RawStdoutDigest = "sha256:raw-second"
	second.RawStderrDigest = "sha256:raw-second"
	second.StdoutDigest = "sha256:canonical-second"
	second.StderrDigest = "sha256:canonical-second"
	secondEvidence.Write.Events[0].Temporary = "/tmp/second-work/.extract.tmp"
	if mustSplitReplayDigest(t, firstEvidence, first, first, first) != mustSplitReplayDigest(t, secondEvidence, second, second, second) {
		t.Fatal("execution-only paths and process digests changed semantic replay")
	}
}

func TestSplitReplayBindsCanonicalOutputCommandAndArtifact(t *testing.T) {
	evidence := splitReplayEvidenceFixture()
	process := splitReplayProcessFixture()
	baseline := mustSplitReplayDigest(t, evidence, process, process, process)

	changedOutput := evidence
	changedOutput.Candidates = append([]operationconformance.FileEvidence{}, evidence.Candidates...)
	changedOutput.Candidates[0].Data = []byte("changed output")
	if mustSplitReplayDigest(t, changedOutput, process, process, process) == baseline {
		t.Fatal("canonical output drift was accepted")
	}

	changedCommand := process
	changedCommand.Command = append([]string{}, process.Command...)
	changedCommand.Command = append(changedCommand.Command, "--changed")
	if mustSplitReplayDigest(t, evidence, changedCommand, process, process) == baseline {
		t.Fatal("canonical command drift was accepted")
	}

	changedArtifact := evidence
	changedArtifact.Candidates = append([]operationconformance.FileEvidence{}, evidence.Candidates...)
	changedArtifact.Candidates[0].Path = "other.go"
	if mustSplitReplayDigest(t, changedArtifact, process, process, process) == baseline {
		t.Fatal("logical artifact drift was accepted")
	}
}

func mustSplitReplayDigest(t *testing.T, evidence operationconformance.SplitGoEvidence, executor, evaluator, verifier generation.ProcessObservation) string {
	t.Helper()
	digest, err := splitReplayDigest(evidence, executor, evaluator, verifier)
	if err != nil {
		t.Fatal(err)
	}
	return digest
}

func TestSplitReplayProjectionPropagatesMarshalFailure(t *testing.T) {
	_, err := splitReplayProjectionBytesWith(func(any) ([]byte, error) {
		return nil, errors.New("marshal failed")
	}, splitReplayEvidenceFixture(), splitReplayProcessFixture(), splitReplayProcessFixture(), splitReplayProcessFixture())
	if err == nil || err.Error() != "marshal failed" {
		t.Fatalf("marshal failure was not propagated: %v", err)
	}
}

func splitReplayEvidenceFixture() operationconformance.SplitGoEvidence {
	return operationconformance.SplitGoEvidence{
		ExpectedHeadSHA:  "head",
		OperationID:      operationconformance.OperationID,
		EvidenceComplete: true,
		Source: operationconformance.FileEvidence{
			Path: "subject.go", Data: []byte("package p\n"),
		},
		Candidates: []operationconformance.FileEvidence{{
			Path: "subject.go", Data: []byte("package p\n"),
		}},
		BuildContexts: []operationconformance.BuildContext{{
			GOOS: "linux", GOARCH: "amd64",
		}},
		Write: operationconformance.WriteReceipt{
			Complete: true, ExecutionSucceeded: true,
			DeclaredTargets: []string{"subject.go"},
			Events: []operationconformance.WriteEvent{{
				Sequence: 0, Kind: "REPLACE", Target: "subject.go",
				Temporary: "/tmp/first-work/.extract.tmp", Success: true,
			}},
			WritesOutsideDeclaredTargets: 0, TemporaryFilesRemaining: 0,
		},
	}
}

func splitReplayProcessFixture() generation.ProcessObservation {
	return generation.ProcessObservation{
		Command:         []string{"go", "run", "<workspace>"},
		ExitCode:        0,
		StdoutBytes:     100,
		StdoutDigest:    "sha256:stdout",
		StderrBytes:     0,
		StderrDigest:    "sha256:stderr",
		RawStdoutDigest: "sha256:raw-stdout",
		RawStderrDigest: "sha256:raw-stderr",
	}
}
