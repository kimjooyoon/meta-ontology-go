package judge

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/model"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type sourceSemantics struct {
	ActivityID           string
	CaseID               string
	CaseKind             string
	Input                int64
	CandidateOperation   string
	CandidateResult      int64
	Expected             int64
	Invariant            string
	InvariantID          string
	DomainID             string
	OperationID          string
	ReplayRecipe         string
	EffectIntent         string
	SemanticSourceDigest string
}

func parseSourceSemantics(source []byte, caseID string) (sourceSemantics, error) {
	file, diagnostics := syntax.ParseFile(model.SourcePath, string(source))
	if diagnostics.HasErrors() {
		return sourceSemantics{}, fmt.Errorf("judge source syntax: %s", diagnostics.Error())
	}
	semanticSourceDigest, err := semanticSourceDigest(file)
	if err != nil {
		return sourceSemantics{}, err
	}
	for _, declaration := range file.Declarations {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok || len(activity.Parameters) != 0 || activity.Result.Name != "Transformation" || !activity.ValueProgramPresent {
			continue
		}
		fields, err := decodeFixtureProgram(activity.ValueProgram)
		if err != nil {
			return sourceSemantics{}, fmt.Errorf("activity %s: %w", activity.Name, err)
		}
		if fields["case"] != caseID {
			continue
		}
		input, err := parseInt(fields, "input")
		if err != nil {
			return sourceSemantics{}, err
		}
		expected, err := parseInt(fields, "expected")
		if err != nil {
			return sourceSemantics{}, err
		}
		candidateResult, err := evaluateAdd(fields["candidate"], input)
		if err != nil {
			return sourceSemantics{}, err
		}
		if fields["invariant"] != "candidate-output-equals-expected" || fields["invariant-id"] != model.InvariantID || fields["domain"] != model.InputDomainID {
			return sourceSemantics{}, fmt.Errorf("case %q is outside the bounded source contract", caseID)
		}
		return sourceSemantics{ActivityID: activity.Name, CaseID: caseID, CaseKind: fields["kind"], Input: input,
			CandidateOperation: fields["candidate"], CandidateResult: candidateResult, Expected: expected, Invariant: fields["invariant"],
			InvariantID: fields["invariant-id"], DomainID: fields["domain"], OperationID: model.Digest([]string{"operation", fields["candidate"]}),
			ReplayRecipe: fields["replay"], EffectIntent: fields["effect"], SemanticSourceDigest: semanticSourceDigest}, nil
	}
	return sourceSemantics{}, fmt.Errorf("judge source case %q is missing", caseID)
}

func semanticSourceDigest(file *syntax.File) (string, error) {
	ir, err := bidir.Lower(file)
	if err != nil {
		return "", fmt.Errorf("judge source lowering: %w", err)
	}
	return "sha256:" + ir.StableHash(), nil
}

func decodeFixtureProgram(program string) (map[string]string, error) {
	parts := strings.Split(program, ";")
	if len(parts) != 10 {
		return nil, fmt.Errorf("fixture computes value has %d fields, want 10", len(parts))
	}
	fields := make(map[string]string, len(parts))
	for _, part := range parts {
		key, value, ok := strings.Cut(part, "=")
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		if !ok || key == "" || value == "" {
			return nil, fmt.Errorf("fixture field %q is malformed", part)
		}
		if _, exists := fields[key]; exists {
			return nil, fmt.Errorf("fixture field %q is duplicated", key)
		}
		fields[key] = value
	}
	for _, key := range []string{"case", "kind", "input", "candidate", "expected", "invariant", "invariant-id", "domain", "replay", "effect"} {
		if fields[key] == "" {
			return nil, fmt.Errorf("fixture field %q is missing", key)
		}
	}
	return fields, nil
}

func parseInt(fields map[string]string, key string) (int64, error) {
	value, err := strconv.ParseInt(fields[key], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("source field %q is not int64: %w", key, err)
	}
	return value, nil
}

func evaluateAdd(operation string, input int64) (int64, error) {
	name, operandText, ok := strings.Cut(operation, ":")
	if !ok || name != "add" || operandText == "" || strings.Contains(operandText, ":") {
		return 0, fmt.Errorf("operation %q is unsupported", operation)
	}
	operand, err := strconv.ParseInt(operandText, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("operation %q operand is not int64: %w", operation, err)
	}
	const maxInt64 = int64(1<<63 - 1)
	const minInt64 = -maxInt64 - 1
	if (operand > 0 && input > maxInt64-operand) || (operand < 0 && input < minInt64-operand) {
		return 0, fmt.Errorf("operation %q overflows int64", operation)
	}
	return input + operand, nil
}

func executeReplay(recipe string, input int64) (int64, error) {
	if recipe == "unavailable" {
		return 0, fmt.Errorf("REGRESSION_REPLAY_RECIPE_UNAVAILABLE")
	}
	return evaluateAdd(recipe, input)
}
