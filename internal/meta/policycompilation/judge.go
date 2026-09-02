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

type generatedJudgeInput struct {
	ID                           string `json:"id"`
	ProducerAvailable            bool   `json:"producer_available"`
	ConsumerAvailable            bool   `json:"consumer_available"`
	ObservedSourceDigest         string `json:"observed_source_digest"`
	ObservedArtifactSourceDigest string `json:"observed_artifact_source_digest"`
	ObservedGeneratedJudgeDigest string `json:"observed_generated_judge_digest"`
	ObservedIndependentDigest    string `json:"observed_independent_digest"`
	UpperDecision                string `json:"upper_decision"`
}

// GenerateJudge emits a standalone Go program containing the reduction rows
// read from Gooo. It has no import path into this repository, which makes its
// execution boundary independently inspectable by the consumer.
func GenerateJudge(policy CompiledPolicy) []byte {
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
    UpperDecision string %q
}
type result struct {
    CaseID string %q
    Decision string %q
    Stage string %q
    Step int %q
    Reason string %q
    UnknownClass string %q
    NextOperation string %q
    BlockedBy []string %q
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
    UnknownClass string
    NextOperation string
    BlockedBy []string
}

const policyDigest = %q
const semanticDigest = %q
const policyDenominator = %d
const fixedDenominator = %d
var digestPattern = regexp.MustCompile("^sha256:[0-9a-f]{64}$")
var reduction = []decisionRule{%s}

