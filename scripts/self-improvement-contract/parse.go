package main

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
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
	nodes, facts := ir.Graph.Nodes(), ir.Graph.DeterministicFacts()
	model := contractModel{EntityByID: map[string]string{}, Activities: map[string]activitySignature{}, SemanticHash: ir.StableHash()}
	for _, node := range nodes {
		switch node.Kind {
		case semantic.Entity:
			model.EntityByID[node.ID.String()] = node.Name
		case semantic.Activity:
			signature, err := contractSignature(node, facts)
			if err != nil {
				return contractModel{}, err
			}
			model.Activities[node.ID.String()] = signature
		}
	}
	return model, nil
}

func contractSignature(node semantic.Node, facts []semantic.Fact) (activitySignature, error) {
	signature := activitySignature{Name: node.Name, InputOrigins: map[string]semantic.Span{}}
	for _, fact := range facts {
		if fact.Predicate == semantic.Used && fact.Subject == node.ID {
			signature.Inputs = append(signature.Inputs, fact.Object.String())
			signature.InputOrigins[fact.Object.String()] = fact.Span
		}
		if fact.Predicate == semantic.WasGeneratedBy && fact.Object == node.ID {
			if signature.Output != "" {
				return activitySignature{}, fmt.Errorf("contract activity %s requires a single output", node.ID)
			}
			signature.Output = fact.Subject.String()
		}
	}
	// The lowered sequence preserves repeated inputs and their authored order;
	// the PROV used relation alone is an unordered set.
	if len(node.InputSequence) > 0 {
		signature.Inputs = make([]string, len(node.InputSequence))
		for index, input := range node.InputSequence {
			signature.Inputs[index] = input.String()
		}
	}
	return signature, nil
}
