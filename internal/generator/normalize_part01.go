package generator

import (
	"fmt"
)

func normalizeIR(input SemanticIR) (SemanticIR, error) {
	if input.Package == "" {
		return SemanticIR{}, fmt.Errorf("generator: package is required")
	}
	if !isGoIdentifier(input.Package) {
		return SemanticIR{}, fmt.Errorf("generator: invalid Go package %q", input.Package)
	}
	result := copyIR(input)
	if err := normalizeImports(&result); err != nil {
		return SemanticIR{}, err
	}
	types, err := normalizeEntities(&result)
	if err != nil {
		return SemanticIR{}, err
	}
	if err := normalizeActivities(&result, types); err != nil {
		return SemanticIR{}, err
	}
	if err := validateTopLevelNames(result); err != nil {
		return SemanticIR{}, err
	}
	if err := validateStableIDs(result); err != nil {
		return SemanticIR{}, err
	}
	if err := validateGoTypes(result); err != nil {
		return SemanticIR{}, err
	}
	return result, nil
}

func copyIR(input SemanticIR) SemanticIR {
	result := input
	// Copying with a non-nil zero-length seed preserves the canonical empty
	// collection representation while folding canonicalization into the only
	// collection copy pass.
	result.Imports = append([]Import{}, input.Imports...)
	result.Entities = append([]Entity{}, input.Entities...)
	result.Activities = append([]Activity{}, input.Activities...)
	for index := range result.Entities {
		result.Entities[index].Fields = append([]Field{}, input.Entities[index].Fields...)
		for fieldIndex := range result.Entities[index].Fields {
			result.Entities[index].Fields[fieldIndex].Aliases = append([]string(nil), input.Entities[index].Fields[fieldIndex].Aliases...)
		}
	}
	for index := range result.Activities {
		result.Activities[index].Inputs = append([]Port{}, input.Activities[index].Inputs...)
		result.Activities[index].Outputs = append([]Port{}, input.Activities[index].Outputs...)
		result.Activities[index].Slots = append([]Slot{}, input.Activities[index].Slots...)
	}
	return result
}
