package producer

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/model"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type sourceFixture struct {
	CaseID             string
	Input              int64
	CandidateOperation string
	CandidateResult    int64
	Expected           int64
	Invariant          string
	ReplayRecipe       string
	ApprovedArtifact   bool
}

func parseSourceFixture(source []byte, spec model.CaseSpec) (sourceFixture, error) {
	file, diagnostics := syntax.ParseFile(model.SourcePath, string(source))
	if diagnostics.HasErrors() {
		return sourceFixture{}, fmt.Errorf("parse invariant transformation source: %s", diagnostics.Error())
	}
	var found *syntax.ActivityDecl
	for _, declaration := range file.Decls {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok || activity.Name != spec.Activity {
			continue
		}
		if found != nil {
			return sourceFixture{}, fmt.Errorf("duplicate fixture activity %q", spec.Activity)
		}
		found = activity
	}
	if found == nil {
		return sourceFixture{}, fmt.Errorf("fixture activity %q is missing", spec.Activity)
	}
	if len(found.Inputs) != 0 || found.Output != "Transformation" || !found.ValueProgramPresent {
		return sourceFixture{}, fmt.Errorf("fixture activity %q has an unsupported signature or missing computes value", spec.Activity)
	}
	fields, err := parseFixtureProgram(found.ValueProgram)
	if err != nil {
		return sourceFixture{}, fmt.Errorf("fixture activity %q: %w", spec.Activity, err)
	}
	if fields["case"] != spec.ID {
		return sourceFixture{}, fmt.Errorf("fixture activity %q is bound to case %q, want %q", spec.Activity, fields["case"], spec.ID)
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
	if fields["invariant"] != "candidate-output-equals-expected" {
		return sourceFixture{}, fmt.Errorf("unsupported invariant %q", fields["invariant"])
	}
	fixture := sourceFixture{CaseID: fields["case"], Input: input, CandidateOperation: fields["candidate"], CandidateResult: candidateResult,
		Expected: expected, Invariant: fields["invariant"], ReplayRecipe: fields["replay"], ApprovedArtifact: fields["effect"] == "approved-artifact"}
	if fields["replay"] == "" {
		return sourceFixture{}, fmt.Errorf("replay recipe is empty")
	}
	if fields["effect"] != "none" && fields["effect"] != "approved-artifact" {
		return sourceFixture{}, fmt.Errorf("unsupported effect value %q", fields["effect"])
	}
	return fixture, nil
}

func parseFixtureProgram(program string) (map[string]string, error) {
	parts := strings.Split(program, ";")
	if len(parts) != 7 {
		return nil, fmt.Errorf("computes value has %d fields, want 7", len(parts))
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
	for _, key := range []string{"case", "input", "candidate", "expected", "invariant", "replay", "effect"} {
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
