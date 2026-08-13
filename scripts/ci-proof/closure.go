package main

import (
	"encoding/json"
	"fmt"
	"os"
)

const closureSchema = "gooo/ci-closure/v1"

type closureInput struct {
	CanonicalJobs        []failureJob `json:"canonical_jobs"`
	TerminalFailures     []failureJob `json:"terminal_failures"`
	TerminalFailureCodes []string     `json:"terminal_failure_codes"`
}

type closureManifest struct {
	Schema                  string       `json:"schema"`
	Version                 int          `json:"version"`
	Status                  string       `json:"status"`
	Decision                string       `json:"decision"`
	Repository              string       `json:"repository"`
	Event                   string       `json:"event"`
	EventRef                string       `json:"event_ref"`
	CheckoutRef             string       `json:"checkout_ref"`
	BaseRef                 string       `json:"base_ref"`
	BaseSHA                 string       `json:"base_sha"`
	HeadSHA                 string       `json:"head_sha"`
	PRNumber                int64        `json:"pr_number"`
	RunID                   int64        `json:"run_id"`
	RunAttempt              int64        `json:"run_attempt"`
	WorkflowSHA             string       `json:"workflow_sha"`
	OwnerBranch             string       `json:"owner_branch"`
	OwnerRef                string       `json:"owner_ref"`
	CanonicalJobs           []failureJob `json:"canonical_jobs"`
	TerminalFailures        []failureJob `json:"terminal_failures"`
	TerminalFailureCodes    []string     `json:"terminal_failure_codes"`
	WriteEffect             string       `json:"write_effect"`
	NoWriteOutsideGenerated bool         `json:"no_write_outside_generated"`
}

func writeClosureManifest(inputPath, outputPath string) error {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("read closure input: %w", err)
	}
	var input closureInput
	if err := json.Unmarshal(data, &input); err != nil {
		return fmt.Errorf("parse closure input: %w", err)
	}
	binding, err := readFailureBinding()
	if err != nil {
		return err
	}
	manifest, err := buildClosureManifest(input, binding)
	if err != nil {
		return err
	}
	encoded, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal closure manifest: %w", err)
	}
	if err := os.WriteFile(outputPath, append(encoded, '\n'), 0o644); err != nil {
		return fmt.Errorf("write closure manifest: %w", err)
	}
	return nil
}

func buildClosureManifest(input closureInput, binding failureBinding) (closureManifest, error) {
	scope, err := failureScope(binding)
	if err != nil {
		return closureManifest{}, err
	}
	if len(input.TerminalFailures) != 0 || len(input.TerminalFailureCodes) != 0 {
		return closureManifest{}, fmt.Errorf("no-failure closure contains terminal failure data")
	}
	if err := validateCanonicalClosureJobs(input.CanonicalJobs, binding); err != nil {
		return closureManifest{}, err
	}
	manifest := closureManifest{
		Schema: closureSchema, Version: 1, Status: "NO_TERMINAL_FAILURE", Decision: "HEALTH_PASS_ONLY",
		Repository: binding.Repository, Event: binding.Event, EventRef: binding.EventRef, CheckoutRef: binding.CheckoutRef,
		BaseRef: binding.BaseRef, BaseSHA: binding.BaseSHA, HeadSHA: binding.HeadSHA, PRNumber: binding.PRNumber,
		RunID: binding.RunID, RunAttempt: binding.RunAttempt, WorkflowSHA: binding.WorkflowSHA,
		OwnerBranch: binding.OwnerBranch, OwnerRef: failureOwnerRef(binding), CanonicalJobs: append([]failureJob(nil), input.CanonicalJobs...),
		TerminalFailures: []failureJob{}, TerminalFailureCodes: []string{}, WriteEffect: "none", NoWriteOutsideGenerated: true,
	}
	if scope == "" {
		return closureManifest{}, fmt.Errorf("closure scope is empty")
	}
	if err := validateClosureManifest(manifest, binding); err != nil {
		return closureManifest{}, err
	}
	return manifest, nil
}

func validateCanonicalClosureJobs(jobs []failureJob, binding failureBinding) error {
	if len(jobs) != len(proofJobs) {
		return fmt.Errorf("no-failure closure requires exactly six canonical jobs")
	}
	seen := make(map[int64]bool, len(jobs))
	for index, job := range jobs {
		if job.Name != proofJobs[index] || job.ID <= 0 || seen[job.ID] || job.Status != "completed" || job.Conclusion != "success" || job.HeadSHA != binding.HeadSHA || job.RunID != binding.RunID || job.RunAttempt != binding.RunAttempt {
			return fmt.Errorf("canonical closure job is missing, failed, duplicated, or stale")
		}
		seen[job.ID] = true
	}
	return nil
}

func validateClosureManifest(manifest closureManifest, binding failureBinding) error {
	if manifest.Schema != closureSchema || manifest.Version != 1 || manifest.Status != "NO_TERMINAL_FAILURE" || manifest.Decision != "HEALTH_PASS_ONLY" || manifest.WriteEffect != "none" || !manifest.NoWriteOutsideGenerated {
		return fmt.Errorf("closure status is not an explicit health-only no-failure result")
	}
	if manifest.Repository != binding.Repository || manifest.Event != binding.Event || manifest.EventRef != binding.EventRef || manifest.CheckoutRef != binding.CheckoutRef || manifest.BaseRef != binding.BaseRef || manifest.BaseSHA != binding.BaseSHA || manifest.HeadSHA != binding.HeadSHA || manifest.PRNumber != binding.PRNumber || manifest.RunID != binding.RunID || manifest.RunAttempt != binding.RunAttempt || manifest.WorkflowSHA != binding.WorkflowSHA || manifest.OwnerBranch != binding.OwnerBranch || manifest.OwnerRef != failureOwnerRef(binding) {
		return fmt.Errorf("closure tuple is stale or mismatched")
	}
	if len(manifest.TerminalFailures) != 0 || len(manifest.TerminalFailureCodes) != 0 {
		return fmt.Errorf("no-failure closure contains terminal failure data")
	}
	return validateCanonicalClosureJobs(manifest.CanonicalJobs, binding)
}
