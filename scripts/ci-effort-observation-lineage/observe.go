package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publicworkflowlineage"
)

type observation struct {
	Schema       string                           `json:"schema"`
	Decision     string                           `json:"decision"`
	LineageState string                           `json:"lineage_state"`
	Reason       string                           `json:"reason"`
	Trigger      publicworkflowlineage.Trigger    `json:"trigger"`
	Source       publicworkflowlineage.SourceRun  `json:"source_run"`
	Evaluation   publicworkflowlineage.Evaluation `json:"evaluation"`
	PolicyDigest string                           `json:"policy_source_digest"`
}

func runObserve(input runInput) error {
	if input.Source == "" || input.Out == "" || input.Trigger == "" || input.Run == "" || input.Artifacts == "" {
		return errors.New("source, out, trigger, run, and artifacts are required")
	}
	source, err := os.ReadFile(input.Source)
	if err != nil {
		return err
	}
	policy, err := publicworkflowlineage.Load(input.Source, source)
	if err != nil {
		return err
	}
	var trigger publicworkflowlineage.Trigger
	if err := readValue(input.Trigger, &trigger); err != nil {
		return err
	}
	var sourceRun publicworkflowlineage.SourceRun
	if err := readValue(input.Run, &sourceRun); err != nil {
		return err
	}
	var artifacts publicworkflowlineage.ArtifactIndex
	if err := readValue(input.Artifacts, &artifacts); err != nil {
		return err
	}
	expected := fmt.Sprintf("ci-evidence-%d-%d", sourceRun.ID, sourceRun.RunAttempt)
	lineageInput := publicworkflowlineage.Input{Trigger: trigger, Source: sourceRun, Artifacts: artifacts, ExpectedArtifactName: expected, ExpectedRepository: policy.Repository, ExpectedWorkflow: policy.SourceWorkflow, ExpectedSourceAPIKey: policy.SourceAPIKey, ExpectedArtifactSubjectBinding: policy.ArtifactSubjectBinding}
	evaluation := publicworkflowlineage.Evaluate(lineageInput)
	readOnly := policy.EvaluateReadOnlyObservation(lineageInput)
	value := observation{Schema: publicworkflowlineage.ReportSchema, Decision: evaluation.Decision, LineageState: evaluation.LineageState, Reason: evaluation.Reason, Trigger: trigger, Source: sourceRun, Evaluation: evaluation, PolicyDigest: policy.SourceDigest}
	if err := os.MkdirAll(input.Out, 0o755); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(input.Out, "lineage-observation.json"), value); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(input.Out, "read-only-observation.json"), readOnly); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(input.Out, "lineage-human.txt"), []byte(fmt.Sprintf("Decision: %s\nState: %s\nReason: %s\nRead-only observation eligibility: %s\nTiming/operation observation: %t/%t\nEvidence reuse/promotion authority: %t/%t\n", value.Decision, value.LineageState, value.Reason, readOnly.Eligibility, readOnly.TimingObservationEligible, readOnly.OperationObservationEligible, readOnly.EvidenceReuseAllowed, readOnly.PromotionAllowed)), 0o444); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(input.Out, "lineage.env"), []byte(fmt.Sprintf("LINEAGE_DECISION=%s\nLINEAGE_STATE=%s\nLINEAGE_OBSERVATION_ELIGIBILITY=%s\nLINEAGE_TIMING_OBSERVATION_ELIGIBLE=%t\nLINEAGE_OPERATION_OBSERVATION_ELIGIBLE=%t\nLINEAGE_EVIDENCE_REUSE_ALLOWED=%t\nLINEAGE_PROMOTION_ALLOWED=%t\n", value.Decision, value.LineageState, readOnly.Eligibility, readOnly.TimingObservationEligible, readOnly.OperationObservationEligible, readOnly.EvidenceReuseAllowed, readOnly.PromotionAllowed)), 0o444); err != nil {
		return err
	}
	if evaluation.Decision == publicworkflowlineage.DecisionRefuted && readOnly.Eligibility != publicworkflowlineage.ObservationAllowed {
		return errors.New("workflow lineage is REFUTED; refusing to consume mismatched evidence")
	}
	return nil
}

func readValue(path string, value any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, value); err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}
	return nil
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
