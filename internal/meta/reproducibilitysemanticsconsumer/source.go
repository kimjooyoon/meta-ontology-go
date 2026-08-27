package reproducibilitysemanticsconsumer

import (
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type declaredCase struct {
	ID              string
	ByteReference   string
	ByteCandidate   string
	MeaningExpected string
	MeaningObserved string
}

func deriveCases(path string, source []byte) ([]declaredCase, string, error) {
	file, diagnostics := syntax.ParseFile(path, string(source))
	if file == nil || diagnostics.HasErrors() {
		return nil, "", fmt.Errorf("judge parse diagnostics: %d", len(diagnostics))
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return nil, "", fmt.Errorf("judge lower source: %w", err)
	}
	if err := ir.Validate(); err != nil {
		return nil, "", fmt.Errorf("judge validate IR: %w", err)
	}
	var cases []declaredCase
	for _, declaration := range file.Declarations {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok || !activity.ValueProgramPresent {
			continue
		}
		node, ok := ir.Graph.NodeByName(ir.Namespace, activity.Name)
		if !ok || node.ValueProgram != activity.ValueProgram {
			return nil, "", fmt.Errorf("judge lowered value mismatch for %q", activity.Name)
		}
		item, err := parseProgram(node.ValueProgram)
		if err != nil {
			return nil, "", fmt.Errorf("judge activity %q: %w", activity.Name, err)
		}
		cases = append(cases, item)
	}
	if len(cases) != CaseCount {
		return nil, "", fmt.Errorf("judge case count %d, want %d", len(cases), CaseCount)
	}
	return cases, "sha256:" + ir.StableHash(), nil
}

func parseProgram(program string) (declaredCase, error) {
	values := map[string]string{}
	for field := range strings.SplitSeq(program, ";") {
		pair := strings.SplitN(field, "=", 2)
		if len(pair) != 2 || pair[0] == "" {
			return declaredCase{}, fmt.Errorf("malformed source value %q", field)
		}
		if _, exists := values[pair[0]]; exists {
			return declaredCase{}, fmt.Errorf("duplicate source value %q", pair[0])
		}
		values[pair[0]] = pair[1]
	}
	for _, key := range []string{"case", "byte.reference", "byte.candidate", "meaning.expected", "meaning.observed"} {
		if _, ok := values[key]; !ok {
			return declaredCase{}, fmt.Errorf("source value field %q is missing", key)
		}
	}
	if len(values) != 5 || values["case"] == "" {
		return declaredCase{}, fmt.Errorf("source value fields incomplete")
	}
	return declaredCase{ID: values["case"], ByteReference: values["byte.reference"], ByteCandidate: values["byte.candidate"],
		MeaningExpected: values["meaning.expected"], MeaningObserved: values["meaning.observed"]}, nil
}

func digestText(value string) string {
	if value == "" {
		return ""
	}
	return digestBytes([]byte(value))
}

func bindsSource(item Case, declared declaredCase) bool {
	return item.ID == declared.ID && item.Byte.Reference == digestText(declared.ByteReference) &&
		item.Byte.Candidate == digestText(declared.ByteCandidate) && item.Meaning.Expected == digestText(declared.MeaningExpected) &&
		item.Meaning.Observed == digestText(declared.MeaningObserved)
}
