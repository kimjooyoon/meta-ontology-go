package judge

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/model"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type sourceSemantics struct {
	CaseID              string
	Input               int64
	CandidateOperation  string
	CandidateResult     int64
	Expected            int64
	Invariant           string
	RegressionAvailable bool
	ApprovedArtifact    bool
}

func parseSourceSemantics(source []byte, spec model.CaseSpec) (sourceSemantics, error) {
	file, diagnostics := syntax.ParseFile(model.SourcePath, string(source))
	if diagnostics.HasErrors() {
		return sourceSemantics{}, fmt.Errorf("source syntax: %s", diagnostics.Error())
	}
	var fixture *syntax.ActivityDecl
	for _, declaration := range file.Declarations {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok || activity.Name != spec.Activity {
			continue
		}
		if fixture != nil {
			return sourceSemantics{}, fmt.Errorf("duplicate source activity %q", spec.Activity)
		}
		fixture = activity
	}
	if fixture == nil || len(fixture.Parameters) != 0 || fixture.Result.Name != "Transformation" || !fixture.ValueProgramPresent {
		return sourceSemantics{}, fmt.Errorf("source activity %q is not an executable fixture", spec.Activity)
	}
	fields, err := decodeFixtureValue(fixture.ValueProgram)
	if err != nil {
		return sourceSemantics{}, err
	}
	if fields["case"] != spec.ID {
		return sourceSemantics{}, fmt.Errorf("source activity %q case binding is %q", spec.Activity, fields["case"])
	}
	input, err := strconv.ParseInt(fields["input"], 10, 64)
	if err != nil {
		return sourceSemantics{}, fmt.Errorf("source input is not int64: %w", err)
	}
	expected, err := strconv.ParseInt(fields["expected"], 10, 64)
	if err != nil {
		return sourceSemantics{}, fmt.Errorf("source expected value is not int64: %w", err)
	}
	result, err := evaluateAdd(fields["candidate"], input)
	if err != nil {
		return sourceSemantics{}, err
	}
	if fields["invariant"] != "candidate-output-equals-expected" {
		return sourceSemantics{}, fmt.Errorf("source invariant %q is not supported", fields["invariant"])
	}
	semantics := sourceSemantics{CaseID: fields["case"], Input: input, CandidateOperation: fields["candidate"], CandidateResult: result,
		Expected: expected, Invariant: fields["invariant"], RegressionAvailable: fields["replay"] == "present", ApprovedArtifact: fields["effect"] == "approved-artifact"}
	if fields["replay"] != "present" && fields["replay"] != "missing" {
		return sourceSemantics{}, fmt.Errorf("source replay value %q is not supported", fields["replay"])
	}
	if fields["effect"] != "none" && fields["effect"] != "approved-artifact" {
		return sourceSemantics{}, fmt.Errorf("source effect value %q is not supported", fields["effect"])
	}
	return semantics, nil
}

func decodeFixtureValue(program string) (map[string]string, error) {
	parts := strings.Split(program, ";")
	want := []string{"case", "input", "candidate", "expected", "invariant", "replay", "effect"}
	if len(parts) != len(want) {
		return nil, fmt.Errorf("source fixture has %d fields, want %d", len(parts), len(want))
	}
	fields := make(map[string]string, len(parts))
	for _, part := range parts {
		key, value, ok := strings.Cut(part, "=")
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		if !ok || key == "" || value == "" {
			return nil, fmt.Errorf("source fixture field %q is malformed", part)
		}
		if _, exists := fields[key]; exists {
			return nil, fmt.Errorf("source fixture field %q is duplicated", key)
		}
		fields[key] = value
	}
	for _, key := range want {
		if fields[key] == "" {
			return nil, fmt.Errorf("source fixture field %q is missing", key)
		}
	}
	return fields, nil
}

func evaluateAdd(operation string, input int64) (int64, error) {
	name, operandText, ok := strings.Cut(operation, ":")
	if !ok || name != "add" || operandText == "" || strings.Contains(operandText, ":") {
		return 0, fmt.Errorf("source candidate operation %q is unsupported", operation)
	}
	operand, err := strconv.ParseInt(operandText, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("source candidate operand is not int64: %w", err)
	}
	const maxInt64 = int64(1<<63 - 1)
	const minInt64 = -maxInt64 - 1
	if (operand > 0 && input > maxInt64-operand) || (operand < 0 && input < minInt64-operand) {
		return 0, fmt.Errorf("source candidate operation %q overflows int64", operation)
	}
	return input + operand, nil
}
