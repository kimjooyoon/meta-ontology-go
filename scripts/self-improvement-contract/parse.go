package main

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func compileContract(path string, source []byte) (contractModel, error) {
	file, diagnostics := syntax.ParseFile(path, string(source))
	if file == nil || diagnostics.HasErrors() {
		return contractModel{}, fmt.Errorf(
			"syntax diagnostics prevent contract compilation: %d", len(diagnostics),
		)
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return contractModel{}, fmt.Errorf("lower contract: %w", err)
	}
	if err := ir.Validate(); err != nil {
		return contractModel{}, fmt.Errorf("validate semantic IR: %w", err)
	}
	model := contractModel{
		Entities: make(map[string]string), EntityByID: make(map[string]string),
		Activities: make(map[string]activitySignature), SemanticHash: ir.StableHash(),
	}
	for _, declaration := range file.Decls {
		switch value := declaration.(type) {
		case *syntax.EntityDecl:
			if _, exists := model.Entities[value.Name]; exists {
				return contractModel{}, fmt.Errorf("duplicate entity %q", value.Name)
			}
			model.Entities[value.Name] = value.ID
			model.EntityByID[value.ID] = value.Name
		case *syntax.ActivityDecl:
			if _, exists := model.Activities[value.Name]; exists {
				return contractModel{}, fmt.Errorf("duplicate activity %q", value.Name)
			}
			inputs := make([]string, len(value.Inputs))
			for index, input := range value.Inputs {
				inputs[index] = input.Name
			}
			model.Activities[value.Name] = activitySignature{
				Inputs: inputs, Output: value.Output,
			}
		}
	}
	return model, nil
}
