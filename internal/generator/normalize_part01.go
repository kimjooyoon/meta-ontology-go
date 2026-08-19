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
	canonicalizeIRCollections(&result)
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

// canonicalizeIRCollections gives semantically equivalent nil and empty
// collections one wire representation without changing caller-owned input.
// This keeps generated metadata digests independent of how an adapter
// materializes absent optional declarations.
func canonicalizeIRCollections(ir *SemanticIR) {
	ir.Imports = append([]Import{}, ir.Imports...)
	ir.Entities = append([]Entity{}, ir.Entities...)
	for index := range ir.Entities {
		ir.Entities[index].Fields = append([]Field{}, ir.Entities[index].Fields...)
	}
	ir.Activities = append([]Activity{}, ir.Activities...)
	for index := range ir.Activities {
		ir.Activities[index].Inputs = append([]Port{}, ir.Activities[index].Inputs...)
		ir.Activities[index].Outputs = append([]Port{}, ir.Activities[index].Outputs...)
		ir.Activities[index].Slots = append([]Slot{}, ir.Activities[index].Slots...)
	}
}
func copyIR(input SemanticIR) SemanticIR {
	result := input
	result.Imports = append([]Import(nil), input.Imports...)
	result.Entities = append([]Entity(nil), input.Entities...)
	result.Activities = append([]Activity(nil), input.Activities...)
	for index := range result.Entities {
		result.Entities[index].Fields = append([]Field(nil), input.Entities[index].Fields...)
		for fieldIndex := range result.Entities[index].Fields {
			result.Entities[index].Fields[fieldIndex].Aliases = append([]string(nil), input.Entities[index].Fields[fieldIndex].Aliases...)
		}
	}
	for index := range result.Activities {
		result.Activities[index].Inputs = append([]Port(nil), input.Activities[index].Inputs...)
		result.Activities[index].Outputs = append([]Port(nil), input.Activities[index].Outputs...)
		result.Activities[index].Slots = append([]Slot(nil), input.Activities[index].Slots...)
	}
	return result
}
