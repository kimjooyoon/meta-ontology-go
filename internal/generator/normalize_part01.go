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
	canonicalizeIREmptyCollections(&result)
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

// canonicalizeIREmptyCollections gives semantically equivalent nil and empty
// collections one wire representation without allocating replacement slices.
// This keeps generated metadata digests independent of how an adapter
// materializes absent optional declarations after the single deep-copy pass.
func canonicalizeIREmptyCollections(ir *SemanticIR) {
	if len(ir.Imports) == 0 {
		ir.Imports = []Import{}
	}
	if len(ir.Entities) == 0 {
		ir.Entities = []Entity{}
	}
	for index := range ir.Entities {
		if len(ir.Entities[index].Fields) == 0 {
			ir.Entities[index].Fields = []Field{}
		}
	}
	if len(ir.Activities) == 0 {
		ir.Activities = []Activity{}
	}
	for index := range ir.Activities {
		if len(ir.Activities[index].Inputs) == 0 {
			ir.Activities[index].Inputs = []Port{}
		}
		if len(ir.Activities[index].Outputs) == 0 {
			ir.Activities[index].Outputs = []Port{}
		}
		if len(ir.Activities[index].Slots) == 0 {
			ir.Activities[index].Slots = []Slot{}
		}
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
