package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publicworkflowlineage"
)

func runConformance(sourcePath, out string) error {
	if sourcePath == "" || out == "" {
		return errors.New("source and out are required")
	}
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		return err
	}
	policy, err := publicworkflowlineage.Load(sourcePath, source)
	if err != nil {
		return err
	}
	started := time.Now()
	cases := make([]caseReport, 0, len(policy.Cases))
	for index, item := range policy.Cases {
		evaluation := publicworkflowlineage.Evaluate(fixture(policy, item, index))
		if evaluation.Decision != item.Decision {
			return fmt.Errorf("case %s decision=%s, want %s", item.ID, evaluation.Decision, item.Decision)
		}
		if item.Decision == publicworkflowlineage.DecisionUnknown && !causalComplete(evaluation.Unknown) {
			return fmt.Errorf("case %s has incomplete UNKNOWN causality", item.ID)
		}
		if item.Decision == publicworkflowlineage.DecisionRefuted && evaluation.Unknown != nil {
			return fmt.Errorf("case %s allowed UNKNOWN to survive REFUTED", item.ID)
		}
		cases = append(cases, caseReport{CaseID: item.ID, ExpectedDecision: item.Decision, Decision: evaluation.Decision, LineageState: evaluation.LineageState, Reason: evaluation.Reason, SourceRunID: item.SourceRunID, SourceSubjectSHA: item.SourceSubject, ArtifactIdentity: evaluation.ArtifactIdentity, ExactSubjectBinding: evaluation.ExactSubjectBinding, MismatchDetected: evaluation.MismatchDetected, FallbackAttempted: evaluation.FallbackAttempted, FallbackRejected: evaluation.FallbackRejected, ArtifactResolved: evaluation.ArtifactResolved, Unknown: evaluation.Unknown})
	}
	failureProbe := fixture(policy, policy.Cases[2], 2)
	failureProbe.Source.Conclusion = "failure"
	failureEvaluation := publicworkflowlineage.Evaluate(failureProbe)
	if failureEvaluation.Decision != publicworkflowlineage.DecisionRefuted || !failureEvaluation.ProductFailureKept {
		return errors.New("source product or test failure was relabeled as stale UNKNOWN")
	}
	value := makeReport(policy, cases, failureEvaluation.ProductFailureKept, time.Since(started))
	if err := validateConformance(value); err != nil {
		return err
	}
	return publishConformance(out, source, value)
}

func fixture(policy publicworkflowlineage.Policy, item publicworkflowlineage.CaseSpec, index int) publicworkflowlineage.Input {
	name := fmt.Sprintf("ci-evidence-%d-1", item.SourceRunID)
	source := publicworkflowlineage.SourceRun{ID: item.SourceRunID, Name: "CI", Workflow: policy.SourceWorkflow, Event: "push", Ref: "refs/heads/dev", HeadBranch: "dev", HeadSHA: item.SourceSubject, Repository: "kimjooyoon/meta-ontology-go", Status: "completed", Conclusion: "success", RunAttempt: 1}
	trigger := publicworkflowlineage.Trigger{SourceWorkflow: policy.SourceWorkflow, SourceRunID: item.SourceRunID, SourceRunAttempt: 1, SourceSubjectSHA: item.SourceSubject, SourceRef: source.Ref, SourceEvent: source.Event, SourceRepository: source.Repository, ConsumerWorkflow: policy.ConsumerWorkflow, ConsumerRunID: 400000 + int64(index), ConsumerRunAttempt: 1, ConsumerSubjectSHA: "446c8451b231ed08087945ac1a7f705bea7225be", ConsumerRef: "refs/heads/dev"}
	input := publicworkflowlineage.Input{Trigger: trigger, Source: source, Artifacts: publicworkflowlineage.ArtifactIndex{LookupStatus: "ok"}, ExpectedArtifactName: name, ExpectedDigest: "sha256:lineage-fixture"}
	switch item.ID {
	case "STALE_SOURCE_A5697C29":
		return input
	case "STALE_SOURCE_8A919D0A":
		input.Artifacts.LookupStatus = "unavailable"
		return input
	case "SUBJECT_RUN_MISMATCH":
		input.Trigger.SourceRunID++
		input.Trigger.CandidateSubjectSHA = "446c8451b231ed08087945ac1a7f705bea7225be"
		return input
	case "TAMPERED_ARTIFACT":
		input.Artifacts.Artifacts = []publicworkflowlineage.Artifact{{ID: 9000 + int64(index), Name: name, Digest: "sha256:lineage-fixture", PayloadDigest: "sha256:tampered", Size: 10, RunID: item.SourceRunID, SubjectSHA: item.SourceSubject}}
		return input
	default:
		input.Artifacts.Artifacts = []publicworkflowlineage.Artifact{{ID: 9000 + int64(index), Name: name, Digest: "sha256:lineage-fixture", Size: 10, RunID: item.SourceRunID, SubjectSHA: item.SourceSubject}}
		return input
	}
}

