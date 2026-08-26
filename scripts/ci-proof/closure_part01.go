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
