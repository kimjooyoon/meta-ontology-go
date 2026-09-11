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
	"reflect"
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

// generatedJudgeInputFields renders the input ABI instead of maintaining a
// second field-name, type, and JSON-tag list inside the generated template.
func generatedJudgeInputFields(inputType reflect.Type) string {
	var builder strings.Builder
	for field := range inputType.Fields() {
		fmt.Fprintf(&builder, "    %s %s %q\n", field.Name, field.Type.String(), string(field.Tag))
	}
	return builder.String()
}

// GenerateJudge emits a standalone Go program containing the reduction rows
// read from Gooo. It has no import path into this repository, which makes its
// execution boundary independently inspectable by the consumer.
func GenerateJudge(policy CompiledPolicy) []byte {
	return []byte(fmt.Sprintf(`package main

import (
    "bytes"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "errors"
    "io"
    "os"
    "reflect"
    "regexp"
    "sort"
    "strconv"
    "strings"
)

type input struct {
%s}
type result struct {
    CaseID string %q
    Decision string %q
    MatchedCondition string %q
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
// rejectDuplicateObjectKeys mirrors the package decoder's token-level
// preflight while remaining self-contained in the generated consumer.
func rejectDuplicateObjectKeys(data []byte) error {
    decoder := json.NewDecoder(bytes.NewReader(data))
    return scanJSONValue(decoder, "$")
}
func scanJSONValue(decoder *json.Decoder, path string) error {
    token, err := decoder.Token()
    if err != nil { return err }
    delimiter, ok := token.(json.Delim)
    if !ok { return nil }
    switch delimiter {
    case '{':
        seen := make(map[string]int64)
        for decoder.More() {
            keyToken, err := decoder.Token()
            if err != nil { return err }
            key, ok := keyToken.(string)
            if !ok { return errors.New("json object key is not a string") }
            keyOffset := decoder.InputOffset()
            if previousOffset, exists := seen[key]; exists {
                return duplicateObjectKeyError(path, key, keyOffset, previousOffset)
            }
            seen[key] = keyOffset
            if err := scanJSONValue(decoder, jsonKeyPath(path, key)); err != nil { return err }
        }
        closing, err := decoder.Token()
        if err != nil { return err }
        if closing != json.Delim('}') { return errors.New("json object did not close with an object delimiter") }
    case '[':
        for index := 0; decoder.More(); index++ {
            if err := scanJSONValue(decoder, jsonArrayPath(path, index)); err != nil { return err }
        }
        closing, err := decoder.Token()
        if err != nil { return err }
        if closing != json.Delim(']') { return errors.New("json array did not close with an array delimiter") }
    default:
        return errors.New("unexpected JSON delimiter")
    }
    return nil
}
func jsonKeyPath(path, key string) string {
    return path + "[" + strconv.Quote(key) + "]"
}
func jsonArrayPath(path string, index int) string {
    return path + "[" + strconv.Itoa(index) + "]"
}
func duplicateObjectKeyError(path, key string, keyOffset, previousOffset int64) error {
    return errors.New("duplicate JSON object key " + strconv.Quote(key) + " at " + jsonKeyPath(path, key) +
        " (byte offset " + strconv.FormatInt(keyOffset, 10) + "; first occurrence at byte offset " + strconv.FormatInt(previousOffset, 10) + ")")
}
func declaredInputBindings(raw []byte, value input) ([]map[string]any, error) {
    var supplied map[string]json.RawMessage
    if err := json.Unmarshal(raw, &supplied); err != nil { return nil, err }
    if supplied == nil { return nil, errors.New("declared input must be a JSON object") }
    shape := reflect.TypeOf(value)
    values := reflect.ValueOf(value)
    known := make(map[string]bool, shape.NumField())
    fields := make([]map[string]any, 0, shape.NumField())
    for index := 0; index < shape.NumField(); index++ {
        field := shape.Field(index)
        name := strings.Split(field.Tag.Get("json"), ",")[0]
        if name == "" || name == "-" || known[name] {
            return nil, errors.New("generated input has no unique explicit JSON field binding")
        }
        known[name] = true
        effective, err := json.Marshal(values.Field(index).Interface())
        if err != nil { return nil, err }
        binding := "MISSING_DEFAULTED"
        rawValue, present := supplied[name]
        if present {
            binding = "DECLARED"
            if bytes.Equal(bytes.TrimSpace(rawValue), []byte("null")) { binding = "NULL_DEFAULTED" }
        }
        record := map[string]any{
            "json_field": name, "go_type": field.Type.String(), "binding": binding,
            "effective": json.RawMessage(effective),
        }
        if present { record["supplied"] = rawValue }
        fields = append(fields, record)
    }
    unknown := []string{}
    for name := range supplied {
        if !known[name] { unknown = append(unknown, name) }
    }
    if len(unknown) != 0 {
        sort.Strings(unknown)
        return nil, errors.New("unknown declared input fields: " + strings.Join(unknown, ","))
    }
    return fields, nil
}
func declaredInputSchema() (map[string]any, error) {
    fields, err := declaredInputBindings([]byte("{}"), input{})
    if err != nil { return nil, err }
    for _, field := range fields {
        field["default"] = field["effective"]
        field["required"] = false
        field["nullable"] = true
        delete(field, "binding")
        delete(field, "effective")
    }
    return map[string]any{
        "schema": "gooo/generated-policy-input-schema/v1",
        "source_digest": policyDigest, "semantic_digest": semanticDigest,
        "evaluation_mode": "--declared-input",
        "fields": fields, "default_input": input{},
        "policy_evaluation_observed": false,
        "mutation_authority": 0, "promotion_authority": 0,
    }, nil
}
func inputDigest(value []byte) string {
    sum := sha256.Sum256(value)
    return "sha256:" + hex.EncodeToString(sum[:])
}
func main() {
    if len(os.Args) == 2 && os.Args[1] == "--input-schema" {
        schema, err := declaredInputSchema()
        if err != nil {
            io.WriteString(os.Stderr, err.Error()+"\n")
            os.Exit(2)
        }
        if err := json.NewEncoder(os.Stdout).Encode(schema); err != nil { os.Exit(3) }
        return
    }
    declared := len(os.Args) == 2 && os.Args[1] == "--declared-input"
    if len(os.Args) != 1 && !declared {
        io.WriteString(os.Stderr, "usage: judge [--declared-input | --input-schema]\n")
        os.Exit(2)
    }
    raw, err := io.ReadAll(os.Stdin)
    if err != nil { os.Exit(2) }
    if err := rejectDuplicateObjectKeys(raw); err != nil {
        io.WriteString(os.Stderr, err.Error()+"\n")
        os.Exit(2)
    }
    var value input
    decoder := json.NewDecoder(bytes.NewReader(raw))
    decoder.DisallowUnknownFields()
    if err := decoder.Decode(&value); err != nil { os.Exit(2) }
    var trailing any
    if err := decoder.Decode(&trailing); err != io.EOF { os.Exit(2) }
    var fields []map[string]any
    if declared {
        fields, err = declaredInputBindings(raw, value)
        if err != nil {
            io.WriteString(os.Stderr, err.Error()+"\n")
            os.Exit(2)
        }
    }
    output := result{CaseID:value.ID, PolicyDigest:policyDigest, SemanticDigest:semanticDigest, Denominator:policyDenominator, BlockedBy:[]string{}}
    if policyDenominator != fixedDenominator {
        output.Decision, output.Stage, output.Reason = "FAIL_CLOSED", "COMPILE", "FIXED_DENOMINATOR_CHANGED"
    } else {
        matched := false
        for _, row := range reduction {
            if matches(row.Condition, value) {
                output.Decision, output.MatchedCondition, output.Stage, output.Step, output.Reason = row.Decision, row.Condition, row.Stage, row.Step, row.Reason
                output.UnknownClass, output.NextOperation, output.BlockedBy = row.UnknownClass, row.NextOperation, append([]string(nil), row.BlockedBy...)
                matched = true
                break
            }
        }
        if !matched { output.Decision, output.Stage, output.Reason = "FAIL_CLOSED", "COMPILE", "NO_REDUCTION_RULE_MATCHED" }
    }
    if output.Decision != "UNKNOWN" { output.UnknownClass, output.NextOperation, output.BlockedBy = "", "", []string{} }
    if declared {
        effective, err := json.Marshal(value)
        if err != nil { os.Exit(3) }
        report := map[string]any{
            "schema": "gooo/generated-policy-declared-input/v1",
            "source_digest": policyDigest, "semantic_digest": semanticDigest,
            "input_digest": inputDigest(raw), "effective_input_digest": inputDigest(effective),
            "evaluation_scope": "CONDITIONAL_GENERATED_POLICY_EVALUATION",
            "external_evidence_state": "UNKNOWN_NOT_VERIFIED",
            "full_conformance_state": "UNKNOWN_NOT_EXECUTED",
            "generated_execution_observed": true,
            "execution_evidence_state": "SELF_REPORTED_NOT_ATTESTED",
            "mutation_authority": 0, "promotion_authority": 0,
            "generated_decision": output, "fields": fields,
        }
        if err := json.NewEncoder(os.Stdout).Encode(report); err != nil { os.Exit(3) }
        return
    }
    if err := json.NewEncoder(os.Stdout).Encode(output); err != nil { os.Exit(3) }
}
`,
		generatedJudgeInputFields(reflect.TypeFor[generatedJudgeInput]()),
		`json:"case_id"`, `json:"decision"`, `json:"matched_condition"`, `json:"stage"`, `json:"step"`, `json:"reason"`,
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
	var result DecisionResult
	if err := decodeStrictJSON(output, &result); err != nil {
		return DecisionResult{}, fmt.Errorf("decode generated judge: %w", err)
	}
	if result.CaseID != input.ID {
		return DecisionResult{}, errors.New("generated judge changed case identity")
	}
	return result, nil
}