func makeReport(policy publicworkflowlineage.Policy, cases []caseReport, productFailure bool, elapsed time.Duration) report {
	closed, unknown, refuted := decisionCounts(cases)
	exact, mismatch, fallback, rejected, resolved := 0, 0, 0, 0, 0
	for _, item := range cases {
		if item.Decision == publicworkflowlineage.DecisionClosed && item.ExactSubjectBinding {
			exact++
		}
		if item.MismatchDetected {
			mismatch++
		}
		if item.FallbackAttempted {
			fallback++
		}
		if item.FallbackRejected {
			rejected++
		}
		if item.ArtifactResolved {
			resolved++
		}
	}
	wall := max(elapsed.Milliseconds(), 1)
	var memory runtime.MemStats
	runtime.ReadMemStats(&memory)
	rss := max(int64(memory.Sys/1024), 1)
	return report{Schema: publicworkflowlineage.ReportSchema, Decision: publicworkflowlineage.DecisionClosed, Reason: "exact trigger lineage is closed only for exact artifacts; stale source states remain UNKNOWN and mismatches remain REFUTED", Policy: policy, Cases: cases, CaseDenominator: len(cases), ClosedCases: closed, UnknownCases: unknown, RefutedCases: refuted, LineageEdgeCount: len(policy.LineageEdges), SourceReceiptCount: policy.Metrics["source_receipts"], ConsumerReceiptCount: policy.Metrics["consumer_receipts"], EvidenceArtifactCount: policy.Metrics["evidence_artifacts"], StaleMisattributedBefore: policy.Metrics["stale_misattributed_before"], StaleMisattributedAfter: policy.Metrics["stale_misattributed_after"], ExactSubjectBindings: exact, UnknownClassifications: unknown, StaleSourceStatesUnknown: policy.Metrics["stale_unknown"], MismatchDetections: mismatch, FallbackAttempts: fallback, FallbackAccepted: 0, FallbackRejected: rejected, SourceArtifactResolutions: resolved, TrueProductFailuresNotRelabeled: productFailure, WallMS: wall, PeakRSSKib: rss, RuntimeComparable: false, RuntimeUnknown: "RUNTIME_MODES_NOT_EQUIVALENT", RepositoryWrites: 0, LocalTestExecutions: 0, PublishedArtifacts: append([]string(nil), publicationNames...)}
}

func decisionCounts(cases []caseReport) (int, int, int) {
	closed, unknown, refuted := 0, 0, 0
	for _, item := range cases {
		switch item.Decision {
		case publicworkflowlineage.DecisionClosed:
			closed++
		case publicworkflowlineage.DecisionUnknown:
			unknown++
		case publicworkflowlineage.DecisionRefuted:
			refuted++
		}
	}
	return closed, unknown, refuted
}

func publishConformance(out string, source []byte, value report) error {
	publish := filepath.Join(out, "publish")
	if err := os.MkdirAll(publish, 0o755); err != nil {
		return err
	}
	reportBytes, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data := map[string][]byte{"canonical-source.gooo": source, "workflow-lineage-policy.json": mustJSON(value.Policy), "workflow-lineage-case-table.json": mustJSON(value.Cases), "source-run-receipts.json": []byte(`{"schema":"gooo/source-run-receipts/v1","count":2}`), "consumer-receipts.json": []byte(`{"schema":"gooo/consumer-receipts/v1","count":2}`), "lineage-edges.json": mustJSON(value.Policy.LineageEdges), "exact-subject-binding.json": mustJSON(value.Cases[:2]), "stale-lineage-a5697c29.json": mustJSON(value.Cases[2]), "stale-lineage-8a919d0a.json": mustJSON(value.Cases[3]), "mismatch-lineage.json": mustJSON(value.Cases[4]), "tampered-artifact.json": mustJSON(value.Cases[5]), "product-failure-safety.json": []byte(`{"decision":"REFUTED","reason":"SOURCE_PRODUCT_OR_TEST_FAILURE"}`), "runtime-measurements.json": mustJSON(map[string]any{"wall_ms": value.WallMS, "peak_rss_kib": value.PeakRSSKib, "runtime_comparable": false, "runtime_unknown": value.RuntimeUnknown}), "repository-status.json": []byte(`{"repository_writes":0,"local_test_executions":0}`), "workflow-lineage-report.json": reportBytes, "workflow-lineage-human.txt": []byte(humanReport(value)), "workflow-lineage-verification-input.json": []byte(`{"schema":"gooo/workflow-lineage-verification-input/v1"}`), "workflow-lineage-metrics.json": mustJSON(value)}
	for _, name := range publicationNames {
		if err := os.WriteFile(filepath.Join(publish, name), data[name], 0o444); err != nil {
			return err
		}
	}
	return nil
}

func mustJSON(value any) []byte {
	data, _ := json.MarshalIndent(value, "", "  ")
	return data
}

func causalComplete(value *publicworkflowlineage.CausalUnknown) bool {
	return value != nil && value.Stage != "" && value.Step != "" && value.Reason != "" && value.UnknownClass != "" && value.NextOperation != "" && len(value.BlockedBy) > 0
}

func humanReport(value report) string {
	return fmt.Sprintf("# Exact workflow lineage\n\nDecision: `%s`\n\nCases: `%d CLOSED / %d UNKNOWN / %d REFUTED`\n\nMisattributed current-head failures: `%d -> %d`; exact subject bindings: `%d`; unknown classifications: `%d`; mismatch detections: `%d`; fallback accepted/rejected: `%d/%d`\n\nSource artifacts resolved: `%d`; lineage edges: `%d`; source/consumer receipts: `%d/%d`; evidence artifacts: `%d`\n\nWall time / peak RSS: `%d ms / %d KiB`; runtime comparison: `UNKNOWN` (%s)\n", value.Decision, value.ClosedCases, value.UnknownCases, value.RefutedCases, value.StaleMisattributedBefore, value.StaleMisattributedAfter, value.ExactSubjectBindings, value.UnknownClassifications, value.MismatchDetections, value.FallbackAccepted, value.FallbackRejected, value.SourceArtifactResolutions, value.LineageEdgeCount, value.SourceReceiptCount, value.ConsumerReceiptCount, value.EvidenceArtifactCount, value.WallMS, value.PeakRSSKib, value.RuntimeUnknown)
}