func knownDecision(value string) bool {
    return value == "PASS" || value == "FAIL_CLOSED" || value == "UNKNOWN"
}
func matches(condition string, value input) bool {
    sourceOK := digestPattern.MatchString(value.ObservedSourceDigest)
    artifactOK := digestPattern.MatchString(value.ObservedArtifactSourceDigest)
    independentOK := digestPattern.MatchString(value.ObservedIndependentDigest)
    judgeOK := digestPattern.MatchString(value.ObservedGeneratedJudgeDigest)
    complete := sourceOK && artifactOK && independentOK && judgeOK
    empty := value.ObservedSourceDigest == "" || value.ObservedArtifactSourceDigest == "" || value.ObservedGeneratedJudgeDigest == "" || value.ObservedIndependentDigest == ""
    ready := value.ProducerAvailable && value.ConsumerAvailable
    switch condition {
    case "UNRECOGNIZED_TOP_LEVEL_DECISION":
        return value.UpperDecision != "" && !knownDecision(value.UpperDecision)
    case "SOURCE_DIGEST_MISMATCH":
        return sourceOK && value.ObservedSourceDigest != policyDigest
    case "ARTIFACT_SOURCE_MISMATCH":
        return sourceOK && artifactOK && value.ObservedSourceDigest == policyDigest && value.ObservedArtifactSourceDigest != policyDigest
    case "INDEPENDENT_SOURCE_MISMATCH":
        return sourceOK && artifactOK && independentOK && value.ObservedSourceDigest == policyDigest && value.ObservedArtifactSourceDigest == policyDigest && value.ObservedIndependentDigest != semanticDigest
    case "EVIDENCE_UNAVAILABLE":
        return !ready && !sourceOK && !artifactOK && !independentOK && !judgeOK
    case "DIGEST_UNAVAILABLE":
        return ready && empty
    case "MALFORMED_DIGEST":
        return ready && !empty && !complete
    case "SEMANTIC_EQUIVALENCE":
        return ready && complete && value.ObservedSourceDigest == policyDigest && value.ObservedArtifactSourceDigest == policyDigest && value.ObservedIndependentDigest == semanticDigest
    default:
        return false
    }
}
func main() {
    var value input
    decoder := json.NewDecoder(os.Stdin)
    decoder.DisallowUnknownFields()
    if err := decoder.Decode(&value); err != nil { os.Exit(2) }
    output := result{CaseID:value.ID, PolicyDigest:policyDigest, SemanticDigest:semanticDigest, Denominator:policyDenominator, BlockedBy:[]string{}}
    if policyDenominator != fixedDenominator {
        output.Decision, output.Stage, output.Reason = "FAIL_CLOSED", "COMPILE", "FIXED_DENOMINATOR_CHANGED"
    } else {
        matched := false
        for _, row := range reduction {
            if matches(row.Condition, value) {
                output.Decision, output.Stage, output.Step, output.Reason = row.Decision, row.Stage, row.Step, row.Reason
                output.UnknownClass, output.NextOperation, output.BlockedBy = row.UnknownClass, row.NextOperation, append([]string(nil), row.BlockedBy...)
                matched = true
                break
            }
        }
        if !matched { output.Decision, output.Stage, output.Reason = "FAIL_CLOSED", "COMPILE", "NO_REDUCTION_RULE_MATCHED" }
    }
    if output.Decision != "UNKNOWN" { output.UnknownClass, output.NextOperation, output.BlockedBy = "", "", []string{} }
    if err := json.NewEncoder(os.Stdout).Encode(output); err != nil { os.Exit(3) }
}
`,
		`json:"id"`, `json:"producer_available"`, `json:"consumer_available"`,
		`json:"observed_source_digest"`, `json:"observed_artifact_source_digest"`,
		`json:"observed_generated_judge_digest"`, `json:"observed_independent_digest"`, `json:"upper_decision"`,
		`json:"case_id"`, `json:"decision"`, `json:"stage"`, `json:"step"`, `json:"reason"`,
		`json:"unknown_class"`, `json:"next_operation"`, `json:"blocked_by"`, `json:"policy_digest"`,
		`json:"semantic_digest"`, `json:"fixed_denominator"`, policy.SourceDigest, policy.SemanticDigest,
		policy.Denominator, FixedDenominator, reductionLiteral(policy.Reduction)))
}

func reductionLiteral(reduction DecisionReduction) string {
	var builder strings.Builder
	for _, row := range reduction.Rules {
		fmt.Fprintf(&builder, "{Condition:%q,Decision:%q,Stage:%q,Step:%d,Reason:%q,UnknownClass:%q,NextOperation:%q,BlockedBy:%#v},", row.Condition, row.Decision, row.Stage, row.Step, row.Reason, row.UnknownClass, row.NextOperation, row.BlockedBy)
	}
	return builder.String()
}

// ExecuteGenerated runs the generated source in an isolated temporary
// directory. The caller must supply a runner-temp output root in CI; this
// function never writes into the repository.
func ExecuteGenerated(ctx context.Context, judgeSource []byte, input Case) (DecisionResult, error) {
	work, err := os.MkdirTemp("", "gooo-policy-judge-")
	if err != nil {
		return DecisionResult{}, err
	}
	defer os.RemoveAll(work)
	path := filepath.Join(work, "judge.go")
	if err := os.WriteFile(path, judgeSource, 0o600); err != nil {
		return DecisionResult{}, err
	}
	payload, err := json.Marshal(generatedJudgeInput{
		ID:                           input.ID,
		ProducerAvailable:            input.ProducerAvailable,
		ConsumerAvailable:            input.ConsumerAvailable,
		ObservedSourceDigest:         input.ObservedSourceDigest,
		ObservedArtifactSourceDigest: input.ObservedArtifactSourceDigest,
		ObservedGeneratedJudgeDigest: input.ObservedGeneratedJudgeDigest,
		ObservedIndependentDigest:    input.ObservedIndependentDigest,
		UpperDecision:                input.UpperDecision,
	})
	if err != nil {
		return DecisionResult{}, err
	}
	command := exec.CommandContext(ctx, "go", "run", path)
	command.Dir = work
	command.Stdin = bytes.NewReader(payload)
	command.Env = append(os.Environ(), "GO111MODULE=off", "GOTOOLCHAIN=go1.27.0")
	output, err := command.CombinedOutput()
	if err != nil {
		return DecisionResult{}, fmt.Errorf("execute generated judge: %w: %s", err, strings.TrimSpace(string(output)))
	}
	decoder := json.NewDecoder(bytes.NewReader(output))
	decoder.DisallowUnknownFields()
	var result DecisionResult
	if err := decoder.Decode(&result); err != nil {
		return DecisionResult{}, fmt.Errorf("decode generated judge: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err == nil {
		return DecisionResult{}, errors.New("generated judge emitted trailing JSON")
	}
	if result.CaseID != input.ID {
		return DecisionResult{}, errors.New("generated judge changed case identity")
	}
	return result, nil
}
