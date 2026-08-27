package producer

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/model"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type sourceFixture struct {
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

type SourceCase struct {
	CaseID     string `json:"case_id"`
	ActivityID string `json:"activity_id"`
	CaseKind   string `json:"case_kind"`
}

// Discover derives the case inventory from executable Transformation values in
// the Gooo source. Validator expectations do not participate in discovery.
func Discover(source []byte) ([]SourceCase, error) {
	file, diagnostics := syntax.ParseFile(model.SourcePath, string(source))
	if diagnostics.HasErrors() {
		return nil, fmt.Errorf("parse invariant transformation source: %s", diagnostics.Error())
	}
	seen := map[string]bool{}
	result := make([]SourceCase, 0, 4)
	for _, declaration := range file.Declarations {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok || len(activity.Parameters) != 0 || activity.Result.Name != "Transformation" || !activity.ValueProgramPresent {
			continue
		}
		fields, err := parseFixtureProgram(activity.ValueProgram)
		if err != nil {
			return nil, fmt.Errorf("activity %s: %w", activity.Name, err)
		}
		caseID := fields["case"]
		if seen[caseID] {
			return nil, fmt.Errorf("duplicate fixture case %q", caseID)
		}
		seen[caseID] = true
		result = append(result, SourceCase{CaseID: caseID, ActivityID: activity.Name, CaseKind: fields["kind"]})
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("source has no executable Transformation fixtures")
	}
	return result, nil
}

func parseSourceFixture(source []byte, caseID string) (sourceFixture, error) {
	file, diagnostics := syntax.ParseFile(model.SourcePath, string(source))
	if diagnostics.HasErrors() {
		return sourceFixture{}, fmt.Errorf("parse invariant transformation source: %s", diagnostics.Error())
	}
	semanticSourceDigest, err := semanticDigest(file)
	if err != nil {
		return sourceFixture{}, err
	}
	for _, declaration := range file.Declarations {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok || len(activity.Parameters) != 0 || activity.Result.Name != "Transformation" || !activity.ValueProgramPresent {
			continue
		}
		fields, err := parseFixtureProgram(activity.ValueProgram)
		if err != nil {
			return sourceFixture{}, fmt.Errorf("activity %s: %w", activity.Name, err)
		}
		if fields["case"] != caseID {
			continue
		}
		input, err := parseIntField(fields, "input")
		if err != nil {
			return sourceFixture{}, err
		}
		expected, err := parseIntField(fields, "expected")
		if err != nil {
			return sourceFixture{}, err
		}
		candidateResult, err := executeCandidate(fields["candidate"], input)
		if err != nil {
			return sourceFixture{}, err
		}
		if fields["invariant"] != "candidate-output-equals-expected" || fields["invariant-id"] != model.InvariantID || fields["domain"] != model.InputDomainID {
			return sourceFixture{}, fmt.Errorf("case %q is outside the bounded source contract", caseID)
		}
		if fields["effect"] != "none" && fields["effect"] != "approved-artifact" {
			return sourceFixture{}, fmt.Errorf("unsupported effect value %q", fields["effect"])
		}
		if fields["replay"] == "" {
			return sourceFixture{}, fmt.Errorf("case %q has an empty replay recipe", caseID)
		}
		return sourceFixture{
			ActivityID: activity.Name, CaseID: caseID, CaseKind: fields["kind"], Input: input,
			CandidateOperation: fields["candidate"], CandidateResult: candidateResult, Expected: expected,
			Invariant: fields["invariant"], InvariantID: fields["invariant-id"], DomainID: fields["domain"],
			OperationID: model.Digest([]string{"operation", fields["candidate"]}), ReplayRecipe: fields["replay"],
			EffectIntent: fields["effect"], SemanticSourceDigest: semanticSourceDigest,
		}, nil
	}
	return sourceFixture{}, fmt.Errorf("fixture case %q is missing from source", caseID)
}

func semanticDigest(file *syntax.File) (string, error) {
	ir, err := bidir.Lower(file)
	if err != nil {
		return "", fmt.Errorf("canonical semantic lowering: %w", err)
	}
	return "sha256:" + ir.StableHash(), nil
}

func parseFixtureProgram(program string) (map[string]string, error) {
	parts := strings.Split(program, ";")
	if len(parts) != 10 {
		return nil, fmt.Errorf("computes value has %d fields, want 10", len(parts))
	}
	fields := make(map[string]string, len(parts))
	for _, part := range parts {
		key, value, ok := strings.Cut(part, "=")
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		if !ok || key == "" || value == "" || fields[key] != "" {
			return nil, fmt.Errorf("invalid or duplicate field %q", part)
		}
		fields[key] = value
	}
	for _, key := range []string{"case", "kind", "input", "candidate", "expected", "invariant", "invariant-id", "domain", "replay", "effect"} {
		if fields[key] == "" {
			return nil, fmt.Errorf("missing field %q", key)
		}
	}
	return fields, nil
}

func parseIntField(fields map[string]string, key string) (int64, error) {
	value, err := strconv.ParseInt(fields[key], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("field %q is not int64: %w", key, err)
	}
	return value, nil
}

func executeCandidate(operation string, input int64) (int64, error) {
	if !strings.HasPrefix(operation, "add:") || strings.Count(operation, ":") != 1 {
		return 0, fmt.Errorf("unsupported candidate operation %q", operation)
	}
	operand, err := strconv.ParseInt(strings.TrimPrefix(operation, "add:"), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("candidate operation %q has invalid operand: %w", operation, err)
	}
	const maxInt64 = int64(1<<63 - 1)
	const minInt64 = -maxInt64 - 1
	if (operand > 0 && input > maxInt64-operand) || (operand < 0 && input < minInt64-operand) {
		return 0, fmt.Errorf("candidate operation %q overflows int64", operation)
	}
	return input + operand, nil
}

func executeReplay(recipe string, input int64) (int64, error) {
	if recipe == "unavailable" {
		return 0, fmt.Errorf("REGRESSION_REPLAY_RECIPE_UNAVAILABLE")
	}
	return executeCandidate(recipe, input)
}
