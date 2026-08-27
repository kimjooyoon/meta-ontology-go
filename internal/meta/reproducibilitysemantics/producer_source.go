package reproducibilitysemantics

import (
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func deriveProducerCases(path string, source []byte) ([]declaredCase, string, error) {
	file, diagnostics := syntax.ParseFile(path, string(source))
	if file == nil || diagnostics.HasErrors() {
		return nil, "", fmt.Errorf("parse Gooo source: %d diagnostics", len(diagnostics))
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return nil, "", fmt.Errorf("lower Gooo source: %w", err)
	}
	if err := ir.Validate(); err != nil {
		return nil, "", fmt.Errorf("validate Gooo semantic IR: %w", err)
	}
	var cases []declaredCase
	for _, declaration := range file.Decls {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok || !activity.ValueProgramPresent {
			continue
		}
		node, ok := ir.Graph.NodeByName(ir.Namespace, activity.Name)
		if !ok || node.ValueProgram != activity.ValueProgram {
			return nil, "", fmt.Errorf("lowered value mismatch for %q", activity.Name)
		}
		item, err := parseProducerProgram(node.ValueProgram)
		if err != nil {
			return nil, "", fmt.Errorf("activity %q: %w", activity.Name, err)
		}
		cases = append(cases, item)
	}
	if len(cases) != CaseCount {
		return nil, "", fmt.Errorf("Gooo case count %d, want %d", len(cases), CaseCount)
	}
	return cases, "sha256:" + ir.StableHash(), nil
}

func parseProducerProgram(program string) (declaredCase, error) {
	item := declaredCase{}
	seen := make(map[string]bool)
	for _, field := range strings.Split(program, ";") {
		pair := strings.SplitN(field, "=", 2)
		if len(pair) != 2 || seen[pair[0]] {
			return declaredCase{}, fmt.Errorf("invalid value field %q", field)
		}
		seen[pair[0]] = true
		switch pair[0] {
		case "case":
			item.ID = pair[1]
		case "byte.reference":
			item.ByteReference = pair[1]
		case "byte.candidate":
			item.ByteCandidate = pair[1]
		case "meaning.expected":
			item.MeaningExpected = pair[1]
		case "meaning.observed":
			item.MeaningObserved = pair[1]
		default:
			return declaredCase{}, fmt.Errorf("unknown value field %q", pair[0])
		}
	}
	if item.ID == "" || len(seen) != 5 {
		return declaredCase{}, fmt.Errorf("value program does not declare five case fields")
	}
	return item, nil
}
