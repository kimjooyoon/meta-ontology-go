package main

import (
	"encoding/json"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/operationconformance"
)

type splitReplayFile struct {
	Path             string                                    `json:"path"`
	Data             []byte                                    `json:"data"`
	DeclarationOrder []operationconformance.DeclarationOrder `json:"declaration_order,omitempty"`
}

type splitReplayEvent struct {
	Sequence int    `json:"sequence"`
	Kind     string `json:"kind"`
	Target   string `json:"target"`
	Success  bool   `json:"success"`
}

type splitReplayWrite struct {
	Complete                     bool               `json:"complete"`
	ExecutionSucceeded           bool               `json:"execution_succeeded"`
	DeclaredTargets              []string           `json:"declared_targets"`
	Events                       []splitReplayEvent `json:"events"`
	WritesOutsideDeclaredTargets int                `json:"writes_outside_declared_targets"`
	TemporaryFilesRemaining      int                `json:"temporary_files_remaining"`
}

type splitReplayProcess struct {
	Command  []string `json:"command"`
	ExitCode int      `json:"exit_code"`
}

type splitReplayProjection struct {
	ExpectedHeadSHA  string                                   `json:"expected_head_sha"`
	OperationID      string                                   `json:"operation_id"`
	EvidenceComplete bool                                     `json:"evidence_complete"`
	Source           splitReplayFile                          `json:"source"`
	Candidates       []splitReplayFile                        `json:"candidates"`
	BuildContexts    []operationconformance.BuildContext `json:"build_contexts"`
	Write            splitReplayWrite                         `json:"write_receipt"`
	Executor         splitReplayProcess                       `json:"executor"`
	Evaluator        splitReplayProcess                       `json:"evaluator"`
	Verifier         splitReplayProcess                       `json:"verifier"`
}

func splitReplayProjectionFrom(evidence operationconformance.SplitGoEvidence, executor, evaluator, verifier generation.ProcessObservation) splitReplayProjection {
	return splitReplayProjection{
		ExpectedHeadSHA: evidence.ExpectedHeadSHA,
		OperationID: evidence.OperationID,
		EvidenceComplete: evidence.EvidenceComplete,
		Source: splitReplayFileFrom(evidence.Source),
		Candidates: splitReplayFilesFrom(evidence.Candidates),
		BuildContexts: append([]operationconformance.BuildContext{}, evidence.BuildContexts...),
		Write: splitReplayWriteFrom(evidence.Write),
		Executor: splitReplayProcessFrom(executor),
		Evaluator: splitReplayProcessFrom(evaluator),
		Verifier: splitReplayProcessFrom(verifier),
	}
}

func splitReplayFileFrom(file operationconformance.FileEvidence) splitReplayFile {
	return splitReplayFile{Path: file.Path, Data: append([]byte{}, file.Data...), DeclarationOrder: append([]operationconformance.DeclarationOrder{}, file.DeclarationOrder...)}
}

func splitReplayFilesFrom(files []operationconformance.FileEvidence) []splitReplayFile {
	result := make([]splitReplayFile, len(files))
	for index, file := range files {
		result[index] = splitReplayFileFrom(file)
	}
	return result
}

func splitReplayWriteFrom(write operationconformance.WriteReceipt) splitReplayWrite {
	events := make([]splitReplayEvent, len(write.Events))
	for index, event := range write.Events {
		events[index] = splitReplayEvent{Sequence: event.Sequence, Kind: event.Kind, Target: event.Target, Success: event.Success}
	}
	return splitReplayWrite{Complete: write.Complete, ExecutionSucceeded: write.ExecutionSucceeded, DeclaredTargets: append([]string{}, write.DeclaredTargets...), Events: events, WritesOutsideDeclaredTargets: write.WritesOutsideDeclaredTargets, TemporaryFilesRemaining: write.TemporaryFilesRemaining}
}

func splitReplayProcessFrom(observation generation.ProcessObservation) splitReplayProcess {
	return splitReplayProcess{Command: append([]string{}, observation.Command...), ExitCode: observation.ExitCode}
}

func splitReplayProjectionBytes(evidence operationconformance.SplitGoEvidence, executor, evaluator, verifier generation.ProcessObservation) []byte {
	payload, err := json.Marshal(splitReplayProjectionFrom(evidence, executor, evaluator, verifier))
	if err != nil {
		return nil
	}
	return payload
}

func splitReplayDigest(evidence operationconformance.SplitGoEvidence, executor, evaluator, verifier generation.ProcessObservation) string {
	return digestBytes(splitReplayProjectionBytes(evidence, executor, evaluator, verifier))
}
