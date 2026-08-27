package policycompilation

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// GenerateJudge emits a standalone Go program. It has no import path into the
// repository, so the execution evidence is independent of this package.
func GenerateJudge(policy CompiledPolicy) []byte {
	return []byte(fmt.Sprintf(`package main

import (
	"encoding/json"
	"os"
)

type input struct {
	ID string %q
	ProducerAvailable bool %q
	ConsumerAvailable bool %q
	ObservedSourceDigest string %q
	ObservedArtifactSourceDigest string %q
	ObservedIndependentDigest string %q
}

type result struct {
	CaseID string %q
	Decision string %q
	Stage string %q
	Step int %q
	Reason string %q
	PolicyDigest string %q
	SemanticDigest string %q
	Denominator int %q
}

const policyDigest = %q
const semanticDigest = %q
const policyDenominator = %d

func main() {
	var value input
	decoder := json.NewDecoder(os.Stdin)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil { os.Exit(2) }
	output := result{CaseID: value.ID, PolicyDigest: policyDigest, SemanticDigest: semanticDigest, Denominator: policyDenominator}
	if !value.ProducerAvailable || !value.ConsumerAvailable {
		output.Decision, output.Stage, output.Step, output.Reason = "UNKNOWN", "VERIFY", 4, "EVIDENCE_UNAVAILABLE"
	} else if value.ObservedSourceDigest == "" || value.ObservedArtifactSourceDigest == "" || value.ObservedIndependentDigest == "" {
		output.Decision, output.Stage, output.Step, output.Reason = "UNKNOWN", "VERIFY", 4, "DIGEST_UNAVAILABLE"
	} else if value.ObservedSourceDigest != policyDigest {
		output.Decision, output.Stage, output.Step, output.Reason = "FAIL_CLOSED", "REDUCE", 7, "SOURCE_DIGEST_MISMATCH"
	} else if value.ObservedArtifactSourceDigest != policyDigest {
		output.Decision, output.Stage, output.Step, output.Reason = "FAIL_CLOSED", "CONSUME", 2, "ARTIFACT_SOURCE_MISMATCH"
	} else if value.ObservedIndependentDigest != policyDigest {
		output.Decision, output.Stage, output.Step, output.Reason = "FAIL_CLOSED", "VERIFY", 4, "INDEPENDENT_SOURCE_MISMATCH"
	} else if policyDenominator != %d {
		output.Decision, output.Stage, output.Step, output.Reason = "FAIL_CLOSED", "COMPILE", 3, "FIXED_DENOMINATOR_CHANGED"
	} else {
		output.Decision, output.Stage, output.Step, output.Reason = "PASS", "REDUCE", 7, "SEMANTIC_EQUIVALENCE_PROVED"
	}
	if err := json.NewEncoder(os.Stdout).Encode(output); err != nil { os.Exit(3) }
}
`, `json:"id"`, `json:"producer_available"`, `json:"consumer_available"`, `json:"observed_source_digest"`, `json:"observed_artifact_source_digest"`, `json:"observed_independent_digest"`, `json:"case_id"`, `json:"decision"`, `json:"stage"`, `json:"step"`, `json:"reason"`, `json:"policy_digest"`, `json:"semantic_digest"`, `json:"fixed_denominator"`, policy.SourceDigest, policy.SemanticDigest, policy.Denominator, FixedDenominator))
}

func ExecuteGenerated(ctx context.Context, judgeSource []byte, input Case) (DecisionResult, error) {
	work, err := os.MkdirTemp("", "gooo-policy-judge-")
	if err != nil {
		return DecisionResult{}, err
	}
	defer os.RemoveAll(work)
	judgePath := filepath.Join(work, "judge.go")
	if err := os.WriteFile(judgePath, judgeSource, 0o600); err != nil {
		return DecisionResult{}, err
	}
	payload, err := json.Marshal(struct {
		ID                           string `json:"id"`
		ProducerAvailable            bool   `json:"producer_available"`
		ConsumerAvailable            bool   `json:"consumer_available"`
		ObservedSourceDigest         string `json:"observed_source_digest"`
		ObservedArtifactSourceDigest string `json:"observed_artifact_source_digest"`
		ObservedIndependentDigest    string `json:"observed_independent_digest"`
	}{
		ID: input.ID, ProducerAvailable: input.ProducerAvailable, ConsumerAvailable: input.ConsumerAvailable,
		ObservedSourceDigest:         input.ObservedSourceDigest,
		ObservedArtifactSourceDigest: input.ObservedArtifactSourceDigest,
		ObservedIndependentDigest:    input.ObservedIndependentDigest,
	})
	if err != nil {
		return DecisionResult{}, err
	}
	command := exec.CommandContext(ctx, "go", "run", judgePath)
	command.Dir = work
	command.Stdin = bytes.NewReader(payload)
	command.Env = append(os.Environ(), "GO111MODULE=off", "GOTOOLCHAIN=go1.27.0")
	output, err := command.Output()
	if err != nil {
		return DecisionResult{}, fmt.Errorf("execute generated judge: %w", err)
	}
	var result DecisionResult
	decoder := json.NewDecoder(bytes.NewReader(output))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&result); err != nil {
		return DecisionResult{}, fmt.Errorf("decode generated judge: %w", err)
	}
	if result.CaseID != input.ID {
		return DecisionResult{}, errors.New("generated judge changed case identity")
	}
	return result, nil
}
