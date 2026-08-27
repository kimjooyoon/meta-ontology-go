package ambiguitybudget

import (
	"fmt"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const canonicalLowering = "syntax.ParseFile->bidir.Lower"

func observeSource(path string, raw []byte) (SourceObservation, error) {
	file, diagnostics := syntax.ParseFile(path, string(raw))
	if file == nil || diagnostics.HasErrors() {
		return SourceObservation{Path: path, Digest: digestBytes(raw), Lowering: canonicalLowering}, fmt.Errorf("source syntax is unknown")
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return SourceObservation{Path: path, Digest: digestBytes(raw), Lowering: canonicalLowering}, fmt.Errorf("source semantic lowering is unknown: %w", err)
	}
	if err := ir.Validate(); err != nil {
		return SourceObservation{Path: path, Digest: digestBytes(raw), Lowering: canonicalLowering}, fmt.Errorf("source semantic validation is unknown: %w", err)
	}

	observation := SourceObservation{Path: path, Digest: digestBytes(raw), SemanticDigest: "sha256:" + ir.StableHash(), Lowering: canonicalLowering,
		Package: ir.Package, Namespace: ir.Namespace.String()}
	for _, node := range ir.Graph.Nodes() {
		switch node.Kind {
		case semantic.Entity:
			observation.Entities++
		case semantic.Activity:
			observation.Activities++
			if node.ValueProgram == "" {
				return observation, fmt.Errorf("activity %q has no computes program", node.Name)
			}
			program, parseErr := parseComputesProgram(node.Name, node.ValueProgram)
			if parseErr != nil {
				return observation, fmt.Errorf("activity %q computes: %w", node.Name, parseErr)
			}
			if program.Kind == "CASE" {
				program.Class = derivedClass(program, expectedBudget())
				program.InputState = inputState(program)
			}
			observation.Programs = append(observation.Programs, ProgramObservation{
				Activity: program.Activity, Program: program.Text, ProgramKind: program.Kind, ID: program.ID,
				Class: program.Class, InputState: program.InputState, Counts: program.Counts,
				UnobservedDimensions: append([]string(nil), program.UnobservedDimensions...), Digest: digestBytes([]byte(program.Text)),
			})
		}
	}
	sort.Slice(observation.Programs, func(i, j int) bool { return observation.Programs[i].Activity < observation.Programs[j].Activity })
	return observation, nil
}

func findBudget(source SourceObservation, activity string) (ProgramObservation, bool) {
	for _, program := range source.Programs {
		if program.ProgramKind == "BUDGET" && program.Activity == activity {
			return program, true
		}
	}
	return ProgramObservation{}, false
}

func findCase(source SourceObservation, activity string) (ProgramObservation, bool) {
	for _, program := range source.Programs {
		if program.ProgramKind == "CASE" && program.Activity == activity {
			return program, true
		}
	}
	return ProgramObservation{}, false
}
