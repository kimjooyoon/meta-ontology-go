package bidir

import (
	"fmt"
	"go/scanner"
	"go/token"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

// The indexed-input profile explicitly preserves source slots alongside PROV
// relations. Existing documents without this profile retain their representation.
func hasIndexedInputProfile(document Document) bool {
	for _, binding := range document.RuntimeBindings {
		if _, ok := semantic.InputPortIndex(binding.Consumer.Port.Name, int(^uint(0)>>1)); ok {
			return true
		}
	}
	for _, declaration := range document.Declarations {
		if len(declaration.Inputs) < 2 || declaration.Attributes[ActivityValueProgramAttribute] == "" {
			continue
		}
		body := []byte(declaration.Attributes[ActivityValueProgramAttribute])
		var first scanner.Scanner
		first.Init(token.NewFileSet().AddFile("body", -1, len(body)), body, nil, 0)
		_, kind, literal := first.Scan()
		if kind == token.RETURN || kind == token.IF || kind == token.VAR || kind == token.IDENT && literal == "let" {
			return true
		}
	}
	return false
}

func bindSourceInputSequences(model *Model, document Document, names map[string]ID, ids map[ID]struct{}) error {
	if !hasIndexedInputProfile(document) {
		return nil
	}
	for _, declaration := range document.Declarations {
		if declaration.Kind != ActivityKind || len(declaration.Inputs) < 2 {
			continue
		}
		id, err := declarationIdentity(model.Namespace, declaration)
		if err != nil {
			return err
		}
		for i := range model.Nodes {
			if model.Nodes[i].ID != id {
				continue
			}
			model.Nodes[i].InputSequence = nil
			for _, reference := range declaration.Inputs {
				entity, err := resolveReference(reference, model.Namespace, names, ids)
				if err != nil {
					return err
				}
				reference.ID = entity
				model.Nodes[i].InputSequence = append(model.Nodes[i].InputSequence, reference)
			}
		}
	}
	return nil
}

func modelInputEntity(model Model, node Node, port string) (ID, bool) {
	if len(node.InputSequence) == 0 {
		inputs := modelRuntimePorts(model, node.ID, PredicateUsed, true)
		if modelRuntimePortArity(model, node.ID, PredicateUsed, true) != 1 || len(inputs) != 1 || port != "input" {
			return "", false
		}
		return inputs[0], true
	}
	index, ok := semantic.InputPortIndex(port, len(node.InputSequence))
	if !ok {
		return "", false
	}
	for _, input := range node.InputSequence {
		entity, found := model.node(input.ID)
		if !found || entity.Kind != EntityKind {
			return "", false
		}
		used := false
		for _, actual := range modelRuntimePorts(model, node.ID, PredicateUsed, true) {
			used = used || actual == input.ID
		}
		if !used {
			return "", false
		}
	}
	return node.InputSequence[index].ID, true
}

func validateInputSequences(model Model) error {
	for _, node := range model.Nodes {
		if len(node.InputSequence) == 0 {
			continue
		}
		if node.Kind != ActivityKind || len(node.InputSequence) < 2 {
			return fmt.Errorf("input sequence requires a multi-input Activity")
		}
		if arity, present := model.activityInputArity[node.ID]; present && arity != len(node.InputSequence) {
			return fmt.Errorf("input sequence disagrees with source arity for %q", node.ID)
		}
		for _, input := range node.InputSequence {
			if err := input.Span.Validate(); err != nil {
				return err
			}
			entity, ok := model.node(input.ID)
			if !ok || entity.Kind != EntityKind {
				return fmt.Errorf("input sequence references unknown entity %q", input.ID)
			}
		}
		used := modelRuntimePorts(model, node.ID, PredicateUsed, true)
		declared := make(map[ID]bool, len(node.InputSequence))
		for _, input := range node.InputSequence {
			declared[input.ID] = true
		}
		if len(declared) != len(used) {
			return fmt.Errorf("input sequence and used relations disagree for %q", node.ID)
		}
		for _, entity := range used {
			if !declared[entity] {
				return fmt.Errorf("input sequence omits used entity %q", entity)
			}
		}
	}
	return nil
}

func sourceInputSequence(node Node, model Model) []Reference {
	result := append([]Reference(nil), node.InputSequence...)
	for i := range result {
		if entity, exists := model.node(result[i].ID); exists {
			result[i].Name, result[i].Namespace = entity.Name, entity.Namespace
		}
	}
	return result
}
