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
	"strings"
)

// GenerateJudge emits a standalone Go program. It has no import path into the
// repository, so the execution evidence is independent of this package.
func GenerateJudge(policy CompiledPolicy) []byte {
	reduction := makeReductionLiteral(policy.Reduction)
	return []byte(fmt.Sprintf(`package main

import (
	"encoding/json"
	"os"
	"regexp"
)

type input struct {
	ID string %q
	ProducerAvailable bool %q
	ConsumerAvailable bool %q
	ObservedSourceDigest string %q
	ObservedArtifactSourceDigest string %q
	ObservedGeneratedJudgeDigest string %q
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

type decisionRule struct {
	Condition string
	Decision string
	Stage string
	Step int
	Reason string
}

const policyDigest = %q
const semanticDigest = %q
const policyDenominator = %d
const fixedDenominator = %d
var digestPattern = regexp.MustCompile("^sha256:[0-9a-f]{64}$")

var reduction = []decisionRule{%s}

func matches(condition string, value input) bool {
	available := value.ProducerAvailable && value.ConsumerAvailable
	valid := func(value string) bool { return digestPattern.MatchString(value) }
	empty := value.ObservedSourceDigest == "" || value.ObservedArtifactSourceDigest == "" || value.ObservedGeneratedJudgeDigest == "" || value.ObservedIndependentDigest == ""
	malformed := !valid(value.ObservedSourceDigest) || !valid(value.ObservedArtifactSourceDigest) || !valid(value.ObservedGeneratedJudgeDigest) || !valid(value.ObservedIndependentDigest)
	switch condition {
	case "EVIDENCE_UNAVAILABLE":
		return !available
	case "DIGEST_UNAVAILABLE":
		return available && empty
	case "MALFORMED_DIGEST":
		return available && !empty && malformed
	case "SOURCE_DIGEST_MISMATCH":
		return available && !empty && !malformed && value.ObservedSourceDigest != policyDigest
	case "ARTIFACT_SOURCE_MISMATCH":
		return available && !empty && !malformed && value.ObservedSourceDigest == policyDigest && value.ObservedArtifactSourceDigest != policyDigest
	case "INDEPENDENT_SOURCE_MISMATCH":
		return available && !empty && !malformed && value.ObservedSourceDigest == policyDigest && value.ObservedArtifactSourceDigest == policyDigest && value.ObservedIndependentDigest != semanticDigest
	case "SEMANTIC_EQUIVALENCE":
		return available && !empty && !malformed && value.ObservedSourceDigest == policyDigest && value.ObservedArtifactSourceDigest == policyDigest && value.ObservedIndependentDigest == semanticDigest
	default:
		return false
	}
}

func main() {
	var value input
	decoder := json.NewDecoder(os.Stdin)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil { os.Exit(2) }
	output := result{CaseID: value.ID, PolicyDigest: policyDigest, SemanticDigest: semanticDigest, Denominator: policyDenominator}
	if policyDenominator != fixedDenominator {
		output.Decision, output.Stage, output.Step, output.Reason = "FAIL_CLOSED", "COMPILE", 3, "FIXED_DENOMINATOR_CHANGED"
	} else {
		matched := false
		for _, rule := range reduction {
			if matches(rule.Condition, value) {
				output.Decision, output.Stage, output.Step, output.Reason = rule.Decision, rule.Stage, rule.Step, rule.Reason
				matched = true
				break
			}
		}
		if !matched {
			output.Decision, output.Stage, output.Step, output.Reason = "FAIL_CLOSED", "COMPILE", 3, "NO_REDUCTION_RULE_MATCHED"
		}
	}
	if err := json.NewEncoder(os.Stdout).Encode(output); err != nil { os.Exit(3) }
}
`, `json:"id"`, `json:"producer_available"`, `json:"consumer_available"`, `json:"observed_source_digest"`, `json:"observed_artifact_source_digest"`, `json:"observed_generated_judge_digest"`, `json:"observed_independent_digest"`, `json:"case_id"`, `json:"decision"`, `json:"stage"`, `json:"step"`, `json:"reason"`, `json:"policy_digest"`, `json:"semantic_digest"`, `json:"fixed_denominator"`, policy.SourceDigest, policy.SemanticDigest, policy.Denominator, FixedDenominator, reduction))
}

func makeReductionLiteral(reduction DecisionReduction) string {
	var builder strings.Builder
	for _, rule := range reduction.Rules {
		fmt.Fprintf(&builder, "{Condition:%q, Decision:%q, Stage:%q, Step:%d, Reason:%q},", rule.Condition, rule.Decision, rule.Stage, rule.Step, rule.Reason)
	}
	return builder.String()
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
		ObservedGeneratedJudgeDigest string `json:"observed_generated_judge_digest"`
		ObservedIndependentDigest    string `json:"observed_independent_digest"`
	}{
		ID: input.ID, ProducerAvailable: input.ProducerAvailable, ConsumerAvailable: input.ConsumerAvailable,
		ObservedSourceDigest:         input.ObservedSourceDigest,
		ObservedArtifactSourceDigest: input.ObservedArtifactSourceDigest,
		ObservedGeneratedJudgeDigest: input.ObservedGeneratedJudgeDigest,
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
